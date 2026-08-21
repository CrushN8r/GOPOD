# models/

Downloaded model binaries. Gitignored in full except this file — see
`../README.md` for why. Nothing here is GOPOD-authored.

- **`vosk/vosk-model-small-en-us-0.15/`** — Alpha Cephei's small US English
  Vosk STT model. Path referenced in-repo at
  `goverlord/runtime/data_gomad/configs/paths.json`'s `vosk_model_path`; this
  is the local offline STT model the PTT lane loads.

- **`edge_tpu/`** — Coral Edge TPU `.tflite` models and their label files,
  grouped by what they detect:
  - `general_objects/` — COCO object-detection models
    (`ssd_mobilenet_v1`/`v2`, `ssdlite_mobiledet`, `efficientdet_lite0`) plus
    `coco_labels.txt`.
  - `animal_bird/` — iNaturalist-trained bird/insect/plant classifiers
    (`mobilenet_v2_1.0_224_inat_*`) plus their label files.
  - `face_optional/` — a single face-detection SSD MobileNet v2 model.

  These back the cockpit's optional Edge TPU inference pane (`OPTIONAL_OFFLINE`
  by default — see root `README.md`'s "Current state").
