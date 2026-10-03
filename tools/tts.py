# /// script
# requires-python = ">=3.11,<3.14"
# dependencies = ["torch==2.14.1", "numpy==2.5.3"]
# ///
"""Синтезатор Silero для сервера domashka.

Сервер запускает его один раз (`uv run --script tools/tts.py`) и шлёт по строке JSON на запрос:
    {"model": "v5_5_ru", "speaker": "baya", "text": "молок+о", "out": "/путь/к/файлу.wav"}
В ответ — строка {"ok": true} или {"error": "..."}. Модели скачиваются при первом обращении
в ~/.cache/domashka/silero и дальше держатся в памяти. «+» перед гласной — ударение.
"""

import hashlib
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

# Модель Silero — это pickle, то есть код, который выполняется при загрузке. Грузим только
# файлы с известным хэшем: подменённая модель (на сайте или в кэше) не запустится.
# Новая модель или версия — посчитать `shasum -a 256` у скачанного файла и добавить сюда.
MODELS = {
    "v5_5_ru": "50081637b602126ee06cb3bc8a744d25651d2da149ee8864b9a379bfdd934437",
    "v5_cis_base": "ba41b18f6a707ad93605a162998865e7c087153d2e010a26dd02229dab0e672a",
}
models = {}


def log(*args):
    print(*args, file=sys.stderr, flush=True)


def sha256(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def model(name: str):
    if name not in models:
        if name not in MODELS:
            raise ValueError(f"неизвестная модель {name}: её хэша нет в MODELS")
        path = CACHE / f"{name}.pt"
        if not path.exists():
            CACHE.mkdir(parents=True, exist_ok=True)
            url = f"https://models.silero.ai/models/tts/ru/{name}.pt"
            log(f"tts: скачиваю {url}")
            urllib.request.urlretrieve(url, path.with_suffix(".part"))
            path.with_suffix(".part").rename(path)
        if sha256(path) != MODELS[name]:
            path.unlink()
            raise ValueError(f"у модели {name} не тот хэш — файл удалён, при следующем запросе скачается заново")
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
