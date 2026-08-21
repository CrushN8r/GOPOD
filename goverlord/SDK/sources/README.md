# sources/

Vendored upstream SDK/tooling clones GOPOD builds against or references.
Gitignored in full except this file — see `../README.md` for why. Each
subdirectory is its own separate git clone (has its own `.git/`); none of
their contents are GOPOD-authored.

- **`cozmo-python-sdk/`** — Anki's official Cozmo Python SDK
  (`github.com/anki/cozmo-python-sdk`). Used for Cozmo robot control on the
  laptop side of the host split (see `../HOST_SPLIT.md`).

- **`wirepod-vector-python-sdk/`** — `kercre123`'s extended fork of Anki's
  Vector Python SDK, with support for any bot on wire-pod
  (`github.com/kercre123/wirepod-vector-python-sdk`). This is the Vector SDK
  GOPOD actually uses (e.g. the Bingo reactor's `anki_vector` dependency —
  see `goverlord/runtime/songs/102_brobots_bingo_game/bingo_reactor/README.md`).

- **`vector-go-sdk/`** — `fforchino`'s early-alpha Go SDK for Vector
  (`github.com/fforchino/vector-go-sdk`). Reads robot config from
  `~/.anki_vector/sdk_config.ini`, same as the Python SDK.

- **`vectorx/`** — `fforchino`'s VECTORX project: additional Vector voice
  commands/features built on top of a Wire-Pod setup via the Go SDK, no
  Vector firmware changes required (`github.com/fforchino/vectorx`). This is
  the sibling repo the GOPOD Bingo sidecar overlays onto during its build —
  see `goverlord/runtime/songs/102_brobots_bingo_game/README.md`.

- **`Scout-open-source/`** — Moorebot's open-source code for the Scout
  home-monitoring robot (Linux/ROS, object recognition, monocular SLAM)
  (`github.com/Pilot-Labs-Dev/Scout-open-source`). Communication layer is
  deliberately excluded from the upstream open-source release; per
  `../HOST_SPLIT.md`, Scout runs on the Jetson side of the host split.

- **`Viccyware/`** — The Viccyware Group's modified copy of the Vector
  source code, unstable branch (`github.com/The-Viccyware-Group/Viccyware`).
  Firmware/source reference only per `../HOST_SPLIT.md`, not built unless
  explicitly needed.
