// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_interview_script.go
package wirepod_ttr

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/sashabaranov/go-openai"
)

// The draft performance and its replay (operator 2026-10-07: "Do a full dry
// run ... fully featured, fully logged ... I need to replay a 'draft
// performance run' to review & edit"; "I want the best possible responses
// organically, then I tweak minor drifts").
//
// A full dry run (harvest_dry_run.sh, gopod_harvest_test.go) writes
// draft_performance.json: the whole show in order - the reporters' opening,
// each question as its reporter reads it, the chapter title, the Brobot's own
// answer, the pictures, the sign-off, the score, keep or delete, the closing.
// The operator edits an answer's "say" or "cue" there.
//
// Typing interview_script.play_words in the chat plays the newest draft (or
// interview_script.draft) through the live chat: each question goes in as a
// typed line to its Brobot - the reporter reads it, the chapter and pictures
// show, the value point is ticked - and the Brobot says the draft's answer,
// word for word, instead of asking the LLM, through the same path as a live
// answer (cue animation, bubble, highlight, speech, memory, Record Session).
// The next question waits until that answer has been said, plus gap_seconds.
// The last question is the sign-off to both Brobots, which brings the live
// ending: score, keep or delete, the reporters' closing. stop_words stops it
// after the line being said.

type gopodDraftLine struct {
	Who  string `json:"who"`
	Text string `json:"text"`
}

type gopodDraftAnswer struct {
	Brobot     string   `json:"brobot"`
	Points     []string `json:"points"`
	Cue        string   `json:"cue"` // edit to change the animation
	Say        string   `json:"say"` // edit to change the words
	Spoken     []string `json:"spoken"`
	LLMSeconds float64  `json:"llm_seconds"`
	// 2026-10-08: moves at spoken words, and {weather} (gopod_script_moves.go).
	Moves []gopodDraftMove `json:"moves,omitempty"`
}

type gopodDraftQuestion struct {
	N       int                `json:"n"`
	Chapter string             `json:"chapter,omitempty"`
	Topic   string             `json:"topic,omitempty"`
	To      string             `json:"to"`      // "Brobot 1", "Brobot 2" or "Brobots"
	Ask     string             `json:"ask"`     // the operator's line
	ReadBy  string             `json:"read_by"` // the reporter who reads it out
	Before  []gopodDraftLine   `json:"before,omitempty"`
	Answers []gopodDraftAnswer `json:"answers"`
	After   []gopodDraftLine   `json:"after,omitempty"`
	// 2026-10-12: a song's own pictures for its question (knobs.json "pictures").
	Pictures []string `json:"pictures,omitempty"`
}

type gopodDraft struct {
	About     string               `json:"about"`
	Run       string               `json:"run"`
	Model     string               `json:"model"`
	Score     string               `json:"score"`
	Seconds   float64              `json:"seconds"`
	Opening   []gopodDraftLine     `json:"opening"`
	Closing   []gopodDraftLine     `json:"closing,omitempty"` // the reporters' closing (bookends_from_draft)
	Questions []gopodDraftQuestion `json:"questions"`
	Ending    []gopodDraftLine     `json:"ending"`
	// 2026-10-12 (operator's interview score: "Interview closer"): said after
	// the run's complete line, when it played to its end.
	Closer []gopodDraftLine `json:"closer,omitempty"`
}

const gopodDraftAbout = "Draft performance (operator 2026-10-07: \"replay a 'draft performance run' to review & edit\"). The whole show in order, from a full dry run with no robots: opening, each question as its reporter reads it (before: the chapter title; after: pictures), the Brobot's own answer, the sign-off, and the ending. To change an answer, edit its \"say\" (words) or \"cue\" (animation). To play it through the robots: type the interview_script play_words (\"play the draft\") in the Brobots chat. draft_performance.md is the same, to read (harvest_dry_run.sh render-draft redraws it)."

var gopodScript = struct {
	sync.Mutex
	pending map[string]gopodDraftAnswer // esn -> the draft's answer to its next question
	ask     map[string]string           // esn -> that question
	readBy  string                      // the reporter who reads the next question ("" = read_by)
	asking  string                      // the question being put in now
	moves   map[string][]gopodDraftMove // esn -> the moves of the answer being said
	logged  map[string][]gopodDraftMove // esn -> those moves, for the session score
	movesAt []string                    // every move phrase in the playing draft, lower case
	// 2026-10-12: a song that names its chapters shows them, and only its own
	// pictures, instead of the topics' (gopod_interview_chapters.go).
	own      bool
	chapter  string
	pictures []string
}{pending: map[string]gopodDraftAnswer{}, ask: map[string]string{}, moves: map[string][]gopodDraftMove{}, logged: map[string][]gopodDraftMove{}}

