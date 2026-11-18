# GOPOD system help
- **words:** system-help, help, commands, system commands
- **ask:** What system- commands are there?
- **see also:** system-texts (what each command answers)
---

## The help list, in order (system_help.lines)
1. system commands (typed on this keyboard):
2.   system-set <setting> <value>      change a setting (system-set master-volume 35%)
3.   system-set <path.to.setting> <value>   any setting by its path (system-set managers.qr_note Scan it!)
4.   system-add / system-remove <path> <item>   an item in a list, or key = value in a group
5.   system-lingo <brobots|reporters|butler> <word> = <said>   (remove <word>; the lingo alone: its list)
6.   system-lingo-improver             the lingo improver page: pick a voice, hear a spelling, keep it
7.   system-topic [<topic> ...]        a chapter topic: words, points (say lines), label, reminder, pictures
8.   system-skip                       skip the voiced lines playing - in the tour, on to the next scripture (from Voices ready on)
9.   Esc                               exit the run going (a song, the tour, the comm check); with none, skip
10.   system-restart                    a forced restart, like wpr 2 (speaker + mic set); the page reloads itself
11.   system-audit                      the full audit, no robots called: files, Go tests, web page, phone app, every test, the log
12.   system-score <song>               a song's score: every step and its words
13.   system-play <song> [from <step>] [to <step>] [robot 1|2] [dry]   a song or part of it (pha0b)
14.   system-gap <seconds>|off|default  reporter gap for system-play
15.   system-tweaks on|off              apply phcal tweaks before a live system-play
16.   system-calibrate <move> <1|2> [values] [dry]   one move alone (phcal)
17.   system-set <song> tempo <seconds>   the song's tempo (tempo-set); a step's tempo_factor too
18.   system-brobot <stats|volume|eyes|locale|units|location|timezone|say|wake|head|lift|wheels|camera> [1|2|both]
19.   system-faces [list|rename|delete|add] [1|2]   the faces a Brobot knows
20.   system-value-points [lock|<topic>]   the golden value points: locked? changed? a topic's; lock on purpose
21.   system-interview-rebuild            the interview song, rebuilt from the game's topics and value points
22.   system-anim <text> [1|2|both]     the animationWI tester: {{playAnimationWI||angry}} and words, through Wire-Pod's say route
23.   system-usb                        the USB devices plugged in (live in the menu)
24.   system-voices                     numpad voices: KPN > persona [voice] (change them in the menu)
25.   system-voices-report [3d|week|month|year]  voice files made, played, most played (report only)
26.   system-tpu [restart]              the Edge TPU worker's state, or start it again
27.   system-photos [1|2]               a Brobot's photos
28.   system-cams [tpu]                 the Brobots' cameras and the webcam side by side; tpu = the Edge TPU on the selected one
29.   system-chats <words> [operator|brobots|managers]   search every saved session's chat
30.   system-chat-save [session]        a readable chat.md in the session's folder
31.   system-menu-theme <preset>        the menu's look: dark, light, blue, green, purple
32.   system-visual [next|prev|<id>]    the TPU +/- rotation: a feed's frame and its Edge TPU objects (KP+ / KP-)
33.   system-version                    Wire-Pod's commit, the newest, what an update touches
34.   system-native [on|off]            Wire-Pod's own voice answers (was the Native? box)
35.   system-numpad on|off              the numpad keys as push-to-talk (was the Numpad? box)
36.   system-numpad-layout [usb|pc]     the numpad's key layout - USB numpad or PC keyboard (then F5)
37.   system-set <song> <step> <field> <value>   change one step (volume, text, timing)
38.   system-set <song> [step]          a song's steps, or one step
39.   system-set-list                   every setting, song and step field
40.   system-locate <thing>             where it lives, file:line (a setting, a song step, a word)
41.   system-locate-list                every setting's and song's file:line
42.   system-undo                       put back the last system-set
43.   system-status                     Wire-Pod, language model, page, run, Brobots' battery
44.   system-find <word>                where a word shows up: the log (6 hours), settings, songs
45.   system-search <words>             GOPOD's own web search, the results (to the Brobots: "search <words>")
46.   system-check                      every file the settings, chapters and songs name - all there?
47.   system-voices-rebuild [go]        check, back up every voice, then wpr 2 + F5 makes them all again
48.   system-test-list                  the tests: system-test-<name>, several in a row, system-test-all (the audit)
49.   system-test-game                  every topic, value point and Easter egg found by its own words (no robots, no model)
50.   system-record on|off              Record Session on or off (the page's Record? box)
51.   system-brain                      the brain chain: every answer's record (who asked, who answered, the line, its kind), and the session so far
52.   system-robots-off on|off          robots off: nothing reaches the Brobots; their says play on the page in their numpad voices, says and animations captured for a replay
53.   system-run-check off|on|replay    run check: off = sound off (robots off, page muted, the chat drawn as ever); on = sound back, its chat lines saved; replay plays them at normal volume
54.   system-docbs [<pool> [add|remove <line>]]   DocBS's lines - the pools, one pool's lines, or add / remove one
55.   system-probers-make               the probe maker: a prober for every value point, checked - probers_made_001.txt, to mic-record for the demo
56.   system-tour                       the system window tour: Goverlord goes through the Scrolls, scripture by scripture; safe commandments run (again: stop)
57.   system-scrolls                      the GOPOD System Scrolls (GOPOD_SYSTEM_SCROLLS.md, the system's master reference) written again from the system as it is now
58.   system-comm-check                 the pre-show comm check: the whole numpad held in turn (KP0 the Butler, KP1-KP9, * All), live - who they are, about vs all about, each other, the interns
59.   system-test-admin-run             the 100% admin run self-check: every line in order - points not revealed, eggs not found, prizes not given, dry reminders
60.   system-test-replay                the chat page plays every egg kind, both prizes and the Operator voices (no robots)
61.   system-help                       this list
