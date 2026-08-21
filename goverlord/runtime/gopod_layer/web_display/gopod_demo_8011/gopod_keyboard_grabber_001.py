"""GOPOD keyboard grabber - list connected keyboards, grab one for exclusive
input, print keypresses, release cleanly on exit.

Mirrors gopod_ptt_chat_writer_013.py's own real, proven mechanism verbatim -
not a parallel invention:
- raw fd open (os.open, O_RDONLY|O_CLOEXEC) on a /dev/input/eventN node
- EVIOCGRAB (0x40044590) via fcntl.ioctl for the exclusive grab itself
- select.select + os.read + struct.unpack("llHHI", ...) for the read loop
- a try/finally that ALWAYS releases the grab (EVIOCGRAB, 0) and closes the
  fd, on any exit path - normal exit, an exception, Ctrl-C. This is the one
  property this tool must never violate: a keyboard left grabbed after this
  tool exits locks the operator out of normal typing.

Detection differs from the PTT writer on purpose. That tool targets one
specific keycode (KEY_KP1) because it's built for a numpad; this tool needs
to find "is this a real keyboard at all," a different question. Two sources
compared live on the operator's own machine before choosing:
- /proc/bus/input/devices' `Handlers=...kbd...` line: 3 false positives out
  of 9 devices on this machine alone (gpio-keys, a USB audio device with
  media keys, a webcam's own media-key interface all carry a kbd handler).
- /dev/input/by-id/*-event-kbd symlinks: udev's own purpose-built keyboard
  classification. On this machine, exactly one such symlink exists
  (usb-SEM_USB_Keyboard-event-kbd), pointing at the real keyboard's actual
  event node - cleanly excluding the false positives above, and also
  excluding the same keyboard's OWN secondary HID interfaces (its
  Consumer/System Control nodes, which appear as separate kbd-handled
  devices in /proc/bus/input/devices but are not the typing keyboard).
This is chosen as primary. /proc/bus/input/devices is kept as a documented
fallback only, for a machine with no udev by-id nodes (rare, but the
capability-first-then-name-fallback posture the PTT writer already
establishes is worth mirroring here too, not just for the grab mechanism).

Exit gesture: the numpad's own KP0-triple-tap-in-a-window doesn't map onto
a full keyboard - there's no dedicated free key playing KP0's role. ESC
triple-tapped within the same window plays the same role here: universally
present, never a normal typing key on its own, safe to reserve as the
deliberate "let me out" gesture. Same mechanic (N taps within a rolling
window), different key, stated plainly rather than silently substituted.
"""

from __future__ import annotations

import argparse
import fcntl
import os
import select
import struct
import sys
import time
from pathlib import Path

EVIOCGRAB = 0x40044590  # byte-identical to gopod_ptt_chat_writer_013.py's own constant

EV_KEY = 0x01
KEY_RELEASE = 0
KEY_PRESS = 1
KEY_REPEAT = 2

INPUT_EVENT_FORMAT = "llHHI"
INPUT_EVENT_SIZE = struct.calcsize(INPUT_EVENT_FORMAT)

KEY_ESC = 1

EXIT_TAPS = 3
EXIT_WINDOW_SECONDS = 2.0

BY_ID_DIR = Path("/dev/input/by-id")
PROC_INPUT_DEVICES = Path("/proc/bus/input/devices")

# A deliberately small, honest subset - common keys only, not a complete
# Linux keycode table. Anything not listed prints as a bare numeric code
# rather than a wrong or invented name.
KEY_NAMES = {
    1: "ESC", 2: "1", 3: "2", 4: "3", 5: "4", 6: "5", 7: "6", 8: "7", 9: "8",
    10: "9", 11: "0", 14: "BACKSPACE", 15: "TAB", 28: "ENTER", 29: "LCTRL",
    42: "LSHIFT", 54: "RSHIFT", 56: "LALT", 57: "SPACE", 58: "CAPSLOCK",
    16: "Q", 17: "W", 18: "E", 19: "R", 20: "T", 21: "Y", 22: "U", 23: "I",
    24: "O", 25: "P", 30: "A", 31: "S", 32: "D", 33: "F", 34: "G", 35: "H",
    36: "J", 37: "K", 38: "L", 44: "Z", 45: "X", 46: "C", 47: "V", 48: "B",
    49: "N", 50: "M",
}


