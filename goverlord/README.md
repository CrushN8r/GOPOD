# goverlord

GOPOD's working parts.

- `wire_pod_overlay/` — everything GOPOD adds to Wire-Pod: the Go overlay, the chat page,
  the prompts, the topic packs and the build scripts. Start with [WIRED-POD.md](../WIRED-POD.md).
- `runtime/songs/` — the scored songs and the engine that plays them.
- `runtime/data_gomad/` — shared config the songs and the room mic read (paths, audio, the
  speech replacements, the weather fetch).
- `runtime/gopod_layer/web_display/` — the push-to-talk writer the room mic service uses.
- `alias_lib_overlay/` — the alias play studio (`pha0b`, `phcal`), Take Out part 1.
- `rules/` — one audio rule the config loader reads.
