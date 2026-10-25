// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_brobots_settings.go
package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// The Brobots chat's settings live in one file, webroot/gopod_brobots_settings.json,
// read here at Wire-Pod start and by the chat page on load (operator 2026-10-01:
// "offload values like "___" to a file as a single source of truth to refer to,
// rather than getting changed in code by mistake. same with any data sets"). The
// code holds no copy of them. Like animation_vocab.json, the file is required: a
// missing or broken file stops Wire-Pod at start with a clear message, instead of
// running with a marker or a persona nobody chose.
type gopodBrobotsSettingsFile struct {
	AnimationMarker     string            `json:"animation_marker"`
	PlaceholderWord     string            `json:"placeholder_word"`
	PlaceholderThought  string            `json:"placeholder_thought"`
	ReplyExamples       map[string]string `json:"reply_examples"`
	FinalOutputCheck    string            `json:"final_output_check"`
	QuestionFactsHeader string            `json:"question_facts_header"`
	OffTopicNote        string            `json:"off_topic_note"`
	WebLinks            struct {
		KeepWith string `json:"keep_with"`
	} `json:"web_links"`
	StretchedWords struct {
		MinRepeat   int `json:"min_repeat"`
		SaidRepeats int `json:"said_repeats"`
	} `json:"stretched_words"`
	MergedWords struct {
		Dictionaries []string `json:"dictionaries"`
		MinLetters   int      `json:"min_letters"`
		MinPart      int      `json:"min_part"`
		Endings      []string `json:"word_endings"`
		KeepWhole    []string `json:"keep_whole"`
	} `json:"merged_words"`
	ScriptedSnippets struct {
		Lines []struct {
			When   string   `json:"when"`
			Words  []string `json:"words"`
			Brobot string   `json:"brobot"`
			At     string   `json:"at"`
			Lines  []string `json:"lines"`
			Chance float64  `json:"chance"`
		} `json:"lines"`
	} `json:"scripted_snippets"`
	AnimationLeadIn      string   `json:"animation_lead_in"`
	StageDirectionEmojis []string `json:"stage_direction_emojis"`
	SpokenSymbols        struct {
		RepeatingWord string            `json:"repeating_word"`
		HmmSaid       string            `json:"hmm_said"`
		Words         map[string]string `json:"words"`
		// 2026-10-14: "−3" is said "negative 3" (gopodSpeakSymbols).
		NegativeWord string `json:"negative_word"`
	} `json:"spoken_symbols"`
	ScriptedSplit struct {
		SegmentMarker string   `json:"segment_marker"`
		PieceEnds     string   `json:"piece_ends"`
		JoinTrailing  []string `json:"join_trailing"`
	} `json:"scripted_split"`
	LiveSplit struct {
		SentenceEnds string `json:"sentence_ends"`
		CommaEnds    string `json:"comma_ends"`
		// 2026-10-08: short pieces said as one run (gopodJoinPieces).
		JoinUpToWords int  `json:"join_up_to_words"`
		ThoughtLines  bool `json:"thought_lines"`
	} `json:"live_split"`
	Countdown struct {
		Role               string  `json:"role"`
		RattleSoundSeconds float64 `json:"rattle_sound_seconds"`
	} `json:"countdown"`
	Brothers struct {
		AsleepNote string `json:"asleep_note"`
	} `json:"brothers"`
	Addressing struct {
		StartWithinWords int `json:"start_within_words"`
		// 2026-10-21: the first answer in a line to both ends handing over.
		RedirectText      string  `json:"redirect_text"`
		FirstHeardSeconds float64 `json:"first_heard_seconds"`
		QuietAnimation    string  `json:"quiet_animation"`
		// 2026-10-14: "All": Brobot 1 answers, then Brobot 2 (gopodTypedInputAll).
		AllNextWaitSeconds float64 `json:"all_next_wait_seconds"`
		// 2026-10-15: before the question in a follow-up's prompt (gopodFollowingUp).
		FollowUpNote string                         `json:"follow_up_note"`
		Names        map[string]map[string][]string `json:"names"`
	} `json:"addressing"`
	ChatHold struct {
		IdleMinutes   float64  `json:"idle_minutes"`
		ThanksPhrases []string `json:"thanks_phrases"`
		ThanksNote    string   `json:"thanks_note"`
		StuckSeconds  float64  `json:"stuck_seconds"`
	} `json:"chat_hold"`
	MicService struct {
		Script string `json:"script"`
	} `json:"mic_service"`
	NumpadPTT struct {
		MapFile string `json:"map_file"`
		// 2026-10-20: the mic's USB id, and what Goverlord says when it's missing.
		MicUSBIDs      []string `json:"mic_usb_ids"`
		MicMissingText string   `json:"mic_missing_text"`
		// 2026-10-21: the USB watch (gopod_usb_watch.go).
		NumpadUSBIDs       []string          `json:"numpad_usb_ids"`
		USBWatchSeconds    float64           `json:"usb_watch_seconds"`
		MicFoundText       string            `json:"mic_found_text"`
		NumpadMissingText  string            `json:"numpad_missing_text"`
		NumpadFoundText    string            `json:"numpad_found_text"`
		MicDisabledSeconds float64           `json:"mic_disabled_seconds"`
		BodyToBrobot       map[string]string `json:"body_to_brobot"`
		// 2026-10-14: Goverlord says texts.on / texts.exit; made at start (gopodSystemLines).
		Texts map[string]string `json:"texts"`
		// 2026-10-16: 0 held talks to all the Brobots (gopod_numpad_ptt.go).
		AllLabel string `json:"all_label"`
	} `json:"numpad_ptt"`
	RecordSession struct {
		SessionsDir       string   `json:"sessions_dir"`
		FolderSuffix      string   `json:"folder_suffix"`
		MaxMinutes        float64  `json:"max_minutes"`
		AnswerWaitMinutes float64  `json:"answer_wait_minutes"`
		KeepWords         []string `json:"keep_words"`
		DeleteWords       []string `json:"delete_words"`
		AnswerTo          string   `json:"answer_to"`
		AskBrobot         string   `json:"ask_brobot"`
		AskText           string   `json:"ask_text"`
		ConfirmBrobot     string   `json:"confirm_brobot"`
		KeptText          string   `json:"kept_text"`
		DeletedText       string   `json:"deleted_text"`
		TakeSheet         string   `json:"take_sheet"`
	} `json:"record_session"`
	LiveWeatherNote string `json:"live_weather_note"`
	ChatMoves       struct {
		NodEvery       int     `json:"nod_every"`
		NodSpeed       float64 `json:"nod_speed"`
		NodHoldSeconds float64 `json:"nod_hold_seconds"`
		// 2026-10-08: searching while a Brobot waits - at least this long.
		SearchingMinSeconds      float64 `json:"searching_min_seconds"`
		SearchingResendSeconds   float64 `json:"searching_resend_seconds"`
		SearchingIntervalSeconds float64 `json:"searching_interval_seconds"`
		// 2026-10-10: the words said while the model works on a reply.
		ThinkingText string `json:"thinking_text"`
		// 2026-10-09: the moves before "Thinking, please wait." ("arm", "nod").
		ThinkingMoves []string `json:"thinking_moves"`
		// 2026-10-09: speech after an animation, and the wait for its done.
		SpeechAfterAnimationSeconds float64 `json:"speech_after_animation_seconds"`
		AnimationDoneMostSeconds    float64 `json:"animation_done_most_seconds"`
		// 2026-10-24: the golden "Got it" sequence, one copy per Brobot (gopodGotItFor).
		GotIt map[string]gopodGotIt `json:"got_it"`
	} `json:"chat_moves"`
	LLM struct {
		WarmupText string `json:"warmup_text"`
		// 2026-10-16: the opener song (gopod_opener_song.go).
		OpenerSongID       string  `json:"opener_song_id"`
		VoicesText         string  `json:"voices_text"`
		WaitText           string  `json:"wait_text"`
		WarmupCheckSeconds float64 `json:"warmup_check_seconds"`
		ReadyText          string  `json:"ready_text"`
		RunReadyText       string  `json:"run_ready_text"`
		RunReadyEnd        string  `json:"run_ready_end"`
		WarmupDoneText     string  `json:"warmup_done_text"`
		KeepAlive          string  `json:"keep_alive"`
		RefreshSeconds     float64 `json:"refresh_seconds"`
		Temperature        float32 `json:"temperature"`
		SayLineTemperature float32 `json:"say_line_temperature"`
	} `json:"llm"`
	BingoBall struct {
		Role string `json:"role"`
	} `json:"bingo_ball"`
	TouchInterrupt struct {
		OffFile string `json:"off_file"`
	} `json:"touch_interrupt"`
	ExtraAnimations struct {
		MinSentences   int      `json:"min_sentences"`
		AfterSentences []int    `json:"after_sentences"`
		Animations     []string `json:"animations"`
	} `json:"extra_animations"`
	SharedChat struct {
		MaxQuestions          int    `json:"max_questions"`
		MaxMinutes            int    `json:"max_minutes"`
		MaxAnswerChars        int    `json:"max_answer_chars"`
		MaxPerformedLines     int    `json:"max_performed_lines"`
		MaxPerformedLineChars int    `json:"max_performed_line_chars"`
		PerformedNote         string `json:"performed_note"`
		ForgottenNote         string `json:"forgotten_note"`
	} `json:"shared_chat"`
	Bingo struct {
		RunMarker     string            `json:"run_marker"`
		RunGapSeconds float64           `json:"run_gap_seconds"`
		SnapshotNote  string            `json:"snapshot_note"`
		QuestionWord  string            `json:"question_word"`
		NotATest      []string          `json:"not_a_test_phrases"`
		PitchPack     string            `json:"pitch_pack"`
		NoRunNote     string            `json:"no_run_note"`
		Takes         map[string]string `json:"takes"`
	} `json:"bingo"`
	InterviewChapters struct {
		Role     string `json:"role"`
		Label    string `json:"label"`
		Summary  string `json:"summary"`
		Chapters []struct {
			Title   string   `json:"title"`
			Summary string   `json:"summary"`
			Topics  []string `json:"topics"`
		} `json:"chapters"`
		TopicLabels map[string]string `json:"topic_labels"`
	} `json:"interview_chapters"`
	// 2026-10-21: the thought bubbles template (gopod_answer_chain.go).
	AnswerChains map[string]struct {
		From  string   `json:"from"`
		Parts []string `json:"parts"`
	} `json:"answer_chains"`
	// 2026-10-20: the system menu's theme presets (system-menu-theme).
	SystemMenu struct {
		Themes map[string]map[string]string `json:"themes"`
	} `json:"system_menu"`
	// 2026-10-20: the TPU +/- rotation's sources (gopod_visual.go).
	VisualSources struct {
		Ring []struct {
			ID     string `json:"id"`
			Label  string `json:"label"`
			Robot  string `json:"robot"`
			Device string `json:"device"`
		} `json:"ring"`
	} `json:"visual_sources"`
	// 2026-10-19: the 126 sleep beats by family (system-robot beat / beats).
	SleepBeats struct {
		Families map[string][]struct {
			Alias string `json:"alias"`
			Name  string `json:"name"`
			Kind  string `json:"kind"`
		} `json:"families"`
	} `json:"sleep_beats"`
	// 2026-10-18: system-test-all's order (gopod_system_tests.go).
	SystemTests struct {
		Audit []string `json:"audit"`
		Tpu   struct {
			Picture string `json:"picture"`
			Expect  string `json:"expect"`
		} `json:"tpu"`
		Audio struct {
			Sink      string `json:"sink"`
			SourceHas string `json:"source_has"`
		} `json:"audio"`
	} `json:"system_tests"`
	TopicPictures struct {
		Role     string              `json:"role"`
		Folder   string              `json:"folder"`
		Pictures map[string][]string `json:"pictures"`
		// 2026-10-18: eggs never on the egg list (halloween).
		NotListed []string `json:"not_listed"`
		// 2026-10-18: eggs taken out (their key, "chapter:<title>"): never pop up, not on the list.
		NotEggs []string `json:"not_eggs"`
		// 2026-10-22: the only eggs (when set) - gopodNotEgg.
		EggsKeep []string `json:"eggs_keep"`
		// 2026-10-13: a live pop-up's congratulations, shown and said.
		PopupText string `json:"popup_text"`
		// 2026-10-13: how many times a topic's pictures show per Wire-Pod start.
		MaxPerSession int `json:"max_per_session"`
		// 2026-10-24: the silence between the egg voice's pieces (gopod_egg_voice.go).
		EggVoiceGapMS  int `json:"egg_voice_gap_ms"`
		EggVoiceEdgeMS int `json:"egg_voice_edge_ms"`
		// 2026-10-16: each egg's congratulations, {name}: the word that found it.
		EggText string `json:"egg_text"`
		// 2026-10-16: under a zoomable prize (gopodShowEgg).
		EggZoomText string `json:"egg_zoom_text"`
		// 2026-10-16: an egg's name by topic (else the word that found it), and
		// the eggs whose prize is a card, not a picture (the roast battle).
		EggNames map[string]string `json:"egg_names"`
		Cards    map[string]string `json:"cards"`
		// 2026-10-15: pop-up pictures the vision model describes for the reply
		// (gopod_picture_describe.go).
		Describe            []string            `json:"describe"`
		DescribeWaitSeconds float64             `json:"describe_wait_seconds"`
		DescribeNote        string              `json:"describe_note"`
		DescribeOnAsk       map[string][]string `json:"describe_on_ask"`
		DescribeWaitText    string              `json:"describe_wait_text"`
		DescribePrompt      string              `json:"describe_prompt"`
		DescribeDoneText    string              `json:"describe_done_text"`
	} `json:"topic_pictures"`
	ReporterVoices struct {
		Role               string            `json:"role"`
		Python             string            `json:"python"`
		Worker             string            `json:"worker"`
		AudioFolder        string            `json:"audio_folder"`
		Hosts              map[string]string `json:"hosts"`
		SystemVoice        string            `json:"system_voice"`
		Speed              float64           `json:"speed"`
		HFHome             string            `json:"hf_home"`
		SystemSayRole      string            `json:"system_say_role"`
		SongAudioRole      string            `json:"song_audio_role"`
		NarratorRole       string            `json:"narrator_role"`
		NarratorLead       float64           `json:"narrator_lead_seconds"`
		HeardMarker        string            `json:"heard_marker"`
		OperatorHost       string            `json:"operator_host"`
		ReadLinesFrom      []string          `json:"read_lines_from"`
		PageLagSeconds     float64           `json:"page_lag_seconds"`
		MaxWaitSeconds     float64           `json:"max_wait_seconds"`
		HeardMarginSeconds float64           `json:"heard_margin_seconds"`
		// 2026-10-14: "Generating..." while a shown line's voice is made (gopodSayShown: the line waits in its own bubble).
		GeneratingMarker string  `json:"generating_marker"`
		StartWaitSeconds float64 `json:"start_wait_seconds"`
		// 2026-10-17: the words while a voice file is made (also the page's).
		GeneratingText string            `json:"generating_text"`
		SayWords       map[string]string `json:"say_words"`
		ReadBy         map[string]string `json:"read_by"`
		OpeningEndText string            `json:"opening_end_text"`
		GateMaxSeconds float64           `json:"gate_max_seconds"`
		GapSeconds     float64           `json:"gap_seconds"`
		Labels         map[string]string `json:"labels"`
		SystemPersona  string            `json:"system_persona"`
		SystemSayWords map[string]string `json:"system_say_words"`
		// 2026-10-13: the voice that reads a live operator line.
		OperatorVoice string `json:"operator_voice"`
		// 2026-10-13: days an unused Operator voice file is kept.
		OperatorKeepDays  float64 `json:"operator_keep_days"`
		FinishWaitSeconds float64 `json:"finish_wait_seconds"`
		ScriptedKeepDays  float64 `json:"scripted_keep_days"`
		LiveKeepDays      float64 `json:"live_keep_days"`
		// 2026-10-13: the sting before a system line (gopodSystemSting).
		SystemStings map[string]struct {
			File string `json:"file"`
			// 2026-10-17: "" (default) or a percentage of chat_audio.volume ("40%").
			Volume json.RawMessage `json:"volume"`
		} `json:"system_stings"`
		Opening []gopodHostLine `json:"opening"`
		Closing []gopodHostLine `json:"closing"`
	} `json:"reporter_voices"`
	// 2026-10-17 (operator: "One master default knobs setting all other knobs
	// refers to"; "evaluate audio files to normalize all volumes to -3db"): the
	// master volume (the page's), and the level every sound is evened to first.
	// 2026-10-17: GOPOD's own web search (gopod_system_search.go).
	SystemSearch struct {
		URL            string  `json:"url"`
		Results        int     `json:"results"`
		TimeoutSeconds float64 `json:"timeout_seconds"`
		Note           string  `json:"note"`
		NoneNote       string  `json:"none_note"`
		Line           string  `json:"line"`
		ResultsRole    string  `json:"results_role"`
	} `json:"system_search"`
	// 2026-10-17: system-set's terminal bubble (gopod_system_set.go).
	SystemSet struct {
		Role string `json:"role"`
	} `json:"system_set"`
	// 2026-10-18: CHALK and PLAYHEAD take system asks (gopod_managers.go).
	Managers struct {
		Takes          []gopodManagerTake `json:"takes"`
		ReleaseSeconds float64            `json:"release_seconds"`
		TimePlaceText  string             `json:"time_place_text"`
		DisplayHost    string             `json:"display_host"`
		QRNote         string             `json:"qr_note"`
		QRDisplayText  string             `json:"qr_display_text"`
		// 2026-10-18: the reporters' live personas (gopod_personas.go).
		Personas map[string]struct {
			Host      string `json:"host"`
			Reporter  string `json:"reporter"`
			Prompt    string `json:"prompt"`
			LinesNote string `json:"lines_note"`
			WhoLine   string `json:"who_line"`
		} `json:"personas"`
		WhoWords []string `json:"who_words"`
	} `json:"managers"`
	// 2026-10-18: a native intent's silent system bubble (gopodNativeBubble).
	// 2026-10-22: after his weather line, its picture and sound, then a pause.
	WeatherCard struct {
		PauseSeconds float64 `json:"pause_seconds"`
		Picture      bool    `json:"picture"`
	} `json:"weather_card"`
	// 2026-10-22: the weather line's formats, per answer position (gopodWeatherReportLine).
	WeatherLines map[string]string `json:"weather_lines"`
	// 2026-10-22: the answer chain's prompt lines (gopodChainPrompt).
	ChainPrompts map[string]string `json:"chain_prompts"`
	// 2026-10-22: topics that get their own answer chain group (gopodTopicGroupFor).
	TopicGroups map[string]struct {
		Topics   []string `json:"topics"`
		RuleLine string   `json:"rule_line"`
	} `json:"topic_groups"`
	// 2026-10-23: the point bubbles part (gopod_point_bubbles.go).
	PointBubbles struct {
		Prompt   string `json:"prompt"`
		Fallback struct {
			Title string `json:"title"`
			Note  string `json:"note"`
		} `json:"fallback"`
		Hype          map[string][]string `json:"hype"`
		KeepLeadLines int                 `json:"keep_lead_lines"`
	} `json:"point_bubbles"`
	// 2026-10-23 (A3): the game's rules, for the Brobots (prebuild_context {game}).
	Game struct {
		Rules string `json:"rules"`
		// 2026-10-23 (A4): the grand prize line, and when the game resets.
		WinText string `json:"win_text"`
		// 2026-10-24: the Master Grand Prize - hidden eggs too.
		MasterText string `json:"master_text"`
		ResetOn    string `json:"reset_on"`
	} `json:"game"`
	// 2026-10-23: exact answers to golden questions (the golden_line part).
	GoldenAnswers []struct {
		Queries []string `json:"queries"`
		Line    string   `json:"line"`
	} `json:"golden_answers"`
	// 2026-10-23 (A5): the interview song built from the game (gopod_interview_from_game.go).
	InterviewBuild struct {
		Cues       map[string]string `json:"cues"`
		Reporters  []string          `json:"reporters"`
		DefaultAsk string            `json:"default_ask"`
	} `json:"interview_build"`
	// 2026-10-23 (template sweep): the host persona's sentences (gopod_brobot_personas.go).
	HostPersonaFrames struct {
		Name      string `json:"name"`
		Steps     string `json:"steps"`
		Signature string `json:"signature"`
		Dial      string `json:"dial"`
	} `json:"host_persona_frames"`
	// 2026-10-23 (template sweep): words that carry no meaning of their own (gopod_brother_repeat.go).
	BrotherRepeat struct {
		PlainWords []string `json:"plain_words"`
	} `json:"brother_repeat"`
	// 2026-10-24: good morning / afternoon / evening / night (gopod_greetings.go).
	Greetings struct {
		Parts []struct {
			Name string `json:"name"`
			From int    `json:"from"`
			To   int    `json:"to"`
		} `json:"parts"`
		RightText string `json:"right_text"`
		WrongText string `json:"wrong_text"`
	} `json:"greetings"`
	// 2026-10-24: GOPOD as a phone app - the secure page (gopod_https.go).
	App struct {
		HTTPSPort int    `json:"https_port"`
		CertDir   string `json:"cert_dir"`
		CAName    string `json:"ca_name"`
	} `json:"app"`
	// 2026-10-24: visitors and the chat sign-up (gopod_visitors.go).
	Visitors struct {
		Participation      bool    `json:"participation"`
		AdminAddressesFile string  `json:"admin_addresses_file"`
		ProfilesFile       string  `json:"profiles_file"`
		SeenSeconds        float64 `json:"seen_seconds"`
		NameMax            int     `json:"name_max"`
		WaitingText        string  `json:"waiting_text"`
		MutedText          string  `json:"muted_text"`
		ApprovedText       string  `json:"approved_text"`
		ApprovedShowMS     int     `json:"approved_show_ms"`
	} `json:"visitors"`
	// 2026-10-23 (template sweep): the session memory's lines (gopod_session_memory.go).
	SessionMemory struct {
		Line      string `json:"line"`
		Scenario  string `json:"scenario"`
		EmptyText string `json:"empty_text"`
	} `json:"session_memory"`
	// 2026-10-23: the summary answer type (gopod_summary_answer.go).
	Summary struct {
		Queries    []string `json:"queries"`
		MinTopics  int      `json:"min_topics"`
		MaxTopics  int      `json:"max_topics"`
		TopicOrder []string `json:"topic_order"`
		FactsNote  string   `json:"facts_note"`
		Fallback   string   `json:"fallback"`
	} `json:"summary"`
	// 2026-10-25: the response-chain template (gopod_response_chains.go).
	ResponseChains map[string]gopodResponseChain `json:"response_chains"`
	Openers        map[string]string             `json:"openers"`
	Nicknames      struct {
		Titles []string `json:"titles"`
		List   []string `json:"list"`
	} `json:"nicknames"`
	// 2026-10-23: the golden answer shapes (gopod_answer_shapes.go).
	AnswerShapes struct {
		Shapes    map[string][]string `json:"shapes"`
		Marks     map[string]string   `json:"marks"`
		MarkOrder []string            `json:"mark_order"`
	} `json:"answer_shapes"`
	// 2026-10-22: the typed line's All menu - who answers first by default.
	TypedLine struct {
		AllFirst string `json:"all_first"`
	} `json:"typed_line"`
	// 2026-10-22: the colour slots' cap (gopod_answer_chain.go).
	ColourSlots struct {
		MaxSentences int `json:"max_sentences"`
	} `json:"colour_slots"`
	// 2026-10-22: the pre-build context in every Brobot answer (gopodPrebuildContext).
	PrebuildContext struct {
		Header string   `json:"header"`
		Lines  []string `json:"lines"`
	} `json:"prebuild_context"`
	// 2026-10-22: the Brobot Bubble's steps, in order (gopodBubbleStep).
	BrobotBubble struct {
		Steps []string `json:"steps"`
	} `json:"brobot_bubble"`
	// 2026-10-22: what goes in each notice in a Brobot's bubble (gopod_value_points_notice.go).
	BubbleNotices struct {
		ValuePoints struct {
			Title   string   `json:"title"`
			Parts   []string `json:"parts"`
			Points  string   `json:"points"`
			Weather string   `json:"weather"`
			Search  string   `json:"search"`
			None    string   `json:"none"`
			GotIt   string   `json:"got_it"`
			Note    string   `json:"note"`
			// 2026-10-22: one notice per bubble - the egg's line; the game lines.
			Egg      string `json:"egg"`
			Combo    string `json:"combo"`
			Chapter  string `json:"chapter"`
			Recovery string `json:"recovery"`
			Next     string `json:"next"`
			Summary  string `json:"summary"`
			Score    string `json:"score"`
			// 2026-10-22: the notice log's 4th line (gopodNoticeGotItAgain).
			GotItAgain  string   `json:"got_it_again"`
			SecondGotIt string   `json:"second_got_it"`
			HelperFrom  []string `json:"helper_from"`
		} `json:"value_points"`
		Egg          gopodNoticeWords `json:"egg"`
		NativeIntent gopodNoticeWords `json:"native_intent"`
	} `json:"bubble_notices"`
	NativeIntent struct {
		Role        string  `json:"role"`
		Text        string  `json:"text"`
		OnceSeconds float64 `json:"once_seconds"`
	} `json:"native_intent"`
	// 2026-10-18: the engine times itself by the page (gopod_page_sync.go).
	PageSync struct {
		ReplyWaitSeconds float64 `json:"reply_wait_seconds"`
		HeldLogSeconds   float64 `json:"held_log_seconds"`
		MainLeaseSeconds float64 `json:"main_lease_seconds"`
		FeedMax          int     `json:"feed_max"`
	} `json:"page_sync"`
	ChatAudio struct {
		Volume       float64 `json:"volume"`
		NormalizeDB  float64 `json:"normalize_db"`
		LoudnessLUFS float64 `json:"loudness_lufs"`
	} `json:"chat_audio"`
	ValuePoints struct {
		ScoreRole            string            `json:"score_role"`
		ScoreText            string            `json:"score_text"`
		Reminders            map[string]string `json:"reminders"`
		WhatElseTopic        string            `json:"what_else_topic"`
		WhatElseCount        int               `json:"what_else_count"`
		ReminderSkip         int               `json:"reminder_skip_next_topics"`
		ReminderChaptersText string            `json:"reminder_chapters_text"`
		ReminderTopicsText   string            `json:"reminder_topics_text"`
		ReminderTopicCount   int               `json:"reminder_topic_count"`
		// 2026-10-17: the eggs that aren't a topic's (gopod_egg_list.go).
		ReminderEggs []struct {
			Key   string `json:"key"`
			Label string `json:"label"`
		} `json:"reminder_eggs"`
		ReminderEggCount int               `json:"reminder_egg_count"`
		SayLines         map[string]string `json:"say_lines"`
		// 2026-10-22: points in note form, "topic N" -> "title | note | note".
		Notes map[string]string `json:"notes"`
		// 2026-10-22: words left out when matching a question to a point.
		PickSkipWords []string `json:"pick_skip_words"`
		// 2026-10-22: topics the Brobots know, never a value point target.
		BackgroundTopics []string `json:"background_topics"`
		// 2026-10-23: what the game counts - "topics" (one point opens a topic) or "points".
		Target string `json:"target"`
		// 2026-10-23 (A3): the value points themselves, and the seven core topics.
		CorePoints []string            `json:"core_points"`
		CoreTopics map[string][]string `json:"core_topics"`
		// 2026-10-23: the answer chain map - which topics lead to which.
		TopicLinks map[string][]string `json:"topic_links"`
		// 2026-10-23: the funnels - lists of topics in order (gopodTopicLinksFor).
		Funnels [][]string `json:"funnels"`
	} `json:"value_points"`
	ReplyTighten struct {
		BudgetOff           bool           `json:"budget_off"`
		BaseSentences       int            `json:"base_sentences"`
		PerTopicSentences   int            `json:"per_topic_sentences"`
		TopicSentences      map[string]int `json:"topic_sentences"`
		MoreWords           []string       `json:"more_words"`
		MoreSentences       int            `json:"more_sentences"`
		MaxSentences        int            `json:"max_sentences"`
		FreeMaxSentences    int            `json:"free_max_sentences"`
		SayLineSentences    int            `json:"say_line_sentences"`
		DropSentencesWith   []string       `json:"drop_sentences_with"`
		NotSentenceEnds     []string       `json:"not_sentence_ends"`
		DropLists           bool           `json:"drop_lists"`
		DropStageDirections bool           `json:"drop_stage_directions"`
		BrandWords          []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"brand_words"`
	} `json:"reply_tighten"`
	InterviewScript struct {
		// 2026-10-09: the interview's song folder (gopod_interview_song.go).
		SongID     string  `json:"song_id"`
		DraftsDir  string  `json:"drafts_dir"`
		Draft      string  `json:"draft"`
		PlayWords  string  `json:"play_words"`
		StopWords  string  `json:"stop_words"`
		GapSeconds float64 `json:"gap_seconds"`
		// 2026-10-08: the played draft's own opening/closing host lines stand in
		// for reporter_voices.opening/closing (the record cut's script).
		BookendsFromDraft bool `json:"bookends_from_draft"`
		// 2026-10-08: the reporters' lines after "Draft run started.", before question 1.
		StartLines []struct {
			Who  string `json:"who"`
			Text string `json:"text"`
		} `json:"start_lines"`
		LLMReadyText string `json:"llm_ready_text"`
	} `json:"interview_script"`
	Personas map[string]string `json:"personas"`
	// 2026-10-21: the Doc & Pip persona template (gopod_brobot_personas.go).
	HostPersonas map[string]gopodHostPersona `json:"host_personas"`
	// 2026-10-21: each pair of hosts - partner, banter, stereo dial.
	HostPairs map[string]gopodHostPair `json:"host_pairs"`
	// 2026-10-08: lines said after the "thanks Brobots" sign-off.
	SignOff struct {
		Lines []struct {
			Who  string `json:"who"`
			Text string `json:"text"`
		} `json:"lines"`
		SystemRole          string  `json:"system_role"`
		SystemText          string  `json:"system_text"`
		StartedText         string  `json:"started_text"`
		WelcomeText         string  `json:"welcome_text"`
		NothingRecordedText string  `json:"nothing_recorded_text"`
		SafeSpotSeconds     float64 `json:"safe_spot_seconds"`
	} `json:"sign_off"`
	SoundsOn       string   `json:"sounds_on"`
	PageSoundFiles []string `json:"page_sound_files"`
	WordAnimations struct {
		PlayOnWords          []string          `json:"play_on_words"`
		PlayOnWordsAnimation string            `json:"play_on_words_animation"`
		Swap                 map[string]string `json:"swap"`
	} `json:"word_animations"`
	GOPODYourself struct {
		Phrases       []string          `json:"phrases"`
		Lines         map[string]string `json:"lines"`
		RepeatSeconds float64           `json:"repeat_seconds"`
	} `json:"gopod_yourself"`
	SongStop struct {
		StopWords   []string `json:"stop_words"`
		SafeScript  string   `json:"safe_script"`
		WaitSeconds float64  `json:"wait_seconds"`
	} `json:"song_stop"`
	// 2026-10-10: Wire-Pod's own voice answers on or off (gopod_native.go).
	Native struct {
		OnAtStart bool     `json:"on_at_start"`
		AlwaysOn  []string `json:"always_on"`
		// 2026-10-21: native intents a manager takes instead (gopodNativeToManager).
		ToManager map[string]string `json:"to_manager"`
	} `json:"native"`
	// 2026-10-10: the session and its runs (gopod_session.go).
	Session struct {
		BusyText       string   `json:"busy_text"`
		RunGateText    string   `json:"run_gate_text"`
		StoppedText    string   `json:"stopped_text"`
		NoRunText      string   `json:"no_run_text"`
		StartedText    string   `json:"started_text"`
		CompleteText   string   `json:"complete_text"`
		NotFoundText   string   `json:"not_found_text"`
		ReplayWord     string   `json:"replay_word"`
		ReplayName     string   `json:"replay_name"`
		StopSession    []string `json:"stop_session_words"`
		PausedText     string   `json:"paused_text"`
		ContinuesText  string   `json:"continues_text"`
		PauseAfterMS   int      `json:"pause_after_ms"`
		PageHiddenFile string   `json:"page_hidden_file"`
		// 2026-10-17: the chat page away (gopod_page_presence.go).
		AwaitsText         string            `json:"awaits_text"`
		PagePresentSeconds float64           `json:"page_present_seconds"`
		RunNames           map[string]string `json:"run_names"`
	} `json:"session"`
	NativeCube struct {
		Intents []struct {
			Intent  string   `json:"intent"`
			Phrases []string `json:"phrases"`
		} `json:"intents"`
	} `json:"native_cube"`
	// 2026-10-08: the live weather a draft answer can say (gopod_live_weather.go).
	LiveWeather struct {
		Location        string  `json:"location"`
		CacheSeconds    float64 `json:"cache_seconds"`
		OfflineText     string  `json:"offline_text"`
		MilitaryLabel   string  `json:"military_label"`
		FollowUpKitLine string  `json:"follow_up_kit_line"`
	} `json:"live_weather"`
	Photo struct {
		AskPhrases         []string `json:"ask_phrases"`
		RecallPhrases      []string `json:"recall_phrases"`
		VisionModel        string   `json:"vision_model"`
		Prompt             string   `json:"prompt"`
		KeepAlive          string   `json:"keep_alive"`
		DescribeMaxSeconds float64  `json:"describe_max_seconds"`
		AnswerText         string   `json:"answer_text"`
		RecallText         string   `json:"recall_text"`
		NothingText        string   `json:"nothing_text"`
		FailText           string   `json:"fail_text"`
		UnclearText        string   `json:"unclear_text"`
		AwakeSeconds       float64  `json:"awake_seconds"`
		// 2026-10-16: two rounds - the Edge TPU's quick look, then the vision
		// model's details (gopod_photo_quick_look.go).
		QuickLook struct {
			Enabled        bool    `json:"enabled"`
			Python         string  `json:"python"`
			Worker         string  `json:"worker"`
			Model          string  `json:"model"`
			Labels         string  `json:"labels"`
			MinScore       float64 `json:"min_score"`
			TimeoutSeconds float64 `json:"timeout_seconds"`
			// 2026-10-17: the enhanced snapshot and its boxes (gopodQuickLookPictures).
			Pictures bool               `json:"pictures"`
			Enhance  map[string]float64 `json:"enhance"`
		} `json:"quick_look"`
		QuickText           string   `json:"quick_text"`
		QuickNoneText       string   `json:"quick_none_text"`
		GatheringText       string   `json:"gathering_text"`
		DetailText          string   `json:"detail_text"`
		DetailUnclearText   string   `json:"detail_unclear_text"`
		DetailStripPrefixes []string `json:"detail_strip_prefixes"`
		StallLines          []string `json:"stall_lines"`
		DetailPrompt        string   `json:"detail_prompt"`
		EggName             string   `json:"egg_name"`
	} `json:"photo"`
	// 2026-10-16: clean GOPOD - flags and politics (gopod_clean_politics.go).
	CleanPolitics struct {
		AllowedTopics []string          `json:"allowed_topics"`
		Words         []string          `json:"words"`
		Countries     map[string]string `json:"countries"`
		// 2026-10-23: a country alone asks for its capital (gopodWeatherCountry).
		Capitals   map[string]string `json:"capitals"`
		Continents []string          `json:"continents"`
	} `json:"clean_politics"`
	// 2026-10-15: the battery watchdog (gopod_battery_watch.go).
	BatteryWatch struct {
		Enabled         bool                `json:"enabled"`
		IntervalSeconds float64             `json:"interval_seconds"`
		LowVolts        float64             `json:"low_volts"`
		RecoverVolts    float64             `json:"recover_volts"`
		Lines           map[string][]string `json:"lines"`
	} `json:"battery_watch"`
	// 2026-10-14: the Brobots' math, worked out exactly in code (gopod_math.go).
	Math struct {
		Enabled              bool              `json:"enabled"`
		Trigger              string            `json:"trigger"`
		Bullet               string            `json:"bullet"`
		FlavourSentences     int               `json:"flavour_sentences"`
		PopupTitle           string            `json:"popup_title"`
		EggName              string            `json:"egg_name"`
		FractionWords        []string          `json:"fraction_words"`
		DecimalFollowUpLine  string            `json:"decimal_follow_up_line"`
		FollowUpNote         string            `json:"follow_up_note"`
		MaxRepeatDigits      int               `json:"max_repeat_digits"`
		LongDivisionMaxSteps int               `json:"long_division_max_steps"`
		PlaceNames           []string          `json:"place_names"`
		Lines                map[string]string `json:"lines"`
	} `json:"math"`
}

// gopodHostLine is one reporter host's line, said word for word.
type gopodHostLine struct {
	Host string `json:"host"`
	Text string `json:"text"`
}

var gopodBrobotsSettingsPath = gopodPathWithEnvOverride("GOPOD_BROBOTS_SETTINGS_PATH", "webroot/gopod_brobots_settings.json")

var gopodBrobotsSettings = loadGOPODBrobotsSettings()

func loadGOPODBrobotsSettings() gopodBrobotsSettingsFile {
	var s gopodBrobotsSettingsFile
	data, err := os.ReadFile(gopodBrobotsSettingsPath)
	if err != nil {
		panic("GOPOD Brobots settings missing at " + gopodBrobotsSettingsPath + ": " + err.Error())
	}
	if err := json.Unmarshal(data, &s); err != nil {
		panic("GOPOD Brobots settings unreadable at " + gopodBrobotsSettingsPath + ": " + err.Error())
	}
	// 2026-10-08: the topics' parts come from the chapters folder (gopod_chapters_folder.go).
	gopodApplyChaptersFolder(&s)
	var missing []string
	check := func(ok bool, name string) {
		if !ok {
			missing = append(missing, name)
		}
	}
	check(strings.TrimSpace(s.AnimationMarker) != "", "animation_marker")
	check(strings.TrimSpace(s.PlaceholderWord) != "", "placeholder_word")
	check(strings.TrimSpace(s.PlaceholderThought) != "", "placeholder_thought")
	check(strings.TrimSpace(s.ReplyExamples["Brobot 1"]) != "" && strings.TrimSpace(s.ReplyExamples["Brobot 2"]) != "", "reply_examples")
	check(strings.TrimSpace(s.FinalOutputCheck) != "", "final_output_check")
	check(strings.TrimSpace(s.QuestionFactsHeader) != "", "question_facts_header")
	check(strings.TrimSpace(s.OffTopicNote) != "", "off_topic_note")
	check(strings.TrimSpace(s.WebLinks.KeepWith) != "", "web_links.keep_with")
	check(s.StretchedWords.MinRepeat > 1 && s.StretchedWords.SaidRepeats > 0, "stretched_words")
	check(s.MergedWords.MinLetters > 0 && s.MergedWords.MinPart > 0, "merged_words")
	for _, l := range s.ScriptedSnippets.Lines {
		check(strings.TrimSpace(l.When) != "" && (len(l.Lines) > 0 || l.When == "reminder") && (l.At == "" || l.At == "start" || l.At == "end"), "scripted_snippets.lines")
	}
	check(len(s.StageDirectionEmojis) > 0, "stage_direction_emojis")
	check(s.ScriptedSplit.SegmentMarker != "", "scripted_split.segment_marker")
	check(s.ScriptedSplit.PieceEnds != "", "scripted_split.piece_ends")
	check(s.LiveSplit.SentenceEnds != "" && s.LiveSplit.CommaEnds != "", "live_split")
	check(strings.TrimSpace(s.Countdown.Role) != "", "countdown.role")
	check(strings.TrimSpace(s.TouchInterrupt.OffFile) != "", "touch_interrupt.off_file")
	check(strings.TrimSpace(s.BingoBall.Role) != "", "bingo_ball.role")
	check(strings.TrimSpace(s.Brothers.AsleepNote) != "", "brothers.asleep_note")
	check(s.Addressing.StartWithinWords > 0, "addressing.start_within_words")
	check(s.Addressing.FirstHeardSeconds > 0, "addressing.first_heard_seconds")
	check(strings.TrimSpace(s.Addressing.QuietAnimation) != "", "addressing.quiet_animation")
	check(s.ChatHold.IdleMinutes > 0, "chat_hold.idle_minutes")
	check(s.ChatMoves.NodEvery > 0, "chat_moves.nod_every")
	check(s.ChatMoves.NodSpeed > 0, "chat_moves.nod_speed")
	check(s.ChatMoves.NodHoldSeconds > 0, "chat_moves.nod_hold_seconds")
	check(len(s.ChatHold.ThanksPhrases) > 0, "chat_hold.thanks_phrases")
	check(s.ChatHold.StuckSeconds > 0, "chat_hold.stuck_seconds")
	check(strings.TrimSpace(s.ChatHold.ThanksNote) != "", "chat_hold.thanks_note")
	check(strings.TrimSpace(s.RecordSession.SessionsDir) != "", "record_session.sessions_dir")
	check(strings.TrimSpace(s.RecordSession.FolderSuffix) != "", "record_session.folder_suffix")
	check(strings.TrimSpace(s.MicService.Script) != "", "mic_service.script")
	check(strings.TrimSpace(s.NumpadPTT.MapFile) != "", "numpad_ptt.map_file")
	check(len(s.NumpadPTT.BodyToBrobot) > 0, "numpad_ptt.body_to_brobot")
	check(s.RecordSession.MaxMinutes > 0, "record_session.max_minutes")
	check(s.RecordSession.AnswerWaitMinutes > 0, "record_session.answer_wait_minutes")
	check(len(s.RecordSession.KeepWords) > 0, "record_session.keep_words")
	check(len(s.RecordSession.DeleteWords) > 0, "record_session.delete_words")
	check(strings.TrimSpace(s.RecordSession.AnswerTo) != "", "record_session.answer_to")
	check(gopodBrobotESN(s.RecordSession.AskBrobot) != "", "record_session.ask_brobot")
	check(strings.TrimSpace(s.RecordSession.AskText) != "", "record_session.ask_text")
	check(gopodBrobotESN(s.RecordSession.ConfirmBrobot) != "", "record_session.confirm_brobot")
	check(strings.TrimSpace(s.RecordSession.KeptText) != "", "record_session.kept_text")
	check(strings.TrimSpace(s.RecordSession.DeletedText) != "", "record_session.deleted_text")
	check(strings.TrimSpace(s.LLM.WarmupText) != "", "llm.warmup_text")
	_, keepAliveErr := time.ParseDuration(strings.TrimSpace(s.LLM.KeepAlive))
	check(keepAliveErr == nil, "llm.keep_alive")
	check(s.LLM.RefreshSeconds > 0, "llm.refresh_seconds")
	check(s.ExtraAnimations.MinSentences > 0, "extra_animations.min_sentences")
	check(len(s.ExtraAnimations.Animations) > 0, "extra_animations.animations")
	check(s.SharedChat.MaxQuestions > 0, "shared_chat.max_questions")
	check(s.SharedChat.MaxMinutes > 0, "shared_chat.max_minutes")
	check(s.SharedChat.MaxAnswerChars > 0, "shared_chat.max_answer_chars")
	check(s.SharedChat.MaxPerformedLines > 0, "shared_chat.max_performed_lines")
	check(s.SharedChat.MaxPerformedLineChars > 0, "shared_chat.max_performed_line_chars")
	check(strings.TrimSpace(s.SharedChat.PerformedNote) != "", "shared_chat.performed_note")
	check(strings.TrimSpace(s.SharedChat.ForgottenNote) != "", "shared_chat.forgotten_note")
	check(strings.TrimSpace(s.Bingo.RunMarker) != "", "bingo.run_marker")
	check(s.Bingo.RunGapSeconds > 0, "bingo.run_gap_seconds")
	check(strings.TrimSpace(s.Bingo.SnapshotNote) != "", "bingo.snapshot_note")
	check(strings.TrimSpace(s.Bingo.QuestionWord) != "", "bingo.question_word")
	check(strings.TrimSpace(s.Bingo.PitchPack) != "", "bingo.pitch_pack")
	check(strings.TrimSpace(s.Bingo.NoRunNote) != "", "bingo.no_run_note")
	for _, name := range []string{"Brobot 1", "Brobot 2"} {
		check(strings.TrimSpace(s.Bingo.Takes[name]) != "", "bingo.takes."+name)
		check(strings.TrimSpace(s.Personas[name]) != "", "personas."+name)
		check(strings.TrimSpace(s.HostPersonas[name].Name) != "", "host_personas."+name+".name")
		check(len(s.Addressing.Names[name]["brobot"]) > 0, "addressing.names."+name+".brobot")
	}
	check(len(s.ResponseChains["single"].Answers) == 1, "response_chains.single.answers")
	check(len(s.ResponseChains["duo"].Answers) == 2, "response_chains.duo.answers")
	if len(missing) > 0 {
		panic(fmt.Sprintf("GOPOD Brobots settings at %s are missing: %s", gopodBrobotsSettingsPath, strings.Join(missing, ", ")))
	}
	return s
}

// gopodSecondsToDuration turns a settings value in seconds into a time.Duration.
func gopodSecondsToDuration(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}

// ## EOF