def event_name(path: Path) -> str:
    """Same technique as gopod_ptt_chat_writer_013.py's own event_name()."""
    try:
        return (Path("/sys/class/input") / path.name / "device/name").read_text(encoding="utf-8").strip()
    except OSError:
        return ""


def list_keyboards_by_id() -> list[Path]:
    """Primary detection: udev's own *-event-kbd classification. Resolves
    each symlink to a real /dev/input/eventN path, deduped, sorted."""
    if not BY_ID_DIR.is_dir():
        return []
    found = []
    for entry in sorted(BY_ID_DIR.glob("*-event-kbd")):
        try:
            resolved = entry.resolve()
        except OSError:
            continue
        if resolved.exists() and resolved not in found:
            found.append(resolved)
    return found


def list_keyboards_by_proc_fallback() -> list[Path]:
    """Fallback only, for a machine with no udev by-id nodes. Parses
    /proc/bus/input/devices for a Handlers= line naming both "kbd" and an
    eventN node. Known weaker than by-id (confirmed 3 false positives on
    the operator's own machine - gpio-keys, a USB audio device's media
    keys, a webcam's media-key interface all carry a kbd handler) - used
    only when by-id genuinely has nothing."""
    if not PROC_INPUT_DEVICES.is_file():
        return []
    found = []
    for line in PROC_INPUT_DEVICES.read_text(encoding="utf-8", errors="replace").splitlines():
        if not line.startswith("H:"):
            continue
        tokens = line.split()
        if "kbd" not in tokens:
            continue
        for token in tokens:
            if token.startswith("event"):
                path = Path("/dev/input") / token
                if path.exists() and path not in found:
                    found.append(path)
                break
    return found


def discover_keyboards() -> tuple[list[Path], str]:
    by_id = list_keyboards_by_id()
    if by_id:
        return by_id, "by-id"
    return list_keyboards_by_proc_fallback(), "proc-fallback"


def prompt_choice(devices: list[Path]) -> Path | None:
    if len(devices) == 1:
        device = devices[0]
        name = event_name(device) or "UNKNOWN"
        answer = input(f"one keyboard found: {device} ({name}) - continue with this one? [default y]: ").strip()
        if answer == "" or answer.lower() == "y":
            return device
        print("GOPOD_KBGRAB_CANCELLED", flush=True)
        return None

    print("GOPOD_KBGRAB_MULTIPLE_FOUND", flush=True)
    for index, device in enumerate(devices, start=1):
        print(f"  {index}. {device} ({event_name(device) or 'UNKNOWN'})", flush=True)
    print("  0. exit", flush=True)
    choice = input("pick a keyboard [0 to exit]: ").strip()
    if not choice.isdigit() or not (1 <= int(choice) <= len(devices)):
        print("GOPOD_KBGRAB_CANCELLED", flush=True)
        return None
    return devices[int(choice) - 1]


def format_key(code: int) -> str:
    return KEY_NAMES.get(code, f"code={code}")


