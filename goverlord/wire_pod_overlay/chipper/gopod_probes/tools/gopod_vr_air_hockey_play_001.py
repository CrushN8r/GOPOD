#!/usr/bin/env python3
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_vr_air_hockey_play_001.py
#
# GOPOD VR/AR air hockey - the mechanics (2026-11-22, operator: "for now, just 1
# brobot"; "Make the play area fit my table-top, just for now, to get the
# mechanics working"; the ruler's black bar is 12 inches). The overhead camera
# (the RunCam) watches the table; a virtual puck plays on it; the Brobot's cone
# is his body - the puck bounces off him, the walls, and scores past his goal.
#
# The puck's physics are ported from VectAR-Airhokey (MIT, Pavlo Tkachenko -
# github.com/PtPavloTkachenko/VectAR-Airhokey, lens/Assets/Scripts/PuckPhysics.ts
# and GameConfig.ts): circle collisions reflected along the contact normal, walls
# at 0.98, speed held between 150 and 235 mm/s, a serve from the centre at a
# random angle within 30 degrees, first to 5. Theirs tracks the robot by wheel
# counts and Snap Spectacles; ours sees his cone from above (no drift).
#
# This step only watches and plays the puck - it never drives a robot.
#   python gopod_vr_air_hockey_play_001.py --brobot "Brobot 1"            (a window - close it to stop)
#   python gopod_vr_air_hockey_play_001.py --frames 150 --out test.mp4     (no window)
import argparse
import math
import os
import random
import sys
import time

import cv2
import numpy as np

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gopod_vr_air_hockey_001 import find_bar, find_brobots, rink_of  # noqa: E402 - the one finder
from gopod_vr_cone_probe_001 import CONES  # noqa: E402

CAMERA_GLOB = "usb-Wasintek_Wasintek_camera_*-video-index0"  # the RunCam, by name
BAR_MM = 304.8          # the ruler's black bar: 12 inches
# The field, in mm from the bar's middle (x right, y up the table toward the
# operator) - 2026-11-22 (operator: "a simple, like 1:1.3 ratio rectangle,
# portrait orientation from the bar"): as wide as the bar, 1.3 times that tall.
# The Brobot's goal is the bar end; the operator's end is a wall for now.
FIELD_X = (-BAR_MM / 2, BAR_MM / 2)
FIELD_Y = (0.0, 1.3 * BAR_MM)  # capped just short of the orange track, where the operator sits
TRACK_GAP_MM = 20.0
# From VectAR-Airhokey's GameConfig.ts.
PUCK_R = 15.0
VECTOR_BODY_R = 55.0
WALL_RESTITUTION = 0.98
PUCK_SERVE_SPEED = 135.0
PUCK_MIN_SPEED = 150.0
PUCK_MAX_SPEED = 235.0
PADDLE_VEL_TRANSFER = 0.6
WIN_SCORE = 5


def open_camera(width, height):
    import glob
    found = glob.glob(os.path.join("/dev/v4l/by-id", CAMERA_GLOB))
    if not found:
        sys.exit("no RunCam found (" + CAMERA_GLOB + ")")
    cap = cv2.VideoCapture(found[0], cv2.CAP_V4L2)
    cap.set(cv2.CAP_PROP_FOURCC, cv2.VideoWriter_fourcc(*"MJPG"))
    cap.set(cv2.CAP_PROP_FRAME_WIDTH, width)
    cap.set(cv2.CAP_PROP_FRAME_HEIGHT, height)
    cap.set(cv2.CAP_PROP_FPS, 30)
    return cap


