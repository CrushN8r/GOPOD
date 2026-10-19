// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_take_sheet.go
package wirepod_ttr

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The take sheet (operator 2026-10-06: raw runs spliced into one scripted
// interview, "like how a lead guitarist will record several takes and splice
// together"). One entry per value point, in interview order; under it, every
// take: from every harvest dry run (gopod_harvest_test.go) and, since "keep
// this session take?" (2026-10-06), from every live session the operator keeps
// (Record Session's keep, gopod_record_session.go). The operator marks the
// keeper ("keep": true) and edits its words ("say"); adding takes never changes
// those. take_sheet.md, to read, is drawn by
// gopod_probes/tools/gopod_take_sheet_render.py.

type gopodHarvestTake struct {
	Take       string   `json:"take"`
	Chapter    string   `json:"chapter"`
	Topic      string   `json:"topic"`
	Points     []string `json:"points"` // value points this answer was given ("brobot_zero 2/2")
	To         string   `json:"to"`
	Ask        string   `json:"ask"`
	Input      string   `json:"input"`     // typed / numpad / mic
	STTHeard   string   `json:"stt_heard"` // what speech-to-text heard (numpad and mic; empty for typed)
	Cue        string   `json:"cue"`       // the reply's animation cue
	Say        string   `json:"say"`       // the bubble text - edit to change the words
	Spoken     []string `json:"spoken"`    // the pieces as the robot said them
	LLMSeconds float64  `json:"llm_seconds"`
	Keep       bool     `json:"keep"` // mark the keeper
}

type gopodTakeSheetPoint struct {
	Chapter string             `json:"chapter"`
	Topic   string             `json:"topic"`
	Point   string             `json:"point"` // "2/5"
	Takes   []gopodHarvestTake `json:"takes"`
}

type gopodTakeSheet struct {
	About  string                `json:"about"`
	Runs   []string              `json:"runs"`
	Points []gopodTakeSheetPoint `json:"points"`
}

const gopodTakeSheetAbout = "The take sheet (operator 2026-10-06: raw runs spliced into one scripted interview). One entry per value point, in interview order; under it, every take from every harvest run and every kept live session. To pick: set \"keep\": true on the take you want (one per point). To change the words: edit its \"say\". A point with no takes needs another harvest. input is typed, numpad or mic; stt_heard is what speech-to-text heard. Made by gopod_take_sheet.go; adding takes never changes your marks or edits."

// gopodTakeSheetAdd adds a run's takes to the take sheet at path: a take goes
// under every point it was given, and under its own topic's first point when it
// was given none of its own. Takes already there, and their keep marks and
// edits, stay as they are.
func gopodTakeSheetAdd(path, run string, takes []gopodHarvestTake) error {
	var sheet gopodTakeSheet
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &sheet); err != nil {
			return fmt.Errorf("take sheet unreadable, not changed: %v", err)
		}
	}
	sheet.About = gopodTakeSheetAbout
	sheet.Runs = append(sheet.Runs, run)
	_, counts := gopodPackPointCounts()
	index := map[string]int{}
	for i, p := range sheet.Points {
		index[p.Topic+" "+p.Point] = i
	}
	for _, c := range gopodBrobotsSettings.InterviewChapters.Chapters {
		for _, topic := range c.Topics {
			for p := 1; p <= counts[topic]; p++ {
				key := fmt.Sprintf("%s %d/%d", topic, p, counts[topic])
				if _, ok := index[key]; !ok {
					index[key] = len(sheet.Points)
					sheet.Points = append(sheet.Points, gopodTakeSheetPoint{Chapter: c.Title, Topic: topic, Point: fmt.Sprintf("%d/%d", p, counts[topic])})
				}
			}
		}
	}
	for _, take := range takes {
		placed := false
		for _, pt := range take.Points {
			// Under every point it was given - its own topic's and any other
			// topic the question also loaded (live run 18:41: booking 2/3 was
			// only ever given alongside another topic's question).
			if i, ok := index[pt]; ok {
				sheet.Points[i].Takes = append(sheet.Points[i].Takes, take)
				placed = placed || strings.HasPrefix(pt, take.Topic+" ")
			}
		}
		if !placed {
			key := fmt.Sprintf("%s 1/%d", take.Topic, counts[take.Topic])
			if i, ok := index[key]; ok {
				sheet.Points[i].Takes = append(sheet.Points[i].Takes, take)
			}
		}
	}
	data, err := json.MarshalIndent(sheet, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// gopodTakeSheetRender redraws take_sheet.md next to the sheet.
func gopodTakeSheetRender(path string) error {
	tool := filepath.Join(gopodChipperRoot(), "gopod_probes/tools/gopod_take_sheet_render.py")
	out, err := gopodAsOperator(exec.Command("python3", tool, path)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// --- live session takes (Record Session) ---

var gopodInternalCue = regexp.MustCompile(`^\s*([A-Za-z][A-Za-z0-9-]*)\.\.\.\s+`)

// gopodLiveTake is one Brobot answer in a recorded live session, as a take.
// input: how the question came in - typed, numpad or mic (a spoken question
// with no typed or numpad line before it); stt_heard: the question as
// speech-to-text heard it (numpad and mic).
func gopodLiveTake(esn, question, input string, result gopodStreamResult) gopodHarvestTake {
	points := gopodGivenPoints(esn)
	sort.Strings(points)
	topic := ""
	if len(points) > 0 {
		topic = points[0][:strings.LastIndex(points[0], " ")]
	} else if packet, ok := loadScenarioPacketForMessage(question); ok {
		topic = packet.Name
	}
	say, cue := strings.TrimSpace(result.Raw), ""
	if m := gopodInternalCue.FindStringSubmatchIndex(say); m != nil && gopodIsAnimationTokenName(say[m[2]:m[3]]) {
		cue, say = say[m[2]:m[3]], say[m[1]:]
	}
	take := gopodHarvestTake{
		Chapter: gopodChapterOf(topic), Topic: topic, Points: points,
		To: gopodBrobotName(esn), Ask: question, Input: input, Cue: cue, Say: say,
		Spoken: gopodSpokenPieceTexts(result.AudibleChunks),
	}
	if input != "typed" {
		take.STTHeard = question
	}
	return take
}

// gopodTakeSheetPath is record_session.take_sheet, "~" the home of the user who
// owns GOPOD's overlay folder (Wire-Pod runs as root); "" when not set.
func gopodTakeSheetPath() (string, int, int) {
	path := strings.TrimSpace(gopodBrobotsSettings.RecordSession.TakeSheet)
	if env := strings.TrimSpace(os.Getenv("GOPOD_TAKE_SHEET")); env != "" {
		path = env
	}
	_, uid, gid := gopodRecordRoot()
	if strings.HasPrefix(path, "~/") {
		home := ""
		if u, err := user.LookupId(strconv.Itoa(uid)); err == nil {
			home = u.HomeDir
		}
		path = filepath.Join(home, path[2:])
	}
	return path, uid, gid
}

// gopodRecordTakesToSheet adds a kept session's takes to the take sheet, the
// session folder's name as the run, and redraws take_sheet.md.
func gopodRecordTakesToSheet(dir string, takes []gopodHarvestTake) error {
	path, uid, gid := gopodTakeSheetPath()
	if path == "" {
		return nil
	}
	if len(takes) == 0 {
		return errors.New("no Brobot answers in the session")
	}
	run := filepath.Base(dir)
	for i := range takes {
		takes[i].Take = fmt.Sprintf("%s#%02d", run, i+1)
	}
	if err := gopodTakeSheetAdd(path, run, takes); err != nil {
		return err
	}
	mdPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".md"
	if err := gopodTakeSheetRender(path); err != nil {
		logger.Println("GOPOD_TAKE_SHEET render: " + err.Error())
	}
	gopodRecordChown(uid, gid, filepath.Dir(path), path, mdPath)
	logger.Println(fmt.Sprintf("GOPOD_TAKE_SHEET added run=%s takes=%d sheet=%s", run, len(takes), path))
	return nil
}

// ## EOF
