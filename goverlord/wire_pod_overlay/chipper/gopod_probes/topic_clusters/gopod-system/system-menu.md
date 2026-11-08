# The system menu (the Torah)
- **words:** system menu, torah, commandments, scriptures, shell window
- **ask:** What's in the system menu?
- **chain:** Goverlord - the shell window's menu: its scriptures (groups), their tours and commands
- **see also:** system-help (the same commands, as a list), system-texts (what each answers)
- **about:** 2026-11-03 (step 7 of GOPOD_CLUSTERS_PLAN_002.md, the format step; operator: "then a shell window scaffolding loads accordingly"): 06_system.json system_menu.groups. Each scripture is a section; its tour is a line list ("[]", a line each); each command is its own section, headed by the command (about, and run: true to run at once, or fill: the words left in the text box). The window's look, themes and copy buttons stay in the settings (the design). A value in backticks is that text exactly.
---

## 1. Restart and check (system_menu.groups.0)
1. title: 1. Restart and check
2. guide: Just started, or made edits? Restart, then check everything is up.
3. next: All PASS? Go to 2 to play, or 3 to edit.

## 1. Restart and check - tour (system_menu.groups.0.tour[])
1. Just started, or made some edits? This is where every show begins. Say “system-restart” for a fresh start, just like wpr 2: the speaker and mic get set, and the page reloads itself. While the reporters warm up, “system-skip”, or the Esc key, skips their lines.
2. Then make sure everything is up. “system-status” shows Wire-Pod, the language model, the page, any run and the batteries. “system-check” makes sure every file is where it should be, and “system-test-all” runs the full audit, PASS or FAIL for each part.
3. All PASS? Go to Scripture 2 to play, or Scripture 3 to edit.

## system-restart (system_menu.groups.0.items.0)
1. cmd: system-restart
2. about: a forced restart, like wpr 2 - speaker and mic set; the page reloads itself (Enter to confirm)
3. fill: system-restart

## system-skip (system_menu.groups.0.items.1)
1. cmd: system-skip
2. about: skip the voiced lines playing - in the tour, on to the next scripture (Esc exits the run)
3. page: skip

## system-status (system_menu.groups.0.items.2)
1. cmd: system-status
2. about: Wire-Pod, language model, page, run, battery
3. run: true

## system-test-all (system_menu.groups.0.items.3)
1. cmd: system-test-all
2. about: the full audit, PASS / FAIL for each
3. run: true

## system-check (system_menu.groups.0.items.4)
1. cmd: system-check
2. about: every file named is there; story.md / knobs match
3. run: true

## 2. Play (system_menu.groups.1)
1. title: 2. Play
2. guide: Run a song or a saved session. One at a time.
3. runs: true
4. next: Something to change? Go to 3.

## 2. Play - tour (system_menu.groups.1.tour[])
1. You can run one song or one saved session at a time. To start a specific song, say “start bingo test”, “start baby brobots”, or “start fart joke”.
2. You can also play any song (or just part of it) with “system-play” followed by the song name. Optional details include the starting step, ending step, which robot should perform it (1 or 2), and whether to do a dry run.
3. Available steps are listed under “system-score”. To replay a previously saved session, say “replay” then the session name. To stop whatever is currently playing, say “stop run.” If you want to change any of this, go to Scripture 3.

## system-play (system_menu.groups.1.items.0)
1. cmd: system-play
2. about: a song or part of it: <song> [from <step>] [to <step>] [robot 1|2] [dry] (steps: system-score)

## replay (system_menu.groups.1.items.1)
1. cmd: replay
2. about: a saved session - add its name

## stop run (system_menu.groups.1.items.2)
1. cmd: stop run
2. about: stops the song playing
3. run: true

## 3. Edit words and settings (system_menu.groups.2)
1. title: 3. Edit words and settings
2. guide: Change what's said and how. system-undo puts back the last edit.
3. next: Don't know the name or path? Go to 4. Done editing? Go to 5.

