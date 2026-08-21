#!/usr/bin/env python3
"""GOPOD consistency checker - the mechanical half of the `decoupler` skill.

Finds copies of a fact that have drifted from the one real source, across three
places this has actually happened (2026-07-23/24, see gopod_notes/
PHA0B_SONG_LIST_CLEANUP_SURVEY_001.md and the memory-file path fixes from the same
pass): repo-wide markdown links, the live alias-lib shell files vs. their own
registry doc, and gopod_notes/ citations inside the persistent memory files.

Read-only. Never edits anything - prints findings, exit code 0 always (this is a
report tool, not a gate). The `decoupler` skill (SKILL.md, same directory) is the
judgment layer on top: which findings are real drift vs. a known false-positive
(e.g. a collapsed range like "`codex-1`...`codex-6`" in prose), and what to do
about each one.

Run: python3 .claude/skills/gopolisher/gopod_consistency_check_001.py
"""
import os
import re
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
ALIAS_LIB_DIR = Path.home() / ".gopod_alias_lib"
BASHRC = Path.home() / ".bashrc"
BASH_ALIASES = Path.home() / ".bash_aliases"
MEMORY_DIR = (
    Path.home()
    / ".claude"
    / "projects"
    / "-home-goverlord-crushn8r-git-GOPOD"
    / "memory"
)
GOPOD_NOTES_DIR = Path.home() / "crushn8r_git" / "gopod_notes"
ALIAS_LIBRARY_DOC = REPO_ROOT / "tech" / "alias_play_studio" / "ALIAS-LIBRARY.md"

# Vendored/third-party trees this repo doesn't own - never flag their links.
SKIP_DIR_MARKERS = ("/.git/", "/SDK/sources/")

MD_LINK_RE = re.compile(r"\]\(([^)]+\.md)\)")
GOPOD_NOTES_CITE_RE = re.compile(r"gopod_notes/([A-Za-z0-9_.\-]+\.md)")
# Matches the file name regardless of what precedes it - "source ", ". ", a bare
# for-loop list entry, "$HOME/", "~/", or an absolute /home/<user>/ path all lead
# here, and trying to enumerate every valid bash sourcing form separately is
# exactly the kind of copy that drifts (this file already missed one once - see
# gopolisher/SKILL.md's own worked example).
SH_SOURCE_RE = re.compile(r"\.gopod_alias_lib/([A-Za-z0-9_.\-]+\.sh)")
DEF_RE = re.compile(r"^(?:alias\s+([a-zA-Z0-9_-]+)=|([a-zA-Z0-9_-]+)\(\)\s*\{)")


def find_repo_md_files():
    out = []
    for root, _dirs, files in os.walk(REPO_ROOT):
        rootp = root.replace(str(REPO_ROOT), "", 1)
        if any(marker in rootp + "/" for marker in SKIP_DIR_MARKERS):
            continue
        for f in files:
            if f.endswith(".md"):
                out.append(Path(root) / f)
    return out


def check_repo_links():
    print("=== 1. Repo-wide markdown link resolution ===")
    broken = []
    for src in find_repo_md_files():
        try:
            text = src.read_text(encoding="utf-8")
        except (OSError, UnicodeDecodeError):
            continue
        for m in MD_LINK_RE.finditer(text):
            target = m.group(1)
            if target.startswith("http"):
                continue
            resolved = os.path.normpath(os.path.join(src.parent, target))
            if not os.path.isfile(resolved):
                broken.append((src.relative_to(REPO_ROOT), target))
    if not broken:
        print("  clean - every .md-to-.md link resolves (vendored SDK docs excluded)")
    else:
        for src, target in broken:
            print(f"  BROKEN: {src} -> {target}")
    print()
    return broken


def loaded_alias_lib_files():
    """Which .gopod_alias_lib/*.sh files .bashrc/.bash_aliases actually load."""
    files = set()
    for path in (BASHRC, BASH_ALIASES):
        if not path.is_file():
            continue
        text = path.read_text(encoding="utf-8")
        for m in SH_SOURCE_RE.finditer(text):
            files.add(m.group(1))
    return sorted(files)


def defined_names(sh_path):
    names = []
    try:
        text = sh_path.read_text(encoding="utf-8")
    except OSError:
        return names
    for line in text.splitlines():
        m = DEF_RE.match(line)
        if m:
            names.append(m.group(1) or m.group(2))
    return names


def check_alias_registry_drift():
    print("=== 2. .bashrc/.bash_aliases loaded files vs. ALIAS-LIBRARY.md ===")
    loaded = loaded_alias_lib_files()
    if not loaded:
        print("  BLOCKED: no .gopod_alias_lib/*.sh source lines found in .bashrc/.bash_aliases")
        print()
        return []
    if not ALIAS_LIBRARY_DOC.is_file():
        print(f"  BLOCKED: {ALIAS_LIBRARY_DOC} not found")
        print()
        return []
    doc_text = ALIAS_LIBRARY_DOC.read_text(encoding="utf-8")
    print(f"  loaded files: {', '.join(loaded)}")
    missing = []
    for fname in loaded:
        sh_path = ALIAS_LIB_DIR / fname
        if not sh_path.is_file():
            print(f"  BROKEN SOURCE: .bashrc/.bash_aliases loads {fname}, file doesn't exist")
            continue
        for name in defined_names(sh_path):
            if name not in doc_text:
                missing.append((fname, name))
    if not missing:
        print("  clean - every defined alias/function name appears somewhere in the doc")
    else:
        print("  Possible gaps (check for a collapsed-range false positive, e.g.")
        print('  "`codex-1`...`codex-6`", before treating any of these as real drift):')
        for fname, name in missing:
            print(f"  MAYBE MISSING: {name}  (in {fname})")
    print()
    return missing


def check_memory_gopod_notes_citations():
    print("=== 3. Memory files' gopod_notes/ citations ===")
    if not MEMORY_DIR.is_dir():
        print(f"  BLOCKED: {MEMORY_DIR} not found")
        print()
        return []
    stale = []
    for mem_file in sorted(MEMORY_DIR.glob("*.md")):
        text = mem_file.read_text(encoding="utf-8")
        for m in GOPOD_NOTES_CITE_RE.finditer(text):
            fname = m.group(1)
            claimed = GOPOD_NOTES_DIR / fname
            if claimed.is_file():
                continue
            found = list(GOPOD_NOTES_DIR.rglob(fname))
            if found:
                stale.append((mem_file.name, fname, "moved", str(found[0])))
            else:
                stale.append((mem_file.name, fname, "missing", None))
    if not stale:
        print("  clean - every gopod_notes/*.md citation resolves at top level")
    else:
        for mem_name, fname, kind, where in stale:
            if kind == "moved":
                print(f"  STALE (moved): {mem_name} cites {fname} -> now at {where}")
            else:
                print(f"  MISSING ENTIRELY: {mem_name} cites {fname}, not found anywhere under gopod_notes/")
    print()
    return stale


def main():
    broken_links = check_repo_links()
    alias_gaps = check_alias_registry_drift()
    stale_citations = check_memory_gopod_notes_citations()
    total = len(broken_links) + len(alias_gaps) + len(stale_citations)
    print(f"=== Summary: {total} raw finding(s) across all 3 checks ===")
    print("Raw findings need a judgment pass before acting - see gopolisher/SKILL.md.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
