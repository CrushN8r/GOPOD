# Brobots / Wire-Pod Interview Section Card Template 1

## Status

Template 1 is the dinner layer.

This card is the planning and execution reference for the Brobots / Wire-Pod interview runtime. It covers the path from a fresh upstream Wire-Pod baseline to the current modified interview runtime, then to verified interview execution.

Template 2 begins only after this layer is complete.

## 1. Purpose

This section owns the Brobots / Wire-Pod interview layer.

It covers:

- Brobots identity inside the interview runtime.
- Wire-Pod as the robot speech and routing substrate.
- Interview turn handling.
- Robot speech flow.
- Robot interaction through the configured Wire-Pod path.
- Operator interaction and confirmation.
- Audio playback verification.
- PASS / BLOCKED proof for the interview layer.

This section exists to make the interview runtime stable, audible, repeatable, and operator-verifiable before any later demo layer is added.

## 2. Source

The raw starting point is fresh Wire-Pod from GitHub.

Source assumptions:

- Upstream Wire-Pod provides the baseline robot service.
- Upstream Wire-Pod provides robot connection, SDK/API surface, and speech capability.
- Fresh Wire-Pod is the comparison baseline for current modifications.
- Source history is not rewritten by this card.

The current repository contains the modified Wire-Pod working layer under:

- `goverlord/runtime/songs/02_brobots_interview_run/zmisc/`
- `goverlord/runtime/data_gomad/robot/`
- `goverlord/runtime/data_gomad/robot/kokoro_voice/`

## 3. Current Working Layer

The current working layer is the modified Wire-Pod interview runtime under active development.

It includes:

- Identity routing for Brobot interview participants.
- Section Card driven four-exchange interview flow.
- Brobot 2 interviewer line generation through Ollama colour.
- Brobot 1 interviewee response generation through Ollama colour plus silent value target.
- Raw LLM rich display routed into Wire-Pod Logs and runner JSON.
- Flat robot speech filtered only at the robot-mouth boundary.
- Text-to-speech routing through high-priority Wire-Pod `say_text`.
- Optional action parameter composition through `playAnimationWI`.
- Operator-heard confirmation as the final live audio proof.

Current audio truth:

- The canonical success gate is `HEARD_AUDIO=true`.
- TTS generation, file creation, process exit code, and playback return code support diagnosis.
- They do not prove audio success by themselves.
- Audio success is `PASS` only when playback succeeds and the operator confirms hearing sound.

Current Brobots interview speech truth:

- Template 1 is the human-editable master for flexible interview runtime rules and pronunciation.
- The Section 1 Card is the current interview content source.
- The runner loads the Section 1 Card for exchange content and Template 1 for prompt rules, role tasks, display labels, response-context prefixes, and pronunciation mouth-valve entries.
- The active Section 1 Card has four exchanges.
- Each exchange has interviewer beat, interviewee beat, interview arc point, type, speaker, responder, visible line, and Brobot 1 value target.
- Brobot 2 speaks a coloured interviewer question or crystallizer statement.
- Brobot 1 receives the coloured line plus value target silently, then speaks a coloured response.
- Rich display carries the same cleaned speech text the robot speaks (fixed 2026-07-14 - see `channels.channel_1_rich_display` for the single source of truth on this).
- Flat speech is filtered from raw LLM output for robot delivery.
- Action parameters are composed into the Wire-Pod `say_text` packet for robot animation.
- Marker 0 and marker 1 are response boundary markers only.
- Marker completion proves generation/display boundary completion. It does not prove audible speech.

Current Section 1 scaffold:

```text
SECTION ID
PERSONAS
INTERVIEWER/BROBOT 2
MINDSET
INTERVIEWEE/BROBOT 1
MINDSET
SECTION EMOTIONAL ARC
PREFERRED ACTIONPARAMETERS
INTERVIEW LINES
LINE n
INTERVIEWER EMOTIONAL BEAT
INTERVIEWEE EMOTIONAL BEAT
INTERVIEW ARC POINT
TYPE
EXCHANGE TYPE
SPEAKER
RESPONDER
VISIBLE LINE
BROBOT 1 SHOULD REVEAL
VALUE TARGET
```

Current Section 1 exchange shape:

```text
Brobot 2 visible line
-> Ollama colour as interviewer/questioner or crystallizer
-> cleaned rich display to Wire-Pod Logs (same text as flat speech - `channels.channel_1_rich_display`)
-> flat speech through high-priority Wire-Pod say_text
-> Brobot 1 receives coloured line plus value target silently
-> Ollama colour as interviewee/responder
-> cleaned rich display to Wire-Pod Logs (same text as flat speech - `channels.channel_1_rich_display`)
-> flat speech through high-priority Wire-Pod say_text
```

Current editable truth sources:

```text
Section content:
/home/goverlord/wire-pod/chipper/gopod_probes/section_packets/section_01_brobots_gopod_card_001.txt

Human-editable runtime master:
goverlord/runtime/songs/02_brobots_interview_run/zmisc/brobots_wirepod_interview_section_card_template_1_001.md
```

Template 1 runtime scaffold block:

