# data_gomad

Shared facts the code reads instead of hardcoding them.

- `configs/` — paths, audio settings and the chat envelope schema, read through `loader.py`
  (`goverlord.runtime.data_gomad.configs.loader`). Copy `endpoints.example.json` to
  `endpoints.json` (not tracked) with your own addresses.
- `robot/say_replacements/` — text fixes made before a line is spoken.
- `robot/kokoro_voice/` — the Kokoro status announcer the reporter voice worker uses.
- `robot/weather/` — the per-robot weather fetch.
- `models/` — downloaded models (Vosk, Coral Edge TPU). Not tracked.
