#!/usr/bin/env python3
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_draft_from_picks.py
#
# The lead-guitar splice (operator 2026-10-06: "record several takes and
# splice together"; 2026-10-07: "build the draft from best takes"). Builds one
# draft performance from the keepers on the take sheet ("keep": true, one per
# value point, any run): the newest dry run's draft is the frame - opening,
# chapters, pictures, sign-off, ending - and each answer is swapped for the
# keeper of its value point (matched on topic and point, not on question text).
# Writes <harvest root>/<date>/<HHMMSS>_best_takes/draft_performance.json and
# .md; "play the draft" plays the newest one.
#
#   gopod_draft_from_picks.py <take_sheet.json>
import datetime
import glob
import json
import os
import re
import subprocess
import sys

# A reminder from before 2026-10-07 said a value point ("Hey, that reminds me...
# Rumination stops when..."); today's only name chapters ("coming up: ...") or
# topics ("we've still got ... to get to!"). A keeper from an older run loses
# its old reminder, from the phrase to the end.
OLD_REMINDER = re.compile(r"\s*Hey,? that reminds me(?:\.\.\.|,|\u2026)(?!\s*(?:coming up|we've still got))[^\n]*$", re.I)


def drop_old_reminder(text):
    return OLD_REMINDER.sub("", text).rstrip()

sheet_path = sys.argv[1]
root = os.path.dirname(sheet_path)
sheet = json.load(open(sheet_path))
keep = {}
for p in sheet["points"]:
    picked = [t for t in p["takes"] if t.get("keep")]
    if len(picked) > 1:
        sys.exit("more than one keeper for %s %s" % (p["topic"], p["point"]))
    if picked:
        keep["%s %s" % (p["topic"], p["point"])] = picked[0]
bases = sorted(glob.glob(os.path.join(root, "*", "*_harvest", "draft_performance.json")))
if not bases:
    sys.exit("no dry-run draft to use as the frame")
draft = json.load(open(bases[-1]))
frame = os.path.basename(os.path.dirname(bases[-1]))
swapped, missing = 0, []
for q in draft["questions"]:
    for a in q["answers"]:
        if not a.get("points"):
            continue  # the sign-off: the frame's own
        key = a["points"][0]
        t = keep.get(key)
        if not t:
            missing.append(key)
            continue
        a.update({"cue": t.get("cue", ""), "say": drop_old_reminder(t["say"]), "spoken": t.get("spoken") or [], "llm_seconds": t.get("llm_seconds", 0), "take": t["take"]})
        swapped += 1
now = datetime.datetime.now()
out = os.path.join(root, now.strftime("%Y-%m-%d"), now.strftime("%H%M%S") + "_best_takes")
os.makedirs(out, exist_ok=True)
draft["run"] = os.path.basename(out)
draft["about"] = ("The spliced draft (operator: \"build the draft from best takes\"): each answer is the keeper of its value point on the take sheet, "
                  "the frame (opening, chapters, pictures, sign-off, ending) is " + frame + ". Edit an answer's \"say\" or \"cue\"; \"take\" names where it came from. "
                  "Type \"play the draft\" in the Brobots chat to play the newest draft.")
path = os.path.join(out, "draft_performance.json")
json.dump(draft, open(path, "w"), indent=2, ensure_ascii=False)
subprocess.run([sys.executable, os.path.join(os.path.dirname(os.path.abspath(__file__)), "gopod_draft_render.py"), path], check=True)
print("BEST_TAKES swapped=%d missing=%d frame=%s" % (swapped, len(missing), frame))
for m in missing:
    print("  no keeper:", m)
