#!/usr/bin/env python3
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_take_sheet_render.py
#
# Draws take_sheet.md (to read) from take_sheet.json (to edit), next to it.
# Called by Wire-Pod after a kept session adds its takes (gopod_take_sheet.go)
# and by harvest_dry_run.sh.
#
#   gopod_take_sheet_render.py <path/to/take_sheet.json>
import json
import os
import sys

path = sys.argv[1]
sheet = json.load(open(path))
lines = ["# Take sheet", "",
         "Edit `take_sheet.json`: set \"keep\": true on one take per point, edit \"say\" to change the words.", "",
         "Runs: " + ", ".join(sheet.get("runs") or []), ""]
chapter, missing = None, 0
for p in sheet["points"]:
    if p["chapter"] != chapter:
        chapter = p["chapter"]
        lines += ["## " + chapter, ""]
    lines.append("### %s %s" % (p["topic"], p["point"]))
    if not p.get("takes"):
        missing += 1
        lines += ["- (no take yet)", ""]
        continue
    for t in p["takes"]:
        mark = "KEEP " if t.get("keep") else ""
        how = "" if t.get("input") in (None, "", "typed") else " [%s]" % t["input"]
        lines.append("- %s**%s** %s%s - *%s*" % (mark, t["take"], t["to"], how, t["ask"]))
        lines.append("  - %s%s" % (("(" + t["cue"] + ") ") if t.get("cue") else "", t["say"].replace("\n", " ")))
    lines.append("")
lines.insert(5, "Points with no take yet: %d of %d" % (missing, len(sheet["points"])))
lines.insert(6, "")
md = os.path.splitext(path)[0] + ".md"
open(md, "w").write("\n".join(lines) + "\n")
print("HARVEST_SHEET %s (read: %s)" % (path, md))
