# LLM warm schema


---

## Required steps, in order
1. warm_message: the System llm bubble's message, shown and said ("Voice files generation check.")
2. warm_message: the second message ("Voice files generated. Language Model Loading")
3. warm_wait: waits until the language model is loaded
4. llm_song: the _llm_song run - the lines between "Voices ready" and "LLM ready"
5. warm_ready: the ready message, shown and said
6. sting: the system sting

## Per run, after the opener
1. smart warm: a song with `"llm_warmup": true` checks the language model first (Ollama's loaded list). Still loaded from the opener: the loading lines are bypassed - Goverlord says the song's own `llm_ready_text` and `llm.run_ready_end`, its `warmup_lines`, and the song goes on. Not loaded: the full warm-up, as the opener does.
2. a song that needs no language model: `"llm_required": false` - "No Language Model Required" and its ready line, no model called.
3. `llm.smart_warm`: false in the settings turns the bypass off (the full warm-up every run).
