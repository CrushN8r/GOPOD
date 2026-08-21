#!/usr/bin/env python3
"""Watches live_chat_messages.json for bingo_end and fires ONE angry animation
on Brobot 2 when the game ends (deck exhausted or bingo called via the
backpack button) - Pip getting mad about not winning. Best-effort: any
failure logs and continues. Never touches Bingo (Brobot 1) or the chat file
itself."""

import argparse
import json
import signal
import time
from datetime import datetime, timezone
from pathlib import Path

import anki_vector

DEFAULT_BROBOT2_SERIAL = "0dd1d8bf"
DEFAULT_CHAT_FILE = (
    "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/gopod_layer/web_display/"
    "gopod_demo_8011/live_chat_messages.json"
)
DEFAULT_POLL_INTERVAL = 0.5
ANGRY_ANIMATION_NAME = "anim_rtpickup_loop_10"

_stop_requested = False
_stop_reason = None


def _handle_sigint(_signum, _frame):
    global _stop_requested, _stop_reason
    _stop_requested = True
    _stop_reason = "sigint"


def dispatch_angry_animation(serial: str, robot) -> None:
    """Fires the angry animation on an already-connected robot. Held-connection
    (stay-put) version - the caller connects once for the whole session instead
    of this function reconnecting per draw."""
    try:
        robot.anim.play_animation(ANGRY_ANIMATION_NAME)
        print(f"GOPOD_BINGO_REACTOR_ANIM_DONE serial={serial}")
    except Exception as exc:  # noqa: BLE001 - best-effort by design
        print(f"GOPOD_BINGO_REACTOR_ANIM_FAIL serial={serial} reason={exc}")


def read_messages(chat_file: Path):
    with chat_file.open("r") as f:
        data = json.load(f)
    return data.get("messages", [])


def _message_is_stale(message, started_at) -> bool:
    """True if this message's own timestamp predates the reactor's own start -
    i.e. it's leftover history from an earlier session still sitting in the
    chat file's capped last-40 window, not a real event from this run."""
    raw = message.get("timestamp")
    if not raw:
        return False  # no timestamp to judge by - don't silently drop it
    try:
        parsed = datetime.fromisoformat(raw.replace("Z", "+00:00"))
    except ValueError:
        return False
    return parsed <= started_at


def run(args, robot_factory=anki_vector.Robot) -> None:
    global _stop_requested, _stop_reason
    _stop_requested = False
    _stop_reason = None

    chat_file = Path(args.chat_file)
    signal.signal(signal.SIGINT, _handle_sigint)

    started_at = datetime.now(timezone.utc)
    print(f"GOPOD_BINGO_REACTOR_STARTED serial={args.brobot2_serial}")

    # Stay-put: connect once for the whole session instead of reconnecting per
    # draw. Per-draw reconnect was the actual bug behind slow/failed reactions
    # under real gameplay - a fresh connect also reloads the robot's whole
    # animation list from scratch every time, which is too slow to keep up
    # once draws come in close together.
    try:
        robot_ctx = robot_factory(serial=args.brobot2_serial)
        robot = robot_ctx.__enter__()
    except Exception as exc:  # noqa: BLE001 - best-effort by design
        print(f"GOPOD_BINGO_REACTOR_CONNECT_FAIL serial={args.brobot2_serial} reason={exc}")
        print("GOPOD_BINGO_REACTOR_STOPPED reason=connect_fail")
        return

    try:
        last_index = 0

        while not _stop_requested:
            if not chat_file.exists():
                print("GOPOD_BINGO_REACTOR_STOPPED reason=chat_file_missing")
                return

            try:
                messages = read_messages(chat_file)
            except Exception as exc:  # noqa: BLE001 - best-effort by design
                print(f"GOPOD_BINGO_REACTOR_PARSE_FAIL reason={exc}")
                time.sleep(args.poll_interval)
                continue

            new_messages = messages[last_index:]
            last_index = len(messages)

            for message in new_messages:
                if _message_is_stale(message, started_at):
                    continue
                producer = message.get("source_producer")
                if producer == "bingo_end":
                    # Fires once, whether the game ended by deck exhaustion or
                    # by "bingo" called via the backpack button - both paths
                    # write the same bingo_end envelope. Pip getting mad about
                    # not winning, once, at the actual end - not per draw.
                    print(f"GOPOD_BINGO_REACTOR_ANIM_DISPATCH serial={args.brobot2_serial} reason=bingo_end")
                    dispatch_angry_animation(args.brobot2_serial, robot)
                    print("GOPOD_BINGO_REACTOR_STOPPED reason=bingo_end")
                    return

            if _stop_requested:
                break

            time.sleep(args.poll_interval)

        print(f"GOPOD_BINGO_REACTOR_STOPPED reason={_stop_reason}")
    finally:
        try:
            robot_ctx.__exit__(None, None, None)
        except Exception:  # noqa: BLE001 - best-effort by design
            pass


def parse_args(argv=None):
    parser = argparse.ArgumentParser(
        description="Brobot 2 angry-animation reactor for Bingo draws"
    )
    parser.add_argument("--brobot2-serial", default=DEFAULT_BROBOT2_SERIAL)
    parser.add_argument("--chat-file", default=DEFAULT_CHAT_FILE)
    parser.add_argument(
        "--poll-interval", type=float, default=DEFAULT_POLL_INTERVAL
    )
    return parser.parse_args(argv)


def main():
    args = parse_args()
    run(args)


if __name__ == "__main__":
    main()
