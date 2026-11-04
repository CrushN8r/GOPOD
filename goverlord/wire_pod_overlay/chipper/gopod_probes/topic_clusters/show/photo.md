# The photo
- **words:** photo, picture, what do you see, camera
- **ask:** What do the Brobots say about a photo?
- **chain:** Other - the photo egg (what a Brobot sees)
- **see also:** easter_eggs, topic-pictures
- **about:** 2026-11-03 (step 6 of GOPOD_CLUSTERS_PLAN_002.md): from 03_chat_page.json photo. A value in backticks is that text exactly ("\n" a break). The numbers, switches, roles and nested groups stay in the settings for the format step.
---

## Photo (photo)
1. prompt: Describe what you see in one short, simple sentence, naming a few main things.
2. answer_text: I see {saw}
3. recall_text: I saw {saw}
4. nothing_text: I haven't taken a photo yet.
5. fail_text: I couldn't take a photo.
6. unclear_text: I took a photo, but I couldn't make out what's in it.
7. quick_text: I see {objects}.
8. quick_none_text: Hmm, I need a closer look.
9. gathering_text: Gathering details. Please wait!
10. detail_text: I see {saw}
11. detail_unclear_text: That's all I can make out.
12. detail_prompt: ``
13. stall_lines: My eyes are basically potato cams, bro. | Hold on, my lens is buffering. | Zoom and enhance... like on TV.
14. detail_strip_prefixes: The image shows | The image features | The image depicts | The image captures | This image shows | The photo shows | The picture shows | In the image, | In this image, | I see

## The quick look (photo.quick_look)
1. about: The Edge TPU's object model (COCO, 90 kinds of thing). 2026-10-16 test on the kept photos: ssdlite_mobiledet at 0.4 named the most and right (person, chair, keyboard; ssd_mobilenet_v2 only one each; 0.3 added a wrong car). The Coral stick runs one program at a time - this worker holds it while Wire-Pod runs.
2. enabled: true
3. python: ~/miniforge3/envs/gopod-coral39/bin/python
4. worker: gopod_probes/tools/gopod_edge_tpu_quick_look_001.py
5. model: ../../runtime/data_gomad/models/edge_tpu/general_objects/ssdlite_mobiledet_coco_qat_postprocess_edgetpu.tflite
6. labels: ../../runtime/data_gomad/models/edge_tpu/general_objects/coco_labels.txt
7. min_score: 0.4
8. timeout_seconds: 3
9. pictures_about: 2026-10-17 (operator: "3 images displayed * [raw snapshot][snapshot enhancement attempt][TPU object recognition on enhancement] * I see ... [TPU read on enhancement] * I see ... [moondream read of TPU]"; "TPU with bounding boxes"): the worker cleans up the snapshot (enhance: median filter size, auto contrast cutoff %, then brightness, contrast, sharpness - 1 is unchanged), finds the objects in the enhanced one and draws their boxes and names; the egg pops up the photo, the enhanced one and the boxes; the vision model looks at the enhanced one, told the TPU's objects. Kept beside the photo (webroot/photos, _enhanced.jpg, _tpu.jpg).
10. pictures: true

## The quick look - enhance (photo.quick_look.enhance)
1. median: 3
2. autocontrast_cutoff: 1
3. brightness: 1.1
4. contrast: 1.1
5. sharpness: 1.6
