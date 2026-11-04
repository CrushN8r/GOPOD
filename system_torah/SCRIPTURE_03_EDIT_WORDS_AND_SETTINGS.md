# Scripture 3, Edit words and settings

This is where you change what's said and how. “system-topic” edits a chapter topic: its words, points and pictures. “system-lingo” changes how a word is said. “system-set” changes any setting by its path, or a song step's words, volume or timing, and “system-add” and “system-remove” put items into a list or take them out.

The show's character lives here too: the roast level and the Jekyll and Hyde switch, how often Brobot 1 goes full Doc BS, his Doc BS lines with “system-docbs”, the parts of a live answer chain, the run cards, one question at a time, and which robot is Brobot 1. “system-voices” picks a voice for each numpad key, and this window's own look and theme can change as well.

Made a mistake? “system-undo” puts back the last edit. Don't know a name or path? Go to Scripture 4. Done editing? Go to Scripture 5.

## Commandments

- system-topic, a chapter topic: its words, points, label, reminder, pictures (alone: the topics)
- system-lingo, how a word is said: brobots / reporters / butler <word> = <said>
- system-set, a setting by its path, or a song step's text / volume / timing
- system-add, an item into a list (or key = value into a group)
- system-remove, an item out of a list (or a key out of a group)
- system-set chat_moves.thinking_text, the "Got it!" line: said while the model works, and the last line of the Value Points notice (then system-restart)
- system-set bubble_notices, what goes in a notice in a Brobot's bubble: value_points, egg, native_intent
- system-set <song> tempo, a song's tempo in seconds (a step's: system-set <song> <step> tempo_factor <n>)
- system-set system_menu.look, this menu's own look: font_size, background, code_background, code_color, results_background
- system-menu-theme, dark | light | blue | green | purple (then F5)
- system-voices-report, voice files: made, played, most played, not played - past 3d | week | month | year (report only, nothing moves)
- system-voices, numpad voices: KPN > persona [voice] - pick a voice per key (bold = used twice); robot-bodied keys need none
- system-set-list, every setting, song and step field
- system-undo, puts back the last edit
- system-set roast_dial.level, the roast level: 0 clean (never punches down), 1 roast, 2 public 18+, 3 paid 18+ afterburner - then system-restart
- system-set roast_dial.jekyll_hyde, Jekyll/Hyde: true lets the roast through on sensitive topics too - then system-restart
- system-set docbs.share, DocBS: how many of Brobot 1's answers go equation roast, 0-100 (-1: his Doc dial) - then system-restart
- system-docbs, DocBS's lines: the pools, one pool's lines (system-docbs openers_clean), or add / remove one
- system-set answer_chains.brobots.parts, a live answer's chain - its parts in order (system-add / system-remove one part)
- system-set game.runs, a run card - its words, song, lines and pre/post chat (system-set game.runs.1.experience.closing ...)
- system-set line_queue.on, one question at a time: true / false - then system-restart
- system-set brobots_setup.brobot_1, which robot is Brobot 1 (an ESN; empty: Wire-Pod's robots decide) - then system-restart

*Don't know the name or path? Go to scripture 4. Done editing? Go to scripture 5.*

---

[Scripture 2, Play](SCRIPTURE_02_PLAY.md) · [The GOPOD Torah](../GOPOD_TORAH.md) · [Scripture 4, Find](SCRIPTURE_04_FIND.md)
