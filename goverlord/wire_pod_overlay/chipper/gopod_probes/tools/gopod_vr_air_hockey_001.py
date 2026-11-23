#!/usr/bin/env python3
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_vr_air_hockey_001.py
#
# GOPOD VR/AR air hockey, the starter (2026-11-22, operator: "use the black bar on
# the ruler as your visual anchor"; "build me starter GOPOD air hockey VR/AR
# thingy"). One photo from the overhead camera (the RunCam, 3840x2160) -> where
# the rink, the two Brobots and the cube (the puck) are:
#   - the anchor: the ruler's black bar - the longest dark thin bar low in the frame;
#     its length is the unit (bar-lengths), its middle the rink's home line;
#   - the rink: the frame above the bar, between the bar's ends widened by rink_wide;
#   - the Brobots: the cones (gopod_vr_cone_probe_001.py's colours: red = Brobot 1,
#     orange = Brobot 2) - only small, solid blobs inside the rink (the orange
#     track around it is long and hollow, so it's left out);
#   - the puck: the cube - a dark, square, cube-sized blob inside the rink.
# Prints one JSON line (positions in pixels and in bar-lengths from the bar's
# middle: x right, y up); with --out, saves the photo with all of it drawn.
# Run with the Coral env's python (gopod-coral39):
#   python gopod_vr_air_hockey_001.py --photo frame.jpg --out marked.jpg
import argparse
import json
import os
import sys

import cv2
import numpy as np

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gopod_vr_cone_probe_001 import CONES, warm_blobs  # noqa: E402 - the one cone finder

RINK_WIDE = 0.9       # the rink reaches this many bar-lengths past each end of the bar
RINK_TALL = 1.4       # and this many bar-lengths above it
CONE_FILL = 0.35      # a cone fills its box at least this much; a curved track piece less
CONE_MAX = 0.12       # a cone is at most this many bar-lengths wide or tall
RINK_BELOW = 0.15     # the rink reaches this many bar-lengths below the bar - the Brobots park level with it
CONE_RING_DARK = 105  # a cone sits on a dark robot: the ring around it is darker than this (cones ~60-95;
                      # pink tape on the white table ~140 - under bright light the same colour as the red cone)
BAR_DARK = 115        # the bar reads mid-grey under bright light (~90 of 255; the table ~170)
PUCK_ICON_LIGHT = 120 # the cube's icon is lighter (~130 under bright light; its face ~60, the table ~170)
PUCK_ICON_SHARE = 0.06  # the icon is at least this share of the cube's dark shape


