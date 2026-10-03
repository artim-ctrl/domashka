// Готовит дубль к сохранению: моно, тишина по краям обрезана, громкость выровнена,
// 24 кГц / 16 бит WAV — такой файл играет в любом браузере, включая Safari на iPhone.

const RATE = 24000;
const WINDOW = 0.01; // окно для поиска речи, с
const PAD_BEFORE = 0.12; // сколько тишины оставить до звука, с
const PAD_AFTER = 0.2; // и после — у щ, ш, х длинный тихий хвост
const PEAK = 0.89; // −1 дБ
const MAX_GAIN = 10;
const FADE = 0.008;

export class SilenceError extends Error {}

export type Take = { blob: Blob; duration: number };

export async function processTake(recorded: Blob): Promise<Take> {
  const decoded = await new OfflineAudioContext(1, 1, RATE).decodeAudioData(await recorded.arrayBuffer());
  const mono = mixDown(decoded);
  const [from, to] = speechBounds(mono, decoded.sampleRate);
  if (to <= from) throw new SilenceError();

  const out = await resample(mono.subarray(from, to), decoded.sampleRate);
  normalize(out);
  fade(out);
  return { blob: encodeWav(out), duration: out.length / RATE };
}

function mixDown(buf: AudioBuffer): Float32Array {
  const out = new Float32Array(buf.length);
  for (let ch = 0; ch < buf.numberOfChannels; ch++) {
    const data = buf.getChannelData(ch);
    for (let i = 0; i < out.length; i++) out[i] += data[i] / buf.numberOfChannels;
  }
  return out;
}

// Границы звука: первое и последнее окно, где громкость заметно выше тишины.
function speechBounds(x: Float32Array, rate: number): [number, number] {
  const win = Math.max(1, Math.round(rate * WINDOW));
  const rms: number[] = [];
  for (let i = 0; i < x.length; i += win) {
    const end = Math.min(i + win, x.length);
    let sum = 0;
    for (let j = i; j < end; j++) sum += x[j] * x[j];
    rms.push(Math.sqrt(sum / (end - i)));
  }
  const loudest = rms.reduce((a, b) => Math.max(a, b), 0);
  if (loudest < 0.003) return [0, 0];
  const threshold = Math.max(loudest * 0.06, 0.002);
  const first = rms.findIndex((v) => v >= threshold);
  const last = rms.findLastIndex((v) => v >= threshold);
  return [
    Math.max(0, first * win - Math.round(rate * PAD_BEFORE)),
    Math.min(x.length, (last + 1) * win + Math.round(rate * PAD_AFTER)),
  ];
}

async function resample(x: Float32Array, rate: number): Promise<Float32Array> {
  const ctx = new OfflineAudioContext(1, Math.max(1, Math.ceil((x.length * RATE) / rate)), RATE);
  const buf = ctx.createBuffer(1, x.length, rate);
  buf.getChannelData(0).set(x);
  const src = ctx.createBufferSource();
  src.buffer = buf;
  src.connect(ctx.destination);
  src.start();
  return (await ctx.startRendering()).getChannelData(0);
}

function normalize(x: Float32Array) {
  const peak = x.reduce((m, v) => Math.max(m, Math.abs(v)), 0);
  if (peak === 0) return;
  const gain = Math.min(PEAK / peak, MAX_GAIN);
  for (let i = 0; i < x.length; i++) x[i] *= gain;
}

function fade(x: Float32Array) {
  const n = Math.min(Math.round(RATE * FADE), Math.floor(x.length / 2));
  for (let i = 0; i < n; i++) {
    x[i] *= i / n;
    x[x.length - 1 - i] *= i / n;
  }
}

function encodeWav(x: Float32Array): Blob {
  const buf = new ArrayBuffer(44 + x.length * 2);
  const v = new DataView(buf);
  const str = (at: number, s: string) => [...s].forEach((c, i) => v.setUint8(at + i, c.charCodeAt(0)));
  str(0, "RIFF");
  v.setUint32(4, 36 + x.length * 2, true);
  str(8, "WAVE");
  str(12, "fmt ");
  v.setUint32(16, 16, true); // размер блока fmt
  v.setUint16(20, 1, true); // PCM
  v.setUint16(22, 1, true); // моно
  v.setUint32(24, RATE, true);
  v.setUint32(28, RATE * 2, true); // байт в секунду
  v.setUint16(32, 2, true); // байт на отсчёт
  v.setUint16(34, 16, true); // бит на отсчёт
  str(36, "data");
  v.setUint32(40, x.length * 2, true);
  for (let i = 0; i < x.length; i++) {
    const s = Math.max(-1, Math.min(1, x[i]));
    v.setInt16(44 + i * 2, s < 0 ? s * 0x8000 : s * 0x7fff, true);
  }
  return new Blob([buf], { type: "audio/wav" });
}