## 3. Edit words and settings - tour (system_menu.groups.2.tour[])
1. This is where you change what's said and how. “system-topic” edits a chapter topic: its words, points and pictures. “system-lingo” changes how a word is said. “system-set” changes any setting by its path, or a song step's words, volume or timing, and “system-add” and “system-remove” put items into a list or take them out.
2. The show's character lives here too: the roast level and the Jekyll and Hyde switch, how often Brobot 1 goes full Doc BS, his Doc BS lines with “system-docbs”, the parts of a live answer chain, the run cards, one question at a time, and which robot is Brobot 1. “system-voices” picks a voice for each numpad key, and this window's own look and theme can change as well.
3. Made a mistake? “system-undo” puts back the last edit. Don't know a name or path? Go to Scripture 4. Done editing? Go to Scripture 5.

## system-topic (system_menu.groups.2.items.0)
1. cmd: system-topic
2. about: a chapter topic: its words, points, label, reminder, pictures (alone: the topics)

## system-lingo (system_menu.groups.2.items.1)
1. cmd: system-lingo
2. about: how a word is said: brobots / reporters / butler <word> = <said>

## system-set (system_menu.groups.2.items.2)
1. cmd: system-set
2. about: a setting by its path, or a song step's text / volume / timing

## system-add (system_menu.groups.2.items.3)
1. cmd: system-add
2. about: an item into a list (or key = value into a group)

## system-remove (system_menu.groups.2.items.4)
1. cmd: system-remove
2. about: an item out of a list (or a key out of a group)

## system-set chat_moves.thinking_text (system_menu.groups.2.items.5)
1. cmd: system-set chat_moves.thinking_text
2. about: the "Got it!" line: said while the model works, and the last line of the Value Points notice (then system-restart)
3. fill: `system-set chat_moves.thinking_text `

## system-set bubble_notices (system_menu.groups.2.items.6)
1. cmd: system-set bubble_notices
2. about: what goes in a notice in a Brobot's bubble: value_points, egg, native_intent
3. fill: system-set bubble_notices.

## system-set <song> tempo (system_menu.groups.2.items.7)
1. cmd: system-set <song> tempo
2. about: a song's tempo in seconds (a step's: system-set <song> <step> tempo_factor <n>)
3. fill: `system-set `

## system-set system_menu.look (system_menu.groups.2.items.8)
1. cmd: system-set system_menu.look
2. about: this menu's own look: font_size, background, code_background, code_color, results_background
3. fill: system-set system_menu.look.

## system-menu-theme (system_menu.groups.2.items.9)
1. cmd: system-menu-theme
2. about: `dark | light | blue | green | purple (then F5)`

## system-voices-report (system_menu.groups.2.items.10)
1. cmd: system-voices-report
2. about: `voice files: made, played, most played, not played - past 3d | week | month | year (report only, nothing moves)`
3. fill: system-voices-report week

## system-voices (system_menu.groups.2.items.11)
1. cmd: system-voices
2. about: numpad voices: KPN > persona [voice] - pick a voice per key (bold = used twice); robot-bodied keys need none
3. page: voices

## system-set-list (system_menu.groups.2.items.12)
1. cmd: system-set-list
2. about: every setting, song and step field
3. run: true

## system-undo (system_menu.groups.2.items.13)
1. cmd: system-undo
2. about: puts back the last edit
3. run: true

## system-set roast_dial.level (system_menu.groups.2.items.14)
1. cmd: system-set roast_dial.level
2. about: the roast level: 0 clean (never punches down), 1 roast, 2 public 18+, 3 paid 18+ afterburner - then system-restart
3. fill: `system-set roast_dial.level `

## system-set roast_dial.jekyll_hyde (system_menu.groups.2.items.15)
1. cmd: system-set roast_dial.jekyll_hyde
2. about: Jekyll/Hyde: true lets the roast through on sensitive topics too - then system-restart
3. fill: `system-set roast_dial.jekyll_hyde `

## system-set docbs.share (system_menu.groups.2.items.16)
1. cmd: system-set docbs.share
2. about: DocBS: how many of Brobot 1's answers go equation roast, 0-100 (-1: his Doc dial) - then system-restart
3. fill: `system-set docbs.share `