def find_bar(img):
    """The ruler's black bar: the longest dark, thin, level bar in the lower half."""
    h, w = img.shape[:2]
    grey = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    dark = cv2.inRange(grey, 0, BAR_DARK)
    dark[: h // 3, :] = 0  # it sits low in the frame
    dark = cv2.morphologyEx(dark, cv2.MORPH_CLOSE, np.ones((5, 25), np.uint8))
    contours, _ = cv2.findContours(dark, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    best = None
    for c in contours:
        x, y, bw, bh = cv2.boundingRect(c)
        if bw < w * 0.12 or bh == 0 or bw / bh < 12:
            continue
        if cv2.contourArea(c) < 0.5 * bw * bh:  # a bar fills its box
            continue
        if best is None or bw > best[2]:
            best = (x, y, bw, bh)
    if best is None:
        return None
    x, y, bw, bh = best
    return {"x": int(x), "y": int(y), "w": int(bw), "h": int(bh), "cx": int(x + bw / 2), "cy": int(y + bh / 2)}


def rink_of(bar, shape):
    h, w = shape[:2]
    x0 = max(0, int(bar["x"] - RINK_WIDE * bar["w"]))
    x1 = min(w, int(bar["x"] + bar["w"] + RINK_WIDE * bar["w"]))
    y0 = max(0, int(bar["y"] - RINK_TALL * bar["w"]))
    y1 = min(h, int(bar["y"] + bar["h"] + RINK_BELOW * bar["w"]))
    return {"x": x0, "y": y0, "w": x1 - x0, "h": y1 - y0}


def inside(b, r):
    return r["x"] <= b["cx"] <= r["x"] + r["w"] and r["y"] <= b["cy"] <= r["y"] + r["h"]


def bar_units(b, bar):
    """A point in bar-lengths from the bar's middle (x right, y up)."""
    return {"bx": round((b["cx"] - bar["cx"]) / bar["w"], 3), "by": round((bar["cy"] - b["cy"]) / bar["w"], 3)}


def ring_dark(grey, b):
    """How bright the ring around a blob is (a box as wide again on each side, less the blob's)."""
    m = max(b["w"], b["h"])
    x0, y0 = max(0, b["x"] - m), max(0, b["y"] - m)
    big = grey[y0: b["y"] + b["h"] + m, x0: b["x"] + b["w"] + m].astype(np.float64)
    inner = grey[b["y"]: b["y"] + b["h"], b["x"]: b["x"] + b["w"]].astype(np.float64)
    return (big.sum() - inner.sum()) / max(1, big.size - inner.size)


def find_brobots(hsv, grey, bar, rink, min_area):
    cones = [b for b in warm_blobs(hsv, min_area, top=12)
             if inside(b, rink) and b["fill"] >= CONE_FILL and ring_dark(grey, b) < CONE_RING_DARK
             and max(b["w"], b["h"]) <= CONE_MAX * bar["w"]]
    out = {"Brobot 1": None, "Brobot 2": None}
    cones = sorted(cones, key=lambda b: -b["area"])[:2]
    if len(cones) == 2:
        red, orange = sorted(cones, key=lambda b: b["hue"])  # the redder is the red cone
        out["Brobot 1"], out["Brobot 2"] = dict(red, cone="red"), dict(orange, cone="orange")
    elif len(cones) == 1:
        out["Brobot 1" if cones[0]["hue"] < 3.0 else "Brobot 2"] = dict(cones[0])
    return out


def find_puck(img, bar, rink):
    """The cube: a dark, square, cube-sized blob inside the rink."""
    grey = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    dark = cv2.inRange(grey, 0, BAR_DARK)
    mask = np.zeros_like(dark)
    mask[rink["y"]: rink["y"] + rink["h"], rink["x"]: rink["x"] + rink["w"]] = 255
    dark = cv2.bitwise_and(dark, mask)
    # Thin dark lines (cables) cut away; the solid cube stays.
    k = max(9, int(0.03 * bar["w"]))
    dark = cv2.morphologyEx(dark, cv2.MORPH_OPEN, np.ones((k, k), np.uint8))
    dark = cv2.morphologyEx(dark, cv2.MORPH_CLOSE, np.ones((9, 9), np.uint8))
    contours, _ = cv2.findContours(dark, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    # The cube is the dark shape with its light icon inside (2026-11-22: ~12% light;
    # the robots, the speaker and the mouse have none). A cable may join its outline,
    # so its centre is the icon's - the light pixels inside its dark shape.
    most = 0.4 * bar["w"]
    best, best_light = None, 0.0
    for c in contours:
        x, y, w, h = cv2.boundingRect(c)
        if w < 0.06 * bar["w"] or h < 0.06 * bar["w"] or w > most:
            continue
        filled = np.zeros_like(dark)
        cv2.drawContours(filled, [c], -1, 255, -1)
        icon = (filled > 0) & (grey > PUCK_ICON_LIGHT)
        light = float(icon.sum()) / max(1, float((filled > 0).sum()))
        if light < PUCK_ICON_SHARE or light <= best_light:
            continue
        ys, xs = np.nonzero(icon)
        cx, cy = int(xs.mean()), int(ys.mean())
        side = int(0.15 * bar["w"])  # the cube, about this wide around its icon
        best, best_light = {"x": cx - side // 2, "y": cy - side // 2, "w": side, "h": side, "cx": cx, "cy": cy,
                            "icon": round(light, 2)}, light
    return best


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--photo", required=True)
    ap.add_argument("--out", default="")
    ap.add_argument("--min-area", type=int, default=150)
    args = ap.parse_args()
    img = cv2.imread(args.photo)
    if img is None:
        print(json.dumps({"error": "no photo: " + args.photo}))
        return 2
    bar = find_bar(img)
    if bar is None:
        print(json.dumps({"error": "no ruler bar found - is the ruler in view?"}))
        return 1
    rink = rink_of(bar, img.shape)
    hsv = cv2.cvtColor(img, cv2.COLOR_BGR2HSV)
    brobots = find_brobots(hsv, cv2.cvtColor(img, cv2.COLOR_BGR2GRAY), bar, rink, args.min_area)
    puck = find_puck(img, bar, rink)
    for b in list(brobots.values()) + [puck]:
        if b:
            b.update(bar_units(b, bar))
    if args.out:
        cv2.rectangle(img, (rink["x"], rink["y"]), (rink["x"] + rink["w"], rink["y"] + rink["h"]), (255, 255, 0), 4)
        cv2.rectangle(img, (bar["x"], bar["y"]), (bar["x"] + bar["w"], bar["y"] + bar["h"]), (255, 0, 255), 6)
        cv2.putText(img, "anchor", (bar["x"], bar["y"] - 12), cv2.FONT_HERSHEY_SIMPLEX, 1.6, (255, 0, 255), 3)
        for who, b in brobots.items():
            if b:
                col = CONES[who]["draw"]
                cv2.rectangle(img, (b["x"], b["y"]), (b["x"] + b["w"], b["y"] + b["h"]), col, 6)
                cv2.putText(img, who, (b["x"], max(30, b["y"] - 12)), cv2.FONT_HERSHEY_SIMPLEX, 1.6, col, 4)
        if puck:
            cv2.circle(img, (puck["cx"], puck["cy"]), max(puck["w"], puck["h"]), (0, 255, 0), 6)
            cv2.putText(img, "puck (cube)", (puck["x"], max(30, puck["y"] - 12)), cv2.FONT_HERSHEY_SIMPLEX, 1.6, (0, 255, 0), 4)
        cv2.imwrite(args.out, img)
    print(json.dumps({"bar": bar, "rink": rink, "brobots": brobots, "puck": puck, "size": [int(img.shape[1]), int(img.shape[0])]}))
    return 0


if __name__ == "__main__":
    sys.exit(main())
