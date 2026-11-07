# Run schema

- **about:** 2026-11-06 (operator: "Please do: Schemas for run cards and answer chains"): the master run card - every run Goverlord can start (`game.runs`, show/game-runs.md) fills these. A run still `pending: true` needs only its name and title. A run's `song` is a song folder in `goverlord/runtime/songs/`. The system check (`gopodSchemaCheck`, every restart) lists a gap "to fix".

---

## Required
1. name: the run's id ("fart_joke")
2. title: its name in the chat ("Fart Joke")

## Required unless pending
1. song: its song folder ("03_fart_joke") - it must exist
2. start: the words that start it ("start fart joke")
3. stop: the words that stop it ("stop fart joke")
4. said: how Goverlord names it in the run lines ("Fart joke")
