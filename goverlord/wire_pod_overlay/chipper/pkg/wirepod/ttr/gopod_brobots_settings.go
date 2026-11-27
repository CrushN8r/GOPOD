// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_brobots_settings.go
package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// The Brobots chat's settings live in one folder, webroot/gopod_settings/, read here at Wire-Pod
// start and by the chat page on load.
type gopodBrobotsSettingsFile struct {
	AnimationMarker     string            `json:"animation_marker"`
	PlaceholderWord     string            `json:"placeholder_word"`
	PlaceholderThought  string            `json:"placeholder_thought"`
	ReplyExamples       map[string]string `json:"reply_examples"`
	FinalOutputCheck    string            `json:"final_output_check"`
	QuestionFactsHeader string            `json:"question_facts_header"`
	OffTopicNote        string            `json:"off_topic_note"`
	TopicAsks           struct {
		DetailWords   []string `json:"detail_words"`
		SummaryWords  []string `json:"summary_words"`
		BareMaxWords  int      `json:"bare_max_words"`
		TheTopic      bool     `json:"the_topic"`
		PreDefault    string   `json:"pre_default"`
		PostDefault   string   `json:"post_default"`
		SummaryNotice string   `json:"summary_notice"`
		SummaryFacts  string   `json:"summary_facts"`
	} `json:"topic_asks"`
	WebLinks struct {
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
	GameQuestions    []gopodGameQuestion `json:"game_questions,omitempty"`
	ScriptedSnippets struct {
		Lines []gopodSnippetLine `json:"lines"`
	} `json:"scripted_snippets"`
	AnimationLeadIn      string   `json:"animation_lead_in"`
	StageDirectionEmojis []string `json:"stage_direction_emojis"`
	SpokenSymbols        struct {
		RepeatingWord string            `json:"repeating_word"`
		HmmSaid       string            `json:"hmm_said"`
		ShowAsSaid    []string          `json:"show_as_said"`
		Words         map[string]string `json:"words"`
		NegativeWord  string            `json:"negative_word"`
	} `json:"spoken_symbols"`
	ScriptedSplit struct {
		SegmentMarker string   `json:"segment_marker"`
		PieceEnds     string   `json:"piece_ends"`
		JoinTrailing  []string `json:"join_trailing"`
	} `json:"scripted_split"`
	LiveSplit struct {
		SentenceEnds  string   `json:"sentence_ends"`
		KeepWhole     []string `json:"keep_whole"`
		CommaEnds     string   `json:"comma_ends"`
		JoinUpToWords int      `json:"join_up_to_words"`
		ThoughtLines  bool     `json:"thought_lines"`
	} `json:"live_split"`
	Countdown struct {
		Role               string  `json:"role"`
		RattleSoundSeconds float64 `json:"rattle_sound_seconds"`
	} `json:"countdown"`
	Brothers struct {
		AsleepNote string `json:"asleep_note"`
	} `json:"brothers"`
	Addressing struct {
		StartWithinWords   int                            `json:"start_within_words"`
		RedirectText       string                         `json:"redirect_text"`
		FirstHeardSeconds  float64                        `json:"first_heard_seconds"`
		QuietAnimation     string                         `json:"quiet_animation"`
		AllNextWaitSeconds float64                        `json:"all_next_wait_seconds"`
		FollowUpNote       string                         `json:"follow_up_note"`
		Names              map[string]map[string][]string `json:"names"`
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
		MapFile            string            `json:"map_file"`
		Layout             string            `json:"layout"`
		MicUSBIDs          []string          `json:"mic_usb_ids"`
		MicMissingText     string            `json:"mic_missing_text"`
		NumpadUSBIDs       []string          `json:"numpad_usb_ids"`
		USBWatchSeconds    float64           `json:"usb_watch_seconds"`
		MicFoundText       string            `json:"mic_found_text"`
		NumpadMissingText  string            `json:"numpad_missing_text"`
		NumpadFoundText    string            `json:"numpad_found_text"`
		USBHost            string            `json:"usb_host"`
		MicDisabledSeconds float64           `json:"mic_disabled_seconds"`
		BodyToBrobot       map[string]string `json:"body_to_brobot"`
		Texts              map[string]string `json:"texts"`
		AllLabel           string            `json:"all_label"`
		Any                struct {
			Label        string `json:"label"`
			DefaultOwner string `json:"default_owner"`
			RedirectText string `json:"redirect_text"`
		} `json:"any"`
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
		NodEvery                    int                   `json:"nod_every"`
		NodSpeed                    float64               `json:"nod_speed"`
		NodHoldSeconds              float64               `json:"nod_hold_seconds"`
		ThinkingText                string                `json:"thinking_text"`
		ThinkingMoves               []string              `json:"thinking_moves"`
		SpeechAfterAnimationSeconds float64               `json:"speech_after_animation_seconds"`
		AnimationDoneMostSeconds    float64               `json:"animation_done_most_seconds"`
		GotIt                       map[string]gopodGotIt `json:"got_it"`
	} `json:"chat_moves"`
	LLM struct {
		WarmupText         string  `json:"warmup_text"`
		OpenerSongID       string  `json:"opener_song_id"`
		VoicesText         string  `json:"voices_text"`
		VoicesEngineText   string  `json:"voices_engine_text"`
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
	Bingo             gopodExperience   `json:"bingo"`
	Experiences       []gopodExperience `json:"experiences"`
	ExperienceMoves   []string          `json:"experience_moves"`
	ClosingRiffText   string            `json:"closing_riff_text"`
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
	AnswerChains map[string]struct {
		From  string   `json:"from"`
		Parts []string `json:"parts"`
	} `json:"answer_chains"`
	SystemMenu struct {
		Themes map[string]map[string]string `json:"themes"`
		Intro  string                       `json:"intro"` // scripture 0 (gopodTorahIntro)
	} `json:"system_menu"`
	VisualSources struct {
		Ring []struct {
			ID     string `json:"id"`
			Label  string `json:"label"`
			Robot  string `json:"robot"`
			Device string `json:"device"`
		} `json:"ring"`
	} `json:"visual_sources"`
	SleepBeats struct {
		Families map[string][]struct {
			Alias string `json:"alias"`
			Name  string `json:"name"`
			Kind  string `json:"kind"`
		} `json:"families"`
	} `json:"sleep_beats"`
	SystemTests struct {
		Audit    []string `json:"audit"`
		AdminRun struct {
			File string `json:"file"`
		} `json:"admin_run"`
		ProbersList struct {
			File      string  `json:"file"`
			PassShare float64 `json:"pass_share"`
		} `json:"probers_list"`
		LookupsList struct {
			File string `json:"file"`
		} `json:"lookups_list"`
		Tpu struct {
			Picture string `json:"picture"`
			Expect  string `json:"expect"`
		} `json:"tpu"`
		Audio struct {
			Sink      string `json:"sink"`
			SourceHas string `json:"source_has"`
		} `json:"audio"`
	} `json:"system_tests"`
	TopicPictures struct {
		Role                string              `json:"role"`
		Folder              string              `json:"folder"`
		Pictures            map[string][]string `json:"pictures"`
		MaxShown            int                 `json:"max_shown"`
		NotListed           []string            `json:"not_listed"`
		NotEggs             []string            `json:"not_eggs"`
		EggsKeep            []string            `json:"eggs_keep"`
		PopupText           string              `json:"popup_text"`
		MaxPerSession       int                 `json:"max_per_session"`
		EggVoiceGapMS       int                 `json:"egg_voice_gap_ms"`
		EggVoiceEdgeMS      int                 `json:"egg_voice_edge_ms"`
		EggText             string              `json:"egg_text"`
		EggZoomText         string              `json:"egg_zoom_text"`
		EggNames            map[string]string   `json:"egg_names"`
		Cards               map[string]string   `json:"cards"`
		Describe            []string            `json:"describe"`
		DescribeWaitSeconds float64             `json:"describe_wait_seconds"`
		DescribeNote        string              `json:"describe_note"`
		DescribeOnAsk       map[string][]string `json:"describe_on_ask"`
		DescribeWaitText    string              `json:"describe_wait_text"`
		DescribePrompt      string              `json:"describe_prompt"`
		DescribeDoneText    string              `json:"describe_done_text"`
	} `json:"topic_pictures"`
	ReporterVoices struct {
		Role                string            `json:"role"`
		Python              string            `json:"python"`
		Worker              string            `json:"worker"`
		AudioFolder         string            `json:"audio_folder"`
		Hosts               map[string]string `json:"hosts"`
		SystemVoice         string            `json:"system_voice"`
		Speed               float64           `json:"speed"`
		HFHome              string            `json:"hf_home"`
		SystemSayRole       string            `json:"system_say_role"`
		SongAudioRole       string            `json:"song_audio_role"`
		NarratorRole        string            `json:"narrator_role"`
		NarratorLead        float64           `json:"narrator_lead_seconds"`
		HeardMarker         string            `json:"heard_marker"`
		OperatorHost        string            `json:"operator_host"`
		ReadLinesFrom       []string          `json:"read_lines_from"`
		PageLagSeconds      float64           `json:"page_lag_seconds"`
		ChunkWords          int               `json:"chunk_words"`
		ChunkMinWords       int               `json:"chunk_min_words"`
		MaxWaitSeconds      float64           `json:"max_wait_seconds"`
		HeardMarginSeconds  float64           `json:"heard_margin_seconds"`
		GeneratingMarker    string            `json:"generating_marker"`
		StartWaitSeconds    float64           `json:"start_wait_seconds"`
		GeneratingText      string            `json:"generating_text"`
		SayWords            map[string]string `json:"say_words"`
		ReadBy              map[string]string `json:"read_by"`
		OpeningEndText      string            `json:"opening_end_text"`
		GateMaxSeconds      float64           `json:"gate_max_seconds"`
		GapSeconds          float64           `json:"gap_seconds"`
		Labels              map[string]string `json:"labels"`
		NicknameLabel       string            `json:"nickname_label"`
		SystemPersona       string            `json:"system_persona"`
		SystemSayWords      map[string]string `json:"system_say_words"`
		OperatorVoice       string            `json:"operator_voice"`
		UnassignedVoiceHost string            `json:"unassigned_voice_host"`
		VoiceTweaks         map[string]struct {
			Speed      float64 `json:"speed"`
			PitchCents float64 `json:"pitch_cents"`
			RobotFX    string  `json:"robot_fx"`
			Carrier    float64 `json:"carrier"`
		} `json:"voice_tweaks"`
		RobotFXPresets map[string]struct {
			Chain string  `json:"chain"`
			Wet   float64 `json:"wet"`
		} `json:"robot_fx_presets"`
		DemoVoices        map[string]string `json:"demo_voices"`
		OperatorKeepDays  float64           `json:"operator_keep_days"`
		FinishWaitSeconds float64           `json:"finish_wait_seconds"`
		ScriptedKeepDays  float64           `json:"scripted_keep_days"`
		LiveKeepDays      float64           `json:"live_keep_days"`
		LiveOnceKeepDays  float64           `json:"live_once_keep_days"`
		PopularUses       int               `json:"popular_uses"`
		PopularKeepDays   float64           `json:"popular_keep_days"`
		SystemStings      map[string]struct {
			File   string          `json:"file"`
			Volume json.RawMessage `json:"volume"`
		} `json:"system_stings"`
		Opening []gopodHostLine `json:"opening"`
		Closing []gopodHostLine `json:"closing"`
	} `json:"reporter_voices"`
	BubbleSounds struct {
		DefaultFile string `json:"default_file"`
		Goverlord   struct {
			File   string          `json:"file"`
			Volume json.RawMessage `json:"volume"`
			Mute   bool            `json:"mute"`
		} `json:"Goverlord"`
	} `json:"bubble_sounds"`
	SystemSearch struct {
		URL            string  `json:"url"`
		Results        int     `json:"results"`
		TimeoutSeconds float64 `json:"timeout_seconds"`
		Note           string  `json:"note"`
		NoneNote       string  `json:"none_note"`
		Line           string  `json:"line"`
		ResultsRole    string  `json:"results_role"`
		HandoffHost    string  `json:"handoff_host"`
		ResultsHost    string  `json:"results_host"`
	} `json:"system_search"`
	SystemSet struct {
		Role string `json:"role"`
	} `json:"system_set"`
	Managers struct {
		Takes          []gopodManagerTake `json:"takes"`
		ReleaseSeconds float64            `json:"release_seconds"`
		TimePlaceText  string             `json:"time_place_text"`
		DisplayHost    string             `json:"display_host"`
		USBText        string             `json:"usb_text"`
		HandoffText    string             `json:"handoff_text"`
		QRNote         string             `json:"qr_note"`
		QRDisplayText  string             `json:"qr_display_text"`
		Personas       map[string]struct {
			Host      string   `json:"host"`
			Reporter  string   `json:"reporter"`
			Prompt    string   `json:"prompt"`
			LinesNote string   `json:"lines_note"`
			WhoLine   string   `json:"who_line"`
			WIP       []string `json:"wip"`
			VoiceOf   string   `json:"voice_of"`
		} `json:"personas"`
		WIPText          string   `json:"wip_text"`
		AnswerNumPredict int      `json:"answer_num_predict"`
		WhoWords         []string `json:"who_words"`
	} `json:"managers"`
	WeatherCard struct {
		PauseSeconds float64 `json:"pause_seconds"`
		Picture      bool    `json:"picture"`
	} `json:"weather_card"`
	WeatherLines map[string]string `json:"weather_lines"`
	ChainPrompts map[string]string `json:"chain_prompts"`
	TopicGroups  map[string]struct {
		Topics   []string `json:"topics"`
		RuleLine string   `json:"rule_line"`
	} `json:"topic_groups"`
	PointBubbles struct {
		Prompt   string `json:"prompt"`
		Fallback struct {
			Title string `json:"title"`
			Note  string `json:"note"`
		} `json:"fallback"`
		Hype          map[string][]string `json:"hype"`
		KeepLeadLines int                 `json:"keep_lead_lines"`
	} `json:"point_bubbles"`
	Game struct {
		Rules      string         `json:"rules"`
		WinText    string         `json:"win_text"`
		MasterText string         `json:"master_text"`
		ResetOn    string         `json:"reset_on"`
		Runs       []gopodGameRun `json:"runs"`
	} `json:"game"`
	GoldenAnswers []struct {
		Queries []string `json:"queries"`
		Line    string   `json:"line"`
	} `json:"golden_answers"`
	InterviewBuild struct {
		Cues       map[string]string `json:"cues"`
		Reporters  []string          `json:"reporters"`
		DefaultAsk string            `json:"default_ask"`
	} `json:"interview_build"`
	HostPersonaFrames struct {
		Name      string `json:"name"`
		Steps     string `json:"steps"`
		Signature string `json:"signature"`
		Dial      string `json:"dial"`
	} `json:"host_persona_frames"`
	BrotherRepeat struct {
		PlainWords []string `json:"plain_words"`
	} `json:"brother_repeat"`
	Greetings struct {
		Parts []struct {
			Name string `json:"name"`
			From int    `json:"from"`
			To   int    `json:"to"`
		} `json:"parts"`
		RightText string `json:"right_text"`
		WrongText string `json:"wrong_text"`
		DayText   string `json:"day_text"`
	} `json:"greetings"`
	App struct {
		HTTPSPort int      `json:"https_port"`
		CertDir   string   `json:"cert_dir"`
		CAName    string   `json:"ca_name"`
		FenceDNS  []string `json:"fence_dns"`
		FenceIPs  []string `json:"fence_ips"`
	} `json:"app"`
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
	SessionMemory struct {
		Line      string `json:"line"`
		Scenario  string `json:"scenario"`
		EmptyText string `json:"empty_text"`
	} `json:"session_memory"`
	Summary struct {
		Queries    []string `json:"queries"`
		MinTopics  int      `json:"min_topics"`
		MaxTopics  int      `json:"max_topics"`
		TopicOrder []string `json:"topic_order"`
		FactsNote  string   `json:"facts_note"`
		Fallback   string   `json:"fallback"`
	} `json:"summary"`
	ResponseChains map[string]gopodResponseChain `json:"response_chains"`
	Openers        map[string]string             `json:"openers"`
	Nicknames      struct {
		Titles []string `json:"titles"`
		List   []string `json:"list"`
	} `json:"nicknames"`
	AnswerShapes struct {
		Shapes    map[string][]string `json:"shapes"`
		Marks     map[string]string   `json:"marks"`
		MarkOrder []string            `json:"mark_order"`
	} `json:"answer_shapes"`
	TypedLine struct {
		AllFirst string `json:"all_first"`
	} `json:"typed_line"`
	ColourSlots struct {
		MaxSentences int `json:"max_sentences"`
	} `json:"colour_slots"`
	SystemTour struct {
		QuietSeconds float64 `json:"quiet_seconds"`
		StartText    string  `json:"start_text"`
		EndText      string  `json:"end_text"`
	} `json:"system_tour"`
	BrobotsSetup struct {
		Brobot1  string `json:"brobot_1"`
		Brobot2  string `json:"brobot_2"`
		SoloNote string `json:"solo_note"`
	} `json:"brobots_setup"`
	ReadmeCast struct {
		Head       string `json:"head"`
		Rule       string `json:"rule"`
		Row        string `json:"row"`
		NumlockOff string `json:"numlock_off"`
		OnBody     string `json:"on_body"`
		OnLocation string `json:"on_location"`
		AllWho     string `json:"all_who"`
		AllWhat    string `json:"all_what"`
	} `json:"readme_cast"`
	Torah struct {
		Opening     string `json:"opening"`
		IntroTitle  string `json:"intro_title"` // scripture 0's title
		Scripture   string `json:"scripture"`
		Commandment string `json:"commandment"`
		NextFrom    string `json:"next_from"`
		NextTo      string `json:"next_to"`
	} `json:"torah"`
	LineQueue struct {
		On               bool    `json:"on"`
		AnswerMaxSeconds float64 `json:"answer_max_seconds"`
		IdleSeconds      float64 `json:"idle_seconds"`
		PageMaxSeconds   float64 `json:"page_max_seconds"`
	} `json:"line_queue"`
	DocBS struct {
		Folder         string            `json:"folder"`
		Share          int               `json:"share"`
		SlotPrompt     string            `json:"slot_prompt"`
		Fallback       string            `json:"fallback"`
		NumPredict     int               `json:"num_predict"`
		ColourPrompt   string            `json:"colour_prompt"`
		ColourPrompt18 string            `json:"colour_prompt_18"`
		Chains         []gopodDocBSChain `json:"chains"`
	} `json:"docbs"`
	GopodYourselfQr struct {
		KeepUpWords string `json:"keep_up_words"`
		KeepUpLine  string `json:"keep_up_line"`
		ContactLine string `json:"contact_line"`
	} `json:"gopod_yourself_qr"`
	BrainChain struct {
		BubbleHide        []string          `json:"bubble_hide"`
		WhosHereHost      string            `json:"whos_here_host"`
		WhosHereText      string            `json:"whos_here_text"`
		VisitorsLabel     string            `json:"visitors_label"`
		Header            string            `json:"header"`
		ShortHeader       string            `json:"short_header"`
		RecentMinutes     float64           `json:"recent_minutes"`
		DeepMinutes       float64           `json:"deep_minutes"`
		RecordLine        string            `json:"record_line"`
		AnswerPart        string            `json:"answer_part"`
		AskedBefore       string            `json:"asked_before"`
		AnswerChars       int               `json:"answer_chars"`
		LongLine          string            `json:"long_line"`
		FlowLine          string            `json:"flow_line"`
		Flow              map[string]string `json:"flow"`
		IntroOpener       string            `json:"intro_opener"`
		Lines             []string          `json:"lines"`
		IntroFact         []string          `json:"intro_fact"`
		LookedUp          string            `json:"looked_up"`
		LookupFiles       []string          `json:"lookup_files"`
		LookupMax         int               `json:"lookup_max"`
		LookupChars       int               `json:"lookup_chars"`
		LookupChainMax    int               `json:"lookup_chain_max"`
		LookupDefault     string            `json:"lookup_default"`
		LookupDefaultLine string            `json:"lookup_default_line"`
		LookupSkip        []string          `json:"lookup_skip"`
		NonsenseLine      string            `json:"nonsense_line"`
		SameWords         string            `json:"same_words"`
		TwoMatchesLine    string            `json:"two_matches_line"`
		RouteRep          string            `json:"route_rep"`
		RouteOther        string            `json:"route_other"`
		RouteAgain        string            `json:"route_again"`
		RouteHandoff      string            `json:"route_handoff"`
		RouteWelcome      string            `json:"route_welcome"`
		RouteStaffWords   string            `json:"route_staff_words"`
		RouteGoverlord    string            `json:"route_goverlord"`
		RouteWelcomeLine  string            `json:"route_welcome_line"`
		ExperienceYes     string            `json:"experience_yes"`
		ExperienceNo      string            `json:"experience_no"`
		ExperienceAsks    string            `json:"experience_asks"`
		ReminderPull      string            `json:"reminder_pull"`
		UnknownPull       string            `json:"unknown_pull"`
		UnknownPullAfter  string            `json:"unknown_pull_after"`
		Aura              map[string]string `json:"aura"`
		RealWords         string            `json:"real_words"`
		CastItem          string            `json:"cast_item"`
		Roster            string            `json:"roster"`
		CastMe            string            `json:"cast_me"`
		RemindersPart     string            `json:"reminders_part"`
		Asked             string            `json:"asked"`
		None              string            `json:"none"`
		QuestionChars     int               `json:"question_chars"`
		GomadsRawChat     bool              `json:"gomads_raw_chat"`
		OperatorWho       string            `json:"operator_who"`
		Performers        string            `json:"performers"`
	} `json:"brain_chain"`
	SystemBubble struct {
		Label        string `json:"label"`
		VoicePersona string `json:"voice_persona"`
	} `json:"system_bubble"`
	IntroRecipe struct {
		Words           []string `json:"words"`
		Text            string   `json:"text"`
		CastLine        string   `json:"cast_line"`
		Relations       []string `json:"relations"`
		Opening         string   `json:"opening"`
		NicknamePart    string   `json:"nickname_part"`
		WhyPart         string   `json:"why_part"`
		WeatherText     string   `json:"weather_text"`
		WeatherTextNext string   `json:"weather_text_next"`
	} `json:"intro_recipe"`
	PersonaProfiles struct {
		Folder        string   `json:"folder"`
		Sections      []string `json:"sections"`
		IntroSections []string `json:"intro_sections"`
		Text          string   `json:"text"`
	} `json:"persona_profiles"`
	PersonaChains struct {
		Folder   string                        `json:"folder"`
		Personas map[string]gopodPersonaChains `json:"personas"`
	} `json:"persona_chains"`
	CommCheck struct {
		Keys   []string `json:"keys"`
		Rounds []struct {
			Text  string   `json:"text"`
			Keys  []string `json:"keys"`
			Asker string   `json:"asker"`
			Self  bool     `json:"self"`
		} `json:"rounds"`
		Crew           []string `json:"crew"`
		QuietSeconds   float64  `json:"quiet_seconds"`
		LineMaxSeconds float64  `json:"line_max_seconds"`
		RecordSeconds  float64  `json:"record_seconds"`
		StartText      string   `json:"start_text"`
		DoneText       string   `json:"done_text"`
		AllName        string   `json:"all_name"`
		Song           string   `json:"song"`
		Wake           bool     `json:"wake"`
		GameWelcome    struct {
			By       string              `json:"by"`
			Base     string              `json:"base"`
			Flavours map[string][]string `json:"flavours"`
		} `json:"game_welcome"`
	} `json:"comm_check"`
	RoastDial struct {
		Level  int `json:"level"`
		Levels []struct {
			Name        string `json:"name"`
			Text        string `json:"text"`
			HumanSwears bool   `json:"human_swears"`
		} `json:"levels"`
		SensitiveWords []string `json:"sensitive_words"`
		SensitiveText  string   `json:"sensitive_text"`
		JekyllHyde     bool     `json:"jekyll_hyde"`
	} `json:"roast_dial"`
	Spice struct {
		Text               string   `json:"text"`
		Pick               int      `json:"pick"`
		PickExclamations   int      `json:"pick_exclamations"`
		RobotSwears        []string `json:"robot_swears"`
		HumanSwears        []string `json:"human_swears"`
		HumanSwearsPrivate string   `json:"human_swears_private"`
		Exclamations       []string `json:"exclamations"`
	} `json:"spice"`
	PrebuildContext struct {
		Header  string   `json:"header"`
		RepRole string   `json:"rep_role"`
		Lines   []string `json:"lines"`
	} `json:"prebuild_context"`
	BrobotBubble struct {
		Steps []string `json:"steps"`
	} `json:"brobot_bubble"`
	BubbleNotices struct {
		ValuePoints struct {
			Title       string            `json:"title"`
			Parts       []string          `json:"parts"`
			Points      string            `json:"points"`
			Weather     string            `json:"weather"`
			Search      string            `json:"search"`
			None        string            `json:"none"`
			GotIt       string            `json:"got_it"`
			Note        string            `json:"note"`
			Guide       map[string]string `json:"guide"`
			Notes       string            `json:"notes"`
			NotesJoin   string            `json:"notes_join"`
			Egg         string            `json:"egg"`
			Combo       string            `json:"combo"`
			Chapter     string            `json:"chapter"`
			Recovery    string            `json:"recovery"`
			Next        string            `json:"next"`
			Summary     string            `json:"summary"`
			Score       string            `json:"score"`
			GotItAgain  string            `json:"got_it_again"`
			SecondGotIt string            `json:"second_got_it"`
			HelperFrom  []string          `json:"helper_from"`
		} `json:"value_points"`
		Egg          gopodNoticeWords `json:"egg"`
		NativeIntent gopodNoticeWords `json:"native_intent"`
	} `json:"bubble_notices"`
	NativeIntent struct {
		Role        string  `json:"role"`
		Text        string  `json:"text"`
		OnceSeconds float64 `json:"once_seconds"`
	} `json:"native_intent"`
	SystemText map[string]string  `json:"system_text"`
	Timings    map[string]float64 `json:"timings"`
	Lightning  struct {
		On          bool    `json:"on"`
		TopTopics   int     `json:"top_topics"`
		TriggerPull float64 `json:"trigger_pull"`
		MinWordPull float64 `json:"min_word_pull"`
		KeepShare   float64 `json:"keep_share"`
		OrderMargin float64 `json:"order_margin"`
	} `json:"lightning"`
	RobotsOff struct {
		On          bool              `json:"on"`
		CaptureFile string            `json:"capture_file"`
		Voices      map[string]string `json:"voices"`
	} `json:"robots_off"`
	HelpLookup struct {
		Words  []string `json:"words"`
		Skip   []string `json:"skip"`
		Max    int      `json:"max"`
		Header string   `json:"header"`
		Item   string   `json:"item"`
	} `json:"help_lookup"`
	RunCheck struct {
		File              string   `json:"file"`
		MuteMarker        string   `json:"mute_marker"`
		UnmuteMarker      string   `json:"unmute_marker"`
		SkipPrefixes      []string `json:"skip_prefixes"`
		ReplayWaitSeconds float64  `json:"replay_wait_seconds"`
	} `json:"run_check"`
	ProbersMake struct {
		File          string            `json:"file"`
		Shapes        []string          `json:"shapes"`
		WrongGreeting map[string]string `json:"wrong_greeting"`
	} `json:"probers_make"`
	SystemHelp struct {
		Lines []string `json:"lines"`
	} `json:"system_help"`
	PageSync struct {
		ReplyWaitSeconds float64 `json:"reply_wait_seconds"`
		HeldLogSeconds   float64 `json:"held_log_seconds"`
		MainLeaseSeconds float64 `json:"main_lease_seconds"`
		FeedMax          int     `json:"feed_max"`
		SecondAdminText  string  `json:"second_admin_text"`
	} `json:"page_sync"`
	ChatAudio struct {
		Volume       float64 `json:"volume"`
		NormalizeDB  float64 `json:"normalize_db"`
		LoudnessLUFS float64 `json:"loudness_lufs"`
	} `json:"chat_audio"`
	ValuePoints struct {
		ScoreRole string `json:"score_role"`
		ScoreText string `json:"score_text"`
		ScoreBar  struct {
			Done string `json:"done"`
			Left string `json:"left"`
		} `json:"score_bar"`
		Reminders            map[string]string `json:"reminders"`
		WhatElseTopic        string            `json:"what_else_topic"`
		WhatElseCount        int               `json:"what_else_count"`
		ReminderSkip         int               `json:"reminder_skip_next_topics"`
		ReminderChaptersText string            `json:"reminder_chapters_text"`
		ReminderTopicsText   string            `json:"reminder_topics_text"`
		ReminderProberText   string            `json:"reminder_prober_text"`
		ReminderTopicCount   int               `json:"reminder_topic_count"`
		ReminderEggs         []struct {
			Key   string `json:"key"`
			Label string `json:"label"`
		} `json:"reminder_eggs"`
		ReminderEggCount int                 `json:"reminder_egg_count"`
		SayLines         map[string]string   `json:"say_lines"`
		Notes            map[string]string   `json:"notes"`
		PickSkipWords    []string            `json:"pick_skip_words"`
		BackgroundTopics []string            `json:"background_topics"`
		Target           string              `json:"target"`
		CorePoints       []string            `json:"core_points"`
		CoreTopics       map[string][]string `json:"core_topics"`
		TopicLinks       map[string][]string `json:"topic_links"`
		Funnels          [][]string          `json:"funnels"`
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
		LinkHosts           []string       `json:"link_hosts"`
		BrandWords          []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"brand_words"`
	} `json:"reply_tighten"`
	InterviewScript struct {
		SongID            string  `json:"song_id"`
		DraftsDir         string  `json:"drafts_dir"`
		HarvestInput      string  `json:"harvest_input"`
		HarvestEndText    string  `json:"harvest_end_text"`
		Draft             string  `json:"draft"`
		PlayWords         string  `json:"play_words"`
		StopWords         string  `json:"stop_words"`
		GapSeconds        float64 `json:"gap_seconds"`
		BookendsFromDraft bool    `json:"bookends_from_draft"`
		StartLines        []struct {
			Who  string `json:"who"`
			Text string `json:"text"`
		} `json:"start_lines"`
		LLMReadyText string `json:"llm_ready_text"`
	} `json:"interview_script"`
	Personas     map[string]string `json:"personas"`
	PersonaCards map[string]struct {
		Group string `json:"group"`
		Body  string `json:"body"`
	} `json:"persona_cards"`
	HostPersonas map[string]gopodHostPersona `json:"host_personas"`
	GemmaEyes    struct {
		Device      []string `json:"device"`
		Size        string   `json:"size"`
		SkipFrames  int      `json:"skip_frames"`
		KeepMinutes float64  `json:"keep_minutes"`
		Safeguard   bool     `json:"safeguard"`
		Prompt      string   `json:"prompt"`
		SayText     string   `json:"say_text"`
		FailText    string   `json:"fail_text"`
	} `json:"gemma_eyes"`
	CompanyIntro struct {
		Lines []string `json:"lines"`
	} `json:"company_intro"`
	PersonaColours map[string]struct {
		Background string `json:"background"`
		Border     string `json:"border"`
		Text       string `json:"text"`
	} `json:"persona_colours"`
	HostPairs map[string]gopodHostPair `json:"host_pairs"`
	SignOff   struct {
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
	Native struct {
		OnAtStart bool              `json:"on_at_start"`
		AlwaysOn  []string          `json:"always_on"`
		ToManager map[string]string `json:"to_manager"`
	} `json:"native"`
	Session struct {
		BusyText             string            `json:"busy_text"`
		RunGateText          string            `json:"run_gate_text"`
		StoppedText          string            `json:"stopped_text"`
		NoRunText            string            `json:"no_run_text"`
		StartedText          string            `json:"started_text"`
		CompleteText         string            `json:"complete_text"`
		NotFoundText         string            `json:"not_found_text"`
		ReplayWord           string            `json:"replay_word"`
		ReplayName           string            `json:"replay_name"`
		StopSession          []string          `json:"stop_session_words"`
		PausedText           string            `json:"paused_text"`
		ContinuesText        string            `json:"continues_text"`
		PauseAfterMS         int               `json:"pause_after_ms"`
		PageHiddenFile       string            `json:"page_hidden_file"`
		AwaitsText           string            `json:"awaits_text"`
		PagePresentSeconds   float64           `json:"page_present_seconds"`
		CommCheckWaitSeconds float64           `json:"comm_check_wait_seconds"`
		RunNames             map[string]string `json:"run_names"`
	} `json:"session"`
	NativeCube struct {
		Intents []struct {
			Intent  string   `json:"intent"`
			Phrases []string `json:"phrases"`
		} `json:"intents"`
	} `json:"native_cube"`
	LiveWeather struct {
		Location        string  `json:"location"`
		CacheSeconds    float64 `json:"cache_seconds"`
		OfflineText     string  `json:"offline_text"`
		MilitaryLabel   string  `json:"military_label"`
		FollowUpKitLine string  `json:"follow_up_kit_line"`
	} `json:"live_weather"`
	Photo struct {
		Camera struct {
			Manual        bool    `json:"manual"`
			ExposureMs    uint32  `json:"exposure_ms"`
			Gain          float32 `json:"gain"`
			SettleSeconds float64 `json:"settle_seconds"`
		} `json:"camera"`
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
		QuickLook          struct {
			Enabled        bool               `json:"enabled"`
			Python         string             `json:"python"`
			Worker         string             `json:"worker"`
			Model          string             `json:"model"`
			Labels         string             `json:"labels"`
			MinScore       float64            `json:"min_score"`
			TimeoutSeconds float64            `json:"timeout_seconds"`
			Pictures       bool               `json:"pictures"`
			Enhance        map[string]float64 `json:"enhance"`
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
	CleanPolitics struct {
		AllowedTopics []string          `json:"allowed_topics"`
		Words         []string          `json:"words"`
		NeverWords    []string          `json:"never_words"`
		NeverInLine   string            `json:"never_in_line"`
		Countries     map[string]string `json:"countries"`
		Capitals      map[string]string `json:"capitals"`
		StateCapitals map[string]string `json:"state_capitals"`
		Continents    []string          `json:"continents"`
	} `json:"clean_politics"`
	HereNow struct {
		Text string `json:"text"`
		Cast []struct {
			Who  string `json:"who"`
			Line string `json:"line"`
		} `json:"cast"`
		CastJoin    string            `json:"cast_join"`
		BrotherLine string            `json:"brother_line"`
		BatteryText string            `json:"battery_text"`
		ClockFormat string            `json:"clock_format"`
		Place       string            `json:"place"`
		ChatNote    string            `json:"chat_note"`
		ChatAnswer  string            `json:"chat_answer"`
		BatteryNote string            `json:"battery_note"`
		Genders     map[string]string `json:"genders"`

		VoiceGenders map[string]string `json:"voice_genders"`
	} `json:"here_now"`
	TopicClusters struct {
		Folder   string   `json:"folder"`
		RootDocs []string `json:"root_docs"`
	} `json:"topic_clusters"`
	NumpadMap     json.RawMessage            `json:"numpad_map"`
	NumpadMaps    map[string]json.RawMessage `json:"numpad_maps"`
	BrobotsPrompt struct {
		Blocks         [][]string `json:"blocks"`
		DialZeroBlocks []string   `json:"dial_zero_blocks"`
	} `json:"brobots_prompt"`
	BatteryWatch struct {
		Enabled         bool                `json:"enabled"`
		IntervalSeconds float64             `json:"interval_seconds"`
		LowVolts        float64             `json:"low_volts"`
		RecoverVolts    float64             `json:"recover_volts"`
		DipVolts        float64             `json:"dip_volts"`
		Lines           map[string][]string `json:"lines"`
		HomeLines       map[string]string   `json:"home_lines"`
		BackLines       map[string]string   `json:"back_lines"`
		BackAfterMisses int                 `json:"back_after_misses"`
		Guards          map[string]string   `json:"guards"`
		GuardWarning    string              `json:"guard_warning"`
		GuardHandover   string              `json:"guard_handover"`
		SweepClear      string              `json:"sweep_clear"`
		SweepMissing    string              `json:"sweep_missing"`
		SweepLow        string              `json:"sweep_low"`
		SweepFull       string              `json:"sweep_full"`
		SSHKey          string              `json:"ssh_key"`
		SweepStop       string              `json:"sweep_stop"`
		Reply           string              `json:"reply"`
	} `json:"battery_watch"`
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

var gopodBrobotsSettings = loadGOPODBrobotsSettings()

func loadGOPODBrobotsSettings() gopodBrobotsSettingsFile {
	var s gopodBrobotsSettingsFile
	data, err := gopodSettingsMerged()
	if err != nil {
		panic("GOPOD Brobots settings missing at " + gopodSettingsDir + ": " + err.Error())
	}
	data = gopodClustersOverlay(data)
	if err := json.Unmarshal(data, &s); err != nil {
		panic("GOPOD Brobots settings unreadable at " + gopodSettingsDir + ": " + err.Error())
	}
	gopodApplyChaptersFolder(&s)
	gopodApplyCommCheckSong(&s)
	gopodApplyGameRuns(&s)
	gopodApplyPersonaProfiles(&s)
	gopodApplyNicknames(&s)
	gopodApplyNameVars(&s)
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
	check(len(s.NumpadMaps[strings.TrimSpace(s.NumpadPTT.Layout)]) > 2, "numpad_maps."+s.NumpadPTT.Layout+" (clusters/system/numpad-mapper.md, numpad_ptt.layout)")
	check(len(s.NumpadPTT.BodyToBrobot) > 0, "numpad_ptt.body_to_brobot")
	check(s.RecordSession.MaxMinutes > 0, "record_session.max_minutes")
	check(s.RecordSession.AnswerWaitMinutes > 0, "record_session.answer_wait_minutes")
	check(len(s.RecordSession.KeepWords) > 0, "record_session.keep_words")
	check(len(s.RecordSession.DeleteWords) > 0, "record_session.delete_words")
	check(strings.TrimSpace(s.RecordSession.AnswerTo) != "", "record_session.answer_to")
	check(gopodBrobotESN(s.RecordSession.AskBrobot) != "" || (s.SignOff.SystemRole != "" && s.RecordSession.AskBrobot == s.SignOff.SystemRole), "record_session.ask_brobot")
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
		panic(fmt.Sprintf("GOPOD Brobots settings at %s are missing: %s", gopodSettingsDir, strings.Join(missing, ", ")))
	}
	return s
}

// gopodSecondsToDuration turns a settings value in seconds into a time.Duration.
func gopodSecondsToDuration(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}

// gopodExperience: a scripted song's pre/post chat (the duo answer-chain template,
// gopod_experiences.go): before its run, a question about it gets its <name>_not_played lines and
// no_run_note; after, the snapshot note, each Brobot's take and his closing remark.
type gopodExperience struct {
	Name          string            `json:"name"`
	RunMarker     string            `json:"run_marker"`
	RunGapSeconds float64           `json:"run_gap_seconds"`
	SnapshotNote  string            `json:"snapshot_note"`
	QuestionWord  string            `json:"question_word"`
	NotATest      []string          `json:"not_a_test_phrases"`
	PitchPack     string            `json:"pitch_pack"`
	NoRunNote     string            `json:"no_run_note"`
	Takes         map[string]string `json:"takes"`
	Closing       map[string]string `json:"closing"`
	Pack          string            `json:"pack"`
	RunPack       string            `json:"run_pack"`
	RunName       string            `json:"run_name"`
	PreAttitude   map[string]string `json:"pre_attitude"`
	PostAttitude  map[string]string `json:"post_attitude"`
	ClosingRiff   map[string]string `json:"closing_riff"`
}

// gopodSnippetLine: one scripted_snippets entry (a run card's lines fill some at load -
// gopod_game_engine.go).
type gopodSnippetLine struct {
	When   string   `json:"when"`
	Words  []string `json:"words"`
	Brobot string   `json:"brobot"`
	At     string   `json:"at"`
	Lines  []string `json:"lines"`
	Chance float64  `json:"chance"`
}

// ## EOF
