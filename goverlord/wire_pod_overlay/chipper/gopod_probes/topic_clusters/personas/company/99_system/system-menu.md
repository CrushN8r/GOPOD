# The system menu (the Scrolls)
- **words:** system menu, scrolls, commandments, scriptures, shell window
- **ask:** What's in the system menu?
- **chain:** Goverlord - the shell window's menu: its scriptures (groups), their tours and commands
- **shape:** each scripture - its head, its tour, its commandments: one line each, “cmd, about”, then “ · key: value” for each more field (like the alias cards)
- **see also:** system-help (the same commands, as a list), system-texts (what each answers)
---

## 1. Restart and check (system_menu.groups.0)
1. title: 1. Restart and check
2. guide: Just started, or made edits? Restart, then check everything is up.
3. next: All PASS? Go to 2 to play, or 3 to edit.

## 1. Restart and check - tour (system_menu.groups.0.tour[])
1. Just started, or made some edits? This is where every show begins. Say “system-restart” for a fresh start, just like wpr 2: the speaker and mic get set, and the page reloads itself. While the reporters warm up, “system-skip”, or the Esc key, skips their lines.
2. Then make sure everything is up. “system-status” shows Wire-Pod, the language model, the page, any run and the batteries. “system-check” makes sure every file is where it should be, and “system-test-all” runs the full audit, PASS or FAIL for each part.
3. All PASS? Go to Scripture 2 to play, or Scripture 3 to edit.

## 1. Restart and check - commandments (system_menu.groups.0.items[])
1. system-restart, a forced restart, like wpr 2 - speaker and mic set; the page reloads itself (Enter to confirm) · fill: system-restart
2. system-skip, skip the voiced lines playing - in the tour, on to the next scripture (Esc exits the run) · page: skip
3. system-status, Wire-Pod, language model, page, run, battery · run: true
4. system-test-all, the full audit, PASS / FAIL for each · run: true
5. system-check, every file named is there; story.md / knobs match · run: true

## 2. Play (system_menu.groups.1)
1. title: 2. Play
2. guide: Run a song or a saved session. One at a time.
3. runs: true
4. next: Something to change? Go to 3.

## 2. Play - tour (system_menu.groups.1.tour[])
1. You can run one song or one saved session at a time. To start a specific song, say “start bingo test”, “start baby brobots”, or “start fart joke”.
2. You can also play any song (or just part of it) with “system-play” followed by the song name. Optional details include the starting step, ending step, which robot should perform it (1 or 2), and whether to do a dry run.
3. Available steps are listed under “system-score”. To replay a previously saved session, say “replay” then the session name. To stop whatever is currently playing, say “stop run.” If you want to change any of this, go to Scripture 3.

## 2. Play - commandments (system_menu.groups.1.items[])
1. system-play, a song or part of it: <song> [from <step>] [to <step>] [robot 1|2] [dry] (steps: system-score)
2. replay, a saved session - add its name
3. stop run, stops the song playing · run: true

## 3. Edit words and settings (system_menu.groups.2)
1. title: 3. Edit words and settings
2. guide: Change what's said and how. system-undo puts back the last edit.
3. next: Don't know the name or path? Go to 4. Done editing? Go to 5.

## 3. Edit words and settings - tour (system_menu.groups.2.tour[])
1. This is where you change what's said and how. “system-topic” edits a chapter topic: its words, points and pictures. “system-lingo” changes how a word is said, and “system-lingo-improver” opens the lingo improver: pick a voice, type a line, hear it, and keep the spelling that sounds right. “system-set” changes any setting by its path, or a song step's words, volume or timing, and “system-add” and “system-remove” put items into a list or take them out.
2. The show's character lives here too: the roast level and the Jekyll and Hyde switch, how often Brobot 1 goes full Doc BS, his Doc BS lines with “system-docbs”, the parts of a live answer chain, the run cards, one question at a time, and which robot is Brobot 1. “system-voices” picks a voice for each numpad key, and this window's own look and theme can change as well.
3. Made a mistake? “system-undo” puts back the last edit. Don't know a name or path? Go to Scripture 4. Done editing? Go to Scripture 5.

