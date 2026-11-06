// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_brain_chain.go
package wirepod_ttr

// The brain chain (2026-10-31, operator: "go for the brain chain * brain chain
// should be apparent everywhere * KP test runs, should be all about mechanical
// diagnostics that feed the llm the cleanest sufficient data set for short-term
// memory & orientation, while maintaining long term memory & orientation"; "flow
// guide for short & long term, self & chat memory handling, within the live chat
// session"; "Like a chat flow guide/helper line or 2"; "Big Brains = code so llm
// doesn't think & is more free to add persona styles").
//
// Every answer in the live chat adds one record, made by code - never by the
// language model, so nothing an answer coloured is carried on: who asked, who
// answered, the line asked, its kind (intro, answer). From the chain:
//
//	short-term   the last brain_chain.short_term records, newest last
//	long-term    the session so far - the game's score, who has spoken, who's
//	             introduced themselves
//	flow         one or two helper lines for what comes next
//
// One brain template (brain_chain.lines, the operator's 2026-10-31 scaffolding:
// Self, Asked by, In the chat, Session, Guided Task Flow): every brain bubble
// shows its lines, and every participant's prompt gets the same lines (after
// brain_chain.header - the facts are true, the style is theirs), then the recent
// chat and the long-term line; the session
// keeps them (brain_chain.jsonl in its folder); the thanks ending starts a new
// chain. GOPOD_BRAIN lines in the journal; system-brain shows the chain.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

type gopodBrainRecord struct {
	Seq      int    `json:"seq"`
	At       string `json:"at"`
	Asked    string `json:"asked"`
	To       string `json:"to"`
	Question string `json:"question"`
	Kind     string `json:"kind"`
}

var gopodBrainChain = struct {
	sync.Mutex
	recs []gopodBrainRecord
}{}

