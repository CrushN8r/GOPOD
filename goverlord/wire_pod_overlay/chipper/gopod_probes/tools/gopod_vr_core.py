#!/usr/bin/env python3
"""GOPOD VR/AR core - the overhead camera, the arena (the ruler bar's scale) and each Brobot's
marker by its colour. Apps (the air hockey, the stage) sit on top; the core never names a game.
Values: the settings folder's vr_markers block."""
import glob
import json
import os
import sys

import cv2
import numpy as np

CAMERA_GLOB = "usb-Wasintek_Wasintek_camera_*-video-index0"  # the RunCam, by name
ANCHOR_FILE = "~/.gopod_air_hockey_anchor.json"  # the arena, locked (the camera moved: find it again)
BAR_MM = 304.8            # the ruler's black bar: 12 inches
BAR_DARK = 115            # the bar reads mid-grey under this light (~90; the table ~170)
TRACK_HSV = [((168, 100, 150), (179, 255, 255)), ((0, 100, 150), (11, 255, 255))]  # the orange track (its red corners too): the arena's wall
SETTINGS_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "webroot", "gopod_settings")


def settings(block):
    """One block of the settings folder (its files read as one)."""
    out = {}
    for f in sorted(glob.glob(os.path.join(SETTINGS_DIR, "*.json"))):
        try:
            out.update(json.load(open(f)))
        except (OSError, ValueError):
            pass
    return out.get(block, {})


def open_camera():
    found = glob.glob(os.path.join("/dev/v4l/by-id", CAMERA_GLOB))
    if not found:
        sys.exit("no RunCam found")
    cap = cv2.VideoCapture(found[0], cv2.CAP_V4L2)
    cap.set(cv2.CAP_PROP_FOURCC, cv2.VideoWriter_fourcc(*"MJPG"))
    cap.set(cv2.CAP_PROP_FRAME_WIDTH, 1920)  # 1080p: the bar stands clear of the speaker
    cap.set(cv2.CAP_PROP_FRAME_HEIGHT, 1080)
    cap.set(cv2.CAP_PROP_FPS, 30)
    return cap


def one_frame(cap, frames=15):
    """A frame once the exposure has settled (the first ones come before it does)."""
    ok, img = False, None
    for _ in range(frames):
        ok, img = cap.read()
    return img if ok else None


def find_bar(img):
    """The ruler's black bar: the longest dark, thin, level bar low in the frame."""
    h, w = img.shape[:2]
    dark = cv2.inRange(cv2.cvtColor(img, cv2.COLOR_BGR2GRAY), 0, BAR_DARK)
    dark[: h // 3, :] = 0
    dark = cv2.morphologyEx(dark, cv2.MORPH_OPEN, np.ones((1, max(20, int(0.06 * w))), np.uint8))  # long level runs only
    dark = cv2.morphologyEx(dark, cv2.MORPH_CLOSE, np.ones((2, 12), np.uint8))
    best = None
    for c in cv2.findContours(dark, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)[0]:
        x, y, bw, bh = cv2.boundingRect(c)
        if bw < w * 0.12 or bh == 0 or bw / bh < 12 or cv2.contourArea(c) < 0.5 * bw * bh:
            continue
        if best is None or bw > best[2]:
            best = (x, y, bw, bh)
    if best is None:
        return None
    x, y, bw, bh = best
    return {"x": int(x), "y": int(y), "w": int(bw), "h": int(bh), "cx": int(x + bw / 2), "cy": int(y + bh / 2)}


def arena(img, recalibrate=False):
    """The locked arena (the ruler bar), found in img when there's none or when asked."""
    anchor = os.path.expanduser(ANCHOR_FILE)
    if not recalibrate and os.path.exists(anchor):
        return json.load(open(anchor))
    return find_bar(img)


def to_mm(bar, px, py):
    """A pixel as mm from the ruler's middle: x right, y up the table."""
    mm_px = BAR_MM / bar["w"]
    return (px - bar["cx"]) * mm_px, (bar["cy"] - py) * mm_px


def find_markers(img, bar, markers=None):
    """Each Brobot's marker (vr_markers: his colour's hsv ranges, a compact piece of size_mm):
    {name: (x_mm, y_mm)} for each one seen."""
    cfg = markers if markers is not None else settings("vr_markers")
    lo_mm, hi_mm = cfg.get("size_mm", [10, 60])
    fill_min = cfg.get("fill_min", 0.45)
    mm_px = BAR_MM / bar["w"]
    hsv = cv2.cvtColor(img, cv2.COLOR_BGR2HSV)
    seen = {}
    for name, m in cfg.get("brobots", {}).items():
        mask = np.zeros(hsv.shape[:2], np.uint8)
        for lo, hi in m["hsv"]:
            mask |= cv2.inRange(hsv, np.array(lo), np.array(hi))
        mask = cv2.morphologyEx(mask, cv2.MORPH_OPEN, np.ones((3, 3), np.uint8))
        best = None
        for c in cv2.findContours(mask, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)[0]:
            (cx, cy), (rw, rh), _ = cv2.minAreaRect(c)
            long_mm, short_mm = max(rw, rh) * mm_px, min(rw, rh) * mm_px
            fill = cv2.contourArea(c) / max(1.0, rw * rh)
            if not (lo_mm <= long_mm <= hi_mm and short_mm >= lo_mm / 2 and fill >= fill_min):
                continue  # a strip of the track, a speck: not his piece
            if best is None or cv2.contourArea(c) > best[0]:
                best = (cv2.contourArea(c), cx, cy)
        if best:
            seen[name] = to_mm(bar, best[1], best[2])
    return seen


# ## EOF
