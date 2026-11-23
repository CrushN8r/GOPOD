#!/usr/bin/env python3
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_vr_air_hockey.py
#
# GOPOD VR/AR air hockey. The overhead camera (the RunCam) watches the table, a
# virtual puck plays on it, and two Brobots take part:
#   Brobot A - the goalie: he wears the pink square. He parks himself in
#              front of his net (like parallel parking) and slides along his goal
#              line to block the puck.
#   Brobot B - the referee and the Brobot team's cheerleader: he starts each
#              round (a nod, "Round 1, go!"), keeps and calls the score (first to
#              5 wins), and raises and lowers his arm as the puck nears A's net.
# The operator plays with his black deflector in his half (a round paddle on its
# front edge). Run it with the alias: gohockey (Brobot 1 plays, Brobot 2 is the
# referee) or gohockey 2 (the other way round). Close the window to stop.
#
# The arena is locked to the ruler's black bar (12 inches) and the orange track:
# found once, kept in ~/.gopod_air_hockey_anchor.json (this machine's, never the
# repo's); --recalibrate finds them again if the camera moves. Each run's frames
# (where A was seen, the puck, his wheels) go to ~/.gopod_air_hockey_last_run.log.
#
# The puck's physics are ported from VectAR-Airhokey (MIT, Pavlo Tkachenko -
# github.com/PtPavloTkachenko/VectAR-Airhokey, PuckPhysics.ts and GameConfig.ts).
# The referee's moves are GOPOD's golden "Got it!" moves (gopod_arm_cue.go and
# chat_moves in 03_chat_page.json): the arm per Brobot, the nod at speed 4, 0.18 s.
import argparse
import glob
import json
import math
import os
import random
import subprocess
import sys
import threading
import time
import urllib.parse
import urllib.request

import cv2
import numpy as np

