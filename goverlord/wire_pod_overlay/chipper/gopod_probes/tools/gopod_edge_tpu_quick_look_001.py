#!/usr/bin/env python3
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_edge_tpu_quick_look_001.py
"""The photo egg's quick look on the Coral Edge TPU (2026-10-16, operator: "Edge TPU
list shown first while moondream is still looking (both, not a swap) * round 1 is
from TPU quick look"). Loads a COCO detection model on the TPU once, then for each
photo path read from stdin (one per line) finds the objects in it and prints one JSON
line: {"objects": ["person", "cup"], "seconds": 0.05} - each object once, most sure
first. Run in the gopod-coral39 environment (pycoral 2.0.0); Wire-Pod starts it once
and keeps it (gopod_photo_quick_look.go, settings photo.quick_look). A photo path as
a 4th argument looks at that one photo and exits (a test).

2026-10-17 (operator: "3 images displayed * [raw snapshot][snapshot enhancement
attempt][TPU object recognition on enhancement]"): a line may instead be JSON,
{"photo", "enhanced", "boxes", "enhance": {...}} - the photo is enhanced first
(photo.quick_look.enhance: a light clean-up, then auto contrast, brightness,
contrast and sharpness) and saved as enhanced, the objects are found in the
enhanced one, and a copy with each object's box and name is saved as boxes."""
import json
import sys
import time

from PIL import Image, ImageDraw, ImageEnhance, ImageFilter, ImageFont, ImageOps
from pycoral.adapters import common, detect
from pycoral.utils.dataset import read_label_file
from pycoral.utils.edgetpu import make_interpreter


def enhanced(image, knobs):
    """The snapshot cleaned up: a median filter for the camera's grain, then auto
    contrast, brightness, contrast and sharpness (1 = unchanged)."""
    if knobs.get("median", 0) >= 3:
        image = image.filter(ImageFilter.MedianFilter(int(knobs["median"]) | 1))
    image = ImageOps.autocontrast(image, cutoff=knobs.get("autocontrast_cutoff", 1))
    for name, kind in (("brightness", ImageEnhance.Brightness), ("contrast", ImageEnhance.Contrast),
                       ("sharpness", ImageEnhance.Sharpness)):
        if knobs.get(name, 1) != 1:
            image = kind(image).enhance(knobs[name])
    return image


def boxed(image, objects, labels):
    """A copy with each object's box and name."""
    out = image.copy()
    draw = ImageDraw.Draw(out)
    size = max(14, out.size[1] // 28)
    try:
        font = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", size)
    except OSError:
        font = ImageFont.load_default()
    for obj in objects:
        name = labels.get(obj.id, "")
        if not name or name == "n/a":
            continue
        b = obj.bbox
        draw.rectangle([b.xmin, b.ymin, b.xmax, b.ymax], outline=(0, 255, 120), width=max(2, size // 6))
        text = "%s %d%%" % (name, round(obj.score * 100))
        tw = draw.textlength(text, font=font)
        top = max(0, b.ymin - size - 4)
        draw.rectangle([b.xmin, top, b.xmin + tw + 8, top + size + 4], fill=(0, 255, 120))
        draw.text((b.xmin + 4, top + 1), text, fill=(0, 0, 0), font=font)
    return out


def look(interpreter, labels, min_score, photo, request=None):
    started = time.time()
    image = Image.open(photo).convert("RGB")
    if request and request.get("enhanced"):
        image = enhanced(image, request.get("enhance") or {})
        image.save(request["enhanced"], quality=92)
    _, scale = common.set_resized_input(
        interpreter, image.size, lambda size: image.resize(size, Image.LANCZOS))
    interpreter.invoke()
    found = []
    objects = sorted(detect.get_objects(interpreter, min_score, scale), key=lambda o: -o.score)
    for obj in objects:
        name = labels.get(obj.id, "")
        if name and name != "n/a" and name not in found:
            found.append(name)
    if request and request.get("boxes"):
        boxed(image, objects, labels).save(request["boxes"], quality=92)
    return {"objects": found, "seconds": round(time.time() - started, 3)}


def main():
    if len(sys.argv) < 4:
        print(json.dumps({"error": "usage: model.tflite labels.txt min_score [photo.jpg]"}))
        return 2
    model, labels_path, min_score = sys.argv[1], sys.argv[2], float(sys.argv[3])
    labels = read_label_file(labels_path)
    interpreter = make_interpreter(model)
    interpreter.allocate_tensors()
    if len(sys.argv) > 4:
        print(json.dumps(look(interpreter, labels, min_score, sys.argv[4])))
        return 0
    print(json.dumps({"ready": True}), flush=True)
    for line in sys.stdin:
        photo, request = line.strip(), None
        if not photo:
            continue
        try:
            if photo.startswith("{"):
                request = json.loads(photo)
                photo = request["photo"]
            out = look(interpreter, labels, min_score, photo, request)
        except Exception as e:  # one bad photo never stops the worker
            out = {"objects": [], "error": str(e)}
        print(json.dumps(out), flush=True)
    return 0


if __name__ == "__main__":
    sys.exit(main())

# ## EOF
