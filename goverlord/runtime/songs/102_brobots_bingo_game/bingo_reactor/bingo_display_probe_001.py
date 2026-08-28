#!/usr/bin/env python3
"""Here's what I did to isolate the Error 915 crash, and every handshake I
tried that did NOT work — 7/7 crashed on real hardware, both robots. This
reproduces 915 cleanly; it does not solve it.

Standalone, gameplay-uninvolved Scenario 2 display probe: does a bingo
number render on Robot 2's OWN face, over its OWN Python anki_vector
connection - separate robot, separate process, separate SDK/language from
Robot 1's Go-side rattle/SayText path entirely, so the same-connection
Error 915 race documented from an earlier probe on Robot 1 does not apply
here by construction. That was the hypothesis, anyway - it turned out not
to matter (see below).

Revision history, honestly:
- First version of this probe (no explicit wire-pod release, DEFAULT
  priority) hit Error 915 twice in a row on real hardware, confirmed both
  times. That falsified "separation alone avoids 915."
- Outside research on vic-engine 914/915 (NO_ENGINE_PROCESS/NO_ENGINE_COMMS
  - a known DisplayFaceImageRGB stressor, especially with a second client
  contending for BehaviorControl) suggested two concrete gaps: no explicit
  release of wire-pod's own held session first, and requesting DEFAULT
  priority instead of HIGH/OVERRIDE. This revision adds both:
  1. Explicitly release wire-pod's OWN held REST session on this robot
     first (`/api-sdk/release_behavior_control`) - wire-pod.service stays
     connected to both robots even when nothing is using them; it never
     tears its own session down on its own.
  2. Wait a NAMED, TUNABLE drain longer than wire-pod's own internal
     release-polling window (~500ms, observed) before this probe's own
     connection ever requests control.
  3. Request HIGH priority (OVERRIDE_BEHAVIORS_PRIORITY) on THIS probe's
     own connection instead of DEFAULT_PRIORITY - matching the same
     high-priority handshake already tried on Robot 1's own connection.
- **This exact fixed handshake was then run live, twice, on this robot.
  Both runs completed a fully clean SDK-side handshake - release
  succeeded, drain completed, display RPC accepted - and the robot
  crashed with Error 915 anyway, both times.** The same fixed handshake
  was also tried on Robot 1 with no audio at all: 915 both times there
  too. Total score across every configuration tried on both robots:
  **7 attempts, 7 crashes.** Session contention, audio adjacency,
  BehaviorControl priority, and release/drain timing are all ruled out as
  the fix - none of them changed the outcome. The one constant across
  every attempt is the DisplayFaceImageRGB call itself.

So: run this and you WILL reproduce Error 915 cleanly and quickly, on a
clean, isolated connection, with no game logic and no guessing involved.
What you get is a solid repro, not a cure. If you find a handshake that
actually avoids it, that's real news - none of the combinations above did.

Sequence: release wire-pod's session (HTTP) -> drain -> connect
(anki_vector.Robot, OVERRIDE_BEHAVIORS_PRIORITY) -> render "L-NN" text to a
184x96 RGB565 image (PIL) -> push via
robot.screen.set_screen_with_image_data() (anki_vector/screen.py's own
ScreenComponent -> DisplayFaceImageRGB) -> hold for a named, tunable
duration -> done. NO audio, NO arm/lift/gesture RPC anywhere in this file,
NO Robot 1, not wired into any game loop - a standalone bench tool only.

Uses this project's own vendored anki_vector SDK checkout - run this with
that checkout's own python3/venv, not system python. Reads wire-pod's base
URL from GOPOD_WIREPOD_BASE_URL (must be exported in the calling shell) -
no network identifier hardcoded here. The robot serial is NOT hardcoded
either: it's read from an optional, untracked
`bingo_display_probe_001_private.json` sidecar (`{"serial": "..."}`) next
to this file if present, otherwise `--serial` is a required argument.
"""

import argparse
import json
import os
import time
import urllib.parse
import urllib.request

import anki_vector
from anki_vector.connection import ControlPriorityLevel
from PIL import Image, ImageDraw, ImageFont

PRIVATE_CONFIG_PATH = os.path.join(
    os.path.dirname(os.path.abspath(__file__)), "bingo_display_probe_001_private.json"
)

