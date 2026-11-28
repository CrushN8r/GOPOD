#!/usr/bin/env python3
"""GOPOD's virtual stage - an app on the VR/AR core (gopod_vr_core.py). Each Brobot is spotted by
his marker (Brobot 1 red, Brobot 2 yellow, vr_markers). His spot on stage (vr_stage): either side
of the cube when it's near the centre spot, else either side of the centre spot with room for the
cube between. --go: each in turn, off his charger and driven there, the air hockey's way (short legs,
small turns, looking again); his facing is the way his marker moves."""
import argparse
import math
import os
import sys
import threading
import time

import cv2
import numpy as np

import gopod_vr_core as core
import gopod_vr_air_hockey as hockey

CUBE_MM = (35.0, 75.0)  # the cube as the camera sees it, its sides
MOVED_MM = 15.0         # his marker moved this far: the way it went is his facing


def find_cube(img, bar, centre, near):
    """The cube: a dark, square piece near the centre spot (mm), or None."""
    mm_px = core.BAR_MM / bar["w"]
    r = int(near / mm_px)
    cx, cy = int(bar["cx"] + centre[0] / mm_px), int(bar["cy"] - centre[1] / mm_px)
    x0, y0 = max(0, cx - r), max(0, cy - r)
    grey = cv2.cvtColor(img[y0:cy + r, x0:cx + r], cv2.COLOR_BGR2GRAY)
    dark = cv2.inRange(grey, 0, int(0.65 * np.median(grey)))
    best = None
    for c in cv2.findContours(dark, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)[0]:
        (px, py), (w, h), _ = cv2.minAreaRect(c)
        if CUBE_MM[0] <= min(w, h) * mm_px and max(w, h) * mm_px <= CUBE_MM[1] and cv2.contourArea(c) >= 0.6 * w * h:
            p = core.to_mm(bar, x0 + px, y0 + py)
            d = math.hypot(p[0] - centre[0], p[1] - centre[1])
            if d <= near and (best is None or d < best[0]):
                best = (d, p)
    return best[1] if best else None


def spots(img, bar):
    """Each Brobot's spot on stage (mm), and the cube (or None)."""
    st = core.settings("vr_stage")
    centre = tuple(st.get("centre_mm", [0, 150]))
    cube = find_cube(img, bar, centre, st.get("cube_near_mm", 120))
    mid = cube or centre
    return {name: (mid[0] + side * st.get("side_mm", 75), mid[1]) for name, side in st.get("sides", {}).items()}, cube


class Camera:
    """The newest frame, read all the time in the background."""

    def __init__(self):
        self.cap, self.img, self.on = core.open_camera(), None, True
        self.img = core.one_frame(self.cap)
        threading.Thread(target=self.run, daemon=True).start()

    def run(self):
        while self.on:
            ok, img = self.cap.read()
            if ok:
                self.img = img

    def close(self):
        self.on = False
        time.sleep(0.2)
        self.cap.release()


def follower(cam, bar, name):
    """frame_pos for the air hockey's drive_path: his marker's spot from the newest frame; his facing
    (seen["heading"]) the way it last moved, up the table to start (on his charger he faces up)."""
    seen = {"heading": math.pi / 2, "pos": None}

    def frame_pos():
        p = core.find_markers(cam.img, bar).get(name)
        if p is None:
            return seen["pos"]
        if seen["pos"] is None:
            seen["pos"] = p
        elif math.hypot(p[0] - seen["pos"][0], p[1] - seen["pos"][1]) >= MOVED_MM:
            going = math.atan2(p[1] - seen["pos"][1], p[0] - seen["pos"][0])
            seen["heading"] = going + (math.pi if frame_pos.backing else 0.0)
            seen["pos"] = p
        return p

    frame_pos.seen, frame_pos.backing = seen, False
    return frame_pos


class StageRobot(hockey.Robot):
    """The air hockey's Brobot; a turn also turns the facing he's followed by, until his marker
    moves again."""

    def turn(self, err):
        super().turn(err)
        self.frame_pos.seen["heading"] += (err + math.pi) % (2 * math.pi) - math.pi


def draw(img, bar, seen, places, cube):
    mm_px = core.BAR_MM / bar["w"]
    px = lambda p: (int(bar["cx"] + p[0] / mm_px), int(bar["cy"] - p[1] / mm_px))
    out = img.copy()
    for name, p in seen.items():
        cv2.circle(out, px(p), 28, (255, 255, 255), 3)
        cv2.putText(out, name, (px(p)[0] + 32, px(p)[1]), cv2.FONT_HERSHEY_SIMPLEX, 0.9, (255, 255, 255), 2)
    for name, p in places.items():
        cv2.drawMarker(out, px(p), (255, 255, 255), cv2.MARKER_CROSS, 40, 3)
    if cube:
        cv2.rectangle(out, (px(cube)[0] - 20, px(cube)[1] - 20), (px(cube)[0] + 20, px(cube)[1] + 20), (255, 255, 255), 2)
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--photo", default="", help="a saved overhead picture instead of the camera (tests)")
    ap.add_argument("--out", default="", help="save the marked picture here")
    ap.add_argument("--go", action="store_true", help="each Brobot off his charger and onto his spot - drives the robots")
    args = ap.parse_args()
    cam = None
    if args.photo:
        img = cv2.imread(args.photo)
    else:
        os.dup2(os.open(os.devnull, os.O_WRONLY), 2)  # the camera's "Corrupt JPEG data" notes are harmless
        cam = Camera()
        img = cam.img
    if img is None:
        print("no picture - is the RunCam plugged in?")
        return 1
    bar = core.arena(img)
    if bar is None:
        print("no ruler bar in view - clear anything dark off it")
        return 1
    seen = core.find_markers(img, bar)
    places, cube = spots(img, bar)
    print("the cube: %s" % ("%.0f, %.0f mm" % cube if cube else "not near the centre spot - room left for it"))
    for name, p in places.items():
        at = seen.get(name)
        print("%s: %s -> his spot %.0f, %.0f mm" % (name, "%.0f, %.0f mm" % at if at else "not seen", p[0], p[1]))
    if args.out:
        cv2.imwrite(args.out, draw(img, bar, seen, places, cube))
        print("marked: " + args.out)
    if args.go and cam:
        for name, place in places.items():
            if name not in seen:
                print("%s: not seen - skipped" % name)
                continue
            robot = StageRobot(name)
            robot.frame_pos = follower(cam, bar, name)
            robot.hold()
            robot.charger("off")
            ok = robot.drive_path(robot.frame_pos, [place])
            robot.let_go()
            print("%s: %s" % (name, "on his spot" if ok else "near his spot"))
    if cam:
        cam.close()
    return 0


if __name__ == "__main__":
    sys.exit(main())

# ## EOF
