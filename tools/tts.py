# /// script
# requires-python = ">=3.11,<3.14"
# dependencies = ["torch", "numpy", "scipy"]
# ///
"""Синтезатор Silero для сервера domashka.

Сервер запускает его один раз (`uv run --script tools/tts.py`) и шлёт по строке JSON на запрос:
    {"model": "v5_5_ru", "speaker": "baya", "text": "молок+о", "out": "/путь/к/файлу.wav"}
В ответ — строка {"ok": true} или {"error": "..."}. Модели скачиваются при первом обращении
в ~/.cache/domashka/silero и дальше держатся в памяти. «+» перед гласной — ударение.
"""

import json
import re
import sys
import urllib.request
import warnings
import wave
from pathlib import Path

warnings.filterwarnings("ignore")

import numpy as np  # noqa: E402
import torch  # noqa: E402

RATE = 48000
CACHE = Path.home() / ".cache" / "domashka" / "silero"
models = {}


def log(*args):
    print(*args, file=sys.stderr, flush=True)


def model(name: str):
    if name not in models:
        path = CACHE / f"{name}.pt"
        if not path.exists():
            CACHE.mkdir(parents=True, exist_ok=True)
            url = f"https://models.silero.ai/models/tts/ru/{name}.pt"
            log(f"tts: скачиваю {url}")
            urllib.request.urlretrieve(url, path.with_suffix(".part"))
            path.with_suffix(".part").rename(path)
        log(f"tts: загружаю модель {name}")
        models[name] = torch.package.PackageImporter(str(path)).load_pickle("tts_models", "model")
    return models[name]


def silero_text(text: str) -> str:
    # Silero считает текстом только буквы из диапазона «а-я», а «ё» (U+0451) в него не входит:
    # строка из одной «ё» для него пустая и падает с ValueError без сообщения. Отдельная «ё»
    # и звучит как /jo/, так что отдаём «йо», ударение — на гласную.
    if not re.search(r"[а-я]", text.lower()):
        text = re.sub(r"\+?ё", lambda m: "й+о" if m.group().startswith("+") else "йо", text, flags=re.IGNORECASE)
    return text


def synthesize(req: dict):
    audio = model(req["model"]).apply_tts(text=silero_text(req["text"]), speaker=req["speaker"], sample_rate=RATE)
    pcm = (np.clip(audio.numpy(), -1, 1) * 32767).astype("<i2")
    out = Path(req["out"])
    part = out.with_suffix(".part")
    with wave.open(str(part), "wb") as w:
        w.setnchannels(1)
        w.setsampwidth(2)
        w.setframerate(RATE)
        w.writeframes(pcm.tobytes())
    part.rename(out)


def main():
    torch.set_num_threads(4)
    for line in sys.stdin:
        try:
            synthesize(json.loads(line))
            reply = {"ok": True}
        except Exception as e:  # ответить и ждать следующий запрос
            reply = {"error": f"{type(e).__name__}: {e}"}
        print(json.dumps(reply, ensure_ascii=False), flush=True)


if __name__ == "__main__":
    main()