var gopodScriptPlaying, gopodScriptStop, gopodPageHidden atomic.Bool

// gopodScriptRan: a draft has played since the last ending (gopodDraftRunCompleted).
var gopodScriptRan atomic.Bool

// GopodPageShown: the chat page is showing (true) or hidden - another tab,
// minimised (operator 2026-10-07: "make the run pause on web-page not the
// focus, and continue upon focus again" - bubbles piled up out of order while
// the tab was away). A playing draft waits before its next question while the
// page is hidden; the line being said finishes.
func GopodPageShown(shown bool) {
	if shown {
		gopodPageSeenMS.Store(time.Now().UnixMilli())
	}
	changed := gopodPageHidden.Swap(!shown) != !shown
	if changed && gopodScriptPlaying.Load() {
		if shown {
			logger.Println("GOPOD_SCRIPT page shown: playing on")
		} else {
			logger.Println("GOPOD_SCRIPT page hidden: paused before the next question")
		}
	}
	if changed {
		gopodPageHiddenFlag(!shown)
	} else if shown && gopodBrobotsSettings.Session.PageHiddenFile != "" {
		// A flag left from before a restart never holds a song.
		os.Remove(filepath.Join(gopodChipperRoot(), gopodBrobotsSettings.Session.PageHiddenFile))
	}
}

// gopodPageHiddenFlag (2026-10-12, operator: "Upon page unfocus = pause at next
// safe spot until focus, then continue * add system messages: run paused, run
// continues"): the song engine waits before its next step while
// session.page_hidden_file exists; Goverlord says the run's paused_text /
// continues_text (a song, the draft, or a session replay).
func gopodPageHiddenFlag(hidden bool) {
	ss := gopodBrobotsSettings.Session
	if ss.PageHiddenFile != "" {
		flag := filepath.Join(gopodChipperRoot(), ss.PageHiddenFile)
		if hidden {
			os.WriteFile(flag, nil, 0o666)
		} else {
			os.Remove(flag)
		}
	}
	name := ""
	if r, ok := gopodRunNow(); ok {
		name = gopodRunText("{run}", r.Name)
	} else if gopodReplayPlaying.Load() {
		name, _ = gopodReplayNow.Load().(string)
	}
	if name == "" {
		return
	}
	text := ss.ContinuesText
	if hidden {
		// 2026-10-17 (survey E8: "Run paused" was said to a page that was away,
		// heard only on its return, right before "Run continues"): shown, not said.
		text = strings.ReplaceAll(ss.PausedText, "{run}", name)
		logger.Println("GOPOD_SESSION " + text)
		GopodChatFeedAdd(gopodBrobotsSettings.SignOff.SystemRole, text)
		return
	}
	gopodSessionSay(strings.ReplaceAll(text, "{run}", name))
}

// gopodScriptPending: a draft answer waits for this Brobot's question (not
// taken) - it's ready, so there's no "Thinking, please wait" to say.
func gopodScriptPending(esn, question string) bool {
	gopodScript.Lock()
	defer gopodScript.Unlock()
	_, ok := gopodScript.pending[esn]
	return ok && gopodSameQuestion(gopodAddressWords(gopodScript.ask[esn]), gopodAddressWords(question))
}

// gopodScriptCommand: a typed line that plays or stops the draft. True when it
// was one.
// gopodScriptCommandMatch: text is one of the interview's own commands.
func gopodScriptCommandMatch(text string) bool {
	s := gopodBrobotsSettings.InterviewScript
	lower := strings.ToLower(strings.TrimSpace(text))
	return (s.PlayWords != "" && gopodHasWholeWords(lower, strings.ToLower(s.PlayWords))) ||
		(s.StopWords != "" && gopodHasWholeWords(lower, strings.ToLower(s.StopWords)))
}

