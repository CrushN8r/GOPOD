# GOPOD system texts
- **words:** system text, system messages, system window, terminal bubble
- **ask:** What does the system say back to a system- command?
- **see also:** system-help (the commands these answer), dizzy-brobots, weather_run
- **about:** 2026-11-03 (operator: "the whole GOPOD system can be that, correct? * then a shell window scaffolding loads accordingly * let's do that as a planned group of steps"; step 1 of GOPOD_CLUSTERS_PLAN_001.md). Every line the system window answers with, by command; {name} is filled by the code. A " | " here is words, never a list. Was 06_system.json system_text.
---

## Comm check (system_text)
1. comm_check.system_comm_check_a_run: system-comm-check: a run is going - try again after it
2. comm_check.system_comm_check_already_running: system-comm-check: already running
3. comm_check.system_comm_check_comm_check: system-comm-check: comm_check.rounds is empty

## Interview from game (system_text)
1. interview_from_game.interview_rebuilt_from_the_game: Interview rebuilt from the game: {questions} questions, one per value point, in {core_topics} chapters (system-undo puts the old one back).

## Robot call (system_text)
1. robot_call.didn_t_wake_no_behavior: didn't wake (no behavior control in 8 s - asleep on the charger, or off?)
2. robot_call.not_connected_to_wire_pod: not connected to Wire-Pod

