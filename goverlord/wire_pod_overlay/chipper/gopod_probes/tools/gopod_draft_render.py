#!/usr/bin/env python3
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_draft_render.py
#
# Draws draft_performance.md (to read) from draft_performance.json (to edit,
# and to play: "play the draft" in the Brobots chat), next to it. Called by
# harvest_dry_run.sh after a full dry run, and by its render-draft mode.
#
#   gopod_draft_render.py <path/to/draft_performance.json>
import json
import os
import sys

path = sys.argv[1]
d = json.load(open(path))
out = ["# Draft performance - %s" % d["run"], "",
       "Model %s, dry run %d min %d s. %s." % (d.get("model", ""), d.get("seconds", 0) // 60, d.get("seconds", 0) % 60, d.get("score", "")), "",
       "To change an answer: edit its `say` (words) or `cue` (animation) in `draft_performance.json`. "
       "To play it through the robots: type **play the draft** in the Brobots chat (**stop the draft** stops it).", "",
       "## Opening", ""]
for l in d.get("opening") or []:
    out += ["**%s:** %s" % (l["who"], l["text"]), ""]


def line(l):
    if l["who"] == "Chapter":
        return ["## " + l["text"], ""]
    if l["who"] == "Picture":
        return ["*Picture:* " + ", ".join(os.path.basename(p) for p in l["text"].split("|")), ""]
    return ["**%s:** %s" % (l["who"], l["text"]), ""]


for q in d["questions"]:
    for l in q.get("before") or []:
        out += line(l)
    out += ["**%d. %s** reads Operator to %s: *%s*" % (q["n"], q["read_by"], q["to"], q["ask"]), ""]
    for a in q.get("answers") or []:
        cue = "(%s) " % a["cue"] if a.get("cue") else ""
        out += ["> **%s:** %s%s" % (a["brobot"], cue, a["say"].replace("\n", " ")), ">",
                "> <sub>%s · %.1f s</sub>" % (", ".join(a.get("points") or []) or "no new point", a.get("llm_seconds", 0)), ""]
    for l in q.get("after") or []:
        out += line(l)
out += ["## Ending", ""]
for l in d.get("ending") or []:
    out += line(l)
md = os.path.join(os.path.dirname(path), "draft_performance.md")
open(md, "w").write("\n".join(out) + "\n")
print("DRAFT %s (read: %s)" % (path, md))