// gopodBrainAdd: one answer's record - who answered (his call name), the line, its kind.
func gopodBrainAdd(to, question, kind string) {
	bc := gopodBrobotsSettings.BrainChain
	q := strings.Join(strings.Fields(question), " ")
	if bc.QuestionChars > 0 && len(q) > bc.QuestionChars {
		q = gopodCut(q, bc.QuestionChars)
	}
	gopodBrainChain.Lock()
	rec := gopodBrainRecord{Seq: len(gopodBrainChain.recs) + 1, At: time.Now().Format(time.RFC3339), Asked: bc.Asked, To: to, Question: q, Kind: kind}
	gopodBrainChain.recs = append(gopodBrainChain.recs, rec)
	gopodBrainChain.Unlock()
	logger.Println("GOPOD_BRAIN #" + strconv.Itoa(rec.Seq) + " " + rec.Asked + " to " + rec.To + " (" + rec.Kind + "): " + rec.Question)
	if dir := gopodSessionLogDir(); dir != "" {
		if line, err := json.Marshal(rec); err == nil {
			if f, err := os.OpenFile(filepath.Join(dir, "brain_chain.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
				f.Write(append(line, '\n'))
				f.Close()
			}
		}
	}
}

// gopodBrainReset: a new chain (the thanks ending, the comm check's end). The runs
// the Brobots performed are kept until the session ends (keepPerformed).
func gopodBrainReset(keepPerformed ...bool) {
	gopodBrainChain.Lock()
	var kept []gopodBrainRecord
	if len(keepPerformed) > 0 && keepPerformed[0] {
		for _, r := range gopodBrainChain.recs {
			if r.Kind == gopodBrainPerformedKind {
				r.Seq = len(kept) + 1
				kept = append(kept, r)
			}
		}
	}
	gopodBrainChain.recs = kept
	gopodBrainChain.Unlock()
}

// gopodBrainPerformedKind: a run the Brobots performed (a song, the interview).
const gopodBrainPerformedKind = "performed"

// gopodBrainPerformed: a run's record - who performed it (brain_chain.performers)
// and its name; stopped runs say so.
func gopodBrainPerformed(run, reason string) {
	what := run
	if reason == "stopped" {
		what += " (stopped)"
	}
	who := gopodBrobotsSettings.BrainChain.Performers
	if who == "" {
		who = "Doc and Pip"
	}
	gopodBrainAdd(who, what, gopodBrainPerformedKind)
}

func gopodBrainSnapshot() []gopodBrainRecord {
	gopodBrainChain.Lock()
	defer gopodBrainChain.Unlock()
	return append([]gopodBrainRecord(nil), gopodBrainChain.recs...)
}

// gopodBrainFacts: the values every brain template fills from.
func gopodBrainFacts(self, question string) map[string]string {
	bc := gopodBrobotsSettings.BrainChain
	recs := gopodBrainSnapshot()
	var spoken, intro []string
	seenS, seenI := map[string]bool{}, map[string]bool{}
	var performed []string
	for _, r := range recs {
		if r.Kind == gopodBrainPerformedKind {
			performed = append(performed, r.Question)
			continue
		}
		if !seenS[r.To] {
			seenS[r.To] = true
			spoken = append(spoken, r.To)
		}
		if r.Kind == "intro" && !seenI[r.To] {
			seenI[r.To] = true
			intro = append(intro, r.To)
		}
	}
	none := bc.None
	score, _ := gopodCoverageScoreText()
	join := func(l []string) string {
		if len(l) == 0 {
			return none
		}
		return strings.Join(l, ", ")
	}
	// The flow: what comes next, from what the chain shows - an intro's task, then
	// the first to speak, or the recent chat (flow.recent; 2026-10-31, operator:
	// "Guided Task Flow: Does the recent chat help me answer the question asked").
	// 2026-11-02 (operator's brain bubble: "Guided Task Flow: a) Introduce
	// yourself. b) Name who's in the chat right now. c) If you're the first to
	// speak this session, open it up."; live: Doc's intro to All went to the
	// weather, steered by "Does the recent chat help..."): an intro's flow is the
	// intro task alone.
	var flow []string
	switch {
	case gopodIsIntro(question):
		// 2026-11-03 (operator's group chat intro: "c) Keys will be the first to
		// speak, to open up the session"; Keys's own: "You're first - open it").
		if o := strings.TrimSpace(bc.IntroOpener); o != "" && strings.EqualFold(gopodCallName(self), o) && bc.Flow["intro_first"] != "" {
			flow = append(flow, bc.Flow["intro_first"])
		} else {
			flow = append(flow, bc.Flow["intro"])
		}
	case len(recs) == 0:
		flow = append(flow, bc.Flow["first"])
	default:
		flow = append(flow, strings.ReplaceAll(bc.Flow["recent"], "{asked}", bc.Asked))
	}
	// Self (2026-10-31, operator: "I am Pip (Brobot 2), a party-bro GOPOD sales rep.
	// I'm clueless & happy. Brobot 1 & I are robot brothers."): his call name, his
	// name when it's another, then his profile's "In short" (else his job title).
	self0 := gopodCallName(self)
	if name := strings.TrimSuffix(gopodPersonaProfileField(self, "My name"), "."); name != "" && !strings.EqualFold(name, self0) {
		self0 += " (" + name + ")"
	}
	if short := gopodBrainDescription(self, true); short != "" {
		self0 += ", " + short
	}
	// In the chat (operator: "In the chat: Doc [description from md] (me), Pip
	// [description from md], ..."): everyone here but the operator, in cast order,
	// each with his profile's job title (brain_chain.cast_item); he is "(me)".
	var described []string
	for _, c := range gopodBrobotsSettings.HereNow.Cast {
		if c.Who == bc.OperatorWho {
			continue
		}
		me := ""
		if c.Who == self || gopodCallName(c.Who) == gopodCallName(self) {
			me = bc.CastMe
		}
		described = append(described, strings.TrimSpace(strings.NewReplacer("{name}", gopodCallName(c.Who),
			"{description}", gopodBrainDescription(c.Who, false), "{me}", me).Replace(bc.CastItem)))
	}
	castLine := strings.Join(described, ", ")
	if g := gopodCastGrouped(self); g != "" {
		castLine = g
	}
	// Reminders: the topics his reminder names this answer (gopodBrainReminders).
	reminders := ""
	if v, ok := gopodBrainReminders.Load(self); ok {
		if names := v.([]string); len(names) > 0 {
			reminders = strings.ReplaceAll(bc.RemindersPart, "{reminders}", gopodBrainList(names))
		}
	}
	return map[string]string{
		"{self}": self0, "{cast}": gopodIntroCast(self), "{cast_described}": castLine, "{reminders_part}": reminders,
		"{spoken}": join(spoken), "{introduced}": join(intro), "{score}": score,
		"{spoken_count}": strconv.Itoa(len(spoken)), "{introduced_count}": strconv.Itoa(len(intro)),
		"{flow}": strings.Join(flow, " "), "{asked}": bc.Asked, "{performed}": join(performed),
		"{relationships}": gopodBrainRelationships(self),
		"{full_self}":     gopodBrainFullSelf(self),
	}
}

// gopodBrainReminders: the reminder topics of each Brobot's answer now (his name -> []string).
var gopodBrainReminders sync.Map

// gopodBrainDescription: who's description - "In short" from his profile when
// self (and he has one); else his line in the crew roster (gopodCrewRoster),
// else his profile's job title, its last full stop off.
func gopodBrainDescription(who string, self bool) string {
	if self {
		if v := strings.TrimSpace(gopodPersonaProfileField(who, "In short")); v != "" {
			return v
		}
	}
	roster := gopodCrewRoster()
	if v := roster[strings.ToLower(gopodCallName(who))]; v != "" {
		return v
	}
	if v := roster[strings.ToLower(who)]; v != "" {
		return v
	}
	return strings.TrimSuffix(strings.TrimSpace(gopodPersonaProfileField(who, "My job title")), ".")
}

// gopodCrewRoster (2026-10-31, operator: "This can be a full GOPOD crew roster
// (sub) topic-cluster? * Crew roster please"): the crew roster - one topic-cluster
// in the master format (brain_chain.roster, its "## Crew" points "Name:
// description") - call name (small letters) -> description.
func gopodCrewRoster() map[string]string {
	out := map[string]string{}
	for _, g := range gopodRosterGroups() {
		for i, n := range g.Names {
			out[strings.ToLower(n)] = g.Descs[i]
		}
	}
	return out
}

// gopodRosterGroup: one of the crew roster's groups - its "## " label, then its
// points "Name: description", in order.
type gopodRosterGroup struct {
	Label string
	Names []string
	Descs []string
}

// gopodRosterGroups (2026-11-02, operator: "GOPOD party-brobot reps: Brobot 1
// [grumpy, impatient], Brobot 2 [clueless, happy]. CRUSHN8R CREW'd news station
// Gomads mentoring interns: ..."): the roster's groups in file order.
func gopodRosterGroups() []gopodRosterGroup {
	path := gopodTopicPath(gopodBrobotsSettings.BrainChain.Roster) // moved, found by its file name
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var out []gopodRosterGroup
	inBody := false
	for _, line := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(line)
		if !inBody {
			inBody = t == "---"
			continue
		}
		if strings.HasPrefix(t, "## ") {
			out = append(out, gopodRosterGroup{Label: strings.TrimSpace(t[3:])})
			continue
		}
		m := gopodProfilePointRE.FindStringSubmatch(t)
		if m == nil || len(out) == 0 || strings.HasSuffix(m[1], "(to come)") {
			continue
		}
		if at := strings.Index(m[1], ":"); at > 0 {
			g := &out[len(out)-1]
			g.Names = append(g.Names, strings.TrimSpace(m[1][:at]))
			g.Descs = append(g.Descs, strings.TrimSuffix(strings.TrimSpace(m[1][at+1:]), "."))
		}
	}
	return out
}