# The value under test's neighbor - how long the drawn number stays on
# screen once shown. Named, tunable, not a magic literal. Sane default: long
# enough to be visually confirmed by the operator, short enough not to hold
# control open needlessly.
DISPLAY_HOLD_SECONDS_DEFAULT = 4.0

# How long to wait after telling wire-pod to release its own BehaviorControl
# session before this probe requests its own - must clear wire-pod's own
# internal release-polling window (~500ms observed). 2.0s is a real margin
# above that, not the bare minimum - matches this project's own existing
# settle-margin convention (a real margin above the measured/documented
# floor, not the floor itself). Named, tunable via --release-drain, not a
# magic literal.
RELEASE_DRAIN_SECONDS_DEFAULT = 2.0

SCREEN_WIDTH, SCREEN_HEIGHT = anki_vector.screen.dimensions()


def load_private_serial():
    """Reads an optional, untracked sidecar JSON for a default robot serial
    - keeps any real hardware identifier out of this published file. Returns
    None if the sidecar is absent or unreadable; callers must then require
    --serial explicitly."""
    if not os.path.exists(PRIVATE_CONFIG_PATH):
        return None
    try:
        with open(PRIVATE_CONFIG_PATH, "r", encoding="utf-8") as f:
            data = json.load(f)
        serial = data.get("serial")
        return serial if serial else None
    except Exception:  # noqa: BLE001 - a bad/missing sidecar just means no default
        return None


def column_letter(number: int) -> str:
    """Standard 75-ball BINGO column ranges: B 1-15, I 16-30, N 31-45,
    G 46-60, O 61-75. Clamped to O above 75 (90-ball has no column-letter
    concept - not this probe's concern, a sample-text helper only)."""
    if number <= 15:
        return "B"
    if number <= 30:
        return "I"
    if number <= 45:
        return "N"
    if number <= 60:
        return "G"
    return "O"


def format_draw(number: int) -> str:
    return f"{column_letter(number)}-{number:02d}"


def clock_prefix(state):
    now = time.time()
    stamp = time.strftime("%H:%M:%S", time.localtime(now)) + f".{int(now % 1 * 1000):03d}"
    if state.get("last") is None:
        state["last"] = now
        return f"[{stamp} +0.000s] "
    delta = now - state["last"]
    state["last"] = now
    return f"[{stamp} +{delta:.3f}s] "


def render_face_image(text: str) -> Image.Image:
    """184x96 RGB image, black background, centered white text. No external
    font file dependency (avoids font-path fragility) - PIL's built-in
    default bitmap font is enough for a short "L-NN" label."""
    img = Image.new("RGB", (SCREEN_WIDTH, SCREEN_HEIGHT), color=(0, 0, 0))
    draw = ImageDraw.Draw(img)
    font = ImageFont.load_default()
    bbox = draw.textbbox((0, 0), text, font=font)
    text_w = bbox[2] - bbox[0]
    text_h = bbox[3] - bbox[1]
    x = (SCREEN_WIDTH - text_w) // 2 - bbox[0]
    y = (SCREEN_HEIGHT - text_h) // 2 - bbox[1]
    draw.text((x, y), text, font=font, fill=(255, 255, 255))
    return img


def release_wirepod_session(serial, state):
    """Explicitly releases wire-pod's OWN held REST BehaviorControl session
    on this robot before the probe's own connection ever requests control -
    the first of the two handshake steps the first probe run skipped (see
    this file's own header comment). Best-effort: prints the outcome either
    way, never raises - a failed release attempt is informative (wire-pod
    may not be running, or may not currently hold a session on this robot)
    but should not block the probe from still trying its own connection."""
    base_url = os.getenv("GOPOD_WIREPOD_BASE_URL")
    if not base_url:
        print(f"{clock_prefix(state)}PROBE_WIREPOD_RELEASE_SKIPPED reason=GOPOD_WIREPOD_BASE_URL not set")
        return
    url = base_url.rstrip("/") + "/api-sdk/release_behavior_control?" + urllib.parse.urlencode({"serial": serial})
    print(f"{clock_prefix(state)}PROBE_WIREPOD_RELEASE serial={serial}")
    try:
        with urllib.request.urlopen(url, timeout=10) as resp:
            body = resp.read().decode("utf-8", errors="replace")
        print(f"{clock_prefix(state)}PROBE_WIREPOD_RELEASE_DONE response={body!r}")
    except Exception as exc:  # noqa: BLE001 - informative, not fatal
        print(f"{clock_prefix(state)}PROBE_WIREPOD_RELEASE_FAILED error={exc!r}")


