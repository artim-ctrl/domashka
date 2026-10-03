// Микрофон открывается один раз на всю сессию записи: тогда начало слова не съедается
// задержкой включения. Шумоподавление и автогромкость выключены — они портят шипящие
// (ш, щ, ж) и выравнивают то, что мы и так выровняем сами.

export type Recorder = { stop(): Promise<Blob>; cancel(): void };

export class Mic {
  private constructor(
    private readonly stream: MediaStream,
    private readonly ctx: AudioContext,
    private readonly analyser: AnalyserNode,
    private readonly samples: Float32Array<ArrayBuffer>,
  ) {}

  static async open(): Promise<Mic> {
    const stream = await navigator.mediaDevices.getUserMedia({
      audio: { echoCancellation: false, noiseSuppression: false, autoGainControl: false, channelCount: 1 },
    });
    const ctx = new AudioContext();
    await ctx.resume();
    const analyser = ctx.createAnalyser();
    analyser.fftSize = 1024;
    ctx.createMediaStreamSource(stream).connect(analyser);
    return new Mic(stream, ctx, analyser, new Float32Array(analyser.fftSize));
  }

  // Пиковый уровень последних ~20 мс, 0…1 — для индикатора.
  level(): number {
    this.analyser.getFloatTimeDomainData(this.samples);
    return this.samples.reduce((m, v) => Math.max(m, Math.abs(v)), 0);
  }

  record(): Recorder {
    const rec = new MediaRecorder(this.stream);
    const chunks: Blob[] = [];
    rec.ondataavailable = (e) => chunks.push(e.data);
    const done = new Promise<Blob>((resolve) => {
      rec.onstop = () => resolve(new Blob(chunks, { type: rec.mimeType }));
    });
    rec.start();
    return {
      stop: () => {
        if (rec.state !== "inactive") rec.stop();
        return done;
      },
      cancel: () => {
        if (rec.state !== "inactive") rec.stop();
      },
    };
  }

  close() {
    this.stream.getTracks().forEach((t) => t.stop());
    void this.ctx.close();
  }
}
