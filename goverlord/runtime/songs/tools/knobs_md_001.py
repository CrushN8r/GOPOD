"""A song's knobs as a topic-cluster md file: zKnobs.md (2026-11-04, step 9 of
gopod_notes/GOPOD_CLUSTERS_PLAN_002.md; operator: "I have zKnobs.json (work file)
that saves to knobs.json * so zKnobs.json (work file) to md instead? Yes if it fits
topic-cluster format"; "zKnobs.md is the one source").

zKnobs.md is the one source of a song's knobs: knobs.json is built from it before
each run (knobs_envelope_001.resolve_active_knobs_path), and the tools that write
knobs (phcal apply, tempo-set, robot pick) write it back here
(knobs_envelope_001.save_knobs_envelope). The format is the topic-cluster one the
Brobots chat reads (gopod_topic_clusters.go):

    # <title>, its head ("- **words:** ..."), then "---", then sections:
    ## <label> (<place>)      a section fills a place: "." is the top level,
                              "steps.3" the fourth step, "steps.3.moves" inside it
    1. key: value             a point; the value is:
                                `text`        exactly (spaces kept, "\\n" a break)
                                12 / 0.5      a number
                                true / false  a switch; null, [] and {} as written
                                a | b | c     a list ("a |" a list of one)
                                anything else text
    ## <label> (<place>[])    a line list: each point one line, as written

Only the points carry knobs; any other line in the file is a note for the reader.
"""
import json
import re

_NUM = re.compile(r"^-?(0|[1-9][0-9]*)(\.[0-9]+)?$")


# --- writing -------------------------------------------------------------------

def _enc(v):
    """A point's value."""
    if isinstance(v, bool):
        return "true" if v else "false"
    if v is None:
        return "null"
    if isinstance(v, (int, float)):
        return json.dumps(v)
    if v == []:
        return "[]"
    if v == {}:
        return "{}"
    if "`" in v:
        raise ValueError("a backtick in a knob's text: " + v[:60])
    if v != v.strip() or "\n" in v or " | " in v or v.endswith(" |") or v == "" or _NUM.match(v) or v in ("true", "false", "null", "[]", "{}"):
        return "`" + v.replace("\n", "\\n") + "`"
    return v


def _short_list(v):
    return (v and all(isinstance(i, str) and i and i == i.strip() and "\n" not in i and " | " not in i
                      and not i.endswith(" |") and "`" not in i and not (i.startswith("`") or i.endswith("`")) for i in v)
            and len(" | ".join(v)) < 160)


def _sections(path, obj, label):
    out, pts, later = [], [], []
    for k, v in obj.items():
        if isinstance(v, (str, bool, int, float)) or v is None or v == [] or v == {}:
            pts.append("%s: %s" % (k, _enc(v)))
        elif isinstance(v, list) and all(isinstance(i, str) for i in v) and _short_list(v):
            pts.append("%s: %s" % (k, " | ".join(v) + (" |" if len(v) == 1 else "")))
        else:
            later.append((k, v))
    if pts:
        out += ["## %s (%s)" % (label, path)] + ["%d. %s" % (i, p) for i, p in enumerate(pts, 1)] + [""]
    for k, v in later:
        sub = k if path == "." else path + "." + k
        if isinstance(v, list) and all(isinstance(i, str) for i in v):
            for i in v:
                if i != i.rstrip() or "\n" in i:
                    raise ValueError("a line with an edge space or a break: " + sub)
            out += ["## %s - %s (%s[])" % (label, k, sub)] + ["%d. %s" % (n, i) for n, i in enumerate(v, 1)] + [""]
        elif isinstance(v, list):
            for n, item in enumerate(v):
                if not isinstance(item, dict):
                    raise ValueError("a list entry that isn't a group: %s.%d" % (sub, n))
                name = item.get("step_id") or item.get("name") or item.get("id") or "%s %d" % (k, n + 1)
                out += _sections("%s.%d" % (sub, n), item, str(name))
        elif isinstance(v, dict):
            out += _sections(sub, v, "%s - %s" % (label, k) if path != "." else k)
        else:
            raise ValueError("can't write " + sub)
    return out


def md_from_knobs(knobs, title=""):
    """The zKnobs.md text for a song's knobs."""
    song = knobs.get("song_id") or title or "the song"
    head = [
        "# %s - knobs" % (title or song),
        "- **words:** knobs, zKnobs, %s" % song,
        "- **ask:** What are %s's knobs?" % song,
        "- **about:** This song's zKnobs.md - the one source of its knobs. knobs.json is built from it before "
        "each run; phcal apply, tempo-set and robot pick write back here. A value in backticks is that text "
        "exactly; a bare number or true/false is that value; a | b is a list; a heading ending in [] is a "
        "line list; (.) is the top level, (steps.3) the fourth step.",
        "---",
        "",
    ]
    return "\n".join(head + _sections(".", knobs, "The song")).rstrip("\n") + "\n"


