# GOPOD Interview — Vamp (Video 1 of 2)

> Two hosts. A hallway backstage. Something is being written, live, right now.

**Naming note:** the interview is a two-video pair — this doc covers VAMP, video 1, the
pre-show banter that covers the live generation wait. Video 2 is the seven-exchange
performance itself:
[INTERVIEW RUN.md](SONG_02_BROBOTS_1_2_INTERVIEW_RUN.md).

---

## What Interview Vamp is

Two hosts, Brobot 3 (female voice, Kokoro `af_bella`) and Brobot 4 (male voice, Kokoro
`am_puck`), cover the GOPOD warm-up live while [the interview](SONG_02_BROBOTS_1_2_INTERVIEW_RUN.md)'s
seven exchanges are being written backstage — Ollama generating video 2's content in the
background while these two banter through it out front. Four movements: cold open,
Brobot 1 wakes, Brobot 2 wakes, handoff — plus a vamp the hosts fall back into if the
interview isn't ready yet when they reach the handoff. Robots are referred to by role,
never by gender.

A sample of the banter, close to verbatim:

> **Host (female):** *—and THAT, right there, is exactly why you never show up to a robot
> warm-up without backup.*
>
> **Host (male):** *She's not wrong. She's never wrong, I've learned that the hard way.
> Okay— hi, hello, we are LIVE, backstage, at what is shaping up to be a genuinely
> unhinged night for GOPOD.*

Backstage, Brobot 1 wakes first, then Brobot 2 — each getting its own hyped-up
introduction from the hosts before the handoff line closes video 1 out: *"Interview...
starts... now."*

---

## Why this is its own video

Video 1 and video 2 are conceptually one interview, split across two independently
fireable songs on disk since 2026-08-19
(`01_brobots_interview_vamp/`/`02_brobots_interview_run/`,
`gopod_notes/INTERVIEW_VAMP_SPLIT_001.md`) — a 2-video playlist rather than one combined
performance. Video 1's own no-generation fire path (`interview-vamp-play`) plays only
this banter, nothing more; the take-rolling path (`interview-vamp`) plays the same
banter WITH video 2's generation running alongside it, for when a fresh take actually
needs writing. See `SONG_SCAFFOLD.md`'s "The vamp — a detachable pre-show module"
section for the full model.

---

## History

Built as a reusable pre-show module — the gate, the loader, and the reporter delivery
are already song-agnostic, ready for a second generating song the day one exists. Only
[the interview](SONG_02_BROBOTS_1_2_INTERVIEW_RUN.md) actually uses it today, since
it's the only shelf song that generates its content live.

---

## GOPOD YAHMM (You Are Here Mall Map)

Part of GOPOD — see [tech/README.md](../README.md) for everything else in this folder, or [the root map](../../README.md) for the rest of GOPOD.