CAMERA_GLOB = "usb-Wasintek_Wasintek_camera_*-video-index0"  # the RunCam, by name
ANCHOR_FILE = "~/.gopod_air_hockey_anchor.json"
RUN_LOG = "~/.gopod_air_hockey_last_run.log"
DRIVE_LOG = "~/.gopod_air_hockey_last_drive.log"  # his parking drive: time, where, which point, wheels
BAR_MM = 304.8            # the ruler's black bar: 12 inches
BAR_DARK = 115            # the bar reads mid-grey under this light (~90; the table ~170)
TRACK_HSV = [((168, 100, 150), (179, 255, 255)), ((0, 100, 150), (11, 255, 255))]  # the orange track
TRACK_GAP_MM = 20.0       # the field stops this short of the track
# The field, mm from the bar's middle (x right, y up the table toward the operator):
# the bar's width, up to just short of the track (1.3 bars at most).
FIELD_X = (-BAR_MM / 2, BAR_MM / 2)
FIELD_Y = (0.0, 1.3 * BAR_MM)
GOAL_SHARE = 0.5          # each end's goal mouth, this share of the field's width
# From VectAR-Airhokey's GameConfig.ts, the speeds halved (operator: "slow down the
# puck, so I and the brobot have a chance to react") - raise PUCK_PACE as the arena grows.
PUCK_R = 15.0
PUCK_PACE = 0.5
PUCK_SERVE_SPEED = 135.0 * PUCK_PACE
PUCK_MIN_SPEED = 150.0 * PUCK_PACE
PUCK_MAX_SPEED = 235.0 * PUCK_PACE
WALL_RESTITUTION = 0.98
PADDLE_VEL_TRANSFER = 0.6
WIN_SCORE = 5
PHYSICS_STEP = 1 / 120.0  # the puck moves in small even steps
FRAME_DT = 1 / 30.0       # a steady 1/30 s a frame (real frame timing jittered it)
# The operator's paddle: the dark compact thing in his half that isn't the Brobot;
# its centre on the deflector's front edge (half and half on the border).
PADDLE_R = 25.0
PADDLE_DARK = 0.6
PADDLE_SIZE = (0.04, 0.3)  # its width, in bar-lengths
# Brobot A's marker: a pink square (measured 2026-11-23: hue 168-169, saturation 115-126,
# brightness 219-234) - three times the green note's colour; the red track reads pale
# (saturation ~42), the orange track's hue is ~2. Was a light-green note (hue ~29,
# saturation ~28); skin (hue ~17, saturation 60-120) stays out.
NOTE_HSV = ((160, 85, 170), (176, 175, 255))
NOTE_SIZE = (0.08, 0.25)  # its width, in bar-lengths
# His whole body: the dark shape the note sits on.
BODY_SEARCH_MM = 90.0     # his body is looked for within a Vector's size around the note (~10 cm long) -
BODY_DARK = 0.55          # so dark things beside him (the LEGO tower, his charger) can't join in
BODY_HALF = (25.0, 75.0)  # half-sizes, at least / at most (more: joined to something else)
BODY_HALF_DEFAULT = 33.0
BODY_SMOOTH = 0.3         # camera jitter isn't motion
PADDLE_SMOOTH = 0.4
STILL_MM = 4.0            # a move smaller than this is the camera's own wobble - it stays put
BODY_V_MIN = 40.0         # mm/s - slower than this he's standing still
JUMP_MM = 80.0            # seen this far from where he just was: not him...
JUMP_FRAMES = 5           # ...unless he's seen there this many frames running
# His driving, measured live on Brobots 1 and 2 (2026-11-23): 80 mm/s for 1 s moves
# him ~30 mm; wheels at 100 opposite ways turn him ~34 degrees a second.
BURST_SPEED = 80.0
TURN_SPEED = 100.0
TURN_RATE = math.radians(50.0)  # Brobot 2, measured live 2026-11-23: 52 deg in 1 s (earlier 34 - the legs correct either way)
SETTLE_AFTER_ASSUME = 1.5  # a fresh hold needs a moment before the first move
WAKE_RELEASE_SETTLE = 2.5  # the golden wake's settle (BROBOTS_WAKE_RELEASE_SETTLE_SECONDS)
STOP_AGAIN = 0.5          # standing: "stop" is said again this often (a lost one let him roll on)
# Parking, like driving a car (operator: "a simple 3 point parallel park maneuver ...
# a road/lane wide enough to drive within ... small adjustments are part of driving"):
# short straight legs, a small turn toward the next point between them (each leg shows
# where he really points); within the lane of a point, on to the next. Three points: out from the charger into open table, angled toward
# his net, then the spot in front of it - arriving sideways, along his goal line.
PATH_SPEED = 70.0         # mm/s
PATH_REAL_MM_S = 50.0     # how far he really gets per second at PATH_SPEED, starting from still
PATH_LEG_MIN, PATH_LEG_MAX = 0.6, 1.5  # seconds of each straight leg
PATH_TURN_DEG = 12.0      # off by more than this: turn toward the point before the leg
PATH_LANE = 45.0          # mm - this close to a point: on to the next
PATH_LEAVE_SECONDS = 1.5  # straight off the charger first - his direction is read from it
PATH_TIMEOUT = 45.0       # s, the whole drive at most
CHARGER_OUT_MM = 90.0     # going home: the point this far out from his charger, then backed in
CHARGER_BACK_SECONDS = 2.5
NET_Y = 55.0              # his spot, mm up from the bar (70 parked him a little far out, live)
# Goalie.
GOALIE_GAIN = 1.5
GOALIE_MAX = 100.0
GOALIE_DEADBAND = 12.0
DRIVE_EVERY = 0.12
LINE_GAIN = 0.6           # steer back onto his line
LINE_DAMP = 0.3
LINE_MAX = 35.0
STUCK_SECONDS = 0.5       # the puck pressed against him this long: trapped
BACK_OFF_SECONDS = 0.6
# The referee's golden moves (gopod_arm_cue.go: the arm per Brobot; chat_moves: the nod).
ARM_CUE = {"Brobot 1": (2.0, 1.2), "Brobot 2": (3.0, 0.5)}  # speed, hold seconds
NOD = (4.0, 0.18)
CHEER_NEAR_MM = 110.0     # the puck this near A's net, coming at it: B cheers
CHEER_EVERY = 3.0
ROUND_TEXT = "Round {n}, go!"
SCORE_TEXT = "{a}, {s}. Room, {r}."
WIN_TEXT = "{w} wins the game!"


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