// gopodCastGrouped: In the chat by the roster's groups - "Group: Name
// [description], Name [description]. Next group: ..." - everyone here but the
// operator, self marked (brain_chain.cast_me). "" with one group (the flat line).
func gopodCastGrouped(self string) string {
	groups := gopodRosterGroups()
	if len(groups) < 2 {
		return ""
	}
	bc := gopodBrobotsSettings.BrainChain
	here := map[string]string{} // a roster name, small letters -> who
	for _, c := range gopodBrobotsSettings.HereNow.Cast {
		if c.Who == bc.OperatorWho {
			continue
		}
		here[strings.ToLower(c.Who)] = c.Who
		here[strings.ToLower(gopodCallName(c.Who))] = c.Who
	}
	var parts []string
	for _, g := range groups {
		var items []string
		for i, n := range g.Names {
			who, ok := here[strings.ToLower(n)]
			if !ok {
				continue
			}
			me := ""
			if who == self || (self != "" && gopodCallName(who) == gopodCallName(self)) {
				me = bc.CastMe
			}
			items = append(items, strings.TrimSpace(strings.NewReplacer("{name}", n, "{description}", g.Descs[i], "{me}", me).Replace(bc.CastItem)))
		}
		if len(items) > 0 {
			parts = append(parts, g.Label+": "+strings.Join(items, ", "))
		}
	}
	// The approved visitors talking now (participation ON, box ticked).
	if names := gopodVisitorsTalking(); len(names) > 0 {
		label := strings.TrimSpace(bc.VisitorsLabel)
		if label == "" {
			label = "Visitors in the chat"
		}
		parts = append(parts, label+": "+strings.Join(names, ", "))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ". ") + "."
}