## System edit (system_text)
1. system_edit.brobots_lingo: brobots lingo: {word} = {said}{after_restart}
2. system_edit.brobots_lingo_no_in_it: brobots lingo: no {word} in it.
3. system_edit.brobots_lingo_taken_out: brobots lingo: {word} taken out{after_restart}
4. system_edit.is_a_list_or_a: {dotted} is a list or a group - system-add / system-remove change its items, or system-set {dotted}.<key> one value.
5. system_edit.lingo: {lingo} lingo: {word} = {said}{after_restart}
6. system_edit.lingo_taken_out: {lingo} lingo: {word} taken out{after_restart}
7. system_edit.native_voice_answers_wire_pod: Native voice answers (Wire-Pod's own: "Hey Vector, what time is it" and the rest): {state} - system-native on|off
8. system_edit.no_setting: no setting {dotted}.
9. system_edit.no_setting_system_locate_word: no setting {dotted} (system-locate <word> finds one).
10. system_edit.no_topic_system_topic_lists: no topic {arg1} - system-topic lists them.
11. system_edit.not_changed: {dotted}: not changed - {error}
12. system_edit.not_changed_2: {arg0} {dotted}: not changed - {error}
13. system_edit.not_restarted_wire_pod_wouldn: Not restarted - Wire-Pod wouldn't start with the settings as they are: {error}
14. system_edit.numpad_off: Numpad off: {value}
15. system_edit.numpad_on: Numpad on: {value}
16. system_edit.path_item_e_g_managers: {arg0} <path> <item> - e.g. {arg0} managers.who_words what's your name
17. system_edit.restarting_wire_pod_forced_like: Restarting Wire-Pod (forced, like wpr 2; {audio}) - the page reloads itself when it's back; click Log once for sound.
18. system_edit.system_lingo_brobots_not_changed: system-lingo brobots: not changed - the list wouldn't read
19. system_edit.system_lingo_brobots_reporters_butler: system-lingo <brobots|reporters|butler> <word> = <said> / remove <word> / (just the lingo: its list)
20. system_edit.system_lingo_brobots_reporters_or: system-lingo: brobots, reporters or butler.
21. system_edit.system_lingo_brobots_the_pronunciation: system-lingo brobots: the pronunciation list isn't found in {file}
22. system_edit.system_lingo_not_changed: system-lingo {lingo}: not changed - {error}
23. system_edit.system_numpad_on_off_the: system-numpad on|off - the numpad keys as push-to-talk; alone: each key.
24. system_edit.system_score_song: system-score <song> - {names}
25. system_edit.system_topic_topic_words_add: system-topic <topic> [words add|remove <word> | words <a, b, c> | point <text> | point <n> <text> | point <n> remove | label|reminder|pictures <text>]\ntopics: {names}
26. system_edit.in_a_cluster: {dotted} is in the topic-cluster {file} now - edit it there, then system-restart.
topics: {names}
26. system_edit.system_topic_words_point_label: system-topic {arg1} words | point | label | reminder | pictures ...
27. system_edit.words: {arg1} words: {words}
28. system_edit.topic_point: point {n}: {point}
29. system_edit.topic_point_taken_out: point {n} taken out
30. system_edit.topic_words: words: {words}

## System extras (system_text)
1. system_extras.edge_tpu_worker: Edge TPU worker: {value}
2. system_extras.edge_tpu_worker_not_ready: Edge TPU worker: not ready after 30 s.
3. system_extras.edge_tpu_worker_ready: Edge TPU worker: ready.
4. system_extras.system_chat_save_no_saved: system-chat-save: no saved sessions.
5. system_extras.system_chat_save_no_session: system-chat-save: no session {arg1} (its folder name, e.g. 154136_session)
6. system_extras.system_chats_words_operator_brobots: system-chats <words> [operator|brobots|managers]
7. system_extras.system_menu_theme: system-menu-theme <{names}>
8. system_extras.system_menu_theme_f5_to: system-menu-theme: {arg1} - F5 to see it (system-undo puts it back one value at a time)
9. system_extras.system_menu_theme_then_f5: system-menu-theme <{names}> (then F5)
10. system_extras.system_numpad_the_key_map: system-numpad: the key map can't be read - {error}
11. system_extras.chat_saved: system-chat-save: {n} lines -> {file}
12. system_extras.chats_found: system-chats "{words}": {found} found in {sessions} sessions

## System locate (system_text)
1. system_locate.can_t_read_it: {value} (can't read it)
2. system_locate.no_step_in_system_set: no step {id} in {dir} - system-set {dir} lists them.
3. system_locate.not_in_it: {value} (not in it)
4. system_locate.nothing_named_in_the_settings: nothing named "{word}" in the settings or songs - system-locate shows how to use it.

## System more (system_text)
1. system_more.system_undo_not_undone_has: system-undo: not undone - {path} has changed since "{what}" ({at}). Edit it by hand: system-locate.
2. system_more.system_undo_nothing_to_undo: system-undo: nothing to undo.
3. system_more.system_undo_undid: system-undo: undid "{what}" ({at}){left}

## System play (system_text)
1. system_play.apply_phcal_tweaks_before_a: Apply phcal tweaks before a live system-play: {value} - system-tweaks on|off
2. system_play.calibrating_its_values_tmp_gopod: Calibrating: {args} ({value}). Its values: /tmp/gopod_system_play.log
3. system_play.isn_t_a_golden_engine: {dir} isn't a golden-engine song - start interview plays the interview.
4. system_play.reporter_gap_for_system_play: Reporter gap for system-play: {now} - system-gap <seconds> | off | default
5. system_play.system_calibrate_arm_nod_rattle: system-calibrate <arm|nod|rattle|danger|cube> <1|2> [reps] [--hold S] [--speed N] [--cycles N] [--volume N] [dry]
6. system_play.system_calibrate_words_numbers_and: system-calibrate: "{w}"? - words, numbers and --options only.
7. system_play.system_gap_seconds_off_default: system-gap <seconds> | off | default
8. system_play.system_play_needs_a_step: system-play: {w} needs a step.
9. system_play.system_play_robot_1_or: system-play: robot 1 or robot 2.
10. system_play.system_play_song_from_step: system-play <song> [from <step>] [to <step>] [robot 1|2] [dry] - steps by id or number (system-score <song>).
11. system_play.system_play_system_play_song: system-play: "{value}"? - system-play <song> [from <step>] [to <step>] [robot 1|2] [dry]
12. system_play.no_step: no step {step} in {song} (system-score {song} lists them)
13. system_play.playing: Playing {song}, {what}, robot {robot}, {mode}. stop run stops it. (log: /tmp/gopod_system_play.log)

## System robots (system_text)
1. system_robots.animation_tokens: animation tokens: {value}
2. system_robots.doing: doing {arg}
3. system_robots.is_learning_s_face_look: {brobot} is learning {arg1}'s face - look at him.
4. system_robots.knows: {brobot} knows: {names}
5. system_robots.knows_no_faces_yet_system: {brobot} knows no faces yet - system-faces add <name> {value}
6. system_robots.no_face: no face {arg1}
7. system_robots.no_stats_yet_the_robot: no stats yet (the robot keeps none)
8. system_robots.played: played {arg}
9. system_robots.played_2: played {name}
10. system_robots.said: said ({value})
11. system_robots.set_to: {what} set to {arg}
12. system_robots.system_anim_text_with_1: system-anim <text with {{playAnimationWI||token}}> [1|2|both] - system-anim list for the tokens
13. system_robots.system_brobot_anim_name_1: system-brobot anim <name> [1|2|both] - system-robot anims <word> finds names (anims sleep).
14. system_robots.system_brobot_beat_alias_1: system-brobot beat <alias> [1|2|both]\n{out2}
{out2}
15. system_robots.system_brobot_beat_alias_1_2: system-brobot beat <alias> [1|2|both] - system-brobot beats lists them.
16. system_robots.system_brobot_cube_1_2: system-brobot cube <{names}> [1|2|both]
17. system_robots.system_brobot_eyes_1_2: system-brobot eyes <{eye_colors}> [1|2|both]
18. system_robots.system_brobot_head_lift_speed: system-brobot head|lift <speed, e.g. 3 or -3> | wheels <left> <right> [1|2|both]
19. system_robots.system_brobot_numbers_only: system-brobot {what}: numbers only.
20. system_robots.system_brobot_say_text_1: system-brobot say <text> [1|2|both]
21. system_robots.system_brobot_stats_anims_word: system-brobot stats | anims <word> | anim <name> | cube <name> | volume <0-5> | eyes <{eye_colors}> | locale <en-US> | units c|f | location <place> | timezone <zone> | say <text> | wake | head|lift <speed> | wheels <left> <right> | camera   [1|2|both]
22. system_robots.system_brobot_units_c_f: system-brobot units c|f [1|2|both]
23. system_robots.system_brobot_volume_0_5: system-brobot volume <0-5> [1|2|both]
24. system_robots.system_faces_add_name_1: system-faces add <name> [1|2] - then look at {brobot}
25. system_robots.system_faces_delete_name_1: system-faces delete <name> [1|2]
26. system_robots.system_faces_done: system-faces {what}: done
27. system_robots.system_faces_list_rename_old: system-faces [list | rename <old> <new> | delete <name> | add <name>] [1|2]
28. system_robots.system_faces_rename_old_new: system-faces rename <old> <new> [1|2]
29. system_robots.system_robot_system_brobot_alone: system-robot: "{what}"? - system-brobot alone lists them.

## System search (system_text)
1. system_search.system_search_no_answer_from: system-search "{q}": no answer from the search ({error}) - is gopod-searxng running?
2. system_search.system_search_nothing_found: system-search "{q}": nothing found.
3. system_search.system_search_words_gopod_s: system-search <words> - GOPOD's own web search (no LLM). To the Brobots: "search <words>".

## System set (system_text)
1. system_set.a_combo_step_its_parts: {where}: a combo step - its parts are edited in knobs.json for now.
2. system_set.give_it_a_value_system: Give it a value: system-set {arg0} {arg1} {arg2} <value>
3. system_set.isn_t_a_field_system: {field} isn't a field system-set writes - system-set lists them.
4. system_set.no_step_in_system_set: no step {id} in {dir} - system-set {dir} lists them.
5. system_set.not_changed: {name}: not changed - {error}
6. system_set.not_changed_2: {where}: not changed - {error}
7. system_set.not_changed_doesn_t_read: {where}: not changed - {name} doesn't read
8. system_set.not_changed_in: {where}: not changed in {name} - {error}
9. system_set.system_set_song_tempo_seconds: system-set <song> tempo <seconds, 0 or more>
10. system_set.takes_effect_after: {name}: {old} -> {raw} ({about}; takes effect after {when})
11. system_set.takes_effect_after_wpr_2: takes effect after wpr 2
12. system_set.takes_effect_at_the_next: takes effect at the next run
13. system_set.tempo_the_next_run_of: {dir} tempo: {out} (the next run of it)
14. system_set.keyboard_only: system commands work from the chat page's keyboard only.

## System tests (system_text)
1. system_tests.no_speaker_named_system_tests: no speaker named (system_tests.audio.sink)
2. system_tests.speaker_and_mic_set: speaker and mic set
3. system_tests.speaker_set_no_mic_with: speaker set, no mic with {source_has} found
4. system_tests.system_audit_is_already_running: system-audit is already running - its result comes here when it's done.
5. system_tests.system_audit_started_everything_but: system-audit started - everything but the robots (a few minutes; keep this page open). Its result comes here.
6. system_tests.audit_last: system-audit: {last}
7. system_tests.didnt_run: didn't run: {why}
8. system_tests.head: system-test: {pass} of {tests} pass
9. system_tests.missing: missing: {missing}
10. system_tests.no_report: no report: {out}
11. system_tests.voices_timed: worker ready {ready}, {timed} of {used} voices in use timed by Vosk

## System tour (system_text)
1. system_tour.system_tour_a_run_is: system-tour: a run is going - try again after it
2. system_tour.system_tour_no_system_voice: system-tour: no system voice (sign_off.system_role)
3. system_tour.system_tour_stopping_at_the: system-tour: stopping at the next commandment.

## System voices (system_text)
1. system_voices.system_voices_rebuild_no_backup: system-voices-rebuild: no backup folder - {error}
2. system_voices.system_voices_rebuild_the_voice: system-voices-rebuild: the voice folder can't be read - {error}
3. system_voices.rebuild_paused: Paused - system-check found {problems} to look at ({checked} checked):
4. system_voices.rebuild_ready: Backed up {moved} voices. Ready for a restart: wpr 2, then F5 and click Log - every voice is made again at the start.

## System window more (system_text)
1. system_window_more.docbs_pools_system_docbs_pool: DocBS pools: {out} - system-docbs <pool> shows one; add|remove <line> edits it.
2. system_window_more.lines_now: {pool}: {arg2} "{line}" - {value} lines now.
3. system_window_more.record_session: Record Session {arg1}.
4. system_window_more.system_docbs_no_pool: system-docbs: no pool {pool}
5. system_window_more.system_docbs_no_pool_system: system-docbs: no pool {pool} (system-docbs lists them)
6. system_window_more.system_record_on_off_record: system-record on|off - Record Session (the page's Record? box)

## Thinking line (system_text)
1. thinking_line.thinking_please_wait: Thinking, please wait.

## Torah (system_text)
1. torah.the_gopod_torah_is_written: The GOPOD Torah is written again: GOPOD_TORAH.md and its scriptures (torah/) match the system as it is now.

## Value points golden (system_text)
1. value_points_golden.no_topic_system_topic_lists: no topic {arg1} (system-topic lists them)
2. value_points_golden.how: system-value-points <topic> lists a topic's; system-value-points lock locks changes
3. value_points_golden.locked: value points locked: {n} (system-undo puts the old index back)
4. value_points_golden.all_locked: all locked, unchanged
5. value_points_golden.to_look_at: {n} to look at
6. value_points_golden.head: value points: {points} in {topics} topics - {state}

## Voice profiles (system_text)
1. voice_profiles.system_voices_report_3d_week: system-voices-report [3d | week | month | year]

## Typed system (system_text)
1. typed_system.not_a_command: 0 is the system: system-<command>, stop run, start <run>, replay <session>, stop session (system-help lists them). Pick 1-9 to talk.
2. typed_system.dial_zero: Did you mean to dial '0' for the operator?

## System check (system_text)
1. system_check.all_good: system-check: all good - {checked} things checked (settings, chapters, songs).
2. system_check.to_fix: system-check: {problems} to fix ({checked} things checked):
3. system_check.words_left: {n} texts in the settings files, {allowed} allowed - put the new ones in a topic-cluster: {paths}

## System status (system_text)
1. system_status.admin_pages:   admin pages: {n} open - keep one (visitor pages can be many)
2. system_status.model_cold:   language model: {model} not loaded (cold)
3. system_status.model_none:   language model: no endpoint/model in the knowledge settings
4. system_status.model_warm:   language model: {model} loaded (warm)
5. system_status.page_away:   chat page: away
6. system_status.page_here:   chat page: here
7. system_status.run_going:   run: {run} going
8. system_status.run_none:   run: none
9. system_status.warmup_done:   warm-up: done
10. system_status.warmup_going:   warm-up: going - lines wait for LLM Ready

## Voices report (system_text)
1. voices_report.head: Voice files, past {period} (report only - nothing moves; plays counted since {since}):

## Probers make (system_text)
1. probers_make.made: The probe maker made {made} probers for {points} value points - {file}. Mic-record them for the demo.
2. probers_make.not_written: system-probers-make: not written - {error}

## Robots off (system_text)
1. robots_off.now: Robots off: {state}. system-robots-off on|off switches it.
2. robots_off.how: system-robots-off on|off - on: nothing reaches the Brobots; their says play on the page in their numpad voices, says and animations captured
3. robots_off.switched: Robots off: {state}.
4. robots_off.not_saved: system-robots-off: not saved - {error}