func gopodScriptCommand(text string) bool {
	s := gopodBrobotsSettings.InterviewScript
	lower := strings.ToLower(strings.TrimSpace(text))
	switch {
	case s.PlayWords != "" && gopodHasWholeWords(lower, strings.ToLower(s.PlayWords)):
		// 2026-10-08 (operator: "only after they play out & the llm is ready is
		// when 'start draft/stop draft' have any meaning"): before the reporters'
		// opening has played and LLM Ready shows, it does nothing.
		if gopodWarmGateClosed.Load() {
			logger.Println("GOPOD_SCRIPT start draft ignored: the opening and LLM Ready aren't done yet")
			return true
		}
		path, draft, err := gopodLoadDraft()
		if err != nil {
			logger.Println("GOPOD_SCRIPT not played: " + err.Error())
			return true
		}
		// One run at a time, checked in with Goverlord (gopod_session.go).
		if ok, _ := GopodRunStart("interview", "interview", 0); !ok {
			return true
		}
		logger.Println("GOPOD_SCRIPT play " + path)
		go gopodPlayDraft(draft)
		return true
	case s.StopWords != "" && gopodHasWholeWords(lower, strings.ToLower(s.StopWords)):
		if gopodScriptPlaying.Load() {
			gopodScriptStop.Store(true)
			logger.Println("GOPOD_SCRIPT stopping after the line being said")
		}
		return true
	}
	return false
}

// gopodLoadDraft reads interview_script.draft, or the newest
// draft_performance.json under interview_script.drafts_dir.
func gopodLoadDraft() (string, gopodDraft, error) {
	s := gopodBrobotsSettings.InterviewScript
	// 2026-10-09: the interview's song, when set and GOPOD_DRAFT isn't
	// (gopod_interview_song.go). 2026-10-10: found by its song_id, wherever its
	// folder is (gopod_song_folders.go).
	if id := strings.TrimSpace(s.SongID); id != "" && strings.TrimSpace(os.Getenv("GOPOD_DRAFT")) == "" {
		dir, err := gopodFindSong(id)
		if err != nil {
			return id, gopodDraft{}, err
		}
		draft, err := gopodInterviewSongDraft(dir)
		if err != nil {
			return dir, draft, errors.New(dir + ": " + err.Error())
		}
		return dir, draft, gopodCheckDraft(dir, draft)
	}
	path := gopodHomePath(s.Draft)
	if env := strings.TrimSpace(os.Getenv("GOPOD_DRAFT")); env != "" {
		path = env
	}
	if path == "" {
		// The spliced best-takes draft, when there is one (the show); else
		// the newest dry run's.
		found, _ := filepath.Glob(filepath.Join(gopodHomePath(s.DraftsDir), "*", "*_best_takes", "draft_performance.json"))
		if len(found) == 0 {
			found, _ = filepath.Glob(filepath.Join(gopodHomePath(s.DraftsDir), "*", "*", "draft_performance.json"))
		}
		if len(found) == 0 {
			return "", gopodDraft{}, errors.New("no draft_performance.json under " + s.DraftsDir)
		}
		sort.Strings(found) // <date>/<time>_harvest: the newest last
		path = found[len(found)-1]
	}
	// 2026-10-09: the plain script beside it, when there is one, is what plays
	// (gopod_draft_script.go).
	if md := filepath.Join(filepath.Dir(path), gopodDraftScriptName); !strings.HasSuffix(path, ".md") {
		if _, err := os.Stat(md); err == nil {
			path = md
		}
	}
	var draft gopodDraft
	data, err := os.ReadFile(path)
	if err != nil {
		return path, draft, err
	}
	if strings.HasSuffix(path, ".md") {
		if draft, err = gopodParseDraftScript(string(data)); err != nil {
			return path, draft, errors.New(path + ": " + err.Error())
		}
	} else if err := json.Unmarshal(data, &draft); err != nil {
		return path, draft, err
	}
	return path, draft, gopodCheckDraft(path, draft)
}

// gopodCheckDraft: every question has its words and an answer, every answer a speaker.
func gopodCheckDraft(path string, draft gopodDraft) error {
	if len(draft.Questions) == 0 {
		return errors.New("no questions in " + path)
	}
	for _, q := range draft.Questions {
		if strings.TrimSpace(q.Ask) == "" || len(q.Answers) == 0 {
			return errors.New("question " + strconv.Itoa(q.N) + " needs ask and an answer")
		}
		for _, a := range q.Answers {
			if (gopodBrobotESN(a.Brobot) == "" && gopodReporterHost(a.Brobot) == "") || strings.TrimSpace(a.Say) == "" {
				return errors.New("question " + strconv.Itoa(q.N) + ": an answer needs brobot and say")
			}
		}
	}
	return nil
}