class Puck:
    def __init__(self):
        self.serve()

    def serve(self):
        self.x, self.y = 0.0, (FIELD_Y[0] + FIELD_Y[1]) / 2
        a = math.radians(random.uniform(-30, 30))
        self.vx = PUCK_SERVE_SPEED * math.sin(a)
        self.vy = -PUCK_SERVE_SPEED * math.cos(a)  # toward A's net
        self.held = True  # until the referee says go

    def clamp_speed(self):
        s = math.hypot(self.vx, self.vy)
        if s:
            k = min(max(s, PUCK_MIN_SPEED), PUCK_MAX_SPEED) / s
            self.vx, self.vy = self.vx * k, self.vy * k

    def advance(self, dt, movers):
        if self.held:
            return ""
        steps = max(1, int(math.ceil(dt / PHYSICS_STEP)))
        for _ in range(steps):
            goal = self.step(dt / steps, movers)
            if goal:
                return goal
        return ""

    def step(self, dt, movers):
        """Bounce off the sides, the ends beside the goal mouths, A's body (a box)
        and the paddle (a circle); "a" past A's net, "room" past the operator's."""
        self.x += self.vx * dt
        self.y += self.vy * dt
        if self.x - PUCK_R < FIELD_X[0]:
            self.x, self.vx = FIELD_X[0] + PUCK_R, abs(self.vx) * WALL_RESTITUTION
        if self.x + PUCK_R > FIELD_X[1]:
            self.x, self.vx = FIELD_X[1] - PUCK_R, -abs(self.vx) * WALL_RESTITUTION
        in_mouth = abs(self.x) < GOAL_SHARE * (FIELD_X[1] - FIELD_X[0]) / 2 - PUCK_R
        if self.y + PUCK_R > FIELD_Y[1] and not in_mouth:
            self.y, self.vy = FIELD_Y[1] - PUCK_R, -abs(self.vy) * WALL_RESTITUTION
        if self.y - PUCK_R < FIELD_Y[0] and not in_mouth:
            self.y, self.vy = FIELD_Y[0] + PUCK_R, abs(self.vy) * WALL_RESTITUTION
        for at, v, r in movers:
            if isinstance(r, tuple):  # his body's box: its nearest point
                px = max(at[0] - r[0], min(at[0] + r[0], self.x))
                py = max(at[1] - r[1], min(at[1] + r[1], self.y))
                reach = PUCK_R
            else:
                px, py, reach = at[0], at[1], PUCK_R + r
            dx, dy = self.x - px, self.y - py
            d = math.hypot(dx, dy) or 1.0
            if d < reach:
                nx, ny = dx / d, dy / d
                self.x, self.y = px + nx * reach, py + ny * reach
                dot = self.vx * nx + self.vy * ny
                if dot < 0:  # moving into it: reflect along the contact normal, once
                    self.vx -= 2 * dot * nx
                    self.vy -= 2 * dot * ny
                    self.vx += PADDLE_VEL_TRANSFER * v[0]
                    self.vy += PADDLE_VEL_TRANSFER * v[1]
                    self.clamp_speed()
        # A goal is the scorer's point (operator: "if the puck scores in my net, its the
        # opponent's point, not mine").
        if self.y < FIELD_Y[0]:
            return "room"  # past Brobot A, into his net
        if self.y > FIELD_Y[1]:
            return "a"     # past the operator, into his net
        return ""