## 3. Edit words and settings - commandments (system_menu.groups.2.items[])
1. system-topic, a chapter topic: its words, points, label, reminder, pictures (alone: the topics)
2. system-lingo, how a word is said: brobots / reporters / butler <word> = <said>
3. system-set, a setting by its path, or a song step's text / volume / timing
4. system-add, an item into a list (or key = value into a group)
5. system-remove, an item out of a list (or a key out of a group)
6. system-set chat_moves.thinking_text, the "Got it!" line: said while the model works, and the last line of the Value Points notice (then system-restart) · fill: `system-set chat_moves.thinking_text `
7. system-set bubble_notices, what goes in a notice in a Brobot's bubble: value_points, egg, native_intent · fill: system-set bubble_notices.
8. system-set <song> tempo, a song's tempo in seconds (a step's: system-set <song> <step> tempo_factor <n>) · fill: `system-set `
9. system-set system_menu.look, this menu's own look: font_size, background, code_background, code_color, results_background · fill: system-set system_menu.look.
10. system-menu-theme, `dark | light | blue | green | purple (then F5)`
11. system-voices-report, `voice files: made, played, most played, not played - past 3d | week | month | year (report only, nothing moves)` · fill: system-voices-report week
12. system-voices, numpad voices: KPN > persona [voice] - pick a voice per key (bold = used twice); robot-bodied keys need none · page: voices
13. system-set-list, every setting, song and step field · run: true
14. system-undo, puts back the last edit · run: true
15. system-set roast_dial.level, the roast level: 0 clean (never punches down), 1 roast, 2 public 18+, 3 paid 18+ afterburner - then system-restart · fill: `system-set roast_dial.level `
16. system-set roast_dial.jekyll_hyde, Jekyll/Hyde: true lets the roast through on sensitive topics too - then system-restart · fill: `system-set roast_dial.jekyll_hyde `
17. system-set docbs.share, DocBS: how many of Brobot 1's answers go equation roast, 0-100 (-1: his Doc dial) - then system-restart · fill: `system-set docbs.share `
18. system-docbs, DocBS's lines: the pools, one pool's lines (system-docbs openers_clean), or add / remove one · run: true
19. system-set answer_chains.brobots.parts, a live answer's chain - its parts in order (system-add / system-remove one part) · fill: `system-set answer_chains.brobots.parts `
20. system-set game.runs, a run card - its words, song, lines and pre/post chat (system-set game.runs.1.experience.closing ...) · fill: system-set game.runs.
21. system-set line_queue.on, one question at a time: true / false - then system-restart · fill: `system-set line_queue.on `
22. system-set brobots_setup.brobot_1, which robot is Brobot 1 (an ESN; empty: Wire-Pod's robots decide) - then system-restart · fill: `system-set brobots_setup.brobot_1 `
23. system-lingo-improver, the lingo improver: pick a voice, type a line, hear it, keep the spelling that sounds right

## 4. Find (system_menu.groups.3)
1. title: 4. Find
2. guide: Don't know where something lives? Find it first.
3. next: Found it? Back to 3 to change it.

## 4. Find - tour (system_menu.groups.3.tour[])
1. Don't know where something lives? Find it first. “system-locate” tells you where a setting, a song step or a word lives, and “system-find” shows everywhere a word shows up. “system-score” prints a song's score, every step and its words.
2. “system-search” is GOPOD's own web search, “system-chats” searches every saved session's chat, and “system-chat-save” writes the newest session as a readable chat.
3. Found it? Go back to Scripture 3 to change it.

## 4. Find - commandments (system_menu.groups.3.items[])
1. system-locate, where a setting, song step or word lives, file:line
2. system-find, where a word shows up: the log, settings, songs
3. system-score, a song's score: every step and its words (alone: the songs)
4. system-locate-list, every setting's and song's file:line · run: true
5. system-search, GOPOD's own web search
6. system-chats, <words> [operator|brobots|managers] - search every saved session's chat
7. system-chat-save, a readable chat.md of the newest session (or name one) · run: true

## 5. Apply (system_menu.groups.4)
1. title: 5. Apply
2. guide: Edits take effect after a restart (group 1); edited story lines get new voices.
3. next: Then restart: Go to 1.

## 5. Apply - tour (system_menu.groups.4.tour[])
1. Edits take effect after a restart, from Scripture 1. Edited story lines get their new voices made then.
2. “system-voices-rebuild” checks and backs up every voice, and the restart after it makes them all again. Then restart: go to Scripture 1.

## 5. Apply - commandments (system_menu.groups.4.items[])
1. system-voices-rebuild, check, back up every voice, then restart makes them all again · fill: system-voices-rebuild

## 6. Switches (system_menu.groups.5)
1. title: 6. Switches
2. guide: Turn parts on or off for this session.
3. next: Close the menu with the speaker icon.

## 6. Switches - tour (system_menu.groups.5.tour[])
1. Turn parts on or off for this session. “system-numpad” shows who each numpad key talks to; “on” makes the keys push-to-talk, “off” lets them type as usual. “system-gap” sets the reporter gap for system-play.
2. “system-tweaks on” plays the moves with their tuned values, “system-native” turns Wire-Pod's own voice answers on or off, and “system-record” turns Record Session on or off. When you're done, close the window with the speaker icon.

## 6. Switches - commandments (system_menu.groups.5.items[])
1. system-numpad, each numpad key: who it talks to, live or not · run: true
2. system-numpad on, numpad keys are push-to-talk · run: true
3. system-numpad off, numpad keys type as usual · run: true
4. system-gap, `reporter gap for system-play: <seconds> | off | default`
5. system-tweaks on, apply phcal tweaks before a live system-play · run: true
6. system-tweaks off, play the moves as written · run: true
7. system-native on, Wire-Pod's own voice answers on · run: true
8. system-native off, Wire-Pod's own voice answers off · run: true
9. system-record on, Record Session on (the page's Record? box) · run: true
10. system-record off, Record Session off, kept zipped · run: true
11. system-robots-off on, Robots off: the Brobots' says on the page in their numpad voices, says and animations captured · run: false
12. system-robots-off off, Robots back on · run: false
13. system-run-check off, sound off for a run check: robots off and the page muted, the chat drawn as ever; start any run · run: false
14. system-run-check on, sound back on; the run check's chat lines saved · run: false
15. system-run-check replay, plays the saved run check again on the page, at normal volume · run: false

## 7. Brobots (system_menu.groups.6)
1. title: 7. Brobots
2. guide: Wire-Pod's bot page, here: add 1 or 2 at the end for one Brobot (both when left out).
3. next: Done? Close the menu with the speaker icon.

## 7. Brobots - tour (system_menu.groups.6.tour[])
1. Wire-Pod's bot page, right here. Add 1 or 2 at the end for one Brobot, or leave it out for both. “system-visual” steps through the camera feeds and what the Edge TPU sees, “system-cams” shows every feed at once, and “system-usb” lists what's plugged in.
2. “system-brobot” does everything else: their stats, volume, eye colour, units, place and time zone, their cube tricks, sleep beats and animations, a line said, a wake, or a small nudge of the head, lift or wheels. “system-faces” and “system-photos” show the faces they know and the photos they took. Done? Close the window with the speaker icon.

## 7. Brobots - commandments (system_menu.groups.6.items[])
1. system-visual next, the TPU +/- rotation: next feed, photo, enhancement, TPU boxes and "I see..." (KP+) · run: true
2. system-visual prev, previous feed (KP-) · run: true
3. system-cams, every feed's photo in a row, then the TPU boxes on the selected one · run: true
4. system-tpu, the Edge TPU worker: ready or not · run: true
5. system-tpu restart, start the Edge TPU worker again · run: true
6. system-usb, the USB devices plugged in, live (hot-swap) · page: usb
7. system-brobot stats, both Brobots' statistics · run: true
8. system-version, Wire-Pod's commit, the newest, what an update touches · run: true
9. system-brobot volume, <0-5>
10. system-brobot eyes, <teal|orange|yellow|lime|sapphire|purple|green>
11. system-brobot locale, <en-US ...>
12. system-brobot units, `c | f`
13. system-brobot location, <place>
14. system-brobot timezone, <America/Toronto ...>
15. system-faces, `the faces a Brobot knows (list | rename <old> <new> | delete <name> | add <name>)` · fill: `system-faces list `
16. system-photos, a Brobot's photos · fill: `system-photos `
17. system-brobot cube, <pickupcube|fetchcube|rollcube|popawheelie|keepaway|findcube> - his own cube moves
18. system-brobot beats, the 126 sleep beats by family: sleep, rts-asleep, held-on-palm, petting · run: true
19. system-brobot beat, <alias> - plays one sleep beat (snores, warbles, sleep moves...)
20. system-brobot anims sleep, every sleep animation on the robot (any word: anims happy) · fill: `system-brobot anims sleep `
21. system-brobot anim, <name> - plays one (names from anims)
22. system-brobot say, <text> - said by the Brobot
23. system-brobot wake, "Hey Vector" without the words
24. system-brobot head, <speed, e.g. 3 or -3> - a short nudge
25. system-brobot lift, <speed>
26. system-brobot wheels, <left> <right>
27. system-brobot camera, live video from the Brobot's camera · run: true
28. system-alias, the terminal's aliases from their cards; ✓ ones run here (system-alias gofps 30) · run: true

## 8. Tune moves (system_menu.groups.7)
1. title: 8. Tune moves
2. guide: A move looks off? Try it alone on one Brobot (tweaks on: group 6), then play it in its song.
3. next: Play that part: system-play <song> from <step> to <step> (group 2).

## 8. Tune moves - tour (system_menu.groups.7.tour[])
1. A move looks off? Try it alone on one Brobot first. “system-calibrate” plays one move, an arm, a nod, the rattle or a cube light, with its speed, hold and cycles. Turn tweaks on in Scripture 6 to keep the tuned values.
2. “system-anim” tests an animation token with words. Then play that part in its song with system-play, from Scripture 2.

## 8. Tune moves - commandments (system_menu.groups.7.items[])
1. system-calibrate, <arm|nod|rattle|danger|cube> <1|2> [reps] [--hold S] [--speed N] [--cycles N] [--volume N] [dry]
2. system-anim list, the animation tokens · run: true
3. system-anim, the animationWI tester (WIP): a token and words, e.g. angry with a count · fill: system-anim {{playAnimationWI||angry}} wait 1, 2, 3, 4, 5, 6, 7, 8, 9, 0

## 9. Test one thing (system_menu.groups.8)
1. title: 9. Test one thing
2. guide: Something acting up? Test just that part.
3. next: A FAIL says what it saw - fix it in 3, then 5.

## 9. Test one thing - tour (system_menu.groups.8.tour[])
1. Something acting up? Test just that part. “system-test-list” lists the tests, and “system-test” followed by a name runs one. “system-test-game” finds every topic, value point and Easter egg by its own words, and “system-test-admin-run” plays the full admin run, every line in order.
2. “system-comm-check” holds every numpad key in turn before the show, live. “system-tour” is this tour, “system-scrolls” writes the Scrolls again, and “system-audit” runs everything with no robots called. A FAIL says what it saw: fix it in Scripture 3, then Scripture 5.

## 9. Test one thing - commandments (system_menu.groups.8.items[])
1. system-test-list, the tests, then system-test-<name> · run: true
2. system-test-game, every topic, value point and Easter egg found by its own words - no robots, no model · run: true
3. system-tour, the system window tour: Goverlord goes through the Scrolls, scripture by scripture; safe commandments run (again: stop) · run: true
4. system-scrolls, the GOPOD System Scrolls (GOPOD_SYSTEM_SCROLLS.md, the master system reference) written again from the system as it is now · run: true
5. system-comm-check, the pre-show comm check: the whole numpad held in turn (KP0 the Butler, KP1-KP9, * All), live - who they are, about vs all about, each other, the interns · run: true
6. system-test-admin-run, the 100% admin run self-check: every line in order - points not revealed, eggs not found, prizes not given, dry reminders · run: true
7. system-audit, the full audit, no robots called - files, Go tests, web page, phone app, every test, the log (a few minutes) · run: true
8. system-test-replay, the chat page plays every egg kind, both prizes and the Operator voices - no robots (run it after any page change) · run: true
9. system-test-, one test - add its name (weather-live, qr, voices...) · fill: system-test-

## 10. Wrap up (system_menu.groups.9)
1. title: 10. Wrap up
2. tour: End the show. Say “thanks Brobots” for the ending, or “stop session” to end the session. “system-help” lists every command, a line each. |
3. guide: End the show.

## 10. Wrap up - commandments (system_menu.groups.9.items[])
1. thanks brobots, the ending · fill: thanks brobots
2. stop session, ends the session · fill: stop session
3. system-help, every command, a line each · run: true

## 11. Local network visitors and the phone app (system_menu.groups.10)
1. title: 11. Local network visitors and the phone app
2. tour: Phones and computers on your local network can join the chat - each shows up under "Local network visitors", with the count. Visitor chat participation starts OFF; turn it on to let visitors chat. Each visitor is listed with its name, like "Phone 2", then its network address, the device and the browser - and a Clear button to remove it. A visitor's first line is its username request: Approve or Clear it. An approved visitor has a box: unticked, it's muted. | The phone app is the same chat, on your phone, as a visitor on your network. Once: in Chrome, open GOPOD's address with /gopod_ca.crt and download it; in your phone's settings, search "CA certificate", Install anyway and pick it; restart Chrome. Then open GOPOD's secure address (port 8443) with ?app - no warning - and in Chrome's menu pick "Install app". Tap the GOPOD icon. On a new network, open the new address once and install again. |
3. guide: Who's on your network, who may chat, and GOPOD on your phone.
