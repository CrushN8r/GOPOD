# web_display

Local cockpit/display servers that render live GOPOD session state to a
browser page — the visual surface an operator or bystander watches during a
demo.

## Contents

- `gopod_demo_8011/` — the main, live GOPOD demo display server
  (`gopod_demo_8011.py`), serving a big-font chat display, live chat
  message log, QR cache, and camera/state snapshots on port 8011.
- `ptt_gominion_test_001/` — removed in Stage B cleanup (Rule 6, zero live
  callers).

`gopod_demo_8011.py` reads from small live-state JSON files
(`camera_state.json`, `weather_state.json`, etc.) that other subsystems
(`robot/`, `wire-pod/`) write to, rather than talking to hardware directly.
