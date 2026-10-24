#!/usr/bin/env python3
# goverlord/wire_pod_overlay/chipper/gopod_probes/tools/gopod_dry_run_001.py
"""GOPOD's own test run (2026-10-24, operator: "Do I have a full complete dry run
file? * make me a new one to test on my behalf * Make me a fucking agent to whoop
your sorry AI ass").

Types a full session into GOPOD the way the operator does (the chat's own typed
line, from this machine = the admin), one line at a time, waiting for each
answer to finish. Then it checks everything that has gone wrong before:

  - every Operator line shows before the Brobot's answer to it
  - every line gets an answer
  - Wire-Pod's log: fallback line used, answer shape drift, voices not heard,
    replies not shown, a stuck baton, stuck replies, refused visitor calls

and writes one short report to gopod_notes/. Live robots answer (they speak).
Run it after LLM Ready, with the PC's chat page open and clicked once:

    python3 gopod_dry_run_001.py
"""
import datetime
import json
import subprocess
import sys
import time
import urllib.parse
import urllib.request
from pathlib import Path

BASE = "http://localhost:8080"
NOTES = Path.home() / "crushn8r_git" / "gopod_notes"
ESN = {"1": "0dd1b9e9", "2": "0dd1d8bf", "all": "all"}

# The session: (to, line, what it checks, words the answer must have).
# 2026-10-24 (operator: "Ask questions to finish the game, not tell a joke * ask
# harder math questions, use fraction, decimal, repeating to test the fucker"):
# the game's own value-point questions come in after the openers (game_steps).
STEPS = [
    ("all", "hey brobots", "a line to both: duo order, hand-over", []),
    ("1", "good morning", "greeting golden line, robot clock", []),
    ("2", "good evening", "greeting checked against the time", []),
    ("1", "what's the weather", "weather take (PLAYHEAD), weather line", []),
    ("2", "what time is it?", "time take, robot clock", []),
    ("1", "convert 0.75 to a fraction, explain the math", "popup: the math Easter egg pop-up", ["3/4"]),
    ("2", "convert 7/8 to a decimal, explain the math", "math: fraction to decimal", ["0.875"]),
    ("1", "convert 0.375 to a fraction, explain the math", "math: decimal to fraction", ["3/8"]),
    ("2", "what is 5/6 as a decimal? explain the math", "math: repeating decimal", ["0.83"]),
    ("1", "convert 0.272727 repeating to a fraction, explain the math", "math: repeating back to a fraction", ["3/11"]),
    ("2", "is 0.33 the same as 1/3? explain the math", "math: close is not equal", ["no"]),
    ("2", "what do you see?", "popup: the photo Easter egg pop-up", []),
]
END_STEPS = [
    ("1", "thank you. Interview over. What have we learned about GOPOD?", "the wrap-up", []),
    ("1", "thanks bros", "the thanks answer (fell back twice before)", []),
]
CHIPPER = Path(__file__).resolve().parents[2]


def game_steps():
    """Each core value point's golden question (value_points.core_points, the
    questions in harvest_questions.json numbered per topic), in the game's order."""
    settings = json.loads((CHIPPER / "webroot/gopod_brobots_settings.json").read_text())
    core = settings.get("value_points", {}).get("core_points", [])
    questions = json.loads((CHIPPER / "gopod_probes/interview_harvest/harvest_questions.json").read_text())["questions"]
    seen, out = {}, []
    for q in questions:
        seen[q["topic"]] = seen.get(q["topic"], 0) + 1
        if f"{q['topic']}#{seen[q['topic']]}" in core:
            out.append((q["to"].split()[-1], q["ask"], f"value point {q['topic']} #{seen[q['topic']]}", []))
    return out


LOG_PROBLEMS = {
    "BROBOT_FALLBACK": "the fallback line was said instead of the answer",
    "DRIFT": "an answer broke its shape",
    "not heard after": "a voice wasn't heard by the page",
    "NOT shown": "a reply bubble wasn't shown",
    "baton held 10 s+": "a bubble waited 10 s+ for the one above",
    "GOPOD_STUCK_REPLY": "a robot dropped off mid-answer",
    "GOPOD_FEED_CHECK": "the page played past a reply with no bubble",
}


def feed():
    with urllib.request.urlopen(BASE + "/gopod_chat_feed.json", timeout=10) as r:
        return json.loads(r.read())


def send(to, text):
    body = urllib.parse.urlencode({"serial": ESN[to], "text": text}).encode()
    with urllib.request.urlopen(urllib.request.Request(BASE + "/api-sdk/gopod_typed_input", data=body), timeout=30) as r:
        return r.read().decode()


def brobot_answer(e):
    who = e.get("who", "")
    return who in ("Brobot 1", "Brobot 2") and not e.get("text", "").startswith(("generating:", "notice:"))


