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
# With --drive (2026-11-22, operator: "The brobot doesn't react to the puck,
# wanders outside the boundary"), the Brobot is the goalie: Wire-Pod holds him
# (no wandering), and he slides along his goal line - sideways, as VectAR's
# goalie does - to meet the puck, never past the field's sides. Place him on the
# red line, turned sideways; a short test nudge learns which way he faces.
# Without --drive it only watches.
#   python gopod_vr_air_hockey_play_001.py --brobot "Brobot 1"            (a window - close it to stop)
#   python gopod_vr_air_hockey_play_001.py --frames 150 --out test.mp4     (no window)
import argparse
import json
import math
import urllib.parse
import urllib.request
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
# 2026-11-22 (operator: "give both sides goal posts"): a goal mouth at each end,
# this share of the field's width, a post at each edge; the rest of the end a wall.
GOAL_SHARE = 0.5
PHYSICS_STEP = 1 / 120.0  # the puck moves in even small steps - no jumps on a slow frame
BODY_SMOOTH = 0.5         # his seen position, smoothed (camera jitter isn't motion)
BODY_V_MIN = 40.0         # mm/s - slower than this he's standing still (no push)
# The goalie (--drive): wheel speed toward the puck's line, held to the field.
GOALIE_GAIN = 2.5         # mm/s of wheel speed per mm he's off the target
GOALIE_MAX = 120.0        # mm/s
GOALIE_DEADBAND = 12.0    # mm - this close, he stops
DRIVE_EVERY = 0.12        # s between wheel commands
SETTLE_AFTER_ASSUME = 1.5 # s - a fresh hold needs a moment before the first move (CLAUDE.md gotcha)


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

    def advance(self, dt, body, body_v):
        """dt seconds in even small steps; "his" (past him, the bar end), "far"
        (past the operator's end) or "" - no goal."""
        steps = max(1, int(math.ceil(dt / PHYSICS_STEP)))
        for _ in range(steps):
            goal = self.step(dt / steps, body, body_v)
            if goal:
                return goal
        return ""

    def step(self, dt, body, body_v):
        """Move dt seconds; bounce off the side walls, the far (operator's) end and
        the robot's body. True when it crosses his goal line, the bar end."""
        self.x += self.vx * dt
        self.y += self.vy * dt
        if self.x - PUCK_R < FIELD_X[0]:
            self.x, self.vx = FIELD_X[0] + PUCK_R, abs(self.vx) * WALL_RESTITUTION
        if self.x + PUCK_R > FIELD_X[1]:
            self.x, self.vx = FIELD_X[1] - PUCK_R, -abs(self.vx) * WALL_RESTITUTION
        mouth = GOAL_SHARE * (FIELD_X[1] - FIELD_X[0]) / 2
        in_mouth = abs(self.x) < mouth - PUCK_R
        if self.y + PUCK_R > FIELD_Y[1] and not in_mouth:
            self.y, self.vy = FIELD_Y[1] - PUCK_R, -abs(self.vy) * WALL_RESTITUTION
        if self.y - PUCK_R < FIELD_Y[0] and not in_mouth:
            self.y, self.vy = FIELD_Y[0] + PUCK_R, abs(self.vy) * WALL_RESTITUTION
        if body is not None:
            dx, dy = self.x - body[0], self.y - body[1]
            d = math.hypot(dx, dy)
            if 0 < d < PUCK_R + VECTOR_BODY_R:
                nx, ny = dx / d, dy / d
                self.x, self.y = body[0] + nx * (PUCK_R + VECTOR_BODY_R), body[1] + ny * (PUCK_R + VECTOR_BODY_R)
                dot = self.vx * nx + self.vy * ny
                if dot < 0:  # moving into him: reflect along the contact normal, once
                    self.vx -= 2 * dot * nx
                    self.vy -= 2 * dot * ny
                    self.vx += PADDLE_VEL_TRANSFER * body_v[0]
                    self.vy += PADDLE_VEL_TRANSFER * body_v[1]
                    self.clamp_speed()
        if self.y < FIELD_Y[0]:
            return "his"
        if self.y > FIELD_Y[1]:
            return "far"
        return ""