func gopodPlayDraft(draft gopodDraft) {
	gopodScriptPlaying.Store(true)
	gopodScriptStop.Store(false)
	gopodScriptRan.Store(true)
	gopodChapterReset()
	var at []string
	for _, q := range draft.Questions {
		for _, a := range q.Answers {
			for _, m := range a.Moves {
				if m.At != "" {
					at = append(at, strings.ToLower(m.At))
				}
			}
		}
	}
	own := false
	for _, q := range draft.Questions {
		own = own || q.Chapter != ""
	}
	gopodScript.Lock()
	gopodScript.movesAt = at
	gopodScript.own, gopodScript.chapter, gopodScript.pictures = own, "", nil
	gopodScript.Unlock()
	defer func() {
		gopodScript.Lock()
		gopodScript.pending, gopodScript.ask = map[string]gopodDraftAnswer{}, map[string]string{}
		gopodScript.movesAt = nil
		gopodScript.own, gopodScript.chapter, gopodScript.pictures = false, "", nil
		gopodScript.Unlock()
		gopodScriptPlaying.Store(false)
		reason := "done"
		if gopodScriptStop.Load() {
			reason = "stopped"
		}
		var closer func()
		if len(draft.Closer) > 0 {
			closer = func() { gopodSayDraftLines(draft.Closer) }
		}
		GopodRunEndThen("interview", reason, closer)
	}()
	// Nothing reaches the Brobots before the green light (gopod_llm_warmup.go).
	for gopodWarmGateClosed.Load() {
		time.Sleep(time.Second)
	}
	// 2026-10-08 (operator: "upon 'start draft' add in a system draft started
	// message"): in Goverlord's bubble and voice.
	if so := gopodBrobotsSettings.SignOff; so.SystemRole != "" && so.StartedText != "" {
		GopodChatFeedAdd(so.SystemRole, so.StartedText)
		gopodSystemSay(so.StartedText, true)
	}
	// 2026-10-08: the reporters' handoff into question 1 (interview_script.start_lines),
	// inside the opener's warm-up (2026-10-12, gopodRunWarmup).
	gopodRunWarmup(gopodRunReadyLine(gopodBrobotsSettings.InterviewScript.LLMReadyText), func() {
		for _, l := range gopodBrobotsSettings.InterviewScript.StartLines {
			if host := gopodReporterHost(l.Who); host != "" {
				gopodReporterSay(host, l.Text, true, true)
			}
		}
	})
	gap := time.Duration(gopodBrobotsSettings.InterviewScript.GapSeconds * float64(time.Second))
	logger.Println("GOPOD_SCRIPT start run=" + draft.Run + " questions=" + strconv.Itoa(len(draft.Questions)))
	for _, q := range draft.Questions {
		for gopodPageHidden.Load() && !gopodScriptStop.Load() {
			time.Sleep(500 * time.Millisecond)
		}
		if gopodScriptStop.Load() {
			logger.Println("GOPOD_SCRIPT stopped before question " + strconv.Itoa(q.N))
			return
		}
		// 2026-10-08 (the record cut's script): the reporters' lines before the
		// question (their banter; a chapter or picture line is shown live, not here).
		for _, l := range q.Before {
			if gopodScriptStop.Load() {
				break
			}
			if host := gopodReporterHost(l.Who); host != "" {
				gopodReporterSay(host, l.Text, true, true)
			}
		}
		if gopodScriptStop.Load() {
			logger.Println("GOPOD_SCRIPT stopped before question " + strconv.Itoa(q.N))
			return
		}
		// The question's own answers; an answer by the other Brobot is his
		// handoff, said after them ("Brobot 2 only answers what he just heard").
		var answers, handoffs []gopodDraftAnswer
		for _, a := range q.Answers {
			// 2026-10-08: a reporter's line among the answers is said in its
			// place between them, through the page (operator: "apply").
			if gopodReporterHost(a.Brobot) != "" {
				handoffs = append(handoffs, a)
				continue
			}
			if q.To == "Brobots" || a.Brobot == q.To {
				answers = append(answers, a)
			} else {
				handoffs = append(handoffs, a)
			}
		}
		gopodScript.Lock()
		for _, a := range answers {
			esn := gopodBrobotESN(a.Brobot)
			gopodScript.pending[esn], gopodScript.ask[esn] = a, q.Ask
		}
		gopodScript.readBy = gopodReporterHost(q.ReadBy)
		gopodScript.asking = q.Ask
		gopodScript.chapter, gopodScript.pictures = q.Chapter, q.Pictures
		gopodScript.Unlock()
		to := gopodBrobotESN(q.To)
		if q.To == "Brobots" || to == "" {
			to = gopodTypedAll
		}
		logger.Println("GOPOD_SCRIPT question " + strconv.Itoa(q.N) + " to " + q.To + ": " + q.Ask)
		if err := gopodOperatorLine(to, q.Ask, "Typed text"); err != nil {
			logger.Println("GOPOD_SCRIPT question " + strconv.Itoa(q.N) + ": " + err.Error())
		}
		gopodScript.Lock()
		gopodScript.asking = ""
		gopodScript.Unlock()
		gopodScriptWaitAnswered(len(answers))
		// 2026-10-08: an answer to "Brobots" that wasn't asked for - the brother
		// stayed quiet (live 14:10: Brobot 2's sign-off never played) - is said
		// as his handoff.
		gopodScript.Lock()
		for esn, a := range gopodScript.pending {
			handoffs = append(handoffs, a)
			delete(gopodScript.pending, esn)
		}
		gopodScript.Unlock()
		for _, a := range handoffs {
			// 2026-10-09: a thanks or "stop run" stops between lines too.
			if gopodScriptStop.Load() {
				break
			}
			time.Sleep(gap)
			say := gopodScriptFill(gopodBrobotESN(a.Brobot), a.Say)
			if strings.TrimSpace(say) == "" {
				continue
			}
			logger.Println("GOPOD_SCRIPT handoff " + a.Brobot + ": " + say)
			if host := gopodReporterHost(a.Brobot); host != "" {
				gopodReporterSay(host, say, true, true)
				continue
			}
			gopodRecordSay(a.Brobot, say)
		}
		time.Sleep(gap)
	}
	logger.Println("GOPOD_SCRIPT done")
}