## system-docbs (system_menu.groups.2.items.17)
1. cmd: system-docbs
2. about: DocBS's lines: the pools, one pool's lines (system-docbs openers_clean), or add / remove one
3. run: true

## system-set answer_chains.brobots.parts (system_menu.groups.2.items.18)
1. cmd: system-set answer_chains.brobots.parts
2. about: a live answer's chain - its parts in order (system-add / system-remove one part)
3. fill: `system-set answer_chains.brobots.parts `

## system-set game.runs (system_menu.groups.2.items.19)
1. cmd: system-set game.runs
2. about: a run card - its words, song, lines and pre/post chat (system-set game.runs.1.experience.closing ...)
3. fill: system-set game.runs.

## system-set line_queue.on (system_menu.groups.2.items.20)
1. cmd: system-set line_queue.on
2. about: one question at a time: true / false - then system-restart
3. fill: `system-set line_queue.on `

## system-set brobots_setup.brobot_1 (system_menu.groups.2.items.21)
1. cmd: system-set brobots_setup.brobot_1
2. about: which robot is Brobot 1 (an ESN; empty: Wire-Pod's robots decide) - then system-restart
3. fill: `system-set brobots_setup.brobot_1 `

## 4. Find (system_menu.groups.3)
1. title: 4. Find
2. guide: Don't know where something lives? Find it first.
3. next: Found it? Back to 3 to change it.

## 4. Find - tour (system_menu.groups.3.tour[])
1. Don't know where something lives? Find it first. “system-locate” tells you where a setting, a song step or a word lives, and “system-find” shows everywhere a word shows up. “system-score” prints a song's score, every step and its words.
2. “system-search” is GOPOD's own web search, “system-chats” searches every saved session's chat, and “system-chat-save” writes the newest session as a readable chat.
3. Found it? Go back to Scripture 3 to change it.

## system-locate (system_menu.groups.3.items.0)
1. cmd: system-locate
2. about: where a setting, song step or word lives, file:line

## system-find (system_menu.groups.3.items.1)
1. cmd: system-find
2. about: where a word shows up: the log, settings, songs

## system-score (system_menu.groups.3.items.2)
1. cmd: system-score
2. about: a song's score: every step and its words (alone: the songs)

## system-locate-list (system_menu.groups.3.items.3)
1. cmd: system-locate-list
2. about: every setting's and song's file:line
3. run: true

## system-search (system_menu.groups.3.items.4)
1. cmd: system-search
2. about: GOPOD's own web search

## system-chats (system_menu.groups.3.items.5)
1. cmd: system-chats
2. about: <words> [operator|brobots|managers] - search every saved session's chat

## system-chat-save (system_menu.groups.3.items.6)
1. cmd: system-chat-save
2. about: a readable chat.md of the newest session (or name one)
3. run: true

## 5. Apply (system_menu.groups.4)
1. title: 5. Apply
2. guide: Edits take effect after a restart (group 1); edited story lines get new voices.
3. next: Then restart: Go to 1.

## 5. Apply - tour (system_menu.groups.4.tour[])
1. Edits take effect after a restart, from Scripture 1. Edited story lines get their new voices made then.
2. “system-voices-rebuild” checks and backs up every voice, and the restart after it makes them all again. Then restart: go to Scripture 1.

## system-voices-rebuild (system_menu.groups.4.items.0)
1. cmd: system-voices-rebuild
2. about: check, back up every voice, then restart makes them all again
3. fill: system-voices-rebuild

## 6. Switches (system_menu.groups.5)
1. title: 6. Switches
2. guide: Turn parts on or off for this session.
3. next: Close the menu with the speaker icon.

## 6. Switches - tour (system_menu.groups.5.tour[])
1. Turn parts on or off for this session. “system-numpad” shows who each numpad key talks to; “on” makes the keys push-to-talk, “off” lets them type as usual. “system-gap” sets the reporter gap for system-play.
2. “system-tweaks on” plays the moves with their tuned values, “system-native” turns Wire-Pod's own voice answers on or off, and “system-record” turns Record Session on or off. When you're done, close the window with the speaker icon.

## system-numpad (system_menu.groups.5.items.0)
1. cmd: system-numpad
2. about: each numpad key: who it talks to, live or not
3. run: true

## system-numpad on (system_menu.groups.5.items.1)
1. cmd: system-numpad on
2. about: numpad keys are push-to-talk
3. run: true

## system-numpad off (system_menu.groups.5.items.2)
1. cmd: system-numpad off
2. about: numpad keys type as usual
3. run: true

## system-gap (system_menu.groups.5.items.3)
1. cmd: system-gap
2. about: `reporter gap for system-play: <seconds> | off | default`

## system-tweaks on (system_menu.groups.5.items.4)
1. cmd: system-tweaks on
2. about: apply phcal tweaks before a live system-play
3. run: true

## system-tweaks off (system_menu.groups.5.items.5)
1. cmd: system-tweaks off
2. about: play the moves as written
3. run: true

## system-native on (system_menu.groups.5.items.6)
1. cmd: system-native on
2. about: Wire-Pod's own voice answers on
3. run: true

## system-native off (system_menu.groups.5.items.7)
1. cmd: system-native off
2. about: Wire-Pod's own voice answers off
3. run: true

## system-record on (system_menu.groups.5.items.8)
1. cmd: system-record on
2. about: Record Session on (the page's Record? box)
3. run: true

## system-record off (system_menu.groups.5.items.9)
1. cmd: system-record off
2. about: Record Session off, kept zipped
3. run: true

## system-robots-off on (system_menu.groups.5.items.10)
1. cmd: system-robots-off on
2. about: Robots off: the Brobots' says on the page in their numpad voices, says and animations captured
3. run: false

## system-robots-off off (system_menu.groups.5.items.11)
1. cmd: system-robots-off off
2. about: Robots back on
3. run: false

## system-run-check on (system_menu.groups.5.items.12)
1. cmd: system-run-check on
2. about: a silent run-check: robots off and the page muted, the chat drawn as ever; start any run
3. run: false

## system-run-check off (system_menu.groups.5.items.13)
1. cmd: system-run-check off
2. about: ends the run check and saves its chat lines
3. run: false

## system-run-check replay (system_menu.groups.5.items.14)
1. cmd: system-run-check replay
2. about: plays the saved run check again on the page, at normal volume
3. run: false

## 7. Brobots (system_menu.groups.6)
1. title: 7. Brobots
2. guide: Wire-Pod's bot page, here: add 1 or 2 at the end for one Brobot (both when left out).
3. next: Done? Close the menu with the speaker icon.

## 7. Brobots - tour (system_menu.groups.6.tour[])
1. Wire-Pod's bot page, right here. Add 1 or 2 at the end for one Brobot, or leave it out for both. “system-visual” steps through the camera feeds and what the Edge TPU sees, “system-cams” shows every feed at once, and “system-usb” lists what's plugged in.
2. “system-brobot” does everything else: their stats, volume, eye colour, units, place and time zone, their cube tricks, sleep beats and animations, a line said, a wake, or a small nudge of the head, lift or wheels. “system-faces” and “system-photos” show the faces they know and the photos they took. Done? Close the window with the speaker icon.

## system-visual next (system_menu.groups.6.items.0)
1. cmd: system-visual next
2. about: the TPU +/- rotation: next feed, photo, enhancement, TPU boxes and "I see..." (KP+)
3. run: true

## system-visual prev (system_menu.groups.6.items.1)
1. cmd: system-visual prev
2. about: previous feed (KP-)
3. run: true

## system-cams (system_menu.groups.6.items.2)
1. cmd: system-cams
2. about: every feed's photo in a row, then the TPU boxes on the selected one
3. run: true

## system-tpu (system_menu.groups.6.items.3)
1. cmd: system-tpu
2. about: the Edge TPU worker: ready or not
3. run: true

## system-tpu restart (system_menu.groups.6.items.4)
1. cmd: system-tpu restart
2. about: start the Edge TPU worker again
3. run: true

## system-usb (system_menu.groups.6.items.5)
1. cmd: system-usb
2. about: the USB devices plugged in, live (hot-swap)
3. page: usb

## system-brobot stats (system_menu.groups.6.items.6)
1. cmd: system-brobot stats
2. about: both Brobots' statistics
3. run: true

## system-version (system_menu.groups.6.items.7)
1. cmd: system-version
2. about: Wire-Pod's commit, the newest, what an update touches
3. run: true

## system-brobot volume (system_menu.groups.6.items.8)
1. cmd: system-brobot volume
2. about: <0-5>

## system-brobot eyes (system_menu.groups.6.items.9)
1. cmd: system-brobot eyes
2. about: <teal|orange|yellow|lime|sapphire|purple|green>

## system-brobot locale (system_menu.groups.6.items.10)
1. cmd: system-brobot locale
2. about: <en-US ...>

## system-brobot units (system_menu.groups.6.items.11)
1. cmd: system-brobot units
2. about: `c | f`

## system-brobot location (system_menu.groups.6.items.12)
1. cmd: system-brobot location
2. about: <place>

## system-brobot timezone (system_menu.groups.6.items.13)
1. cmd: system-brobot timezone
2. about: <America/Toronto ...>

## system-faces (system_menu.groups.6.items.14)
1. cmd: system-faces
2. about: `the faces a Brobot knows (list | rename <old> <new> | delete <name> | add <name>)`
3. fill: `system-faces list `

## system-photos (system_menu.groups.6.items.15)
1. cmd: system-photos
2. about: a Brobot's photos
3. fill: `system-photos `

## system-brobot cube (system_menu.groups.6.items.16)
1. cmd: system-brobot cube
2. about: <pickupcube|fetchcube|rollcube|popawheelie|keepaway|findcube> - his own cube moves

## system-brobot beats (system_menu.groups.6.items.17)
1. cmd: system-brobot beats
2. about: the 126 sleep beats by family: sleep, rts-asleep, held-on-palm, petting
3. run: true

## system-brobot beat (system_menu.groups.6.items.18)
1. cmd: system-brobot beat
2. about: <alias> - plays one sleep beat (snores, warbles, sleep moves...)

## system-brobot anims sleep (system_menu.groups.6.items.19)
1. cmd: system-brobot anims sleep
2. about: every sleep animation on the robot (any word: anims happy)
3. fill: `system-brobot anims sleep `

## system-brobot anim (system_menu.groups.6.items.20)
1. cmd: system-brobot anim
2. about: <name> - plays one (names from anims)

## system-brobot say (system_menu.groups.6.items.21)
1. cmd: system-brobot say
2. about: <text> - said by the Brobot

## system-brobot wake (system_menu.groups.6.items.22)
1. cmd: system-brobot wake
2. about: "Hey Vector" without the words

## system-brobot head (system_menu.groups.6.items.23)
1. cmd: system-brobot head
2. about: <speed, e.g. 3 or -3> - a short nudge

## system-brobot lift (system_menu.groups.6.items.24)
1. cmd: system-brobot lift
2. about: <speed>

## system-brobot wheels (system_menu.groups.6.items.25)
1. cmd: system-brobot wheels
2. about: <left> <right>

## system-brobot camera (system_menu.groups.6.items.26)
1. cmd: system-brobot camera
2. about: live video from the Brobot's camera
3. run: true

## 8. Tune moves (system_menu.groups.7)
1. title: 8. Tune moves
2. guide: A move looks off? Try it alone on one Brobot (tweaks on: group 6), then play it in its song.
3. next: Play that part: system-play <song> from <step> to <step> (group 2).

## 8. Tune moves - tour (system_menu.groups.7.tour[])
1. A move looks off? Try it alone on one Brobot first. “system-calibrate” plays one move, an arm, a nod, the rattle or a cube light, with its speed, hold and cycles. Turn tweaks on in Scripture 6 to keep the tuned values.
2. “system-anim” tests an animation token with words. Then play that part in its song with system-play, from Scripture 2.

## system-calibrate (system_menu.groups.7.items.0)
1. cmd: system-calibrate
2. about: <arm|nod|rattle|danger|cube> <1|2> [reps] [--hold S] [--speed N] [--cycles N] [--volume N] [dry]

## system-anim list (system_menu.groups.7.items.1)
1. cmd: system-anim list
2. about: the animation tokens
3. run: true

## system-anim (system_menu.groups.7.items.2)
1. cmd: system-anim
2. about: the animationWI tester (WIP): a token and words, e.g. angry with a count
3. fill: system-anim {{playAnimationWI||angry}} wait 1, 2, 3, 4, 5, 6, 7, 8, 9, 0

## 9. Test one thing (system_menu.groups.8)
1. title: 9. Test one thing
2. guide: Something acting up? Test just that part.
3. next: A FAIL says what it saw - fix it in 3, then 5.

## 9. Test one thing - tour (system_menu.groups.8.tour[])
1. Something acting up? Test just that part. “system-test-list” lists the tests, and “system-test” followed by a name runs one. “system-test-game” finds every topic, value point and Easter egg by its own words, and “system-test-admin-run” plays the full admin run, every line in order.
2. “system-comm-check” holds every numpad key in turn before the show, live. “system-tour” is this tour, “system-torah” writes the Torah again, and “system-audit” runs everything with no robots called. A FAIL says what it saw: fix it in Scripture 3, then Scripture 5.

## system-test-list (system_menu.groups.8.items.0)
1. cmd: system-test-list
2. about: the tests, then system-test-<name>
3. run: true

## system-test-game (system_menu.groups.8.items.1)
1. cmd: system-test-game
2. about: every topic, value point and Easter egg found by its own words - no robots, no model
3. run: true

## system-tour (system_menu.groups.8.items.2)
1. cmd: system-tour
2. about: the system window tour: Goverlord goes through the Torah, scripture by scripture; safe commandments run (again: stop)
3. run: true

## system-torah (system_menu.groups.8.items.3)
1. cmd: system-torah
2. about: the GOPOD Torah (GOPOD_TORAH.md, the master system reference) written again from the system as it is now
3. run: true

## system-comm-check (system_menu.groups.8.items.4)
1. cmd: system-comm-check
2. about: the pre-show comm check: the whole numpad held in turn (KP0 the Butler, KP1-KP9, * All), live - who they are, about vs all about, each other, the interns
3. run: true

## system-test-admin-run (system_menu.groups.8.items.5)
1. cmd: system-test-admin-run
2. about: the 100% admin run self-check: every line in order - points not revealed, eggs not found, prizes not given, dry reminders
3. run: true

## system-audit (system_menu.groups.8.items.6)
1. cmd: system-audit
2. about: the full audit, no robots called - files, Go tests, web page, phone app, every test, the log (a few minutes)
3. run: true

## system-test-replay (system_menu.groups.8.items.7)
1. cmd: system-test-replay
2. about: the chat page plays every egg kind, both prizes and the Operator voices - no robots (run it after any page change)
3. run: true

## system-test- (system_menu.groups.8.items.8)
1. cmd: system-test-
2. about: one test - add its name (weather-live, qr, voices...)
3. fill: system-test-

## 10. Wrap up (system_menu.groups.9)
1. title: 10. Wrap up
2. tour: End the show. Say “thanks Brobots” for the ending, or “stop session” to end the session. “system-help” lists every command, a line each. |
3. guide: End the show.

## thanks brobots (system_menu.groups.9.items.0)
1. cmd: thanks brobots
2. about: the ending
3. fill: thanks brobots

## stop session (system_menu.groups.9.items.1)
1. cmd: stop session
2. about: ends the session
3. fill: stop session

## system-help (system_menu.groups.9.items.2)
1. cmd: system-help
2. about: every command, a line each
3. run: true
