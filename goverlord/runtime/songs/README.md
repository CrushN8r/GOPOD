# songs

Each folder here is one song: `story.md` (its lines) and `knobs.json` (what each step does, and
when). The folders are the list.

- `_new_song/` - the template: copy it to start a song (`_new_song/_llm_warm/` for an LLM warm
  song). `_new_song/song-schema.md` - what every song must have.
- `global_knobs.json` - the move values and waits every song starts from; a song's own
  `knobs.json` overrides them, a step's own value over that.
- `tools/` - the song engine and its helpers.
- How a song is made, played and edited: [song making](../../../alias_play_studio/02_SONGS.md).

Not in the repo: `zKnobs.json`, `runs/`, `notation/` (each song's private working files).

<details><summary>Game card: our songs</summary>
<!-- gopod topic: 02_the_songs -->

# our songs
- **chapter:** Show and tell

- **words:** songs, song, your routines, what shows, your shows, performances, scripted, sequences
- **reminder:** Hey, that reminds me... We perform songs too - the Bingo test, the nap, the fart joke.
- **say 1:** The Bingo test: the two of us trade banter and call three Bingo balls - B-1, I-9 and O-0.
- **notes 1:** the Bingo test | two Brobots, three Bingo balls | B-1, I-9, O-0 | "start bingo test"
- **notes 2:** the nap | "Do Baby Robots Dream?" | we sleep through the narrated poem | wake up together
- **notes 3:** the fart joke | stuck in an elevator | three thinking pauses, three farts | "start fart joke"
- **notes 4:** the interview | scripted questions | the same every time | "start interview"
- **ask:** Tell me about your songs
- **summary:** the Bingo test: two Brobots, three Bingo balls | the nap: we sleep through the narrated poem, wake up together | the fart joke: an elevator, three farts | the interview: scripted questions, the same every time
- **source:** goverlord/runtime/songs/README.md
- **ask 1:** Brobot 2, what songs do you do?
- **ask 2:** Brobot 1, and the song where you two fall asleep?
- **ask 3:** Brobot 2, what's the song in the elevator?
- **ask 4:** Brobot 1, and the interview song?

---

Your songs - scored sequences you perform together, typed in the chat to start ("start bingo test", "start baby brobots", "start fart joke", "start interview"). Reveal these, in order, in your own words:
1. The Bingo test: the two of you trade banter and call three Bingo balls - B-1, I-9 and O-0 - a test before the real Bingo game.
2. The nap, "Do Baby Robots Dream?": a lyric video - you both sleep through a narrated poem with music, then wake up together.
3. The fart joke: Brobot 2 sets up an elevator, three thinking pauses each end in a fart, and Brobot 1 gets frustrated, then angry, then leaves.
4. The interview: scripted questions, the same every time, ready to record.

</details>