// gopodSayDraftLines says lines in order: a reporter's through the page, a
// Brobot's by the robot with its bubble.
func gopodSayDraftLines(lines []gopodDraftLine) {
	gap := time.Duration(gopodBrobotsSettings.InterviewScript.GapSeconds * float64(time.Second))
	for _, l := range lines {
		time.Sleep(gap)
		if host := gopodReporterHost(l.Who); host != "" {
			gopodReporterSay(host, l.Text, true, true)
			continue
		}
		gopodRecordSay(l.Who, l.Text)
	}
}

// gopodScriptWaitAnswered waits for each answer to start (the reporter reads
// the question first) and then to be said; with two answers (the sign-off), the
// brother's too, if he hasn't already answered.
func gopodScriptWaitAnswered(answers int) {
	deadline := time.Now().Add(8 * time.Minute)
	waitStart := 3 * time.Minute
	for i := 0; i < answers; i++ {
		start := time.Now().Add(waitStart)
		for !gopodReplyBusy() && time.Now().Before(start) {
			time.Sleep(200 * time.Millisecond)
		}
		for gopodReplyBusy() && time.Now().Before(deadline) {
			time.Sleep(200 * time.Millisecond)
		}
		waitStart = 10 * time.Second
	}
}

// gopodScriptAnswer: the draft's answer for this Brobot's question, while the
// draft is playing ("cue___ words"); taken once.
func gopodScriptAnswer(esn, question string) (string, bool) {
	gopodScript.Lock()
	defer gopodScript.Unlock()
	a, ok := gopodScript.pending[esn]
	if !ok {
		return "", false
	}
	if !gopodSameQuestion(gopodAddressWords(gopodScript.ask[esn]), gopodAddressWords(question)) {
		logger.Println("GOPOD_SCRIPT not the draft's question, the LLM answers: " + question)
		return "", false
	}
	delete(gopodScript.pending, esn)
	gopodScript.moves[esn] = append([]gopodDraftMove(nil), a.Moves...)
	gopodScript.logged[esn] = append([]gopodDraftMove(nil), a.Moves...)
	say := gopodScriptFill(esn, a.Say)
	if a.Cue == "" {
		return say, true
	}
	return a.Cue + gopodAnimationMarker + " " + say, true
}

