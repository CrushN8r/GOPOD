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
ANCHOR_FILE = "~/.gopod_air_hockey_anchor.json"  # the locked bar - this machine's, never the repo's
RUN_LOG = "~/.gopod_air_hockey_last_run.log"     # each frame: where he was seen, the puck, his wheels (private)
JUMP_MM = 80.0   # a marker seen this far from where he just was is a mistake, not him
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
# 2026-11-23 (operator: "slow down the puck, so I and the brobot have a chance to
# react"): VectAR's speeds times PUCK_PACE - raise it as the arena grows.
PUCK_PACE = 0.5
PUCK_SERVE_SPEED = 135.0 * PUCK_PACE
PUCK_MIN_SPEED = 150.0 * PUCK_PACE
PUCK_MAX_SPEED = 235.0 * PUCK_PACE
PADDLE_VEL_TRANSFER = 0.6
WIN_SCORE = 5
# 2026-11-22 (operator: "give both sides goal posts"): a goal mouth at each end,
# this share of the field's width, a post at each edge; the rest of the end a wall.
GOAL_SHARE = 0.5
# The operator's paddle (2026-11-22, operator: "a sliding rounded puck deflector,
# that tracks my fingers ... on my side of the arena"): a bright blue marker on a
# fingertip, found in his half; a round paddle the size of VectAR's.
PADDLE_R = 25.0
# 2026-11-23 (operator: "my black object I'll be using for my deflector"): the
# paddle is the dark compact thing in his half that isn't the Brobot.
PADDLE_DARK = 0.6           # darker than this share of the table's brightness
PADDLE_SIZE = (0.04, 0.3)   # its width, in bar-lengths
# 2026-11-23 (operator: "I put a light green square sticky note on the brobot"
# - the red cone and the flag tracked poorly live): Brobot 1's marker. Under this
# light it reads pale yellow-green (hue ~25, saturation ~20, bright).
NOTE_HSV = ((15, 12, 150), (45, 255, 255))
NOTE_SIZE = (0.08, 0.25)    # its width, in bar-lengths
# --home: off the charger and to the middle of his goal line, in short bursts.
HOME_Y = VECTOR_BODY_R + 15.0   # mm up from the bar - just in front of his net
HOME_NEAR = 30.0                # mm - close enough
BURST_SPEED = 60.0              # mm/s forward bursts
BURST_SECONDS = 0.5
TURN_SPEED = 40.0               # mm/s each wheel, opposite ways
TURN_RATE = 2 * TURN_SPEED / 48.0  # rad/s - Vector's wheels are about 48 mm apart
PHYSICS_STEP = 1 / 120.0  # the puck moves in even small steps - no jumps on a slow frame
BODY_SMOOTH = 0.5         # his seen position, smoothed (camera jitter isn't motion)
BODY_V_MIN = 40.0         # mm/s - slower than this he's standing still (no push)
# The goalie (--drive): wheel speed toward the puck's line, held to the field.
GOALIE_GAIN = 2.5         # mm/s of wheel speed per mm he's off the target
GOALIE_MAX = 120.0        # mm/s
GOALIE_DEADBAND = 12.0    # mm - this close, he stops
DRIVE_EVERY = 0.12        # s between wheel commands
LINE_GAIN = 0.6           # wheel-speed difference per mm off his goal line
LINE_DAMP = 0.3           # less as he's already heading back
LINE_MAX = 35.0           # mm/s at most between the wheels
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

    def advance(self, dt, movers):
        """dt seconds in even small steps; "his" (past him, the bar end), "far"
        (past the operator's end) or "" - no goal."""
        steps = max(1, int(math.ceil(dt / PHYSICS_STEP)))
        for _ in range(steps):
            goal = self.step(dt / steps, movers)
            if goal:
                return goal
        return ""

    def step(self, dt, movers):
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
        for body, body_v, r in movers:  # the Brobot, the operator's paddle
            dx, dy = self.x - body[0], self.y - body[1]
            d = math.hypot(dx, dy)
            if 0 < d < PUCK_R + r:
                nx, ny = dx / d, dy / d
                self.x, self.y = body[0] + nx * (PUCK_R + r), body[1] + ny * (PUCK_R + r)
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
        self.sent = (0, 0)

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

    def wheels(self, lw, rw=None):
        rw = lw if rw is None else rw
        self.sent = (int(lw), int(rw))
        self.send("/api-sdk/move_wheels", {"lw": int(lw), "rw": int(rw)})

    def stop(self):
        if self.speed != (0, 0):
            self.wheels(0)
            self.speed = (0, 0)

    def home(self, seen):
        """Off the charger and to the middle of his goal line, then sideways -
        short forward bursts show which way he faces; turns between them."""
        def burst():
            before = seen()
            self.wheels(BURST_SPEED)
            time.sleep(BURST_SECONDS)
            self.wheels(0)
            time.sleep(0.3)
            after = seen()
            if before and after and math.hypot(after[0] - before[0], after[1] - before[1]) > 8:
                return after, math.atan2(after[1] - before[1], after[0] - before[0])
            return after, None

        def turn(err):
            err = (err + math.pi) % (2 * math.pi) - math.pi
            sgn = 1 if err > 0 else -1  # left (counter-clockwise, seen from above) when positive
            self.wheels(-sgn * TURN_SPEED, sgn * TURN_SPEED)
            time.sleep(abs(err) / TURN_RATE)
            self.wheels(0)
            time.sleep(0.3)

        for _ in range(25):
            pos, heading = burst()
            if pos is None:
                print("home: he's out of sight - stopping")
                return False
            if math.hypot(pos[0], pos[1] - HOME_Y) < HOME_NEAR:
                break
            if heading is not None:
                turn(math.atan2(HOME_Y - pos[1], -pos[0]) - heading)
        # Sideways, along his goal line: the nearer of facing right or left.
        for _ in range(4):
            pos, heading = burst()
            if heading is None:
                continue
            err = min((0 - heading, math.pi - heading), key=lambda e: abs((e + math.pi) % (2 * math.pi) - math.pi))
            if abs((err + math.pi) % (2 * math.pi) - math.pi) < math.radians(12):
                self.facing = 1 if abs(heading) < math.pi / 2 else -1
                print("home: in front of his net, facing %s" % ("right" if self.facing > 0 else "left"))
                return True
            turn(err)
        print("home: close - couldn't square him up")
        return True

    def drive_to(self, body, body_v, target_x, now):
        """Slide along his goal line to target_x - and steer back onto the line
        (2026-11-23, live: "Brobot is all over the place" - not quite sideways,
        each slide took him up or down the field and nothing brought him back)."""
        if now - self.last < DRIVE_EVERY:
            return
        self.last = now
        err = target_x - body[0]
        speed = 0.0 if abs(err) < GOALIE_DEADBAND else max(-GOALIE_MAX, min(GOALIE_MAX, GOALIE_GAIN * err))
        # 2026-11-23 (live: "going back & forth blindly off boundaries"): never
        # outward past the field's sides - only back in.
        room = (FIELD_X[1] - FIELD_X[0]) / 2 - VECTOR_BODY_R / 2
        if (body[0] >= room and err > 0) or (body[0] <= -room and err < 0):
            speed = 0.0
        speed *= self.facing
        # Off his line: one wheel a little faster bends him back.
        steer = 0.0
        if speed:
            steer = LINE_GAIN * (HOME_Y - body[1]) - LINE_DAMP * body_v[1]
            # Moving toward +x, turning counter-clockwise lifts him; toward -x, it lowers him.
            steer = max(-LINE_MAX, min(LINE_MAX, steer)) * (1 if speed * self.facing > 0 else -1)
        lw, rw = speed - steer, speed + steer
        if self.speed is None or abs(lw - self.speed[0]) > 4 or abs(rw - self.speed[1]) > 4:
            self.wheels(lw, rw)
            self.speed = (lw, rw)

    def let_go(self):
        self.wheels(0)
        self.send("/api-sdk/release_behavior_control", {})


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--brobot", default="Brobot 1", choices=["Brobot 1", "Brobot 2"])
    ap.add_argument("--width", type=int, default=1920)  # 1080p: the bar stands clear of the speaker
    ap.add_argument("--height", type=int, default=1080)
    ap.add_argument("--drive", action="store_true", help="he plays goalie (live robot)")
    ap.add_argument("--home", action="store_true", help="with --drive: off the charger and to his net first")
    ap.add_argument("--recalibrate", action="store_true", help="find the ruler bar again (after the camera moved)")
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
    # 2026-11-22 (live: a Brobot parked at the bar's end made it read too long):
    # the bar is found once, while clear, and locked (the camera doesn't move);
    # --recalibrate finds it again.
    anchor = os.path.expanduser(ANCHOR_FILE)
    bar, frame = None, None
    if os.path.exists(anchor) and not args.recalibrate:
        bar = json.load(open(anchor))
    for _ in range(90):
        ok, frame = cap.read()
        if ok and bar is None:
            bar = find_bar(frame)
            if bar is not None:
                json.dump(bar, open(anchor, "w"))
                print("anchor: the ruler bar locked (%d px wide) - --recalibrate if the camera moves" % bar["w"])
        if ok and bar is not None:
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
    # 2026-11-23 (live: "arena is off the table again" - an arm over the track at
    # start hid it, so the field ran off the table): found once and locked with
    # the bar (the anchor file); never looked for again unless --recalibrate.
    global FIELD_Y
    if "top_mm" not in bar:
        from gopod_vr_cone_probe_001 import WARM
        warm = np.zeros(frame.shape[:2], np.uint8)
        for lo, hi in WARM:
            warm |= cv2.inRange(hsv0, np.array(lo), np.array(hi))
        cols = warm[: bar["y"], bar["x"]: bar["x"] + bar["w"]]
        rows = np.nonzero(cols.sum(axis=1) > 0.3 * 255 * bar["w"])[0]
        if not len(rows):
            print("the orange track isn't in view (an arm over it?) - clear it and start again")
            cap.release()
            return 1
        bar["top_mm"] = min(FIELD_Y[1], (bar["cy"] - rows.max()) * mm_px - TRACK_GAP_MM)
        json.dump(bar, open(anchor, "w"))
        print("anchor: the field's far end locked too")
    FIELD_Y = (0.0, bar["top_mm"])
    puck = Puck()

    def find_note(img):
        """Brobot 1's light-green note: the biggest note-coloured square in the rink."""
        crop = img[cy0:cy1, cx0:cx1]
        small = cv2.resize(crop, (crop.shape[1] // 2, crop.shape[0] // 2))
        m = cv2.inRange(cv2.cvtColor(small, cv2.COLOR_BGR2HSV), np.array(NOTE_HSV[0]), np.array(NOTE_HSV[1]))
        m = cv2.morphologyEx(m, cv2.MORPH_OPEN, np.ones((3, 3), np.uint8))
        cs, _ = cv2.findContours(m, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        best = None
        for c in cs:
            x, y, w, h = cv2.boundingRect(c)
            size = 2 * max(w, h) / bar["w"]
            if NOTE_SIZE[0] <= size <= NOTE_SIZE[1] and 0.5 <= w / max(1, h) <= 2.0:
                if best is None or cv2.contourArea(c) > cv2.contourArea(best[0]):
                    best = (c, x, y, w, h)
        if best is None:
            return None
        _, x, y, w, h = best
        return to_mm(cx0 + (x + w / 2) * 2, cy0 + (y + h / 2) * 2)

    # Only the rink's area is searched (2026-11-22, live: 17 fps on the whole frame).
    cx0, cy0 = rink["x"], rink["y"]
    cx1, cy1 = rink["x"] + rink["w"], rink["y"] + rink["h"]
    shift = lambda d: dict(d, x=(d["x"] - cx0) // 2, y=(d["y"] - cy0) // 2, cx=(d["cx"] - cx0) // 2 if "cx" in d else 0,
                           cy=(d["cy"] - cy0) // 2 if "cy" in d else 0, w=d["w"] // 2, h=d["h"] // 2)
    bar_c, rink_c = shift(bar), shift(rink)

    def seen(img):
        if args.brobot == "Brobot 1":
            return find_note(img)
        crop = img[cy0:cy1, cx0:cx1]
        small = cv2.resize(crop, (crop.shape[1] // 2, crop.shape[0] // 2))
        c = find_brobots(cv2.cvtColor(small, cv2.COLOR_BGR2HSV), cv2.cvtColor(small, cv2.COLOR_BGR2GRAY),
                         bar_c, rink_c, 15).get(args.brobot)  # a cone is small at half size
        return to_mm(c["cx"] * 2 + cx0, c["cy"] * 2 + cy0) if c else None

    goalie = None
    if args.drive:
        goalie = Goalie(args.brobot)
        goalie.hold()
        def seen_now():
            got = None
            for _ in range(6):
                ok, img = cap.read()
                got = seen(img) if ok else got
            return got

        if args.home:
            goalie.home(seen_now)
        else:
            # Which way is forward: a short nudge, watched from above.
            before = seen_now()
            goalie.wheels(60)
            time.sleep(0.6)
            goalie.wheels(0)
            time.sleep(0.3)
            after = seen_now()
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

    last_body, last_paddle, last_t, n = None, None, time.time(), 0
    log = open(os.path.expanduser(RUN_LOG), "w")
    log.write("# t  seen_x seen_y  puck_x puck_y puck_vy  wheels_l wheels_r  (mm, mm/s) - field x %.0f..%.0f y %.0f..%.0f\n"
              % (FIELD_X[0], FIELD_X[1], FIELD_Y[0], FIELD_Y[1]))
    # The operator's half: the paddle is looked for there only.
    hx0, hy0 = to_px(FIELD_X[0], FIELD_Y[1])
    hx1, hy1 = to_px(FIELD_X[1], (FIELD_Y[0] + FIELD_Y[1]) / 2)

    def find_paddle(img, body):
        """The operator's black deflector: the dark compact thing in his half
        (a little past his end too) that isn't the Brobot."""
        top = max(0, hy0 - int(0.15 * bar["w"]))
        crop = img[top:hy1, max(0, hx0):hx1]
        small = cv2.resize(crop, (max(1, crop.shape[1] // 2), max(1, crop.shape[0] // 2)))
        grey = cv2.cvtColor(small, cv2.COLOR_BGR2GRAY)
        m = cv2.inRange(grey, 0, int(PADDLE_DARK * np.percentile(grey, 75)))
        m = cv2.morphologyEx(m, cv2.MORPH_OPEN, np.ones((5, 5), np.uint8))
        cs, _ = cv2.findContours(m, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        best = None
        for c in cs:
            x, y, w, h = cv2.boundingRect(c)
            size = 2 * max(w, h) / bar["w"]
            if not PADDLE_SIZE[0] <= size <= PADDLE_SIZE[1] or not 0.5 <= w / max(1, h) <= 2.0:
                continue
            # 2026-11-23 (operator: "reposition my blue puck-deflector forward a bit,
            # so it sits half & half on the arena border line"): the paddle's centre
            # is the deflector's front edge, the side facing the Brobot.
            p = to_mm(max(0, hx0) + (x + w / 2) * 2, top + (y + h) * 2)
            if body and math.hypot(p[0] - body[0], p[1] - body[1]) < VECTOR_BODY_R * 1.5:
                continue  # that's the Brobot
            if best is None or cv2.contourArea(c) > best[0]:
                best = (cv2.contourArea(c), p)
        return best[1] if best else None
    while True:
        ok, frame = cap.read()
        if not ok:
            break
        now = time.time()
        dt, last_t = min(0.1, now - last_t), now
        body = seen(frame)
        if body and last_body and math.hypot(body[0] - last_body[0], body[1] - last_body[1]) > JUMP_MM:
            body = None  # a jump: not him (2026-11-23)
        if body and last_body:
            body = (last_body[0] + BODY_SMOOTH * (body[0] - last_body[0]), last_body[1] + BODY_SMOOTH * (body[1] - last_body[1]))
        body_v = (0.0, 0.0)
        if body and last_body and dt > 0:
            body_v = ((body[0] - last_body[0]) / dt, (body[1] - last_body[1]) / dt)
            if math.hypot(*body_v) < BODY_V_MIN:
                body_v = (0.0, 0.0)
        last_body = body or last_body
        if goalie and not body:
            goalie.stop()  # out of sight: he stops (2026-11-22, live: he ran on and on)
        elif goalie and abs(body[0]) > (FIELD_X[1] - FIELD_X[0]) / 2 + VECTOR_BODY_R:
            goalie.stop()  # past the field's side: he stops
        elif goalie and puck.vy >= 0:
            # 2026-11-23 (operator: "after the brobot deflects the puck, stop until
            # I hit the puck or it hits my end"): going away from him, he waits.
            goalie.stop()
        elif goalie and body:
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
            goalie.drive_to(body, body_v, target, now)
        fps = 1.0 / dt if dt > 0 else 0
        log.write("%.2f  %s  %.0f %.0f %.0f  %s\n" % (now % 1000, "%.0f %.0f" % body if body else "- -",
                                                     puck.x, puck.y, puck.vy, "%d %d" % goalie.sent if goalie else "- -"))
        paddle = find_paddle(frame, body)
        paddle_v = (0.0, 0.0)
        if paddle and last_paddle and dt > 0:
            paddle_v = ((paddle[0] - last_paddle[0]) / dt, (paddle[1] - last_paddle[1]) / dt)
        last_paddle = paddle
        movers = []
        if body:
            movers.append((body, body_v, VECTOR_BODY_R))
        if paddle:
            movers.append((paddle, paddle_v, PADDLE_R))
        goal = puck.advance(dt, movers)
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
        if paddle:
            cv2.circle(frame, to_px(*paddle), int(PADDLE_R / mm_px), (255, 120, 0), -1)
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
    log.close()
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