def grab_and_read(device: Path) -> int:
    try:
        fd = os.open(str(device), os.O_RDONLY | os.O_CLOEXEC)
    except PermissionError as exc:
        print(f"GOPOD_KBGRAB_PERMISSION_DENIED device={device} detail={exc}", flush=True)
        return 1
    except OSError as exc:
        print(f"GOPOD_KBGRAB_OPEN_FAILED device={device} detail={exc}", flush=True)
        return 1

    grabbed = False
    try:
        fcntl.ioctl(fd, EVIOCGRAB, 1)
        grabbed = True
        print(f"GOPOD_KBGRAB_EXCLUSIVE device={device} - grabbed, other apps/the OS won't see keys until exit", flush=True)
    except OSError as exc:
        print(f"GOPOD_KBGRAB_UNAVAILABLE {type(exc).__name__}: {exc}", flush=True)
        answer = input("exclusive grab unavailable - continue in non-exclusive read (keys also reach other apps)? [default n]: ").strip()
        if answer.lower() != "y":
            os.close(fd)
            print("GOPOD_KBGRAB_CANCELLED", flush=True)
            return 1
        print("GOPOD_KBGRAB_NONEXCLUSIVE - reading without exclusive grab", flush=True)

    esc_taps: list[float] = []
    down_codes: set[int] = set()
    try:
        print("GOPOD_KBGRAB_READY - press ESC 3x within 2s to exit cleanly", flush=True)
        while True:
            readable, _, _ = select.select([fd], [], [], 1.0)
            if esc_taps and (time.monotonic() - esc_taps[0]) > EXIT_WINDOW_SECONDS:
                esc_taps = []
            if not readable:
                continue
            data = os.read(fd, INPUT_EVENT_SIZE)
            if len(data) != INPUT_EVENT_SIZE:
                continue
            _sec, _usec, event_type, code, value = struct.unpack(INPUT_EVENT_FORMAT, data)
            if event_type != EV_KEY:
                continue

            if value == KEY_RELEASE:
                down_codes.discard(code)
                continue
            if value != KEY_PRESS:
                continue  # KEY_REPEAT ignored, same posture as the PTT writer
            if code in down_codes:
                continue
            down_codes.add(code)

            print(f"GOPOD_KBGRAB_KEY {format_key(code)}", flush=True)

            if code == KEY_ESC:
                now = time.monotonic()
                esc_taps = [t for t in esc_taps if now - t <= EXIT_WINDOW_SECONDS]
                esc_taps.append(now)
                print(f"GOPOD_KBGRAB_EXIT_TAP {len(esc_taps)}/{EXIT_TAPS}", flush=True)
                if len(esc_taps) >= EXIT_TAPS:
                    print("GOPOD_KBGRAB_EXIT", flush=True)
                    return 0
    finally:
        # MANDATORY: this must run on every exit path - normal return,
        # KeyboardInterrupt, or any unhandled exception. A device left
        # grabbed here is a real, immediate problem for the operator.
        if grabbed:
            try:
                fcntl.ioctl(fd, EVIOCGRAB, 0)
                print(f"GOPOD_KBGRAB_RELEASED device={device}", flush=True)
            except OSError as exc:
                print(f"GOPOD_KBGRAB_RELEASE_WARN {type(exc).__name__}: {exc}", flush=True)
        os.close(fd)


def main() -> int:
    parser = argparse.ArgumentParser(description="GOPOD keyboard grabber")
    parser.add_argument("--device", help="skip discovery, grab this /dev/input/eventN directly")
    parser.add_argument("--list-only", action="store_true", help="list detected keyboards and exit, no grab")
    args = parser.parse_args()

    if args.device:
        device = Path(args.device)
        if not device.exists():
            print(f"GOPOD_KBGRAB_DEVICE_NOT_FOUND device={device}", flush=True)
            return 1
        if args.list_only:
            print(f"{device} ({event_name(device) or 'UNKNOWN'})", flush=True)
            return 0
        return grab_and_read(device)

    devices, source = discover_keyboards()
    print(f"GOPOD_KBGRAB_DETECT_SOURCE {source}", flush=True)
    if not devices:
        print("GOPOD_KBGRAB_NONE_FOUND - rerun with --device /dev/input/eventX", flush=True)
        return 1

    if args.list_only:
        for device in devices:
            print(f"{device} ({event_name(device) or 'UNKNOWN'})", flush=True)
        return 0

    chosen = prompt_choice(devices)
    if chosen is None:
        return 0
    return grab_and_read(chosen)


if __name__ == "__main__":
    sys.exit(main())
