#!/usr/bin/env python3
import json
import os
import subprocess
import sys
from datetime import datetime, timezone

LOG_PATH = os.path.expanduser("~/gopod_chat_jsons.txt")

def now():
    return datetime.now(timezone.utc).isoformat()

def write_event(event):
    with open(LOG_PATH, "a", encoding="utf-8") as f:
        f.write(json.dumps(event, ensure_ascii=False) + "\n")

def main():
    if len(sys.argv) < 3:
        print("Usage: gopod_json_capture.py <source> <command...>", file=sys.stderr)
        return 2

    source = sys.argv[1]
    cmd = sys.argv[2:]

    write_event({
        "timestamp": now(),
        "event": "capture_start",
        "source": source,
        "command": cmd,
        "proof_status": "STARTED",
    })

    proc = subprocess.Popen(
        cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        bufsize=1,
    )

    assert proc.stdout is not None
    for line in proc.stdout:
        clean = line.rstrip("\n")
        print(clean, flush=True)
        write_event({
            "timestamp": now(),
            "event": "terminal_line",
            "source": source,
            "text": clean,
        })

    rc = proc.wait()

    write_event({
        "timestamp": now(),
        "event": "capture_done",
        "source": source,
        "returncode": rc,
        "proof_status": "PASS" if rc == 0 else "BLOCKED",
    })

    return rc

if __name__ == "__main__":
    raise SystemExit(main())
