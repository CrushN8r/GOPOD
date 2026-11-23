#!/usr/bin/env python3
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_vr_cone_probe_001.py
#
# GOPOD VR/AR, first test (2026-11-18, operator: "preliminary testing or a first
# test with the C920 + TPU for GOPOD VR/AR * brobot id from red/orange traffic
# cones adhered behind their heads"): one photo -> where each Brobot is.
#   - the cones, by colour: red = Brobot 1, orange = Brobot 2 (no model needed);
#   - the Edge TPU's own boxes (COCO objects), when the TPU is free (Wire-Pod's
#     quick-look worker may hold it - then the cones alone).
# Prints one JSON line; with --out, saves the photo with the boxes drawn.
# Run with the Coral env's python (gopod-coral39):
#   python gopod_vr_cone_probe_001.py --photo in.jpg --out marked.jpg
import argparse
import json
import sys

import cv2
import numpy as np

# The cones, told apart against each other (2026-11-18, two photos): light moves
# both cones' colours together - dim, the red cone is hue 171-179/0-2 and the
# orange 2-8; bright and washed out, red is 178-5 and orange 3-9 - so no fixed
# range holds, but the red cone is always the redder of the two. Both are found
# as warm cone-coloured blobs (hue 168-179 or 0-14, some colour, not dark); the
# redder one is Brobot 1 (red cone), the other Brobot 2 (orange cone). The LEGO
# figures (hue 15+) stay out.
CONES = {
    "Brobot 1": {"colour": "red", "draw": (40, 40, 230)},
    "Brobot 2": {"colour": "orange", "draw": (0, 150, 255)},
}
WARM = [((168, 100, 150), (179, 255, 255)), ((0, 100, 150), (11, 255, 255))]  # paper and wood (colour ~70-95) and the yellow LEGO (hue 12-16) stay out
ALONE_RED_BELOW = 3.0  # one cone only: redder than this (shifted hue) is the red one