// gopodBrainBubbleLines: the brain template's lines as a brain bubble shows them -
// without the lines starting with brain_chain.bubble_hide (the swarm list is
// PLAYHEAD's to say; every prompt still gets it, gopodBrainText).
func gopodBrainBubbleLines(self, question string) []string {
	var out []string
	for _, l := range gopodBrainLines(self, question) {
		hide := false
		for _, h := range gopodBrobotsSettings.BrainChain.BubbleHide {
			if h = strings.TrimSpace(h); h != "" && strings.HasPrefix(l, h) {
				hide = true
				break
			}
		}
		if !hide {
			out = append(out, l)
		}
	}
	return out
}

// gopodWhosHere: the chat's people as PLAYHEAD says them - everyone here by the
// names they're called, then the talking visitors by their approved names.
func gopodWhosHere() []string {
	var names []string
	for _, n := range strings.Split(gopodIntroCast(""), ", ") {
		if n = strings.TrimSpace(n); n != "" {
			names = append(names, n)
		}
	}
	return append(names, gopodVisitorsTalking()...)
}

var gopodWhosHereLast struct {
	sync.Mutex
	line string
}

// gopodWhosHereAnnounce (2026-11-02, operator: "PLAYHEAD Gomad generates a dynamic
// In the chat list as participants join/leave"): whos_here_host says who's here,
// when it's changed since he last said it; wait: until it's been heard.
func gopodWhosHereAnnounce(reason string, wait bool) {
	bc := gopodBrobotsSettings.BrainChain
	host, text := strings.TrimSpace(bc.WhosHereHost), strings.TrimSpace(bc.WhosHereText)
	names := gopodWhosHere()
	if host == "" || text == "" || len(names) == 0 {
		return
	}
	line := strings.ReplaceAll(text, "{names}", gopodBrainList(names))
	gopodWhosHereLast.Lock()
	same := line == gopodWhosHereLast.line
	gopodWhosHereLast.line = line
	gopodWhosHereLast.Unlock()
	if same {
		return
	}
	logger.Println("GOPOD_WHOS_HERE (" + reason + ") " + host + ": " + line)
	gopodReporterSayTo(host, "", line, true, wait)
}

// gopodBrainList: "a, b, c, & d" (operator: "Reminders: [], [], [], & []").
func gopodBrainList(names []string) string {
	if len(names) < 2 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + ", & " + names[len(names)-1]
}

func gopodBrainFill(tmpl string, facts map[string]string) string {
	for k, v := range facts {
		tmpl = strings.ReplaceAll(tmpl, k, v)
	}
	return strings.TrimSpace(tmpl)
}