// gopodScriptTakeLoggedMoves: the moves of this Brobot's scripted answer, once,
// for the session score (gopod_session_log.go).
func gopodScriptTakeLoggedMoves(esn string) []gopodDraftMove {
	gopodScript.Lock()
	defer gopodScript.Unlock()
	m := gopodScript.logged[esn]
	delete(gopodScript.logged, esn)
	return m
}

// gopodChatStream is the LLM's reply stream, or a draft's answer standing in for it.
type gopodChatStream interface {
	Recv() (openai.ChatCompletionStreamResponse, error)
}

type gopodScriptedStream struct {
	text string
	sent bool
}

func (s *gopodScriptedStream) Recv() (openai.ChatCompletionStreamResponse, error) {
	if s.sent {
		return openai.ChatCompletionStreamResponse{}, io.EOF
	}
	s.sent = true
	return openai.ChatCompletionStreamResponse{Choices: []openai.ChatCompletionStreamChoice{
		{Delta: openai.ChatCompletionStreamChoiceDelta{Content: s.text}},
	}}, nil
}

// gopodHomePath: "~/" as the home of the user who owns GOPOD's overlay folder
// (Wire-Pod runs as root).
func gopodHomePath(path string) string {
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	_, uid, _ := gopodRecordRoot()
	home := ""
	if u, err := user.LookupId(strconv.Itoa(uid)); err == nil {
		home = u.HomeDir
	}
	return filepath.Join(home, path[2:])
}

// gopodReporterHost: the reporter host ("Brobot 3") for a name or its label
// ("Reporter 1, Brobot 3"); "" when it isn't one.
func gopodReporterHost(who string) string {
	rv := gopodBrobotsSettings.ReporterVoices
	who = strings.TrimSpace(who)
	if _, ok := rv.Hosts[who]; ok {
		return who
	}
	for host, label := range rv.Labels {
		if strings.EqualFold(label, who) {
			return host
		}
	}
	return ""
}

// gopodScriptQuestionReader: the reporter who reads this line, when it's the
// playing draft's question (the one read_by would pick otherwise), or "".
func gopodScriptQuestionReader(to, text string) string {
	gopodScript.Lock()
	asking, host := gopodScript.asking, gopodScript.readBy
	gopodScript.Unlock()
	if !gopodScriptPlaying.Load() || asking == "" || strings.TrimSpace(text) != strings.TrimSpace(asking) {
		return ""
	}
	rv := gopodBrobotsSettings.ReporterVoices
	if host == "" {
		host = rv.ReadBy[to]
	}
	if host == "" {
		host = rv.OperatorHost
	}
	return host
}

// gopodScriptMovePhrases: the playing draft's move phrases (gopodJoinPieces).
func gopodScriptMovePhrases() []string {
	gopodScript.Lock()
	defer gopodScript.Unlock()
	return gopodScript.movesAt
}

// gopodScriptTakeReader: the reporter the playing draft named for this
// question, taken once.
func gopodScriptTakeReader() string {
	gopodScript.Lock()
	defer gopodScript.Unlock()
	host := gopodScript.readBy
	gopodScript.readBy = ""
	return host
}

// gopodBookends: the reporters' opening (or closing) - the draft's own when
// interview_script.bookends_from_draft is on and it has them, else
// reporter_voices'.
func gopodBookends(opening bool) []gopodHostLine {
	rv := gopodBrobotsSettings.ReporterVoices
	own := rv.Closing
	if opening {
		own = rv.Opening
		// 2026-10-20: the opener's _llm_song run, when the opener uses one.
		if lines, ok := gopodOpenerLLMSongLines(); ok {
			return lines
		}
	}
	if !gopodBrobotsSettings.InterviewScript.BookendsFromDraft {
		return own
	}
	_, draft, err := gopodLoadDraft()
	if err != nil {
		return own
	}
	from := draft.Closing
	if opening {
		from = draft.Opening
	}
	var lines []gopodHostLine
	for _, l := range from {
		if host := gopodReporterHost(l.Who); host != "" {
			lines = append(lines, gopodHostLine{Host: host, Text: l.Text})
		}
	}
	if len(lines) == 0 {
		return own
	}
	return lines
}

// ## EOF
