#!/bin/bash
# goverlord/wire_pod_overlay/build_direct_sdk_tool.sh
# Builds one direct-SDK tool (chipper/gopod_probes/tools/direct_sdk_<name>.go) from
# this overlay folder - 2026-10-02, operator: "keep ~/wire-pod as native as possible".
# The tools compile inside Wire-Pod's own chipper Go module (they import the vendored
# vector-go-sdk and the gopoddirect package by its chipper module path), so the build
# runs in ~/wire-pod/chipper with Go's -overlay flag placing the tool and gopoddirect
# at their module paths; nothing is written into ~/wire-pod. The binary lands next to
# its source here (gitignored).
# Usage: ./build_direct_sdk_tool.sh direct_sdk_cube_blip_001
set -euo pipefail
OVERLAY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WIREPOD_CHIPPER_DIR="/home/goverlord/wire-pod/chipper"
name="${1:?usage: build_direct_sdk_tool.sh direct_sdk_<name>}"
src_dir="$OVERLAY_DIR/chipper/gopod_probes/tools"
virt_dir="$WIREPOD_CHIPPER_DIR/gopod_probes/tools"
overlay_json="$(mktemp --suffix=.json)"
trap 'rm -f "$overlay_json"' EXIT
python3 - "$src_dir" "$virt_dir" "$name" "$overlay_json" <<'PY'
import json, pathlib, sys
src, virt, name, out = sys.argv[1:]
rep = {f"{virt}/{name}.go": f"{src}/{name}.go"}
for f in pathlib.Path(src, "gopoddirect").glob("*.go"):
    rep[f"{virt}/gopoddirect/{f.name}"] = str(f)
json.dump({"Replace": rep}, open(out, "w"))
PY
cd "$WIREPOD_CHIPPER_DIR"
/usr/local/go/bin/go build -overlay="$overlay_json" -o "$src_dir/$name" "$virt_dir/$name.go"
echo "DIRECT_SDK_BUILD_OK -> $src_dir/$name"