// gopodBrainLines: the brain template's lines for self (brain_chain.lines) - the
// one template every brain bubble shows and every prompt gets. A line with
// nothing after its label isn't kept.
func gopodBrainLines(self, question string) []string {
	facts := gopodBrainFacts(self, question)
	var out []string
	for _, l := range gopodBrobotsSettings.BrainChain.Lines {
		if line := gopodBrainFill(l, facts); line != "" && !strings.HasSuffix(line, ":") {
			out = append(out, line)
		}
	}
	return out
}

// gopodBrainText: the brain for self's prompt - the header, the brain template's
// lines (the same as his brain bubble), the recent chat, the long-term line.
func gopodBrainText(self, question string) string {
	bc := gopodBrobotsSettings.BrainChain
	if strings.TrimSpace(bc.Header) == "" {
		return ""
	}
	facts := gopodBrainFacts(self, question)
	lines := append([]string{bc.Header}, gopodBrainLines(self, question)...)
	recs := gopodBrainSnapshot()
	if n := bc.ShortTerm; n > 0 && len(recs) > n {
		recs = recs[len(recs)-n:]
	}
	if len(recs) > 0 {
		lines = append(lines, bc.ShortHeader)
		for _, r := range recs {
			lines = append(lines, strings.NewReplacer("{seq}", strconv.Itoa(r.Seq), "{asked}", r.Asked, "{to}", r.To,
				"{question}", r.Question, "{kind}", r.Kind).Replace(bc.RecordLine))
		}
	}
	lines = append(lines, gopodBrainFill(bc.LongLine, facts))
	text := strings.Join(lines, "\n")
	logger.Println("GOPOD_BRAIN fed " + gopodCallName(self) + ": " + strconv.Itoa(len(recs)) + " short-term, " + strconv.Itoa(len(text)) + " chars")
	return text
}

// gopodSystemBrain: system-brain - the chain as it is now.
func gopodSystemBrain() string {
	recs := gopodBrainSnapshot()
	facts := gopodBrainFacts("", "")
	out := []string{"Brain chain: " + strconv.Itoa(len(recs)) + " records"}
	for _, r := range recs {
		out = append(out, "  #"+strconv.Itoa(r.Seq)+" "+r.Asked+" to "+r.To+" ("+r.Kind+"): "+r.Question)
	}
	out = append(out, gopodBrainFill(gopodBrobotsSettings.BrainChain.LongLine, facts))
	return gopodJoinLines(out)
}

// gopodBrainFullSelf (2026-11-02, operator's brain bubble: "My full name is
// KEYNOTE VALE, known at GOPOD as 'Keys' the Butler ..."): his full name and
// job from his profile, then his profile's "In short".
func gopodBrainFullSelf(who string) string {
	if strings.TrimSpace(who) == "" {
		return ""
	}
	first := func(v string) string {
		if at := strings.Index(v, " - "); at > 0 {
			v = v[:at]
		}
		if at := strings.Index(v, ". "); at > 0 {
			v = v[:at]
		}
		return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(v), "."))
	}
	head, _ := gopodPersonaProfileRead(gopodPersonaProfileFile(who))
	name := first(gopodPersonaProfileField(who, "My name"))
	if name == "" && head != nil {
		name = head["label"]
	}
	if name == "" {
		name = gopodCallName(who)
	}
	job := ""
	if head != nil {
		job = strings.TrimSpace(gopodPersonaFill(who, head["job"]))
	}
	// 2026-11-06 (operator: "go with formal first names on company time * let's
	// keep nicknames for bantering"): the brain says his formal name, not his nickname.
	out := "My full name is " + name
	if job != "" {
		out += ", the " + job
	}
	out += "."
	if short := strings.TrimSpace(gopodPersonaProfileField(who, "In short")); short != "" {
		short = strings.ToUpper(short[:1]) + short[1:]
		if !strings.HasSuffix(short, ".") {
			short += "."
		}
		out += " " + short
	}
	return out
}

// ## EOF