# --- reading -------------------------------------------------------------------

def _value(v):
    if len(v) >= 2 and v.startswith("`") and v.endswith("`"):
        return v[1:-1].replace("\\n", "\n")
    if v in ("true", "false"):
        return v == "true"
    if v == "null":
        return None
    if v == "[]":
        return []
    if v == "{}":
        return {}
    if _NUM.match(v):
        return json.loads(v)
    if " | " in v or v.endswith(" |"):
        return [i.strip() for i in v[:-2].split(" | ") if i.strip()] if v.endswith(" |") else [i.strip() for i in v.split(" | ") if i.strip()]
    return v


def _place(root, parts):
    """The group or list at parts, made as it goes; returns (container, key) of its slot."""
    node, holder, key = root, None, None
    for i, p in enumerate(parts):
        if p.isdigit():
            n = int(p)
            if node is None:
                node = []
                holder[key] = node
            if not isinstance(node, list):
                raise ValueError(".".join(parts[:i]) + " isn't a list")
            if n > len(node):
                raise ValueError("list entry %s comes before entry %d" % (p, len(node)))
            if n == len(node):
                node.append(None)
            holder, key, node = node, n, node[n]
            continue
        if node is None:
            node = {}
            holder[key] = node
        if not isinstance(node, dict):
            raise ValueError(".".join(parts[:i]) + " isn't a group")
        holder, key, node = node, p, node.get(p)
    return holder, key


def knobs_from_md(text):
    """A song's knobs from its zKnobs.md text."""
    root, body, path, lines, points = {}, False, None, False, []
    sections = []
    for raw in text.split("\n"):
        t = raw.strip()
        if t == "---" and not body:
            body = True
            continue
        if not body:
            continue
        if t.startswith("## "):
            o, c = t.rfind("("), t.rfind(")")
            sections.append([t[o + 1:c].strip() if o >= 0 and c > o else "", []])
            continue
        m = re.match(r"^(\d+)\. ", t)
        if m and sections:
            sections[-1][1].append(raw.lstrip()[len(m.group(0)):].rstrip(" \t\r"))
    for path, pts in sections:
        if not path:
            raise ValueError("a section with no place in brackets")
        lines = path.endswith("[]")
        if lines:
            path = path[:-2]
        if path == ".":
            target = root
        else:
            holder, key = _place(root, path.split("."))
            if lines:
                holder[key] = list(pts)
                continue
            if (holder.get(key) if isinstance(holder, dict) else holder[key]) is None:
                holder[key] = {}
            target = holder[key]
            if not isinstance(target, dict):
                raise ValueError(path + " isn't a group")
        for pt in pts:
            at = pt.find(": ")
            if at <= 0:
                raise ValueError('"%s" has no ": "' % pt)
            target[pt[:at].strip()] = _value(pt[at + 2:].strip())
    return root


# --- the command (Wire-Pod calls it after it edits a song's knobs.json) ---------

def main(argv):
    """knobs_md_001.py build <song_dir>: knobs.json built from zKnobs.md (written only
    when it differs). knobs_md_001.py write <song_dir>: zKnobs.md rewritten from
    knobs.json. knobs_md_001.py check <song_dir>: "same" when zKnobs.md builds
    knobs.json as it is, else "differs" (exit 1). Nothing for a song without a zKnobs.md."""
    from pathlib import Path
    if len(argv) != 3 or argv[1] not in ("write", "check", "build"):
        print(main.__doc__)
        return 2
    song = Path(argv[2])
    md, clean = song / "zKnobs.md", song / "knobs.json"
    if not md.exists():
        print("no zKnobs.md")
        return 0
    if argv[1] == "build":
        built = knobs_from_md(md.read_text(encoding="utf-8"))
        try:
            same = json.loads(clean.read_text(encoding="utf-8")) == built
        except (OSError, ValueError):
            same = False
        if not same:
            clean.write_text(json.dumps(built, indent=2) + "\n", encoding="utf-8")
        print("knobs.json " + ("as it was" if same else "built from zKnobs.md"))
        return 0
    knobs = json.loads(clean.read_text(encoding="utf-8"))
    if argv[1] == "write":
        md.write_text(md_from_knobs(knobs, song.name), encoding="utf-8")
        print("zKnobs.md written")
        return 0
    same = knobs_from_md(md.read_text(encoding="utf-8")) == knobs
    print("same" if same else "differs")
    return 0 if same else 1


if __name__ == "__main__":
    import sys
    sys.exit(main(sys.argv))