class Puck:
    def __init__(self):
        self.serve()

    def serve(self):
        self.x, self.y = 0.0, (FIELD_Y[0] + FIELD_Y[1]) / 2
        a = math.radians(random.uniform(-30, 30))
        self.vx = PUCK_SERVE_SPEED * math.sin(a)
        self.vy = -PUCK_SERVE_SPEED * math.cos(a)  # toward his goal, the bar end

    def clamp_speed(self):
        s = math.hypot(self.vx, self.vy)
        if s == 0:
            return
        k = min(max(s, PUCK_MIN_SPEED), PUCK_MAX_SPEED) / s
        self.vx, self.vy = self.vx * k, self.vy * k

    def step(self, dt, body, body_v):
        """Move dt seconds; bounce off the side walls, the far (operator's) end and
        the robot's body. True when it crosses his goal line, the bar end."""
        self.x += self.vx * dt
        self.y += self.vy * dt
        if self.x - PUCK_R < FIELD_X[0]:
            self.x, self.vx = FIELD_X[0] + PUCK_R, abs(self.vx) * WALL_RESTITUTION
        if self.x + PUCK_R > FIELD_X[1]:
            self.x, self.vx = FIELD_X[1] - PUCK_R, -abs(self.vx) * WALL_RESTITUTION
        if self.y + PUCK_R > FIELD_Y[1]:
            self.y, self.vy = FIELD_Y[1] - PUCK_R, -abs(self.vy) * WALL_RESTITUTION
        if body is not None:
            dx, dy = self.x - body[0], self.y - body[1]
            d = math.hypot(dx, dy)
            if 0 < d < PUCK_R + VECTOR_BODY_R:
                nx, ny = dx / d, dy / d
                self.x, self.y = body[0] + nx * (PUCK_R + VECTOR_BODY_R), body[1] + ny * (PUCK_R + VECTOR_BODY_R)
                dot = self.vx * nx + self.vy * ny
                if dot < 0:  # moving into him: reflect along the contact normal
                    self.vx -= 2 * dot * nx
                    self.vy -= 2 * dot * ny
                self.vx += PADDLE_VEL_TRANSFER * body_v[0]
                self.vy += PADDLE_VEL_TRANSFER * body_v[1]
                self.clamp_speed()
        return self.y < FIELD_Y[0]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--brobot", default="Brobot 1", choices=["Brobot 1", "Brobot 2"])
    ap.add_argument("--width", type=int, default=1920)
    ap.add_argument("--height", type=int, default=1080)
    ap.add_argument("--frames", type=int, default=0, help="stop after this many (0: until q)")
    ap.add_argument("--out", default="", help="save the drawn run to this video")
    args = ap.parse_args()
    cap = open_camera(args.width, args.height)
    ok, frame = cap.read()
    for _ in range(10):  # let the exposure settle
        ok, frame = cap.read()
    if not ok:
        sys.exit("the camera gave no frame")
    bar = find_bar(frame)
    if bar is None:
        sys.exit("no ruler bar in view")
    rink = rink_of(bar, frame.shape)
    hsv0 = cv2.cvtColor(frame, cv2.COLOR_BGR2HSV)
    mm_px = BAR_MM / bar["w"]

    def to_mm(px, py):
        return (px - bar["cx"]) * mm_px, (bar["cy"] - py) * mm_px

    def to_px(x, y):
        return int(bar["cx"] + x / mm_px), int(bar["cy"] - y / mm_px)

    # The field stops just short of the orange track (the operator's end): the
    # lowest track pixel above the bar, between its ends - 1.3 bars at most.
    from gopod_vr_cone_probe_001 import WARM
    warm = np.zeros(frame.shape[:2], np.uint8)
    for lo, hi in WARM:
        warm |= cv2.inRange(hsv0, np.array(lo), np.array(hi))
    cols = warm[: bar["y"], bar["x"]: bar["x"] + bar["w"]]
    rows = np.nonzero(cols.sum(axis=1) > 0.3 * 255 * bar["w"])[0]
    if len(rows):
        track_mm = (bar["cy"] - rows.max()) * mm_px
        global FIELD_Y
        FIELD_Y = (0.0, min(FIELD_Y[1], track_mm - TRACK_GAP_MM))
    puck = Puck()
    score = {"saves": 0, "goals": 0}
    writer = None
    if args.out:
        writer = cv2.VideoWriter(args.out, cv2.VideoWriter_fourcc(*"mp4v"), 30, (frame.shape[1], frame.shape[0]))
    # 2026-11-22 (live: this env's OpenCV has no window support): the frames go
    # to ffplay, which shows them; closing its window (or q in it) ends the game.
    window = None
    if not args.out and os.environ.get("DISPLAY"):
        import subprocess
        window = subprocess.Popen(["ffplay", "-loglevel", "quiet", "-window_title", "GOPOD air hockey",
                                   "-f", "rawvideo", "-pixel_format", "bgr24",
                                   "-video_size", "%dx%d" % (frame.shape[1], frame.shape[0]), "-i", "-"],
                                  stdin=subprocess.PIPE)
    # The camera's "Corrupt JPEG data" notes are harmless - kept off the terminal.
    devnull = os.open(os.devnull, os.O_WRONLY)
    os.dup2(devnull, 2)
    last_body, last_t, n = None, time.time(), 0
    while True:
        ok, frame = cap.read()
        if not ok:
            break
        now = time.time()
        dt, last_t = min(0.1, now - last_t), now
        hsv = cv2.cvtColor(frame, cv2.COLOR_BGR2HSV)
        cone = find_brobots(hsv, cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY), bar, rink, 150).get(args.brobot)
        body = to_mm(cone["cx"], cone["cy"]) if cone else None
        body_v = (0.0, 0.0)
        if body and last_body and dt > 0:
            body_v = ((body[0] - last_body[0]) / dt, (body[1] - last_body[1]) / dt)
        last_body = body
        was_vy = puck.vy
        goal = puck.step(dt, body, body_v)
        if was_vy < 0 <= puck.vy and body and abs(puck.y - body[1]) < VECTOR_BODY_R + PUCK_R + 5:
            score["saves"] += 1  # it turned back off him
        if goal:
            score["goals"] += 1  # one Brobot: a goal is past him
            if score["goals"] >= WIN_SCORE:
                score = {"saves": 0, "goals": 0}
            puck.serve()
        # Draw: the field, his goal, him, the puck, the score.
        x0, y0 = to_px(FIELD_X[0], FIELD_Y[1])
        x1, y1 = to_px(FIELD_X[1], FIELD_Y[0])
        cv2.rectangle(frame, (x0, y0), (x1, y1), (255, 255, 0), 3)
        cv2.line(frame, (x0, y1), (x1, y1), (0, 0, 255), 8)  # his goal: the bar end
        if body:
            cv2.circle(frame, to_px(*body), int(VECTOR_BODY_R / mm_px), CONES[args.brobot]["draw"], 3)
        cv2.circle(frame, to_px(puck.x, puck.y), int(PUCK_R / mm_px), (0, 255, 255), -1)
        cv2.putText(frame, "%s - saves %d, goals past him %d" % (args.brobot, score["saves"], score["goals"]), (x0, max(40, y0 - 15)),
                    cv2.FONT_HERSHEY_SIMPLEX, 1.2, (255, 255, 255), 3)
        if writer:
            writer.write(frame)
        if window:
            try:
                window.stdin.write(frame.tobytes())
            except (BrokenPipeError, OSError):
                break  # its window was closed
        n += 1
        if args.frames and n >= args.frames:
            break
    cap.release()
    if window:
        try:
            window.stdin.close()
        except OSError:
            pass
        window.terminate()
    if writer:
        writer.release()
    print("frames %d, score %s, mm per pixel %.3f" % (n, score, mm_px))
    return 0


if __name__ == "__main__":
    sys.exit(main())