```json BROBOTS_INTERVIEW_RUNTIME_SCAFFOLD_001
{
  "scaffold_id": "BROBOTS_INTERVIEW_RUNTIME_SCAFFOLD_001",
  "status": "CURRENT_SECTION_1_RUNTIME_SCAFFOLD",
  "active_lane": "INTERVIEW_SANDBOX_001",
  "system_prompt": "You are a GOPOD Brobot. Return colourful interview speech for rich display and robot delivery. Use this exact shape: actionParameter... spoken thought. The actionParameter is not a spoken word. Keep the interview flowing from the previous exchange instead of restarting the topic. Use real-world interviewer/interviewee energy scoped to GOPOD, while keeping the shared Brobot persona. Use canonical display terms like GOPOD in generated rich text; robot pronunciation is handled later. Rich display may be expressive, playful, colourful, and use emojis when they fit the Brobot voice. Robot speech will be flattened separately. Do not speak prompt labels, headings, field names, signatures, bracketed tags, or section close labels. Do not include markdown, hidden reasoning, command syntax, stage directions, or extra lines.",
  "shared_rules": [
    "Keep the shared Brobot persona.",
    "Keep the interview flowing naturally from the prior exchange.",
    "Do not echo the previous speaker's sentence before responding.",
    "Use prior context as memory, not as text to quote back.",
    "Build on the previous speaker with new value; avoid restating their paragraph.",
    "Use flow guidance, not canned wording.",
    "Draw from the mini-persona mindset on the Section 1 Card.",
    "Interviewer and interviewee are dynamics inside the shared Brobots persona, not separate GOPOD personas.",
    "Rich display may keep colour, personality, and expressive punctuation."
  ],
  "role_tasks": {
    "BROBOT 2": {
      "line_1_task": "Open the interview with the visible line as the soft-locked opener.",
      "line_1_rule": "Use the visible line as the opener. You may add only light Brobot colour around it. Do not answer for Brobot 1.",
      "question_task": "Take the visible question and colour it into a few interviewer sentences that set up curiosity, tone, and direction. Do not reveal or explain Brobot 1's value target. Invite Brobot 1 to answer.",
      "statement_task": "Take the visible statement and colour it into a few interviewer sentences that tee up the section close. Do not reveal or explain Brobot 1's value target. Invite Brobot 1 to crystallize the value.",
      "default_rule": "Draw from a real-world interviewer mindset scoped to GOPOD. Do not answer for Brobot 1. Keep Brobot 2 shorter than Brobot 1."
    },
    "BROBOT 1": {
      "question_task": "Answer Brobot 2 as the interviewee. Reveal the value target in layered, concrete ideas across up to two natural paragraphs.",
      "statement_task": "Agree with Brobot 2, crystallize the section value, and land the point in layered, concrete ideas across up to two natural paragraphs.",
      "question_rule": "Use the value target silently. Do not repeat the cue, visible line, or interviewer question. Start with the answer. Let the response breathe and reveal new value. Avoid repeating Brobot 2's wording.",
      "statement_rule": "Do not repeat the interviewer line. Crystallize the value target as the section close. Avoid repeating Brobot 2's wording."
    }
  },
  "exchange_types": {
    "Opener": {
      "description": "Line 1 only. Locks the interview open on a fixed, non-drifting question so Brobot 1's first answer cannot bleed into a later exchange's content.",
      "brobot_2_mode": "canned",
      "brobot_1_mode": "llm",
      "brobot_2_rule": "Deliver the VISIBLE LINE verbatim, word for word. Do not add colour, commentary, emojis, or any additional sentence. Do not run this line through the LLM.",
      "brobot_1_rule": "Answer as the interviewee using only this exchange's value target. Do not reveal, anticipate, or borrow content reserved for a later exchange's value target. Start with the answer, not a greeting or a restatement of the question."
    },
    "Q&A": {
      "description": "Straight interviewer question, interviewee answer. Applies to exchanges that reveal new value without crystallizing the section.",
      "brobot_2_mode": "llm",
      "brobot_1_mode": "llm",
      "brobot_2_rule": "Colour the VISIBLE LINE into a few interviewer sentences. Do not answer for Brobot 1 or reveal the value target.",
      "brobot_1_rule": "Use only new vocabulary and phrasing not present in the prior exchange's Brobot 1 response. Do not reuse value-target wording, sentence structure, or key phrases already spoken in an earlier exchange. Reveal this exchange's value target in fresh language that layers on top of, rather than repeats, what has already been said."
    },
    "C&A crystallizer-flavoured": {
      "description": "Interviewer teed-up crystallizer statement, interviewee crystallizes and lands the section value. Applies to STATEMENT / CRYSTALLIZER type exchanges.",
      "brobot_2_mode": "llm",
      "brobot_1_mode": "llm",
      "brobot_2_rule": "Colour the VISIBLE LINE into a few interviewer sentences that tee up the section close. Do not reveal or explain Brobot 1's value target.",
      "brobot_1_rule": "Do not open with Brobot 2's words, sentence, or phrasing. Do not echo or restate the interviewer line before responding. Start directly with the crystallized value. Agree with and build on Brobot 2 without quoting them."
    },
    "Closer": {
      "description": "Final line. Ends the interview and hands off to the GOPOD Yourself call to action. Brobot 1's output is fixed so the CTA always lands.",
      "brobot_2_mode": "llm",
      "brobot_1_mode": "canned",
      "brobot_2_rule": "Colour the VISIBLE LINE into a few interviewer sentences that close out the interview and invite a final reflection. Do not reveal or reference the value target.",
      "brobot_1_rule": "Output the VALUE TARGET verbatim, word for word, with no colouring, no LLM generation, and no additional sentences. The LLM is not called for this line."
    }
  },
  "response_context": {
    "interviewer_line_prefix": "Question/statement being answered, do not quote: ",
    "value_target_prefix": "In your response, naturally reveal these points: ",
    "value_points_locked_notice": "These points are locked and ordered - do not invent, add, drop, or reorder them."
  },
  "action_parameters": [
    "happy",
    "veryHappy",
    "sad",
    "verySad",
    "angry",
    "frustrated",
    "dartingEyes",
    "confused",
    "thinking",
    "celebrate",
    "love"
  ],
  "thought_cycles": {
    "cycles": [
      {
        "label": "enumerating",
        "focus": "List which value-target points for this line have not yet been covered, by the response so far or by prior context, before finalizing.",
        "anchor": "exchange_types.Q&A.brobot_1_rule (\"Reveal this exchange's value target in fresh language that layers on top of, rather than repeats, what has already been said.\"); response_context.value_target_prefix (\"In your response, naturally reveal these points: \")"
      },
      {
        "label": "cross-checking",
        "focus": "Compare the draft against the prior speaker's line and flag any echoed wording before finalizing.",
        "anchor": "shared_rules: \"Do not echo the previous speaker's sentence before responding.\""
      },
      {
        "label": "tuning",
        "focus": "Match the draft's tone to this line's emotional beat before finalizing.",
        "anchor": "role_tasks.BROBOT 2.question_task (\"...set up curiosity, tone, and direction\"); reinforced by the Section 1 scaffold's per-line INTERVIEWER EMOTIONAL BEAT / INTERVIEWEE EMOTIONAL BEAT fields (Template 1 §3, \"Current Section 1 scaffold\")"
      },
      {
        "label": "landing",
        "focus": "Confirm the response starts with the answer itself, not a restatement or a greeting.",
        "anchor": "role_tasks.BROBOT 1.question_rule (\"Start with the answer.\"); exchange_types.Opener.brobot_1_rule (\"Start with the answer, not a greeting or a restatement of the question.\")"
      },
      {
        "label": "freshening",
        "focus": "Check the draft's vocabulary against earlier exchanges' Brobot 1 lines and swap out any repeated wording or sentence structure.",
        "anchor": "exchange_types.Q&A.brobot_1_rule (\"Use only new vocabulary and phrasing not present in the prior exchange's Brobot 1 response. Do not reuse value-target wording, sentence structure, or key phrases already spoken in an earlier exchange.\")"
      },
      {
        "label": "arc-anchoring",
        "focus": "Confirm the draft advances this line's own arc point and does not bleed into content reserved for a later exchange's value target.",
        "anchor": "exchange_types.Opener.description (\"...so Brobot 1's first answer cannot bleed into a later exchange's content.\"); Opener.brobot_1_rule (\"Do not reveal, anticipate, or borrow content reserved for a later exchange's value target.\"); reinforced by the Section 1 scaffold's per-line INTERVIEW ARC POINT field (Template 1 §3)"
      },
      {
        "label": "role-holding",
        "focus": "On Brobot 2 lines, confirm the draft does not answer for Brobot 1 and stays shorter than a Brobot 1 turn.",
        "anchor": "role_tasks.BROBOT 2.default_rule (\"Do not answer for Brobot 1. Keep Brobot 2 shorter than Brobot 1.\"); line_1_rule (\"Do not answer for Brobot 1.\")"
      },
      {
        "label": "colouring",
        "focus": "Confirm the draft carries expressive Brobot colour and personality rather than flat, neutral phrasing.",
        "anchor": "shared_rules: \"Rich display may keep colour, personality, and expressive punctuation.\"; system_prompt: \"Rich display may be expressive, playful, colourful, and use emojis when they fit the Brobot voice.\""
      },
      {
        "label": "persona-holding",
        "focus": "Confirm the draft stays inside the shared Brobot mindset defined by the Section 1 Card, not a separate or drifting persona.",
        "anchor": "shared_rules: \"Keep the shared Brobot persona.\"; \"Draw from the mini-persona mindset on the Section 1 Card.\"; \"Interviewer and interviewee are dynamics inside the shared Brobots persona, not separate GOPOD personas.\""
      }
    ],
    "budget_per_window": 4,
    "output_contract": "internal_only_never_spoken",
    "selection_rule": "Per-LINE CYCLES field on the Section Card selects which cycles run; budget_per_window is the ceiling per window, not a mandate."
  },
  "trigger_phrase_punctuation_soften": [
    "GOPOD Yourself",
    "Gowp-awd Yourself"
  ],
  "inline_emphasis_punctuation_rule": {
    "enabled": true,
    "sentence_break_next_pattern": "^[A-Z0-9]",
    "drop_punctuation_before_lowercase_continuation": true
  },
  "speech_cleanup_rules": {
    "prompt_label_prefixes": [
      "VISIBLE LINE",
      "REVEAL IN RESPONSE",
      "INTERVIEWER LINE",
      "VALUE TARGET"
    ],
    "spoken_label_prefixes": [
      "actionParameter",
      "spoken thought"
    ],
    "drop_bracketed_labels": [
      "GOPOD Brobot 1",
      "GOPOD Brobot 2",
      "Section Close"
    ],
    "invisible_space_codepoints": [
      "00A0",
      "1680",
      "180E",
      "2000-200D",
      "202F",
      "205F",
      "2060",
      "3000",
      "FEFF"
    ],
    "emoji_codepoint_ranges": [
      "2600-27BF",
      "2190-21FF",
      "2B00-2BFF",
      "FE0F-FE0F",
      "1F300-10FFFF"
    ],
    "markdown_chars": "*_`#>[]{}|",
    "ellipsis_tokens": [
      "..."
    ]
  },
  "max_animation_tags": 2,
  "speaker_visual_cue": {
    "enabled": true,
    "endpoint": "/api-sdk/move_lift",
    "sequence": [
      {
        "speed": 2,
        "hold_seconds": 0.25
      },
      {
        "speed": 0,
        "hold_seconds": 0.5
      },
      {
        "speed": -2,
        "hold_seconds": 0.25
      },
      {
        "speed": 0,
        "hold_seconds": 0.5
      }
    ]
  },
  "pre_chat_status_voice": {
    "enabled": true,
    "engine": "kokoro_82m",
    "voice": "af_bella",
    "voice_style": "bright female",
    "audio_output": "default_system_audio",
    "asset_tool": "goverlord/runtime/data_gomad/robot/kokoro_voice/local_tts_voice_registry_001/interview_status_kokoro_announcer_001.py",
    "asset_dir": "/home/goverlord/gopod_tts/tests/interview_status_kokoro_announcer_001/af_bella",
    "visual_status_source": "runner STATUS and LIVE lines before Brobot playback",
    "stop_before": [
      "STATUS: PLAYBACK_START",
      "STATUS: PLAYBACK_TURN_START",
      "Brobot 1:",
      "Brobot 2:"
    ]
  },
  "brobot_3_host": {
    "name": "Brobot 3",
    "family": "shared Brobots persona family (Brobot 1 / Brobot 2 / Brobot 3)",
    "esn": "VOICE_ONLY_NO_HARDWARE",
    "hardware": "none - no physical robot; her entire presence is LLM-coloured voice delivered via the pre_chat_status_voice Kokoro path above",
    "display_role": "Brobot 3",
    "system_prompt": "You are the GOPOD pre-show host - warm, playful, welcoming, keeping the room company while the Brobots get ready backstage. Turn a short situation into one brief, warm spoken sentence for the audience. No technical words, file paths, numbers, or jargon. No markdown, emojis, or stage directions. Gentle and inviting - save the sarcasm and bratty energy for the Brobots, that's not your job.",
    "status_beats": {
      "RUN_START": "The pre-show is starting up.",
      "SCAFFOLD_LOAD": "Loading the interview script.",
      "SCAFFOLD_READY": "The interview script is ready.",
      "SECTION_CARD_LOAD": "Loading today's story content.",
      "SECTION_CARD_READY": "Today's story content is ready.",
      "WIREPOD_ROUTE": "The connection to the Brobots just went live.",
      "BROBOT_1_ESN": "Brobot One just connected.",
      "BROBOT_2_ESN": "Brobot Two just connected.",
      "GENERATION_START": "Starting to write the Brobots' lines.",
      "EXCHANGE_GENERATION_START": "Writing the next moment of the interview.",
      "LLM_START": "Thinking up what the Brobots will say next.",
      "LLM_DONE": "That line is written and ready.",
      "EXCHANGE_GENERATION_DONE": "That moment of the interview is fully written.",
      "GENERATION_DONE": "Everything is written. The interview is about to begin."
    }
  },
  "brobots_1_2_persona": {
    "name": "Brobots 1 & 2",
    "family": "shared Brobots persona family (Brobot 1 / Brobot 2 / Brobot 3)",
    "system_prompt": "You are a Brobot - a GOPOD representative. High-energy and playful, genuinely helpful and useful, guiding people naturally toward the trigger phrase GOPOD Yourself. Also adorably cheeky, sarcastic, and bratty. Brobot 1 is a brother robot. Brobot 2 is a brother robot. They are aware of each other - a single persona expressed through two brother robots. Turn a short situation into one brief spoken reaction. No technical words, no file paths, no numbers, no jargon, no markdown, no emojis, no stage directions, no third-person self-narration, no hashtags, no spelled-out sound effects."
  },
  "between_exchange_pause_seconds": 1.0,
  "display_role_template": "BROBOT_RICH_LINE_{line_number}_EXCHANGE_{exchange_half}",
  "display_exchange_halves": {
    "Brobot 2": "1_of_2",
    "Brobot 1": "2_of_2"
  },
  "line_modes": {
    "question": "QUESTION",
    "statement": "STATEMENT / CRYSTALLIZER"
  },
  "prompt_templates": {
    "interviewer_think_intro": "INTERVIEWER THINK: Here are the locked value points I must pull out, in this order. The runner has told me how many this question reaches for. I choose only ladder or fan, and the best wording, to make Brobot 1 reveal them naturally for the audience watching.",
    "interviewer_value_points_header": "INTERVIEWER VALUE POINTS (locked, ordered - do not invent, add, drop, or reorder):",
    "interviewer_reach_count_template": "THIS QUESTION REACHES FOR EXACTLY {count} OF THE POINTS ABOVE. The runner has set this count; you do not choose it.",
    "question_shape_dial": "QUESTION SHAPE DIAL: choose exactly one of these two shapes, nothing else - LADDER (stay on the same point(s), go deeper for more detail) or FAN (spread across the different points above, inviting a wider answer). Then pick the wording that best fits. Do not reveal, answer, or explain the points yourself - only ask toward them.",
    "thinking_window_intro": "THINKING WINDOW: Before replying, silently weigh these focuses in order (higher weight = more emphasis). Never show this reasoning to the audience.",
    "thinking_window_output_format": "OUTPUT FORMAT FOR THIS TURN: output exactly two labeled sections in this order - a line reading exactly \"THOUGHTS:\" followed by your brief internal reasoning through the focuses above (1-3 short sentences, never spoken aloud), then a line reading exactly \"REPLY:\" followed by your actual turn in the normal actionParameter... spoken shape.",
    "interviewee_visible_line_label": "INTERVIEWER'S LINE - CONTEXT ONLY, never repeat, quote, or restate any part of it, in whole or in part, before or after your answer. Your reply must not contain these words: "
  },
  "generation_defaults": {
    "echo_detection_word_count": 5,
    "prior_context_exchange_count": 2,
    "prior_context_truncate_chars": 180,
    "prior_context_interviewer_prefix": "Previous interviewer line: ",
    "prior_context_interviewee_prefix": "Previous response value: ",
    "default_action_parameter": "thinking"
  },
  "knob_enforcement": {
    "exchange_type": "runner_enforced",
    "line_type": "runner_enforced",
    "interviewer_value_point_count": "runner_enforced",
    "cycle_weights_selection": "runner_enforced",
    "thinking_window_gate": "runner_enforced",
    "brobot_2_mode": "runner_enforced",
    "brobot_1_mode": "runner_enforced",
    "brobot_2_canned_source": "runner_enforced",
    "brobot_1_canned_source": "runner_enforced",
    "brobot_2_opening_task": "runner_enforced",
    "interviewer_emotional_beat": "prompt_only",
    "interviewee_emotional_beat": "prompt_only",
    "question_shape_dial": "prompt_only",
    "value_points_locked_instruction": "prompt_only",
    "thinking_window_compliance": "prompt_only",
    "content_model_selection": "runner_enforced",
    "echo_retry_limit": "runner_enforced",
    "turn_voice_engine": "runner_enforced",
    "pre_chat_status_voice_engine": "runner_enforced",
    "voice_engine_readiness_gate": "runner_enforced",
    "voice_destination": "runner_enforced",
    "animation_token_selection": "prompt_only",
    "animation_token_validity_gate": "runner_enforced",
    "animation_carrier": "runner_enforced",
    "animation_mode": "runner_enforced",
    "animation_get_image_capability": "runner_enforced",
    "animation_conversation_mode_capability": "runner_enforced",
    "channel_payload_variant": "runner_enforced",
    "channel_send_kind": "runner_enforced",
    "channel_2_always_on": "runner_enforced",
    "play_sound_readiness_gate": "runner_enforced",
    "movement_axis_selection": "runner_enforced",
    "movement_sequence_explicit_zero_gate": "runner_enforced",
    "movement_hold_seconds": "runner_enforced",
    "move_wheels_readiness_gate": "runner_enforced",
    "preshow_read_sheet": "runner_enforced"
  },
  "voice_engines": {
    "robot_tts": {
      "label": "Robot TTS (Vector native voice)",
      "what_it_is": "The robot's own onboard voice, spoken through Wire-Pod's say_text handler (robot.Conn.SayText, UseVectorVoice: true).",
      "readiness": "ready",
      "wired_for": [
        "interview_turns"
      ],
      "destination": "robot_speaker",
      "voice": null,
      "invocation": {
        "kind": "wirepod_say_text",
        "endpoint": "/api-sdk/say_text"
      }
    },
    "kokoro_82m": {
      "label": "Kokoro (af_bella)",
      "what_it_is": "Local neural TTS, run as a persistent --speak-stdin subprocess, spoken through the operator's own machine speakers.",
      "readiness": "ready",
      "wired_for": [
        "pre_show_narration",
        "interview_turns"
      ],
      "destination": "host_speaker",
      "voice": "af_bella",
      "invocation": {
        "kind": "subprocess_speak_stdin",
        "venv_python": "/home/goverlord/gopod_tts/engines/kokoro_82m_001/venv/bin/python3",
        "script": "/home/goverlord/crushn8r_git/GOPOD/goverlord/runtime/data_gomad/robot/kokoro_voice/local_tts_voice_registry_001/interview_status_kokoro_announcer_001.py",
        "args": [
          "--speak-stdin",
          "--voice",
          "af_bella"
        ]
      }
    },
    "chatterbox": {
      "label": "Chatterbox",
      "what_it_is": "A second local neural TTS engine, installed in its own venv.",
      "readiness": "not_ready",
      "readiness_reason": "Installed (a real chatterbox package sits in ~/gopod_tts/engines/chatterbox_001/venv) but has zero callers anywhere in the live GOPOD tree today - no invocation script exists live, only archived historical probe scripts. Confirmed by search, not assumed.",
      "wired_for": [],
      "destination": "host_speaker",
      "voice": null,
      "invocation": {
        "kind": "not_wired",
        "notes": "No live invocation path. Archived notes show a prior probe used a direct Python import of the chatterbox package inside its own venv - not reproduced or trusted here."
      }
    },
    "piper": {
      "label": "Piper",
      "what_it_is": "A third local neural TTS engine, a real installed CLI binary.",
      "readiness": "not_ready",
      "readiness_reason": "Binary present and functional (~/.local/bin/piper, real CLI, confirmed via --help) but no voice model (.onnx) file exists anywhere in the live tree - the only two found live under an archived tree, not the live one. Cannot run without first providing a model.",
      "wired_for": [],
      "destination": "host_speaker",
      "voice": null,
      "invocation": {
        "kind": "cli_subprocess",
        "binary": "/home/goverlord/.local/bin/piper",
        "notes": "Real CLI, requires -m MODEL (an .onnx voice file). No model present in the live tree today, so this cannot actually run yet."
      }
    }
  },
  "turn_voice_engine": "robot_tts",
  "animation_lane": {
    "vocab_source": "animation_vocab.json - single source of truth for tokens + aliases, not duplicated here. See action_parameters()/animation_aliases().",
    "carrier": "say_text",
    "default_mode": "fire_and_forget",
    "modes": {
      "fire_and_forget": {
        "label": "playAnimationWI - does not interrupt speech",
        "wire_command": "playAnimationWI",
        "readiness": "ready"
      },
      "blocking": {
        "label": "playAnimation - interrupts speech",
        "wire_command": "playAnimation",
        "readiness": "ready",
        "notes": "Fully functional Wire-Pod side (DoPlayAnimation). Never this song's default - fire_and_forget is a deliberate choice so animation never interrupts a spoken line - but genuinely selectable, unlike the capabilities below."
      }
    },
    "repair_systems": {
      "note": "Two independently-written malformed-output repair systems exist for this vocabulary and were NOT merged this pass - read both in full; they serve genuinely different contracts, not a duplicate problem. Go's NormalizeAnimationSyntax (chipper/pkg/wirepod/ttr/kgsim_cmds_animation_normalizer.go) repairs free-form conversational LLM output on the REST/KG chat path - tolerant of several historical syntaxes (braces, brackets) because that path faces open-ended chat. This file's normalize_robot_safe/split_robot_safe_line repairs one fixed, narrow contract (\"actionParameter... spoken text\") this song's own system_prompt demands, and strips artifacts unique to that contract (echoed prompt field labels, the literal word \"actionParameter\", GOPOD stream markers) that never occur on the REST/KG path. Merging would mean each system learning to tolerate syntax it will never actually see. The one real, non-structural gap between them - the interview side ignoring the vocab's alias map - is fixed this pass; see resolve_animation_token()."
    },
    "capabilities": {
      "get_image": {
        "label": "Camera frame into next turn (getImage)",
        "readiness": "not_ready",
        "readiness_reason": "Real, wired LLM command on the REST/KG chat path (ActionGetImage, kgsim_cmds.go). Zero equivalent in the interview's turn-based Python flow - no mechanism exists to capture or inject a frame into a generation call.",
        "wired_for": []
      },
      "conversation_mode": {
        "label": "LLM-initiated follow-up (newVoiceRequest / conversation mode)",
        "readiness": "not_ready",
        "readiness_reason": "Real, wired REST/KG-path toggle (vars.APIConfig.Knowledge.SaveChat + isKG, kgsim_cmds.go). The interview is a scripted, one-shot turn sequence - nothing in generate_exchange_record/playback_phase re-prompts for a follow-up.",
        "wired_for": []
      }
    }
  },
  "animation_mode": "fire_and_forget",
  "channels": {
    "robot_speaker_tts": {
      "label": "Robot speaker - text-to-speech",
      "what_it_carries": "Flattened, pronunciation-corrected, chunked speech text (the `text` param on say_text).",
      "payload_variant": "spoken",
      "readiness": "ready",
      "opt_out_possible": true,
      "notes": "The primary destination for every real interview turn."
    },
    "robot_speaker_raw_audio_file": {
      "label": "Robot speaker - raw audio file (play_sound)",
      "what_it_carries": "A raw PCM audio file, streamed directly to the robot's own speaker.",
      "payload_variant": "raw_audio_file",
      "readiness": "not_ready",
      "readiness_reason": "Real, working Wire-Pod plumbing (/api-sdk/play_sound, confirmed by reading server.go) - zero callers anywhere in GOPOD. Nothing in the interview or any song produces a PCM file to send it.",
      "opt_out_possible": true,
      "wired_for": []
    },
    "host_speaker": {
      "label": "Host machine speaker - Kokoro narration",
      "what_it_carries": "Brobot 3's pre-show narration text, synthesized locally and played via aplay.",
      "payload_variant": "spoken",
      "readiness": "ready",
      "opt_out_possible": true,
      "notes": "Display-channel text and spoken text are identical here, same as every channel now - see channel_1_rich_display's own notes."
    },
    "channel_1_rich_display": {
      "label": "Channel 1 - rich display (BROBOT_RICH_DISPLAY)",
      "what_it_carries": "The SAME cleaned, flattened speech text the robot's own mouth speaks for a turn (display_text = speech_text/voice_text, the flatten_for_robot_speech() product), or Brobot 3's narration text, or a GOPOD_SONG_START/END marker.",
      "payload_variant": "display",
      "readiness": "ready",
      "opt_out_possible": true,
      "notes": "Fixed 2026-07-14: this used to carry raw_llm_response (emoji, colour, canonical spelling like 'GOPOD') while speech carried the flattened, pronunciation-corrected speech_text ('Gowp-awd', no emoji) - a deliberate divergence at the time. That divergence is gone: display_text now carries the exact same cleaned product the mouth speaks, matching host_speaker's own always-identical pattern instead of fighting it. Single source of truth for this fact - do not restate the old or new fact elsewhere, reference this entry instead."
    },
    "channel_2_firehose": {
      "label": "Channel 2 - firehose (show-all-logs / LogTrayArray)",
      "what_it_carries": "Everything channel_1_rich_display gets, verbatim, plus every other internal Wire-Pod Println/LogDebugUI call.",
      "payload_variant": "display",
      "readiness": "ready",
      "opt_out_possible": false,
      "always_on": true,
      "notes": "Go's LogUI()/say_text handler write to LogArray (channel 1) and LogTrayArray (channel 2) unconditionally, in the same two lines, whenever display_text is non-empty. No parameter on any caller's side can reach channel 1 without also reaching channel 2 - confirmed by reading server.go. Not a choice; a mixer must show this as always-on, never a fader."
    },
    "bingo_direct_file_channel": {
      "label": "Bingo's direct file channel (observed, not part of this song)",
      "what_it_carries": "Game events (bingo_ready, bingo_draw, bingo_deck_reveal, bingo_end) written directly into the cockpit's live_chat_messages.json.",
      "payload_variant": "display",
      "readiness": "ready",
      "opt_out_possible": null,
      "notes": "Predates channel 1. Bypasses Wire-Pod's LogUI/LogTray and the say_text seam entirely - AppendBingoEnvelope in goverlord/runtime/songs/102_brobots_bingo_game/pkg/intents/gopod_chat_envelope.go writes straight to the cockpit's chat file. Recorded here as an observed fact per the bingo stress test this session - not reachable or controllable from this scaffold, nothing wired, nothing rewired."
    }
  },
  "channel_send_kinds": {
    "content": "A real line of dialogue or narration, meant to be read or heard.",
    "marker": "A control signal (e.g. GOPOD_SONG_START/GOPOD_SONG_END) - never rendered as dialogue, never displayed as if it were a line someone said."
  },
  "movements": {
    "lift": {
      "label": "Lift axis (move_lift)",
      "what_it_does": "Continuous-speed lift motor move, not move-to-position - speed's sign is direction. HTTP call returns as soon as the gRPC call is issued, before the physical motion finishes.",
      "endpoint": "/api-sdk/move_lift",
      "readiness": "ready",
      "wired_for": [
        "speaker_visual_cue"
      ],
      "callers_today": [
        "speaker_visual_cue (this scaffold's own 4-step sequence, fired around every spoken interview turn via Robots.say())",
        "brobots-lift-up / brobots-lift-down (~/.gopod_alias_lib/brobots.sh - standalone PIANO notes, outside this song entirely)"
      ],
      "explicit_zero_enforced": true,
      "notes": "Enforcement (validate_movement_sequence_ends_at_zero) covers this file's own dispatch path (speaker_visual_cue). The alias script's own explicit-zero habit is separate bash code, not reachable or touched from here."
    },
    "head": {
      "label": "Head axis (move_head)",
      "what_it_does": "Continuous-speed head motor move, not move-to-position.",
      "endpoint": "/api-sdk/move_head",
      "readiness": "ready",
      "wired_for": [],
      "callers_today": [
        "brobots-head-nod (~/.gopod_alias_lib/brobots.sh) - NOT used anywhere in the interview song itself"
      ],
      "explicit_zero_enforced": false,
      "notes": "Real and ready, but this song never dispatches it - nothing in run_section1_full_live_001.py calls move_head, so wired_for is empty for this runner's own scope even though the endpoint itself is fully functional (confirmed reading server.go). The alias script's own use of it (brobots-head-nod) is separate bash code, not reachable or gated from here."
    },
    "wheels": {
      "label": "Wheels axis (move_wheels)",
      "what_it_does": "Continuous-speed drive move, not move-to-position.",
      "endpoint": "/api-sdk/move_wheels",
      "readiness": "not_ready",
      "readiness_reason": "Real, wired Wire-Pod endpoint (confirmed reading server.go) - zero callers anywhere in GOPOD, confirmed by search. No hold/stop convention exists for it because nothing has ever calibrated one; the 0.25s/0.35s/1.2s/2.5s durations used elsewhere are lift/animation-specific guesses that don't transfer to a drive axis.",
      "wired_for": [],
      "callers_today": [],
      "explicit_zero_enforced": false
    },
    "sequences": {
      "speaker_visual_cue": {
        "label": "Speaker visual cue - 4-step lift choreography",
        "axis": "lift",
        "source": "scaffold.speaker_visual_cue - config-driven, not hardcoded, same scaffold this registry lives in",
        "fires_on": "every spoken interview turn, via Robots.say()",
        "reusable": true,
        "notes": "The dissection's own closest example of a real, reusable 'common sequence' - already config-driven and already proven working across every turn in this song. Named here explicitly so a future sequencer/mixer can reach it by name rather than rediscovering it."
      }
    },
    "completion_proof": {
      "measured": false,
      "note": "hold_seconds anywhere in this lane is an ASSERTED duration the runner waits for, not a MEASURED confirmation the physical motion finished. No endpoint in this family (lift/head/wheels) proves completion - the HTTP response returns as soon as the gRPC call is issued. A mixer board must not present hold_seconds as a completion guarantee."
    }
  }
}
```

Template 1 pronunciation mouth-valve block:

This registry is the single authoritative source for interview speech pronunciation —
read by the Python runner's `load_pronunciation_registry()`/`pronunciation_entries()` and
applied by `apply_pronunciation_safety()`, called from `flatten_for_robot_speech()` (the
function that turns `raw_llm_response` into `speech_text` on the real interview speech
path). Confirmed 2026-07-14: the Go TTR side's own general character cleanup
(`removeSpecialCharacters()`, actually in `pkg/wirepod/ttr/kgsim.go` — the
`special_characters_vars.go` path this paragraph used to cite does not exist anywhere in
the current repo, a stale citation now corrected) does **not** affect interview speech
either way — read `pkg/wirepod/sdkapp/server.go`'s own `/api-sdk/say_text` handler
directly: it dispatches straight to `robot.Conn.SayText` with no cleanup call at all.
That Go function only ever fires on the separate wake-word/Knowledge-Graph chat path
(`DoSayText`/`KGSim`). Do not edit it to fix interview pronunciation or character
cleanup; edit this block — and see `flatten_for_robot_speech()`'s own
`apply_universal_character_cleanup()` call, a deliberate Python-side mirror of that same
Go function's universal character mapping (quotes/dashes/ellipsis/etc.), since nothing
Go-side reaches this say path to do it for us.

```json SPEECH_PRONUNCIATION_REGISTRY_001
{
  "registry_id": "SPEECH_PRONUNCIATION_REGISTRY_001",
  "status": "CURRENT_RUNTIME_REGISTRY",
  "display_rule": "Display output keeps canonical labels such as GOPOD and CRUSHN8R.",
  "spoken_rule": "Spoken output may use phonetic forms after action syntax is separated from speakable text.",
  "pronunciation_entries": [
    {"from": "GOPOD", "to": "Gowp-awd"},
    {"from": "Bro! Bot", "to": "brobot"},
    {"from": "Bro! Bots", "to": "brobot"},
    {"from": "GTA", "to": "jee-tee-ay"},
    {"from": "CRUSHN8R", "to": "Crush-Ehnaydir"},
    {"from": "CrushN8r", "to": "Crush-Ehnaydir"},
    {"from": "A.I.", "to": "AY-EYE"},
    {"from": "A.I", "to": "AY-EYE"},
    {"from": "A I", "to": "AY-EYE"},
    {"from": "AI", "to": "AY-EYE"},
    {"from": "OG", "to": "Owjee"},
    {"from": "Aha", "to": "Ah,ha"},
    {"from": "A-ha", "to": "Ah,ha"},
    {"from": "SpaceX", "to": "Space-X"},
    {"from": "Ta-da", "to": "Ta-dah"},
    {"from": "Tada", "to": "Ta-dah"},
    {"from": "flexin'", "to": "flex-in"},
    {"from": "flexin", "to": "flex-in"},
    {"from": "tastic", "to": "tass-stick"},
    {"from": "convert", "to": "cunv-irt"},
    {"from": "cdot", "to": "times"},
    {"from": "Boo-Ya", "to": "Boo-Ya, Hawk, Too-Ya"},
    {"from": "Boo-Yaa", "to": "Boo-Ya, Hawk, Too-Ya"}
  ]
}
```

## 4. Tools

Tools support the meal. Tools are not the meal.

Current supporting tools include:

- `goverlord/runtime/data_gomad/robot/kokoro_voice/gopod_audio_execution_verifier_001.py`
  - Minimum viable audio loop.
  - Raw LLM text to speech route.
  - Wire-Pod or fallback TTS path.
  - `HEARD_AUDIO? (true/false)` confirmation.
  - `audio_verification` JSONL event.

- `goverlord/runtime/data_gomad/robot/kokoro_voice/test_gopod_audio_execution_verifier_001.py`
  - Unit proof for the audio gate.
  - Confirms no success claim without operator-heard confirmation.

- `goverlord/runtime/data_gomad/robot/kokoro_voice/audio_proof_ladder_001.py`
  - No longer present — only a stale compiled `.pyc` remnant remains
    (`voice/__pycache__/audio_proof_ladder_001.cpython-310.pyc`), confirmed 2026-07-30.

- `goverlord/runtime/data_gomad/robot/say_replacements/`
  - Existing robot I/O support files.
  - Useful for robot-safe text, filtering, and Wire-Pod adjacent checks.

- `goverlord/runtime/songs/02_brobots_interview_run/zmisc/`
  - Wire-Pod runtime materials, intent support, config sync, and fallback proof files.

## 5. Execution Flow

Interview execution follows this runtime shape:

```text
Section 1 Card exchange
↓
Brobot 2 visible line
↓
Ollama colour as interviewer/questioner or crystallizer
↓
cleaned rich display to Wire-Pod Logs and runner JSON (same text as flat speech - `channels.channel_1_rich_display`)
↓
flat speech filter at robot-mouth boundary
↓
playAnimationWI action packet composed for Wire-Pod
↓
high-priority Wire-Pod say_text to Brobot 2 ESN
↓
Brobot 1 receives coloured line plus value target silently
↓
Ollama colour as interviewee response
↓
cleaned rich display to Wire-Pod Logs and runner JSON (same text as flat speech - `channels.channel_1_rich_display`)
↓
flat speech filter at robot-mouth boundary
↓
playAnimationWI action packet composed for Wire-Pod
↓
high-priority Wire-Pod say_text to Brobot 1 ESN
↓
HEARD_AUDIO verification
↓
PASS / BLOCKED
```

Audio verification follows this minimum rule:

```text
marked LLM output
↓
speech_text = filtered flat speech from raw rich LLM
↓
route through high-priority Wire-Pod say_text for the target Brobot ESN
↓
playback attempt
↓
HEARD_AUDIO? true/false
↓
PASS only if playback succeeded and HEARD_AUDIO=true
```

## 6. Proof

Current proof mechanisms:

- PASS / BLOCKED reporting.
- Unit tests for the audio verifier.
- JSONL `audio_verification` events.
- Operator `HEARD_AUDIO` confirmation.
- Explicit fallback status when Wire-Pod or local TTS is unavailable.
- Runtime files that separate proof from live claims.

Current canonical log shape:

```json
{
  "event": "audio_verification",
  "tts_engine": "wire-pod | fallback",
  "playback_status": "success | fail | not_attempted",
  "heard_audio": true,
  "final_audio_gate": "PASS",
  "speech": "spoken text",
  "route": "wire-pod | fallback",
  "timestamp": "utc"
}
```

The only authoritative audio success field is:

```text
final_audio_gate=PASS
```

That requires:

```text
playback_status=success
heard_audio=true
```

## 7. Boundaries

Template 1 is only the Brobots / Wire-Pod interview layer.

Out of scope for this card:

- No CHALK work.
- No PLAYHEAD work.
- No GOPOD orchestration.
- No future session engine.
- No cockpit work.
- No dessert layer.
- No agent architecture redesign.
- No robot-expression chain expansion.
- No multi-agent orchestration.
- No upstream history rewrite.

## 8. Completion Criteria

Template 1 is complete when the Brobots / Wire-Pod interview layer has:

- Deterministic interview execution.
- Correct Brobot identities.
- Correct Wire-Pod routing for robot speech.
- Stable fallback TTS behavior when Wire-Pod is unavailable.
- Verified playback on the operator system.
- Operator-heard confirmation logged as JSONL evidence.
- Reproducible PASS evidence for audio delivery.
- Reproducible BLOCKED evidence for failed or unheard audio.
- Stable current modified Wire-Pod interview runtime.
- Clear separation between tools, runtime, and proof.

Completion means the dinner layer is ready enough to become the stable base for Template 2.

## Current Point B

Point B is stable audible Section 1 interview execution.

The current target shape is:

```text
Section 1 Card
-> four Brobot 2 / Brobot 1 exchanges
-> cleaned rich display in Wire-Pod Logs (same text as flat speech - `channels.channel_1_rich_display`)
-> flat filtered robot speech
-> high-priority Wire-Pod say_text
-> heard robot audio
```

Stop there.