def warm_blobs(hsv, min_area, piece_area=25, gap=40, top=2):
    """The two biggest cones: warm pieces (a cone's white stripes and glare cut it
    into several) grouped when they lie within gap pixels of each other."""
    mask = None
    for lo, hi in WARM:
        m = cv2.inRange(hsv, np.array(lo), np.array(hi))
        mask = m if mask is None else cv2.bitwise_or(mask, m)
    mask = cv2.morphologyEx(mask, cv2.MORPH_OPEN, np.ones((3, 3), np.uint8))
    mask = cv2.morphologyEx(mask, cv2.MORPH_CLOSE, np.ones((7, 7), np.uint8))
    contours, _ = cv2.findContours(mask, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    pieces = [c for c in contours if cv2.contourArea(c) >= piece_area]
    # 2026-11-22 (a U-shaped orange track's box covered the whole rink and
    # swallowed the red cone inside it): pieces group when they are actually
    # within gap pixels of each other (the mask grown by gap), never because one's
    # box covers the other.
    near = cv2.dilate(mask, np.ones((gap, gap), np.uint8))
    _, labels = cv2.connectedComponents(near)
    by_label = {}
    for c in pieces:
        x, y, w, h = cv2.boundingRect(c)
        px, py = c[0][0]
        lab = int(labels[py, px])
        g = by_label.get(lab)
        if g is None:
            by_label[lab] = [x, y, x + w, y + h, [c]]
        else:
            g[0], g[1], g[2], g[3] = min(g[0], x), min(g[1], y), max(g[2], x + w), max(g[3], y + h)
            g[4].append(c)
    groups = list(by_label.values())  # [x0, y0, x1, y1, [contours]]
    hue = hsv[:, :, 0].astype(np.float32)
    shifted = np.where(hue > 90, hue - 180, hue)  # 178 -> -2: red sits below orange
    blobs = []
    for x0, y0, x1, y1, cs in groups:
        area = sum(cv2.contourArea(c) for c in cs)
        if area < min_area:
            continue
        one = np.zeros(mask.shape, np.uint8)
        cv2.drawContours(one, cs, -1, 255, -1)
        sel = (one > 0) & (mask > 0)
        blobs.append({"x": int(x0), "y": int(y0), "w": int(x1 - x0), "h": int(y1 - y0), "cx": int((x0 + x1) / 2), "cy": int((y0 + y1) / 2),
                      "area": int(area), "hue": round(float(shifted[sel].mean()), 1),
                      # how solid its box is: a cone fills it, a curved track piece doesn't
                      "fill": round(float(area) / max(1, (x1 - x0) * (y1 - y0)), 2)})
    return sorted(blobs, key=lambda b: -b["area"])[:top]


def find_brobots(hsv, min_area):
    blobs = warm_blobs(hsv, min_area)
    out = {"Brobot 1": None, "Brobot 2": None}
    if len(blobs) == 2:
        red, orange = sorted(blobs, key=lambda b: b["hue"])
        out["Brobot 1"], out["Brobot 2"] = dict(red, cone="red"), dict(orange, cone="orange")
    elif len(blobs) == 1:
        b = blobs[0]
        who = "Brobot 1" if b["hue"] < ALONE_RED_BELOW else "Brobot 2"
        out[who] = dict(b, cone=CONES[who]["colour"])
    return out


def tpu_boxes(photo, model, labels_path, min_score):
    try:
        from PIL import Image
        from pycoral.adapters import common, detect
        from pycoral.utils.dataset import read_label_file
        from pycoral.utils.edgetpu import make_interpreter
        labels = read_label_file(labels_path)
        interpreter = make_interpreter(model)
        interpreter.allocate_tensors()
        image = Image.open(photo).convert("RGB")
        _, scale = common.set_resized_input(interpreter, image.size, lambda size: image.resize(size, Image.LANCZOS))
        interpreter.invoke()
        out = []
        for o in detect.get_objects(interpreter, min_score, scale):
            b = o.bbox
            out.append({"label": labels.get(o.id, str(o.id)), "score": round(float(o.score), 2), "x": int(b.xmin), "y": int(b.ymin), "w": int(b.xmax - b.xmin), "h": int(b.ymax - b.ymin)})
        return out, ""
    except Exception as e:  # the TPU held by Wire-Pod's worker, or no pycoral
        return [], str(e).splitlines()[0][:160] if str(e) else type(e).__name__


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--photo", required=True)
    ap.add_argument("--out", default="")
    ap.add_argument("--model", default="")
    ap.add_argument("--labels", default="")
    ap.add_argument("--min-score", type=float, default=0.4)
    ap.add_argument("--min-area", type=int, default=150)
    args = ap.parse_args()
    img = cv2.imread(args.photo)
    if img is None:
        print(json.dumps({"error": "no photo: " + args.photo}))
        return 2
    hsv = cv2.cvtColor(img, cv2.COLOR_BGR2HSV)
    brobots = find_brobots(hsv, args.min_area)
    boxes, tpu_error = ([], "no model given")
    if args.model and args.labels:
        boxes, tpu_error = tpu_boxes(args.photo, args.model, args.labels, args.min_score)
    if args.out:
        for b in boxes:
            cv2.rectangle(img, (b["x"], b["y"]), (b["x"] + b["w"], b["y"] + b["h"]), (200, 200, 200), 2)
            cv2.putText(img, "%s %.2f" % (b["label"], b["score"]), (b["x"], max(12, b["y"] - 4)), cv2.FONT_HERSHEY_SIMPLEX, 0.5, (200, 200, 200), 1)
        for who, b in brobots.items():
            if b:
                col = CONES[who]["draw"]
                cv2.rectangle(img, (b["x"], b["y"]), (b["x"] + b["w"], b["y"] + b["h"]), col, 3)
                cv2.putText(img, who, (b["x"], max(16, b["y"] - 6)), cv2.FONT_HERSHEY_SIMPLEX, 0.7, col, 2)
        cv2.imwrite(args.out, img)
    print(json.dumps({"brobots": brobots, "tpu_boxes": boxes, "tpu_error": tpu_error, "size": [int(img.shape[1]), int(img.shape[0])]}))
    return 0


if __name__ == "__main__":
    sys.exit(main())
