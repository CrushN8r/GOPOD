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
// Every participant's prompt gets them (brain_chain.header first - the facts
// are true, the style is theirs); every brain bubble shows them; the session
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
	last, lastTo := none, ""
	if len(recs) > 0 {
		r := recs[len(recs)-1]
		last, lastTo = r.To+" - "+r.Kind, r.To
	}
	score, _ := gopodCoverageScoreText()
	join := func(l []string) string {
		if len(l) == 0 {
			return none
		}
		return strings.Join(l, ", ")
	}
	// The flow: what comes next, from what the chain shows.
	var flow []string
	if gopodIsIntro(question) {
		flow = append(flow, bc.Flow["intro"])
	} else {
		flow = append(flow, bc.Flow["answer"])
	}
	switch {
	case len(recs) == 0:
		flow = append(flow, bc.Flow["first"])
	case lastTo != "" && lastTo != gopodCallName(self):
		flow = append(flow, strings.ReplaceAll(bc.Flow["after"], "{last_to}", lastTo))
	}
	self0 := gopodCallName(self)
	if job := gopodPersonaProfileField(self, "My job title"); job != "" {
		self0 += " - " + strings.TrimSuffix(job, ".")
	}
	if role := gopodPersonaProfileField(self, "My role"); role != "" {
		self0 += "; " + role
	}
	return map[string]string{
		"{self}": self0, "{cast}": gopodIntroCast(self),
		"{last}": last, "{spoken}": join(spoken), "{introduced}": join(intro), "{score}": score,
		"{spoken_count}": strconv.Itoa(len(spoken)), "{introduced_count}": strconv.Itoa(len(intro)),
		"{flow}": strings.Join(flow, " "), "{asked}": bc.Asked, "{performed}": join(performed),
	}
}

func gopodBrainFill(tmpl string, facts map[string]string) string {
	for k, v := range facts {
		tmpl = strings.ReplaceAll(tmpl, k, v)
	}
	return strings.TrimSpace(tmpl)
}

// gopodBrainText: the brain for self's prompt - short-term, long-term, flow.
func gopodBrainText(self, question string) string {
	bc := gopodBrobotsSettings.BrainChain
	if strings.TrimSpace(bc.Header) == "" {
		return ""
	}
	facts := gopodBrainFacts(self, question)
	lines := []string{bc.Header}
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
	lines = append(lines, gopodBrainFill(bc.LongLine, facts), gopodBrainFill(bc.FlowLine, facts))
	text := strings.Join(lines, "\n")
	logger.Println("GOPOD_BRAIN fed " + gopodCallName(self) + ": " + strconv.Itoa(len(recs)) + " short-term, " + strconv.Itoa(len(text)) + " chars")
	return text
}

// gopodBrainBubbleLines: the brain bubble's memory lines (brain_chain.bubble_lines).
func gopodBrainBubbleLines(self, question string) []string {
	facts := gopodBrainFacts(self, question)
	var out []string
	for _, l := range gopodBrobotsSettings.BrainChain.BubbleLines {
		if line := gopodBrainFill(l, facts); line != "" {
			out = append(out, line)
		}
	}
	return out
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

// ## EOF