def wait_answers(after_id, want, quiet=12, most=150):
    """Waits until `want` Brobot answers came after after_id and the feed has been
    quiet for `quiet` seconds (the answers said), at most `most` seconds."""
    start = time.time()
    last_id, last_change = after_id, time.time()
    while time.time() - start < most:
        time.sleep(1)
        events = [e for e in feed().get("events", []) if e["id"] > after_id]
        top = max([e["id"] for e in events], default=after_id)
        if top != last_id:
            last_id, last_change = top, time.time()
        answers = [e for e in events if brobot_answer(e)]
        if len(answers) >= want and time.time() - last_change >= quiet:
            return events
    return [e for e in feed().get("events", []) if e["id"] > after_id]


def main():
    started = datetime.datetime.now()
    since = started.strftime("%Y-%m-%d %H:%M:%S")
    rows, problems = [], []
    # python3 gopod_dry_run_001.py photo  - only the steps whose check names "photo".
    only = sys.argv[1].lower() if len(sys.argv) > 1 else ""
    for to, text, what, must in STEPS + game_steps() + END_STEPS:
        if only and only not in what.lower():
            continue
        before = max([e["id"] for e in feed().get("events", [])], default=0)
        send(to, text)
        # A line to both, or the math pop-up (Brobot 2 follows up), has two answers.
        events = wait_answers(before, 2 if to == "all" or what.startswith("popup: the math") else 1)
        op = next((e for e in events if e.get("who") == "Operator" and text.lower() in e.get("text", "").lower()), None)
        answers = [e for e in events if brobot_answer(e)]
        note = []
        if op is None:
            note.append("no Operator bubble")
        if not answers:
            note.append("no answer")
        elif op is not None and answers[0]["id"] < op["id"]:
            note.append("answer came before the Operator line")
        said = " ".join(e["text"] for e in answers).lower().replace(" ", "")
        for w in must:
            if w.lower().replace(" ", "") not in said:
                note.append(f"answer lacks \"{w}\"")
        # A pop-up step: its pop-up came on the feed, and the page showed it.
        if what.startswith("popup:"):
            pops = [e for e in events if e.get("who") == "Picture" and e.get("text", "").startswith("popup:")]
            if not pops:
                note.append("no pop-up")
            else:
                played = subprocess.run(["journalctl", "-u", "wire-pod", "--since", since, "--no-pager"], capture_output=True, text=True).stdout
                if f"GOPOD_PAGE played {pops[0]['id']} " not in played:
                    note.append("pop-up not shown by the page")
        first = answers[0]["text"].replace("⁣", "")[:90] if answers else ""
        rows.append((text, what, "ok" if not note else "; ".join(note), first))
        if note:
            problems.append(f"\"{text}\": " + "; ".join(note))
        print(f"{'ok ' if not note else 'BAD'} {text} -> {first[:60]}", flush=True)
    log = subprocess.run(["journalctl", "-u", "wire-pod", "--since", since, "--no-pager"], capture_output=True, text=True).stdout
    found = {}
    for line in log.splitlines():
        for key, why in LOG_PROBLEMS.items():
            if key in line:
                found.setdefault(key, []).append(line.split(": ", 1)[-1][:220])
    for key, lines in found.items():
        problems.append(f"{LOG_PROBLEMS[key]} ({len(lines)}x)")
    # The game's own score at the end (the last "Topics ... of ..." notice).
    score = ""
    for e in feed().get("events", []):
        if "Topics " in e.get("text", "") and " of " in e.get("text", ""):
            score = e["text"].split("|")[-1]
    stamp = started.strftime("%Y%m%d_%H%M")
    out = NOTES / f"DRY_RUN_REPORT_{stamp}.md"
    md = [f"# GOPOD dry run {started:%Y-%m-%d %H:%M}", "",
          f"**Bottom line:** {'all clean' if not problems else str(len(problems)) + ' problem(s)'}", "", f"**Game score at the end:** {score or 'none shown'}", ""]
    if problems:
        md += ["## Problems", ""] + [f"- {p}" for p in problems] + [""]
    md += ["## Each line", "", "| Line | Checks | Result | First answer |", "|---|---|---|---|"]
    md += [f"| {t} | {w} | {r} | {a.replace('|', '/')} |" for t, w, r, a in rows]
    for key, lines in found.items():
        md += ["", f"## Log: {LOG_PROBLEMS[key]}", ""] + [f"- `{l}`" for l in lines[:8]]
    out.write_text("\n".join(md) + "\n")
    print(f"\n{'ALL CLEAN' if not problems else str(len(problems)) + ' PROBLEM(S)'} - {score} - {out}")
    return 0 if not problems else 1


if __name__ == "__main__":
    sys.exit(main())