class Goalie:
    """The Brobot as goalie, through Wire-Pod's SDK page (its address in
    GOPOD_WIREPOD_BASE_URL; his serial in ~/.gopod_esns.json, private)."""

    def __init__(self, brobot):
        self.base = os.environ.get("GOPOD_WIREPOD_BASE_URL", "").rstrip("/")
        esns = json.load(open(os.path.expanduser("~/.gopod_esns.json")))
        self.serial = esns["ESN01" if brobot == "Brobot 1" else "ESN02"]
        self.facing = 1  # +1: his forward is +x (right on the screen)
        self.last = 0.0
        self.speed = None

    def send(self, path, form):
        form = dict(form, serial=self.serial)
        req = urllib.request.Request(self.base + path + "?" + urllib.parse.urlencode(form), method="POST",
                                     headers={"Content-Type": "application/x-www-form-urlencoded"}, data=b"")
        try:
            urllib.request.urlopen(req, timeout=3).read()
        except OSError as e:
            print("Wire-Pod: " + str(e))

    def hold(self):
        self.send("/api-sdk/assume_behavior_control", {"priority": "high"})
        time.sleep(SETTLE_AFTER_ASSUME)

    def wheels(self, speed):
        self.send("/api-sdk/move_wheels", {"lw": int(speed), "rw": int(speed)})

    def drive_to(self, robot_x, target_x, now):
        if now - self.last < DRIVE_EVERY:
            return
        self.last = now
        err = target_x - robot_x
        speed = 0.0 if abs(err) < GOALIE_DEADBAND else max(-GOALIE_MAX, min(GOALIE_MAX, GOALIE_GAIN * err))
        speed *= self.facing
        if self.speed is None or abs(speed - self.speed) > 5:
            self.wheels(speed)
            self.speed = speed

    def let_go(self):
        self.wheels(0)
        self.send("/api-sdk/release_behavior_control", {})


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--brobot", default="Brobot 1", choices=["Brobot 1", "Brobot 2"])
    ap.add_argument("--width", type=int, default=1920)  # 1080p: the bar stands clear of the speaker
    ap.add_argument("--height", type=int, default=1080)
    ap.add_argument("--drive", action="store_true", help="he plays goalie (live robot)")
    ap.add_argument("--frames", type=int, default=0, help="stop after this many (0: until q)")
    ap.add_argument("--out", default="", help="save the drawn run to this video")
    args = ap.parse_args()
    # The camera's "Corrupt JPEG data" notes are harmless - kept off the terminal
    # (our own lines go to stdout).
    quiet = os.open(os.devnull, os.O_WRONLY)
    os.dup2(quiet, 2)
    cap = open_camera(args.width, args.height)
    # 2026-11-22 (live: "no ruler bar in view" - the first frames, before the
    # exposure settles, are too dark or bright): it keeps looking, up to 3 s.
    bar, frame = None, None
    for _ in range(90):
        ok, frame = cap.read()
        if ok:
            bar = find_bar(frame)
            if bar is not None:
                break
    if frame is None:
        print("the camera gave no frame")
        return 1
    if bar is None:
        print("no ruler bar in view - is the ruler clear of anything dark touching it?")
        return 1
    rink = rink_of(bar, frame.shape)
    # The cones are found on a half-size copy of each frame - the same finder,
    # a quarter of the pixels - so it keeps up at 30 fps.

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

    # Only the rink's area is searched (2026-11-22, live: 17 fps on the whole frame).
    cx0, cy0 = rink["x"], rink["y"]
    cx1, cy1 = rink["x"] + rink["w"], rink["y"] + rink["h"]
    shift = lambda d: dict(d, x=(d["x"] - cx0) // 2, y=(d["y"] - cy0) // 2, cx=(d["cx"] - cx0) // 2 if "cx" in d else 0,
                           cy=(d["cy"] - cy0) // 2 if "cy" in d else 0, w=d["w"] // 2, h=d["h"] // 2)
    bar_c, rink_c = shift(bar), shift(rink)

    def seen(img):
        crop = img[cy0:cy1, cx0:cx1]
        small = cv2.resize(crop, (crop.shape[1] // 2, crop.shape[0] // 2))
        c = find_brobots(cv2.cvtColor(small, cv2.COLOR_BGR2HSV), cv2.cvtColor(small, cv2.COLOR_BGR2GRAY),
                         bar_c, rink_c, 40).get(args.brobot)
        return to_mm(c["cx"] * 2 + cx0, c["cy"] * 2 + cy0) if c else None

    goalie = None
    if args.drive:
        goalie = Goalie(args.brobot)
        goalie.hold()
        # Which way is forward: a short nudge, watched from above.
        before = None
        for _ in range(10):
            ok, img = cap.read()
            before = seen(img) if ok else before
        goalie.wheels(60)
        time.sleep(0.6)
        goalie.wheels(0)
        time.sleep(0.3)
        after = None
        for _ in range(10):
            ok, img = cap.read()
            after = seen(img) if ok else after
        if before and after and abs(after[0] - before[0]) > 5:
            goalie.facing = 1 if after[0] > before[0] else -1
            print("goalie: forward is %s" % ("right" if goalie.facing > 0 else "left"))
        else:
            print("goalie: couldn't tell which way he faces - is he turned sideways, on the red line?")
    score = {"Brobot": 0, "Room": 0}
    writer = None
    if args.out:
        writer = cv2.VideoWriter(args.out, cv2.VideoWriter_fourcc(*"mp4v"), 30, (frame.shape[1] // 2, frame.shape[0] // 2))
    # 2026-11-22 (live: this env's OpenCV has no window support): the frames go
    # to ffplay, which shows them; closing its window (or q in it) ends the game.
    window = None
    if not args.out and os.environ.get("DISPLAY"):
        import subprocess
        window = subprocess.Popen(["ffplay", "-loglevel", "quiet", "-window_title", "GOPOD air hockey",
                                   "-f", "rawvideo", "-pixel_format", "bgr24",
                                   "-video_size", "%dx%d" % (frame.shape[1] // 2, frame.shape[0] // 2), "-i", "-"],
                                  stdin=subprocess.PIPE)

    last_body, last_t, n = None, time.time(), 0
    while True:
        ok, frame = cap.read()
        if not ok:
            break
        now = time.time()
        dt, last_t = min(0.1, now - last_t), now
        body = seen(frame)
        if body and last_body:
            body = (last_body[0] + BODY_SMOOTH * (body[0] - last_body[0]), last_body[1] + BODY_SMOOTH * (body[1] - last_body[1]))
        body_v = (0.0, 0.0)
        if body and last_body and dt > 0:
            body_v = ((body[0] - last_body[0]) / dt, (body[1] - last_body[1]) / dt)
            if math.hypot(*body_v) < BODY_V_MIN:
                body_v = (0.0, 0.0)
        last_body = body or last_body
        if goalie and body:
            # Where the puck will cross his line; held inside the field's sides.
            target = puck.x
            if puck.vy < 0:
                t = (puck.y - body[1]) / -puck.vy
                target = puck.x + puck.vx * t
                lo, hi = FIELD_X[0], FIELD_X[1]
                while target < lo or target > hi:  # bank shots off the side walls
                    target = 2 * lo - target if target < lo else 2 * hi - target
            room = (FIELD_X[1] - FIELD_X[0]) / 2 - VECTOR_BODY_R / 2
            target = max(-room, min(room, target))
            goalie.drive_to(body[0], target, now)
        fps = 1.0 / dt if dt > 0 else 0
        goal = puck.advance(dt, body, body_v)
        if goal:
            score["Room" if goal == "his" else "Brobot"] += 1
            if max(score.values()) >= WIN_SCORE:
                score = {"Brobot": 0, "Room": 0}
            puck.serve()
        # Draw: the field, his goal, him, the puck, the score.
        x0, y0 = to_px(FIELD_X[0], FIELD_Y[1])
        x1, y1 = to_px(FIELD_X[1], FIELD_Y[0])
        cv2.rectangle(frame, (x0, y0), (x1, y1), (255, 255, 0), 3)
        # The goal mouths, a post at each edge: his at the bar end (red), the far one (green).
        mouth = GOAL_SHARE * (FIELD_X[1] - FIELD_X[0]) / 2
        for gy, col in ((FIELD_Y[0], (0, 0, 255)), (FIELD_Y[1], (0, 200, 0))):
            a, b = to_px(-mouth, gy), to_px(mouth, gy)
            cv2.line(frame, a, b, col, 8)
            cv2.circle(frame, a, 8, (255, 255, 255), -1)
            cv2.circle(frame, b, 8, (255, 255, 255), -1)
        if body:
            cv2.circle(frame, to_px(*body), int(VECTOR_BODY_R / mm_px), CONES[args.brobot]["draw"], 3)
        cv2.circle(frame, to_px(puck.x, puck.y), int(PUCK_R / mm_px), (0, 255, 255), -1)
        cv2.putText(frame, "%s %d - %d Room - %.0f fps" % (args.brobot, score["Brobot"], score["Room"], fps), (x0, max(40, y0 - 15)),
                    cv2.FONT_HERSHEY_SIMPLEX, 1.2, (255, 255, 255), 3)
        # Shown at half size (2026-11-22: full 1080p frames to the window held it to 17 fps).
        shown = cv2.resize(frame, (frame.shape[1] // 2, frame.shape[0] // 2))
        if writer:
            writer.write(shown)
        if window:
            try:
                window.stdin.write(shown.tobytes())
            except (BrokenPipeError, OSError):
                break  # its window was closed
        n += 1
        if args.frames and n >= args.frames:
            break
    if goalie:
        goalie.let_go()
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
