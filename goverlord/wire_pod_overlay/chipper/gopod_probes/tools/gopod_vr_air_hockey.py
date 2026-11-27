#!/usr/bin/env python3
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
FIELD_X = (-BAR_MM / 2, BAR_MM / 2)
FIELD_Y = (0.0, 1.3 * BAR_MM)
GOAL_SHARE = 0.5          # each end's goal mouth, this share of the field's width
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
PADDLE_R = 25.0
PADDLE_DARK = 0.6
PADDLE_SIZE = (0.04, 0.3)  # its width, in bar-lengths
NOTE_HSV = ((160, 85, 170), (179, 175, 255))  # 2026-11-23 brighter room: hue 177 (176 lost him all game)
NOTE_SIZE = (0.08, 0.25)  # its width, in bar-lengths
VECTOR_IN_SQUARES = (1.4, 2.3)  # his 60 x 100 mm (width, length), in pink-square sides (measured 2026-11-23)
SQUARE_MM = 43.0          # the square's side as the camera sees it, until it's seen
BODY_DARK = 0.7           # darker than this share of the arena's own light (each frame): him (0.6 fit him only ~0.35)
BODY_AREA = (1.5, 8.0)    # his dark shape's area, in pink squares (measured ~2.4: the flag covers part)
BODY_HALF_DEFAULT = 33.0
FIT_TURN_DEG = (-24, -16, -8, 0, 8, 16, 24)  # facings tried around his last one
FIT_MARGIN_MM = 30.0      # how far round his expected spot the box is tried (finding him: twice)
FIT_MIN = 0.3             # the box at best this much more dark than light, or it isn't him
TRACK_GAIN = (0.6, 0.15)  # how fast his position / his speed follow what's seen
HEADING_GAIN = 0.4        # how fast his facing follows
TRACK_LOST_S = 0.5        # unseen this long: found afresh
SPEED_MAX = 400.0         # mm/s - faster is a mis-sighting, not him
ON_CHARGER_MM = 60.0      # his flag this near his charger's middle: he's on it
STRIPE_MM = (85.0, 25.0)  # its length, width as seen (measured 2026-11-23)
STRIPE_LOOK_MM = 80.0     # looked for this far round its last spot
PADDLE_SMOOTH = 0.4
STILL_MM = 4.0            # a move smaller than this is the camera's own wobble - it stays put
BODY_V_MIN = 40.0         # mm/s - slower than this he's standing still
JUMP_MM = 80.0            # seen this far from where he just was: not him...
JUMP_FRAMES = 5           # ...unless he's seen there this many frames running
BURST_SPEED = 80.0
TURN_SPEED = 100.0
TURN_RATE = math.radians(50.0)  # Brobot 2, measured live 2026-11-23: 52 deg in 1 s (earlier 34 - the legs correct either way)
SETTLE_AFTER_ASSUME = 1.5  # a fresh hold needs a moment before the first move
WAKE_RELEASE_SETTLE = 2.5  # the golden wake's settle (BROBOTS_WAKE_RELEASE_SETTLE_SECONDS)
STOP_AGAIN = 0.5          # standing: "stop" is said again this often (a lost one let him roll on)
PATH_SPEED = 70.0         # mm/s
PATH_REAL_MM_S = 50.0     # how far he really gets per second at PATH_SPEED, starting from still
PATH_LEG_MIN, PATH_LEG_MAX = 0.6, 1.5  # seconds of each straight leg
PATH_TURN_DEG = 12.0      # off by more than this: turn toward the point before the leg
PATH_LANE = 45.0          # mm - this close to a point: on to the next
PATH_TIMEOUT = 45.0       # s, the whole drive at most
PARK_OUT = (322.0, 82.0)   # point 1, mm (x mirrored to his home side) - where the flag sat
PARK_BACK_IN_X = 100.0     # mm from the middle: where he backs in to the ruler
TURN_STEP_MAX = math.radians(50.0)  # one turn at most ~1 s; more is done over legs
DOCK_NEAR = 130.0         # this close to his charger: his own dock move takes over
NET_Y = 55.0              # his spot, mm up from the bar (70 parked him a little far out, live)
CREASE = (100.0, 100.0)   # half-width, depth (mm)
CREASE_ROOM = CREASE[0] - 50.0  # his middle stays this near the middle
GOALIE_GAIN = 1.5
GOALIE_MAX = 100.0
GOALIE_DEADBAND = 12.0
DRIVE_EVERY = 0.12
LINE_GAIN = 0.6           # steer back onto his line
LINE_DAMP = 0.3
LINE_MAX = 35.0
STUCK_SECONDS = 0.5       # the puck pressed against him this long: trapped
BACK_OFF_SECONDS = 0.6
ARM_CUE = {"Brobot 1": (2.0, 1.2), "Brobot 2": (3.0, 0.5)}  # speed, hold seconds
NOD = (4.0, 0.18)
ARM_NEAR_MM = 110.0       # the puck this near A's net, coming at it: B's arm goes up; going away: down
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
        self.arm_up = False

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
        self.send("/api-sdk/" + endpoint, {"speed": int(round(speed))})
        time.sleep(hold)
        self.send("/api-sdk/" + endpoint, {"speed": 0})

    def nod(self):
        self.axis("move_head", -NOD[0], NOD[1])
        self.axis("move_head", NOD[0], NOD[1])

    def arm(self, up):
        """Arm up (the puck nears A's net) or down to rest (it goes away, or a goal)."""
        speed, hold = ARM_CUE[self.name]
        self.axis("move_lift", speed if up else -speed, hold)
        self.arm_up = up

    def say(self, text):
        self.send("/api-sdk/say_text", {"text": text})


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

    def drive_path(self, frame_pos, points, lane=PATH_LANE):
        """Through the points in short straight legs, like steering a car in small
        corrections: a small turn toward the point, a leg, look again. His facing is
        the camera's (his flag on his back). A point (x, y, "back") is backed into, as
        in parallel parking. Within lane of a point: on to the next. (Steering while
        rolling made him spin in place, live.)"""
        self.plan = [p[:2] for p in points]
        log = open(os.path.expanduser(DRIVE_LOG), "a")
        log.write("# drive to %s\n" % ", ".join("%.0f,%.0f%s" % (p[0], p[1], " back" if len(p) > 2 else "") for p in points))
        k, still, t0 = 0, 0, time.time()
        pos = self.watch(frame_pos, 0.2)
        while k < len(points) and time.time() - t0 < PATH_TIMEOUT and pos is not None:
            target, back = points[k], len(points[k]) > 2
            dist = math.hypot(target[0] - pos[0], target[1] - pos[1])
            if dist < lane:
                print("drive: point %d of %d reached at %.0f, %.0f mm" % (k + 1, len(points), pos[0], pos[1]))
                k += 1
                continue
            err, heading = 0.0, frame_pos.seen["heading"]
            if heading is not None:
                going = heading + (math.pi if back else 0.0)  # backing: his back leads
                err = (math.atan2(target[1] - pos[1], target[0] - pos[0]) - going + math.pi) % (2 * math.pi) - math.pi
                if abs(err) > math.radians(PATH_TURN_DEG):
                    self.turn(max(-TURN_STEP_MAX, min(TURN_STEP_MAX, err)))
                    pos = self.watch(frame_pos, 0.2) or pos
            leg = min(PATH_LEG_MAX, max(PATH_LEG_MIN, dist / PATH_REAL_MM_S))
            self.wheels(-PATH_SPEED if back else PATH_SPEED)
            self.watch(frame_pos, leg)
            self.wheels(0)
            after = self.watch(frame_pos, 0.3)
            if after is None:
                break
            moved = math.hypot(after[0] - pos[0], after[1] - pos[1])
            log.write("%.1f  at %.0f,%.0f  point %d%s  turned %.0f deg  leg %.1f s  moved %.0f mm\n"
                      % (time.time() - t0, after[0], after[1], k + 1, " back" if back else "", math.degrees(err), leg, moved))
            still = 0 if moved > 8 else still + 1
            if still >= 3:  # told to drive, he hasn't moved: the wrong Brobot, or stuck
                print("drive: he isn't moving - is the pink square on %s?" % self.name)
                break
            pos = after
        self.wheels(0)
        log.close()
        return k >= len(points)

    def charger(self, move):
        """His own charger move ("off" / "on") over GOPOD's direct robot connection
        (direct_sdk_charger_001 - gopoddirect, as brobots ready uses). The web hold
        is let go for it and taken again after "off"."""
        self.send("/api-sdk/release_behavior_control", {})
        tool = os.path.join(os.path.dirname(os.path.abspath(__file__)), "direct_sdk_charger_001")
        env = dict(os.environ, WIREPOD_HOME=os.path.expanduser("~/wire-pod"))
        out = subprocess.run([tool, move, "ESN01" if self.name == "Brobot 1" else "ESN02"],
                             capture_output=True, text=True, env=env, timeout=70)
        print((out.stdout or out.stderr).strip())
        if move == "off":
            self.send("/api-sdk/assume_behavior_control", {"priority": "high"})
            time.sleep(SETTLE_AFTER_ASSUME)

    def park(self, frame_pos, home):
        """Part 1 (operator's plan): off his charger by himself, out up the table,
        back in toward the ruler's bar on his home side, then back along it to the
        middle of his goal line - facing along it."""
        self.charger("off")
        side = 1 if home[0] >= 0 else -1
        ok = self.drive_path(frame_pos, [(side * PARK_OUT[0], PARK_OUT[1]), (side * PARK_BACK_IN_X, NET_Y, "back"), (0.0, NET_Y, "back")])
        heading = frame_pos.seen["heading"]
        if heading is not None:
            self.facing = 1 if math.cos(heading) > 0 else -1
        print("park: %s, facing %s" % ("in front of his net" if ok else "near his net", "right" if self.facing > 0 else "left"))
        return ok

    def go_home(self, frame_pos, charger):
        """Part 3: drive near his charger (legs, small turns - he ends facing it, so
        he sees it), then his own dock move."""
        self.drive_path(frame_pos, [charger[:2]], lane=DOCK_NEAR)
        self.charger("on")

    def guard(self, body, body_v, target_x, now):
        """Slide along his goal line to target_x, steering back onto the line;
        never outward past the field's sides."""
        if now - self.last < DRIVE_EVERY:
            return
        self.last = now
        err = target_x - body[0]
        speed = 0.0 if abs(err) < GOALIE_DEADBAND else max(-GOALIE_MAX, min(GOALIE_MAX, GOALIE_GAIN * err))
        if (body[0] >= CREASE_ROOM and err > 0) or (body[0] <= -CREASE_ROOM and err < 0):
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
    ap.add_argument("--lock-only", action="store_true", help="lock the arena, then stop before any robot moves")
    args = ap.parse_args()
    referee_name = "Brobot 2" if args.player == "Brobot 1" else "Brobot 1"
    os.dup2(os.open(os.devnull, os.O_WRONLY), 2)  # the camera's "Corrupt JPEG data" notes are harmless
    cap = open_camera()

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
    if args.lock_only:
        print("the arena is locked: ruler at %d,%d, %d px wide" % (bar["cx"], bar["cy"], bar["w"]))
        cap.release()
        return 0
    sx0, sy0 = to_px(FIELD_X[0] - BAR_MM, FIELD_Y[1] + 60)
    sx1, sy1 = to_px(FIELD_X[1] + BAR_MM, -150)  # down past the ruler: his charger sits there
    sx0, sy0 = max(0, sx0), max(0, sy0)
    body_half = {"v": (BODY_HALF_DEFAULT, BODY_HALF_DEFAULT)}
    seen = {"pos": None, "v": (0.0, 0.0), "heading": math.pi / 2, "square": SQUARE_MM}  # on his charger he faces up the table
    ax0, ay0 = to_px(FIELD_X[0], FIELD_Y[1])  # the arena, down to the ruler bar
    ax1, ay1 = to_px(FIELD_X[1], 0)
    ax0, ay0, ay1 = max(0, ax0), max(0, ay0), bar["y"] + bar["h"]
    bar_row = max(0, (bar["y"] - ay0) // 2 - 2)

    def body_size():
        """His box (width, length) in mm: his real 60 x 100, scaled by his pink square."""
        return VECTOR_IN_SQUARES[0] * seen["square"], VECTOR_IN_SQUARES[1] * seen["square"]

    def find_pink(img):
        """His pink square - the flag on his back - near the arena or at home: its
        middle (mm), and its side is kept for his box."""
        crop = img[sy0:sy1, sx0:sx1]
        small = cv2.resize(crop, (crop.shape[1] // 2, crop.shape[0] // 2))
        m = cv2.inRange(cv2.cvtColor(small, cv2.COLOR_BGR2HSV), np.array(NOTE_HSV[0]), np.array(NOTE_HSV[1]))
        m = cv2.morphologyEx(m, cv2.MORPH_OPEN, np.ones((3, 3), np.uint8))
        best = None
        for c in cv2.findContours(m, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)[0]:
            x, y, w, h = cv2.boundingRect(c)
            if NOTE_SIZE[0] <= 2 * max(w, h) / bar["w"] <= NOTE_SIZE[1] and 0.5 <= w / max(1, h) <= 2.0:
                if best is None or cv2.contourArea(c) > best[0]:
                    best = (cv2.contourArea(c), x + w / 2, y + h / 2, max(cv2.minAreaRect(c)[1]))  # its own sides: turned, its outline reads bigger
        if best is None:
            return None
        seen["square"] += 0.1 * (best[3] * 2 * mm_px - seen["square"])  # steady, not jumpy
        return to_mm(sx0 + 2 * best[1], sy0 + 2 * best[2])

    track = {"t": None, "seen": 0.0, "jumps": 0}
    dock = bar.get("charger")  # his charger, from its white stripe

    def fit_box(img, guess, headings, margin, dark_below):
        """His real-size box slid and turned over the dark pixels near guess, at full
        camera detail: (x, y, facing, fit). The facing is along him, either end."""
        w, l = body_size()
        r = int((l / 2 + margin) / mm_px)
        gx, gy = to_px(*guess)
        x0, y0 = max(0, gx - r), max(0, gy - r)
        patch = img[y0:gy + r, x0:gx + r]
        if patch.shape[0] < 10 or patch.shape[1] < 10:
            return None
        score = np.where(cv2.cvtColor(patch, cv2.COLOR_BGR2GRAY) < dark_below, 1.0, -1.0).astype(np.float32)
        score[cv2.inRange(cv2.cvtColor(patch, cv2.COLOR_BGR2HSV), np.array(NOTE_HSV[0]), np.array(NOTE_HSV[1])) > 0] = 0.0  # his flag: neither
        b0, b1 = bar["y"] - y0 - 4, bar["y"] + bar["h"] - y0 + 4
        score[max(0, b0):max(0, b1), :] = -1.0  # the ruler bar is dark, not him
        ph, pw = score.shape
        size = (max(3, int(l / mm_px)), max(3, int(w / mm_px)))
        best = None
        for h in headings:
            m = cv2.getRotationMatrix2D((pw / 2, ph / 2), -math.degrees(h), 1.0)  # him lengthwise along x
            fit = cv2.boxFilter(cv2.warpAffine(score, m, (pw, ph), borderValue=-1.0), -1, size)
            _, val, _, loc = cv2.minMaxLoc(fit)
            if best is None or val > best[0]:
                best = (val, loc, m, h)
        val, (u, v), m, h = best
        mi = cv2.invertAffineTransform(m)
        x, y = to_mm(x0 + mi[0, 0] * u + mi[0, 1] * v + mi[0, 2], y0 + mi[1, 0] * u + mi[1, 1] * v + mi[1, 2])
        return x, y, h, val

    def find_stripe(img):
        """His charger's white stripe near its last spot: its middle and the way he'd
        face on it (along it, up the table, nearest its last facing) - kept in dock."""
        r = int((STRIPE_LOOK_MM + STRIPE_MM[0] / 2) / mm_px)
        cx, cy = to_px(dock["x"], dock["y"])
        x0, y0 = max(0, cx - r), max(0, cy - r)
        patch = img[y0:cy + r, x0:cx + r]
        if patch.size == 0:
            return
        white = ((cv2.cvtColor(patch, cv2.COLOR_BGR2GRAY) > 170) & (cv2.cvtColor(patch, cv2.COLOR_BGR2HSV)[..., 1] < 40)).astype(np.uint8)
        white = cv2.morphologyEx(white, cv2.MORPH_OPEN, np.ones((5, 5), np.uint8))
        for c in cv2.findContours(white, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)[0]:
            (sx, sy), (w, h), _ = cv2.minAreaRect(c)
            length, width = max(w, h) * mm_px, min(w, h) * mm_px
            if not (0.7 * STRIPE_MM[0] <= length <= 1.3 * STRIPE_MM[0] and 0.5 * STRIPE_MM[1] <= width <= 1.6 * STRIPE_MM[1]):
                continue
            pts = c.reshape(-1, 2).astype(np.float64)
            pts[:, 1] *= -1  # y up, as mm
            vals, vecs = np.linalg.eigh(np.cov(pts.T))
            ang = math.atan2(vecs[1, 1], vecs[0, 1])
            was = math.radians(dock["facing_deg"])
            if math.cos(ang - was) < 0:
                ang += math.pi
            x, y = to_mm(x0 + sx, y0 + sy)
            dock.update(x=x, y=y, facing_deg=math.degrees(ang))
            return

    def find_a(img):
        """Brobot A: his pose (x, y; facing in seen) from the newest frame, steadied.
        On his charger: the charger's own spot and facing. Else his box fitted near
        where he should be - or, finding him afresh, near his flag, or on the dark
        shape in the light arena nearest his net. His flag on his back says which end
        is his front."""
        now = time.time()
        pink = find_pink(img)
        crop = img[ay0:ay1, ax0:ax1]
        small = cv2.resize(crop, (crop.shape[1] // 2, crop.shape[0] // 2))
        grey = cv2.cvtColor(small, cv2.COLOR_BGR2GRAY)
        dark_below = int(BODY_DARK * np.median(grey))  # the arena's own light, each frame
        if track["t"] is not None and now - track["seen"] > TRACK_LOST_S:
            track["t"] = None
        meas = None
        if dock and not (pink and math.hypot(pink[0] - dock["x"], pink[1] - dock["y"]) < ON_CHARGER_MM):
            find_stripe(img)  # he's off it: where it is now
        if pink and dock and math.hypot(pink[0] - dock["x"], pink[1] - dock["y"]) < ON_CHARGER_MM:
            meas = (dock["x"], dock["y"], math.radians(dock["facing_deg"]))
        else:
            if track["t"] is not None:  # where he should be by now
                dt = now - track["t"]
                guess = (seen["pos"][0] + seen["v"][0] * dt, seen["pos"][1] + seen["v"][1] * dt)
                headings, margin = [seen["heading"] + math.radians(d) for d in FIT_TURN_DEG], FIT_MARGIN_MM
            else:  # found afresh: every facing, a wider look
                guess = None
                if pink:
                    guess = pink
                else:
                    dark = cv2.inRange(grey, 0, dark_below)
                    dark[bar_row:, :] = 0
                    best = None
                    for c in cv2.findContours(dark, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)[0]:
                        squares = cv2.contourArea(c) * (2 * mm_px) ** 2 / seen["square"] ** 2
                        if BODY_AREA[0] <= squares <= BODY_AREA[1]:
                            mo = cv2.moments(c)
                            p = to_mm(ax0 + 2 * mo["m10"] / mo["m00"], ay0 + 2 * mo["m01"] / mo["m00"])
                            if best is None or p[1] < best[1]:
                                best = p
                    guess = best
                headings, margin = [math.radians(d) for d in range(0, 180, 15)], 2 * FIT_MARGIN_MM
            fit = guess and fit_box(img, guess, headings, margin, dark_below)
            if fit and fit[3] >= FIT_MIN:
                x, y, h, _ = fit
                if pink:  # his flag is on his back
                    flip = math.cos(h - math.atan2(y - pink[1], x - pink[0])) < 0
                else:
                    flip = math.cos(h - seen["heading"]) < 0
                meas = (x, y, h + math.pi if flip else h)
        if meas is None:
            return None
        if track["t"] is None:
            seen["pos"], seen["v"], seen["heading"], track["jumps"] = meas[:2], (0.0, 0.0), meas[2], 0
        else:
            dt = max(1e-3, now - track["t"])
            pred = (seen["pos"][0] + seen["v"][0] * dt, seen["pos"][1] + seen["v"][1] * dt)
            rx, ry = meas[0] - pred[0], meas[1] - pred[1]
            if math.hypot(rx, ry) > JUMP_MM and track["jumps"] + 1 < JUMP_FRAMES:
                track["jumps"] += 1  # far from where he should be: a mis-sighting, unless it lasts
                return seen["pos"]
            if math.hypot(rx, ry) > JUMP_MM:
                seen["pos"], seen["v"] = meas[:2], (0.0, 0.0)
            else:
                seen["pos"] = (pred[0] + TRACK_GAIN[0] * rx, pred[1] + TRACK_GAIN[0] * ry)
                vx, vy = seen["v"][0] + TRACK_GAIN[1] * rx / dt, seen["v"][1] + TRACK_GAIN[1] * ry / dt
                k = min(1.0, SPEED_MAX / max(1e-6, math.hypot(vx, vy)))
                seen["v"] = (vx * k, vy * k)
            track["jumps"] = 0
            seen["heading"] += HEADING_GAIN * ((meas[2] - seen["heading"] + math.pi) % (2 * math.pi) - math.pi)
        track["t"] = track["seen"] = now
        (w, l), h = body_size(), seen["heading"]
        body_half["v"] = ((abs(l * math.cos(h)) + abs(w * math.sin(h))) / 2, (abs(l * math.sin(h)) + abs(w * math.cos(h))) / 2)
        return seen["pos"]

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
        for gy, d, col in ((FIELD_Y[0], CREASE[1], (0, 0, 255)), (FIELD_Y[1], -CREASE[1], (0, 200, 0))):
            cv2.rectangle(img, to_px(-CREASE[0], gy + d), to_px(CREASE[0], gy), col, 2, cv2.LINE_AA)  # the creases
        return x0, y0

    def draw_body(img, body):
        if body:  # his real size, scaled by his pink square, turned his way; a dot on his front
            (w, l), h = body_size(), seen["heading"] or 0.0
            w, l, c, s_ = w / 2, l / 2, math.cos(h), math.sin(h)
            corners = [to_px(body[0] + u * l * c - v * w * s_, body[1] + u * l * s_ + v * w * c) for u, v in ((1, 1), (1, -1), (-1, -1), (-1, 1))]
            cv2.polylines(img, [np.array(corners, np.int32)], True, (0, 150, 255), 3, cv2.LINE_AA)
            if seen["heading"] is not None:
                cv2.circle(img, to_px(body[0] + l * c, body[1] + l * s_), 8, (0, 150, 255), -1, cv2.LINE_AA)

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
        """His position (his body's middle) from the newest frame, shown with the
        field, his body and the parking points."""
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
        return body
    frame_pos.seen = seen  # his facing, for driving

    a, b = Robot(args.player), Robot(referee_name)
    a.hold()
    b.hold()
    charger = a.watch(frame_pos, 0.5)  # where his charger is: where he starts
    if not charger:
        print("Brobot A isn't in view")
        return 1
    a.park(frame_pos, charger)
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
        arm["t"].join()
        if b.arm_up:  # a goal: his arm back to rest, then the call
            b.arm(False)
        b.nod()
        if max(score.values()) >= WIN_SCORE:
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
    arm = {"t": background(lambda: None)}

    log = open(os.path.expanduser(RUN_LOG), "w")
    log.write("# t  a_x a_y  puck_x puck_y puck_vy  wheels_l wheels_r - field x %.0f..%.0f y %.0f..%.0f\n"
              % (FIELD_X[0], FIELD_X[1], FIELD_Y[0], FIELD_Y[1]))
    last_paddle = None
    last_t, stuck_for, n = time.time(), 0.0, 0
    while True:
        ok, frame = cap.read()
        if not ok:
            break
        now = time.time()
        dt, last_t = min(0.1, now - last_t), now
        body = find_a(frame)  # steadied inside, with his speed
        body_v = seen["v"] if body and math.hypot(*seen["v"]) >= BODY_V_MIN else (0.0, 0.0)
        if not body:
            a.stop(now)
        elif puck.held or puck.vy >= 0:
            a.guard(body, body_v, 0.0, now)  # the puck going away: back to the middle (operator: "that's what I do")
        else:
            target = puck.x + puck.vx * (puck.y - body[1]) / -puck.vy
            while target < FIELD_X[0] or target > FIELD_X[1]:  # banked off the sides
                target = 2 * FIELD_X[0] - target if target < FIELD_X[0] else 2 * FIELD_X[1] - target
            a.guard(body, body_v, max(-CREASE_ROOM, min(CREASE_ROOM, target)), now)
        want = not puck.held and puck.vy < 0 and puck.y < ARM_NEAR_MM
        if want != b.arm_up and not arm["t"].is_alive() and not referee.is_alive():
            arm["t"] = background(b.arm, want)
        paddle = find_paddle(frame, body)
        if paddle:
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
        near = body and math.hypot(puck.x - body[0], puck.y - body[1]) < PUCK_R + max(body_half["v"]) + 4
        stuck_for = stuck_for + dt if near else 0.0
        if stuck_for > STUCK_SECONDS:
            stuck_for = 0.0
            side = -1 if body[0] > 0 else 1
            room = FIELD_X[1] - PUCK_R
            puck.x = max(-room, min(room, body[0] + side * (body_half["v"][0] + PUCK_R + 5)))
            puck.y = max(puck.y, body[1])
            puck.vx, puck.vy = side * PUCK_MIN_SPEED / 2, PUCK_MIN_SPEED
            puck.clamp_speed()
            background(lambda away: (a.wheels(away * BURST_SPEED), time.sleep(BACK_OFF_SECONDS), a.wheels(0)),
                       (-1 if puck.x > body[0] else 1) * a.facing)
        if goal:
            score[goal] += 1
            puck.serve()
            referee = background(call_goal, goal)
        log.write("%.2f  %s  %.0f %.0f %.0f  %d %d\n" % (now % 1000, "%.0f %.0f" % body if body else "- -",
                                                       puck.x, puck.y, puck.vy, a.sent[0], a.sent[1]))
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
        a.go_home(frame_pos, (dock["x"], dock["y"]) if dock else charger)  # where his charger is now
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