def main():
    private_serial = load_private_serial()
    parser = argparse.ArgumentParser(description="Robot 2 standalone face-display probe (Scenario 2) - Error 915 repro, not a fix")
    parser.add_argument(
        "--serial",
        default=private_serial,
        required=(private_serial is None),
        help="robot serial - required unless bingo_display_probe_001_private.json (untracked, next to this file) provides one",
    )
    parser.add_argument("--number", type=int, default=1, help="bingo number to render, formatted L-NN (default 1 -> B-01)")
    parser.add_argument("--hold", type=float, default=DISPLAY_HOLD_SECONDS_DEFAULT, help="DISPLAY_HOLD_SECONDS override - how long the number stays on screen")
    parser.add_argument("--release-drain", type=float, default=RELEASE_DRAIN_SECONDS_DEFAULT, help="RELEASE_DRAIN_SECONDS override - wait after releasing wire-pod's own session, before this probe requests control")
    args = parser.parse_args()

    text = format_draw(args.number)
    state = {}
    print(f"{clock_prefix(state)}PROBE_TARGET serial={args.serial} text={text!r} hold={args.hold:.3f}s release_drain={args.release_drain:.3f}s")

    print(f"{clock_prefix(state)}PROBE_RENDER text={text!r}")
    try:
        face_img = render_face_image(text)
        screen_data = anki_vector.screen.convert_image_to_screen_data(face_img)
    except Exception as exc:  # noqa: BLE001 - this probe's whole point is to surface the real error
        print(f"{clock_prefix(state)}PROBE_RENDER_FAILED error={exc!r}")
        print(f"{clock_prefix(state)}PROBE_RESULT status=FAIL hold={args.hold:.3f}s error={exc!r}")
        raise SystemExit(1)
    print(f"{clock_prefix(state)}PROBE_RENDER_DONE bytes={len(screen_data)}")

    release_wirepod_session(args.serial, state)
    print(f"{clock_prefix(state)}PROBE_RELEASE_DRAIN seconds={args.release_drain:.3f}")
    time.sleep(args.release_drain)
    print(f"{clock_prefix(state)}PROBE_RELEASE_DRAIN_DONE")

    print(f"{clock_prefix(state)}PROBE_CONNECT serial={args.serial} priority=OVERRIDE_BEHAVIORS")
    try:
        connect_start = time.time()
        with anki_vector.Robot(serial=args.serial, behavior_control_level=ControlPriorityLevel.OVERRIDE_BEHAVIORS_PRIORITY) as robot:
            connect_elapsed = time.time() - connect_start
            print(f"{clock_prefix(state)}PROBE_CONNECT_DONE elapsed={connect_elapsed:.3f}s")

            display_start = time.time()
            print(f"{clock_prefix(state)}PROBE_DISPLAY_START serial={args.serial} text={text!r}")
            robot.screen.set_screen_with_image_data(screen_data, args.hold)
            display_elapsed = time.time() - display_start
            print(f"{clock_prefix(state)}PROBE_DISPLAY_DONE elapsed={display_elapsed:.3f}s (RPC accepted - not proof the physical screen rendered; see this probe's own header comment - this is exactly where 915 has hit every time)")

            print(f"{clock_prefix(state)}PROBE_HOLD seconds={args.hold:.3f}")
            time.sleep(args.hold)
            print(f"{clock_prefix(state)}PROBE_HOLD_DONE")
    except Exception as exc:  # noqa: BLE001 - capture/print verbatim, this probe's whole point
        print(f"{clock_prefix(state)}PROBE_FAILED error={exc!r}")
        print(f"{clock_prefix(state)}PROBE_RESULT status=FAIL hold={args.hold:.3f}s error={exc!r}")
        raise SystemExit(1)

    print(f"{clock_prefix(state)}PROBE_RESULT status=SUCCESS text={text!r} hold={args.hold:.3f}s (SDK-side only - check the robot's actual face before trusting this)")


if __name__ == "__main__":
    main()