class Robot:
    """One Brobot through Wire-Pod's SDK page (its address in GOPOD_WIREPOD_BASE_URL;
    the serials in ~/.gopod_esns.json - both private)."""

    def __init__(self, name):
        self.name = name
        self.base = os.environ.get("GOPOD_WIREPOD_BASE_URL", "").rstrip("/")
        self.serial = json.load(open(os.path.expanduser("~/.gopod_esns.json")))["ESN01" if name == "Brobot 1" else "ESN02"]
        self.facing = 1  # +1: his forward is +x (right on the screen)
        self.last = 0.0
        self.speed = None
        self.sent = (0, 0)
        self.last_stop = 0.0

    def send(self, path, form):
        req = urllib.request.Request(self.base + path + "?" + urllib.parse.urlencode(dict(form, serial=self.serial)),
                                     method="POST", data=b"", headers={"Content-Type": "application/x-www-form-urlencoded"})
        try:
            urllib.request.urlopen(req, timeout=5).read()
        except OSError as e:
            print("Wire-Pod: " + str(e))

    def hold(self):
        """Wake him and take control (operator: "add a wake call in case the brobots
        are asleep"): GOPOD's golden wake, run_brobots_wake() in
        run_robot_control_song_001.py - assume, release, settle, assume again."""
        self.send("/api-sdk/conn_test", {})
        self.send("/api-sdk/assume_behavior_control", {"priority": "high"})
        self.send("/api-sdk/release_behavior_control", {})
        time.sleep(WAKE_RELEASE_SETTLE)
        self.send("/api-sdk/assume_behavior_control", {"priority": "high"})
        time.sleep(SETTLE_AFTER_ASSUME)

    def let_go(self):
        self.wheels(0)
        self.send("/api-sdk/release_behavior_control", {})

    def wheels(self, lw, rw=None):
        rw = lw if rw is None else rw
        if lw == 0 and rw == 0:  # a stop is said twice - a lost one lets him roll on
            self.send("/api-sdk/move_wheels", {"lw": 0, "rw": 0})
        self.sent = (int(lw), int(rw))
        self.send("/api-sdk/move_wheels", {"lw": int(lw), "rw": int(rw)})

    def stop(self, now=0.0):
        """Stopped - said again every STOP_AGAIN s while he should stand (live
        2026-11-23: a lost stop let Brobot 2 roll on off the arena)."""
        if self.speed != (0, 0) or now - self.last_stop > STOP_AGAIN:
            self.wheels(0)
            self.speed = (0, 0)
            self.last_stop = now

    def axis(self, endpoint, speed, hold):
        # Wire-Pod reads a whole number here - "2.0" read as 0 and nothing moved (live 2026-11-23).
        self.send("/api-sdk/" + endpoint, {"speed": int(round(speed))})
        time.sleep(hold)
        self.send("/api-sdk/" + endpoint, {"speed": 0})

    def nod(self):
        self.axis("move_head", -NOD[0], NOD[1])
        self.axis("move_head", NOD[0], NOD[1])

    def cheer(self):
        speed, hold = ARM_CUE[self.name]
        self.axis("move_lift", speed, hold)
        self.axis("move_lift", -speed, hold)

    def say(self, text):
        self.send("/api-sdk/say_text", {"text": text})

    # --- Brobot A: parking and goalkeeping ---

    def turn(self, err):
        err = (err + math.pi) % (2 * math.pi) - math.pi
        sgn = 1 if err > 0 else -1  # counter-clockwise, seen from above, when positive
        self.wheels(-sgn * TURN_SPEED, sgn * TURN_SPEED)
        time.sleep(abs(err) / TURN_RATE)
        self.wheels(0)
        time.sleep(0.4)

    def watch(self, frame_pos, seconds):
        """Wait, watching (the window keeps showing him); his last position."""
        end, pos = time.time() + seconds, None
        while time.time() < end:
            pos = frame_pos() or pos
        return pos

    def drive_path(self, frame_pos, points, arrive_along_x=False):
        """Through the points in short straight legs, like steering a car in small
        corrections: each leg shows which way he really points (where he went), then a
        small turn toward the next point, then the next leg. Within PATH_LANE of a
        point: on to the next. (Steering while rolling made him spin in place, live.)"""
        self.plan = points
        log = open(os.path.expanduser(DRIVE_LOG), "a")
        log.write("# drive to %s\n" % ", ".join("%.0f,%.0f" % p for p in points))
        k, heading, still, t0 = 0, None, 0, time.time()
        pos = self.watch(frame_pos, 0.2)
        while k < len(points) and time.time() - t0 < PATH_TIMEOUT and pos is not None:
            target = points[k]
            dist = math.hypot(target[0] - pos[0], target[1] - pos[1])
            if dist < PATH_LANE:
                print("drive: point %d of %d reached at %.0f, %.0f mm" % (k + 1, len(points), pos[0], pos[1]))
                k += 1
                continue
            err = 0.0
            if heading is not None:
                err = (math.atan2(target[1] - pos[1], target[0] - pos[0]) - heading + math.pi) % (2 * math.pi) - math.pi
                if abs(err) > math.radians(PATH_TURN_DEG):
                    self.turn(err)
            leg = min(PATH_LEG_MAX, max(PATH_LEG_MIN, dist / PATH_REAL_MM_S))
            self.wheels(PATH_SPEED)
            self.watch(frame_pos, leg)
            self.wheels(0)
            after = self.watch(frame_pos, 0.3)
            if after is None:
                break
            moved = math.hypot(after[0] - pos[0], after[1] - pos[1])
            log.write("%.1f  at %.0f,%.0f  point %d  turned %.0f deg  leg %.1f s  moved %.0f mm\n"
                      % (time.time() - t0, after[0], after[1], k + 1, math.degrees(err), leg, moved))
            if moved > 8:
                heading, still = math.atan2(after[1] - pos[1], after[0] - pos[0]), 0
            else:
                still += 1
                if still >= 3:  # told to drive, he hasn't moved: the wrong Brobot, or stuck
                    print("drive: he isn't moving - is the pink square on %s?" % self.name)
                    break
            pos = after
        self.wheels(0)
        log.close()
        if arrive_along_x and heading is not None:
            self.facing = 1 if math.cos(heading) > 0 else -1
        return k >= len(points)

    def park(self, frame_pos):
        """In front of his net in three points, like parking a car."""
        self.wheels(PATH_SPEED)
        time.sleep(PATH_LEAVE_SECONDS)  # straight off the charger
        pos = frame_pos()
        if pos is None:
            self.wheels(0)
            print("park: he's out of sight")
            return False
        side = 1 if pos[0] >= 0 else -1
        points = [(side * 0.75 * BAR_MM, 2.2 * NET_Y), (side * 0.4 * BAR_MM, 1.3 * NET_Y), (0.0, NET_Y)]
        ok = self.drive_path(frame_pos, points, arrive_along_x=True)
        print("park: %s, facing %s" % ("in front of his net" if ok else "near his net", "right" if self.facing > 0 else "left"))
        return ok

    def go_home(self, frame_pos, charger):
        """Back to his charger: the same kind of drive to a point out in front of
        it, then backed in (Vector docks backwards)."""
        out = (charger[0], charger[1] + CHARGER_OUT_MM) if abs(charger[0]) < BAR_MM / 2 else \
              (charger[0] - math.copysign(CHARGER_OUT_MM, charger[0]), charger[1] + CHARGER_OUT_MM / 2)
        self.wheels(PATH_SPEED)
        ok = self.drive_path(frame_pos, [out, (out[0] + (charger[0] - out[0]) * 0.3, out[1] + 25)])
        self.wheels(-PATH_SPEED)
        time.sleep(CHARGER_BACK_SECONDS)
        self.wheels(0)
        print("home: %s" % ("backed onto his charger" if ok else "near his charger"))
        return ok

    def guard(self, body, body_v, target_x, now):
        """Slide along his goal line to target_x, steering back onto the line;
        never outward past the field's sides."""
        if now - self.last < DRIVE_EVERY:
            return
        self.last = now
        err = target_x - body[0]
        speed = 0.0 if abs(err) < GOALIE_DEADBAND else max(-GOALIE_MAX, min(GOALIE_MAX, GOALIE_GAIN * err))
        room = (FIELD_X[1] - FIELD_X[0]) / 2 - BODY_HALF_DEFAULT
        if (body[0] >= room and err > 0) or (body[0] <= -room and err < 0):
            speed = 0.0
        speed *= self.facing
        steer = 0.0
        if speed:
            steer = LINE_GAIN * (NET_Y - body[1]) - LINE_DAMP * body_v[1]
            steer = max(-LINE_MAX, min(LINE_MAX, steer)) * (1 if speed * self.facing > 0 else -1)
        lw, rw = speed - steer, speed + steer
        if self.speed is None or abs(lw - self.speed[0]) > 4 or abs(rw - self.speed[1]) > 4:
            self.wheels(lw, rw)
            self.speed = (lw, rw)


