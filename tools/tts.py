# /// script
# requires-python = ">=3.12,<3.15"
# dependencies = ["torch==2.14.1", "numpy==2.5.4"]
# ///
import hashlib
import json
import re
import sys
import urllib.request
import wave
from pathlib import Path

import numpy as np
import torch

RATE = 48000
CACHE = Path.home() / ".cache" / "domashka" / "silero"
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


def download(name: str, path: Path):
    CACHE.mkdir(parents=True, exist_ok=True)
    url = f"https://models.silero.ai/models/tts/ru/{name}.pt"
    log(f"tts: downloading {url}")
    part = path.with_suffix(".part")
    urllib.request.urlretrieve(url, part)
    part.rename(path)


def model(name: str):
    if name in models:
        return models[name]
    if name not in MODELS:
        raise ValueError(f"unknown model {name}: its hash is not in MODELS")
    path = CACHE / f"{name}.pt"
    if not path.exists():
        download(name, path)
    if sha256(path) != MODELS[name]:
        path.unlink()
        raise ValueError(f"model {name} has an unexpected hash; the file was deleted and will be downloaded again")
    log(f"tts: loading model {name}")
    models[name] = torch.package.PackageImporter(str(path)).load_pickle("tts_models", "model")
    return models[name]


def silero_text(text: str) -> str:
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
        except Exception as e:
            reply = {"error": f"{type(e).__name__}: {e}"}
        print(json.dumps(reply, ensure_ascii=False), flush=True)


if __name__ == "__main__":
    main()
