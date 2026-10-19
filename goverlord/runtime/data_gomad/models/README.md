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

## Edge TPU: turn it on (golden, 2026-10-14)

For photo face and object recognition. The Coral USB stick shows as
`1a6e:089a Global Unichip Corp.` until a model runs on it; after that it shows
as `18d1:9302 Google Inc.` - on.

Setup already in place:
- the TPU driver (`libedgetpu1-max`)
- the `gopod-coral39` Python environment (`~/miniforge3/envs/gopod-coral39`, pycoral 2.0.0)
- face and object models in `goverlord/runtime/data_gomad/models/edge_tpu/`

Turn it on (after a replug or a restart):

```
~/miniforge3/envs/gopod-coral39/bin/python -c "from pycoral.utils.edgetpu import make_interpreter as m; i=m('$HOME/crushn8r_git/GOPOD/goverlord/runtime/data_gomad/models/edge_tpu/face_optional/ssd_mobilenet_v2_face_quant_postprocess_edgetpu.tflite'); i.allocate_tensors(); print('TPU on')"
```

Check: `lsusb | grep 18d1` shows `Google Inc.`

Since 2026-10-16 Wire-Pod turns it on itself: the photo egg's quick look
(`gopod_edge_tpu_quick_look_001.py`, `photo.quick_look` in the Brobots
settings) loads `ssdlite_mobiledet` at Wire-Pod start and holds the stick
while Wire-Pod runs - the stick runs one program at a time, so the command
above only works with Wire-Pod stopped.