def background(fn, *a):
    t = threading.Thread(target=fn, args=a, daemon=True)
    t.start()
    return t


def main():
    global FIELD_Y
    ap = argparse.ArgumentParser()
    ap.add_argument("--player", default="Brobot 1", choices=["Brobot 1", "Brobot 2"])
    ap.add_argument("--recalibrate", action="store_true", help="find the ruler and the track again (the camera moved)")
    ap.add_argument("--part", type=int, default=0, choices=[0, 1, 2, 3],
                    help="1: wake, leave home, park in front of his net, stop; 0: the whole game")
    ap.add_argument("--frames", type=int, default=0, help="stop after this many frames (tests)")
    ap.add_argument("--out", default="", help="save the run to this video instead of a window (tests)")
    args = ap.parse_args()
    referee_name = "Brobot 2" if args.player == "Brobot 1" else "Brobot 1"
    os.dup2(os.open(os.devnull, os.O_WRONLY), 2)  # the camera's "Corrupt JPEG data" notes are harmless
    cap = open_camera()

    # The arena, locked once: the ruler bar, then the track's edge.
    anchor = os.path.expanduser(ANCHOR_FILE)
    bar = None if args.recalibrate or not os.path.exists(anchor) else json.load(open(anchor))
    frame = None
    for _ in range(90):  # the first frames come before the exposure settles
        ok, frame = cap.read()
        if ok and bar is None:
            bar = find_bar(frame)
        if ok and bar is not None:
            break
    if bar is None:
        print("no ruler bar in view - clear anything dark off it")
        return 1
    mm_px = BAR_MM / bar["w"]
    to_mm = lambda px, py: ((px - bar["cx"]) * mm_px, (bar["cy"] - py) * mm_px)
    to_px = lambda x, y: (int(bar["cx"] + x / mm_px), int(bar["cy"] - y / mm_px))
    if "top_mm" not in bar:
        hsv = cv2.cvtColor(frame, cv2.COLOR_BGR2HSV)
        track = sum(cv2.inRange(hsv, np.array(lo), np.array(hi)) // 255 for lo, hi in TRACK_HSV)
        rows = np.nonzero(track[: bar["y"], bar["x"]: bar["x"] + bar["w"]].sum(axis=1) > 0.3 * bar["w"])[0]
        if not len(rows):
            print("the orange track isn't in view (an arm over it?) - clear it and start again")
            return 1
        bar["top_mm"] = min(FIELD_Y[1], (bar["cy"] - rows.max()) * mm_px - TRACK_GAP_MM)
        json.dump(bar, open(anchor, "w"))
        print("the arena is locked - --recalibrate only if the camera moves")
    FIELD_Y = (0.0, bar["top_mm"])
    table_level = float(np.percentile(cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY), 75))
    # Searched near the arena only (the whole frame held it to 17 fps).
    sx0, sy0 = to_px(FIELD_X[0] - BAR_MM, FIELD_Y[1] + 60)
    sx1, sy1 = to_px(FIELD_X[1] + BAR_MM, -150)  # down past the ruler: his charger sits there
    sx0, sy0 = max(0, sx0), max(0, sy0)
    body_half = {"v": (BODY_HALF_DEFAULT, BODY_HALF_DEFAULT)}
    seen = {"mark": None}

    def find_body(img, nx, ny, sq):
        """A's whole body: the dark shape his pink square sits on (the ruler masked out)."""
        r = int(BODY_SEARCH_MM / mm_px)
        x0, y0 = max(0, nx - r), max(0, ny - r)
        dark = cv2.inRange(cv2.cvtColor(img[y0:ny + r, x0:nx + r], cv2.COLOR_BGR2GRAY), 0, int(BODY_DARK * table_level))
        bx, by = bar["x"] - x0, bar["y"] - y0
        dark[max(0, by - 4): max(0, by + bar["h"] + 4), max(0, bx): max(0, bx + bar["w"])] = 0
        window = np.zeros_like(dark)
        cv2.circle(window, (nx - x0, ny - y0), r, 255, -1)
        dark &= window  # only within his own size
        # the square itself, a little wider, joins the body it sits on (2026-11-23: the
        # big pink square cut the old small seed off from his body)
        cv2.rectangle(dark, (sq[0] - x0 - 6, sq[1] - y0 - 6), (sq[0] + sq[2] - x0 + 6, sq[1] + sq[3] - y0 + 6), 255, -1)
        dark = cv2.morphologyEx(dark, cv2.MORPH_CLOSE, np.ones((9, 9), np.uint8))
        _, lab, st, _ = cv2.connectedComponentsWithStats(dark)
        x, y, w, h = st[lab[ny - y0, nx - x0]][:4]
        half = (max(BODY_HALF[0], w * mm_px / 2), max(BODY_HALF[0], h * mm_px / 2))
        if max(half) > BODY_HALF[1]:
            body_half["v"] = (BODY_HALF_DEFAULT, BODY_HALF_DEFAULT)
            return to_mm(nx, ny)
        body_half["v"] = half
        return to_mm(x0 + x + w / 2, y0 + y + h / 2)

    def find_a(img, whole=True):
        """Brobot A: his pink square, then his body around it (whole=False: the
        square's own middle - steadier for driving)."""
        crop = img[sy0:sy1, sx0:sx1]
        small = cv2.resize(crop, (crop.shape[1] // 2, crop.shape[0] // 2))
        m = cv2.inRange(cv2.cvtColor(small, cv2.COLOR_BGR2HSV), np.array(NOTE_HSV[0]), np.array(NOTE_HSV[1]))
        m = cv2.morphologyEx(m, cv2.MORPH_OPEN, np.ones((3, 3), np.uint8))
        best = None
        for c in cv2.findContours(m, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)[0]:
            x, y, w, h = cv2.boundingRect(c)
            if NOTE_SIZE[0] <= 2 * max(w, h) / bar["w"] <= NOTE_SIZE[1] and 0.5 <= w / max(1, h) <= 2.0:
                if best is None or cv2.contourArea(c) > best[0]:
                    best = (cv2.contourArea(c), x, y, w, h)
        if best is None:
            return None
        _, x, y, w, h = best
        seen["mark"] = to_mm(sx0 + (x + w / 2) * 2, sy0 + (y + h / 2) * 2)
        if not whole:
            return seen["mark"]
        return find_body(img, int(sx0 + (x + w / 2) * 2), int(sy0 + (y + h / 2) * 2),
                         (sx0 + 2 * x, sy0 + 2 * y, 2 * w, 2 * h))

    hx0, hy0 = to_px(FIELD_X[0], FIELD_Y[1])
    hx1, hy1 = to_px(FIELD_X[1], (FIELD_Y[0] + FIELD_Y[1]) / 2)

    def find_paddle(img, body):
        """The operator's black deflector, in his half (a little past his end too)."""
        top = max(0, hy0 - int(0.15 * bar["w"]))
        crop = img[top:hy1, max(0, hx0):hx1]
        small = cv2.resize(crop, (max(1, crop.shape[1] // 2), max(1, crop.shape[0] // 2)))
        grey = cv2.cvtColor(small, cv2.COLOR_BGR2GRAY)
        note = cv2.inRange(cv2.cvtColor(small, cv2.COLOR_BGR2HSV), np.array(NOTE_HSV[0]), np.array(NOTE_HSV[1]))
        m = cv2.morphologyEx(cv2.inRange(grey, 0, int(PADDLE_DARK * np.percentile(grey, 75))), cv2.MORPH_OPEN, np.ones((5, 5), np.uint8))
        best = None
        for c in cv2.findContours(m, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)[0]:
            x, y, w, h = cv2.boundingRect(c)
            if not PADDLE_SIZE[0] <= 2 * max(w, h) / bar["w"] <= PADDLE_SIZE[1] or not 0.5 <= w / max(1, h) <= 2.0:
                continue
            if cv2.countNonZero(note[max(0, y - 6): y + h + 6, max(0, x - 6): x + w + 6]) > 20:
                continue  # it carries the pink square: that's Brobot A, not the deflector
            p = to_mm(max(0, hx0) + (x + w / 2) * 2, top + (y + h) * 2)  # its front edge
            if body and math.hypot(p[0] - body[0], p[1] - body[1]) < 80:
                continue  # that's Brobot A
            if best is None or cv2.contourArea(c) > best[0]:
                best = (cv2.contourArea(c), p)
        return best[1] if best else None

    def seen_now():
        got = None
        for _ in range(6):
            ok, img = cap.read()
            got = find_a(img) if ok else got
        return got

    window = writer = None
    w2, h2 = frame.shape[1] // 2, frame.shape[0] // 2
    if args.out:
        writer = cv2.VideoWriter(args.out, cv2.VideoWriter_fourcc(*"mp4v"), 30, (w2, h2))
    if os.environ.get("DISPLAY"):  # this env's OpenCV has no window: ffplay shows the frames
        window = subprocess.Popen(["ffplay", "-loglevel", "quiet", "-window_title", "GOPOD air hockey", "-f", "rawvideo",
                                   "-pixel_format", "bgr24", "-video_size", "%dx%d" % (w2, h2), "-i", "-"], stdin=subprocess.PIPE)
    def draw_field(img):
        """The field and both goal mouths with posts."""
        x0, y0 = to_px(FIELD_X[0], FIELD_Y[1])
        x1, y1 = to_px(FIELD_X[1], FIELD_Y[0])
        cv2.rectangle(img, (x0, y0), (x1, y1), (255, 255, 0), 3, cv2.LINE_AA)
        mouth = GOAL_SHARE * (FIELD_X[1] - FIELD_X[0]) / 2
        for gy, col in ((FIELD_Y[0], (0, 0, 255)), (FIELD_Y[1], (0, 200, 0))):
            p, q = to_px(-mouth, gy), to_px(mouth, gy)
            cv2.line(img, p, q, col, 8, cv2.LINE_AA)
            cv2.circle(img, p, 8, (255, 255, 255), -1)
            cv2.circle(img, q, 8, (255, 255, 255), -1)
        return x0, y0

    def draw_body(img, body):
        if body:
            hw, hh = body_half["v"]
            cv2.rectangle(img, to_px(body[0] - hw, body[1] + hh), to_px(body[0] + hw, body[1] - hh), (0, 150, 255), 3, cv2.LINE_AA)

    def show(img):
        """Out to the window (or the test video); False once the window is closed."""
        shown = cv2.resize(img, (w2, h2), interpolation=cv2.INTER_AREA)
        if writer:
            writer.write(shown)
        if window:
            try:
                window.stdin.write(shown.tobytes())
            except OSError:
                return False
        return True

    def frame_pos():
        """His position (his square's middle - steadiest for driving) from the newest
        frame, shown with the field, his body and the parking points."""
        cap.grab()
        ok, img = cap.read()
        if not ok:
            return None
        body = find_a(img)
        draw_field(img)
        draw_body(img, body)
        plan = getattr(a, "plan", [])
        for k, pt in enumerate(plan):
            cv2.circle(img, to_px(*pt), int(PATH_LANE / mm_px), (255, 0, 255), 2, cv2.LINE_AA)
            cv2.putText(img, str(k + 1), to_px(pt[0] - 6, pt[1] - 8), cv2.FONT_HERSHEY_SIMPLEX, 1.0, (255, 0, 255), 2, cv2.LINE_AA)
        show(img)
        return seen["mark"] if body else None

    # The Brobots: A parks (his charger's spot kept for the way home), B gets ready.
    a, b = Robot(args.player), Robot(referee_name)
    a.hold()
    b.hold()
    charger = seen_now()
    a.park(frame_pos)
    if args.part == 1:  # part 1 only: parked, then stop
        a.let_go()
        b.let_go()
        cap.release()
        if window:
            window.stdin.close()
            window.terminate()
        if writer:
            writer.release()
        return 0
    game = {"over": False}

    puck = Puck()
    score = {"a": 0, "room": 0}
    rounds = {"n": 1}

    def start_round():
        b.nod()
        b.say(ROUND_TEXT.format(n=rounds["n"]))
        time.sleep(1.5)
        puck.held = False

    def call_goal(scorer):
        b.nod()
        if max(score.values()) >= WIN_SCORE:
            # The game is won: B announces it, then A drives back to his charger.
            b.say(WIN_TEXT.format(w=args.player if score["a"] > score["room"] else "The room"))
            time.sleep(3)
            game["over"] = True
            return
        else:
            b.say(SCORE_TEXT.format(a=args.player, s=score["a"], r=score["room"]))
            time.sleep(2.5)
        rounds["n"] += 1
        start_round()

    referee = background(start_round)
    last_cheer = 0.0

    log = open(os.path.expanduser(RUN_LOG), "w")
    log.write("# t  a_x a_y  puck_x puck_y puck_vy  wheels_l wheels_r - field x %.0f..%.0f y %.0f..%.0f\n"
              % (FIELD_X[0], FIELD_X[1], FIELD_Y[0], FIELD_Y[1]))
    last_body = last_paddle = None
    last_t, stuck_for, n, jumps = time.time(), 0.0, 0, 0
    while True:
        ok, frame = cap.read()
        if not ok:
            break
        now = time.time()
        dt, last_t = min(0.1, now - last_t), now
        body = find_a(frame)
        if body and last_body and math.hypot(body[0] - last_body[0], body[1] - last_body[1]) > JUMP_MM:
            jumps += 1  # far from where he was: a mistake - unless he stays there (live: lost for good)
            if jumps < JUMP_FRAMES:
                body = None
            else:
                last_body, jumps = None, 0
        else:
            jumps = 0
        if body and last_body and math.hypot(body[0] - last_body[0], body[1] - last_body[1]) < STILL_MM:
            body = last_body  # the camera's own wobble: he's still
        elif body and last_body:
            body = (last_body[0] + BODY_SMOOTH * (body[0] - last_body[0]), last_body[1] + BODY_SMOOTH * (body[1] - last_body[1]))
        body_v = (0.0, 0.0)
        if body and last_body and dt > 0:
            body_v = ((body[0] - last_body[0]) / dt, (body[1] - last_body[1]) / dt)
            if math.hypot(*body_v) < BODY_V_MIN:
                body_v = (0.0, 0.0)
        last_body = body or last_body
        # A: guard while the puck comes at him; wait while it goes away.
        if not body or puck.held or puck.vy >= 0:
            a.stop(now)
        else:
            target = puck.x + puck.vx * (puck.y - body[1]) / -puck.vy
            while target < FIELD_X[0] or target > FIELD_X[1]:  # banked off the sides
                target = 2 * FIELD_X[0] - target if target < FIELD_X[0] else 2 * FIELD_X[1] - target
            room = (FIELD_X[1] - FIELD_X[0]) / 2 - BODY_HALF_DEFAULT
            a.guard(body, body_v, max(-room, min(room, target)), now)
        # B: cheer as the puck nears A's net.
        if not puck.held and puck.vy < 0 and puck.y < CHEER_NEAR_MM and now - last_cheer > CHEER_EVERY and not referee.is_alive():
            last_cheer = now
            background(b.cheer)
        paddle = find_paddle(frame, body)
        if paddle:
            # Held to the operator's half of the arena (operator: "use the arena border as a
            # clamp"), and smoothed - it jittered even standing still.
            paddle = (max(FIELD_X[0], min(FIELD_X[1], paddle[0])),
                      max((FIELD_Y[0] + FIELD_Y[1]) / 2, min(FIELD_Y[1], paddle[1])))
            if last_paddle and math.hypot(paddle[0] - last_paddle[0], paddle[1] - last_paddle[1]) < STILL_MM:
                paddle = last_paddle  # moved less than the camera's own wobble: still
            elif last_paddle:
                paddle = (last_paddle[0] + PADDLE_SMOOTH * (paddle[0] - last_paddle[0]),
                          last_paddle[1] + PADDLE_SMOOTH * (paddle[1] - last_paddle[1]))
        paddle_v = (0.0, 0.0)
        if paddle and last_paddle and dt > 0:
            paddle_v = ((paddle[0] - last_paddle[0]) / dt, (paddle[1] - last_paddle[1]) / dt)
        last_paddle = paddle
        movers = ([(body, body_v, body_half["v"])] if body else []) + ([(paddle, paddle_v, PADDLE_R)] if paddle else [])
        goal = puck.advance(FRAME_DT if dt < 3 * FRAME_DT else dt, movers)
        # The puck trapped against him: he backs away, it's nudged back into play.
        near = body and math.hypot(puck.x - body[0], puck.y - body[1]) < PUCK_R + max(body_half["v"]) + 4
        stuck_for = stuck_for + dt if near else 0.0
        if stuck_for > STUCK_SECONDS:
            stuck_for = 0.0
            puck.vy = abs(puck.vy) + PUCK_MIN_SPEED / 2
            puck.clamp_speed()
            background(lambda away: (a.wheels(away * BURST_SPEED), time.sleep(BACK_OFF_SECONDS), a.wheels(0)),
                       (-1 if puck.x > body[0] else 1) * a.facing)
        if goal:
            score[goal] += 1
            puck.serve()
            referee = background(call_goal, goal)
        log.write("%.2f  %s  %.0f %.0f %.0f  %d %d\n" % (now % 1000, "%.0f %.0f" % body if body else "- -",
                                                       puck.x, puck.y, puck.vy, a.sent[0], a.sent[1]))
        # Draw: the field, both goal mouths with posts, A's body, the paddle, the puck, the score.
        x0, y0 = draw_field(frame)
        draw_body(frame, body)
        if paddle:
            cv2.circle(frame, to_px(*paddle), int(PADDLE_R / mm_px), (255, 120, 0), -1, cv2.LINE_AA)
        cv2.circle(frame, to_px(puck.x, puck.y), int(PUCK_R / mm_px), (0, 255, 255), -1, cv2.LINE_AA)
        cv2.putText(frame, "%s %d - %d Room   round %d" % (args.player, score["a"], score["room"], rounds["n"]),
                    (x0, max(40, y0 - 15)), cv2.FONT_HERSHEY_SIMPLEX, 1.2, (255, 255, 255), 3, cv2.LINE_AA)
        if not show(frame):
            break  # the window was closed
        n += 1
        if game["over"] or (args.frames and n >= args.frames):
            break
    log.close()
    if game["over"] and charger:
        a.go_home(frame_pos, charger)
    a.let_go()
    b.let_go()
    cap.release()
    if window:
        try:
            window.stdin.close()
        except OSError:
            pass
        window.terminate()
    if writer:
        writer.release()
    print("frames %d - %s %d, Room %d" % (n, args.player, score["a"], score["room"]))
    return 0


if __name__ == "__main__":
    sys.exit(main())
