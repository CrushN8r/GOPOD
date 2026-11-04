// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_topic_clusters_test.go
package wirepod_ttr

import (
	"encoding/json"
	"strings"
	"testing"
)

// 2026-10-31 (operator: "Topic-clusters please * Easter eggs * Golden answers *
// The Butler's lines * Nicknames"): the clusters are the one source - the
// settings files don't hold their values, and every value reaches Wire-Pod and
// the page.
func TestTopicClusters(t *testing.T) {
	if problems, n := gopodClustersCheck(); len(problems) > 0 || n < 40 {
		t.Fatalf("clusters: %d points, problems %v", n, problems)
	}
	s := gopodBrobotsSettings
	if !strings.HasPrefix(s.Session.StartedText, "{run} started. Right away, sir.") || !strings.Contains(s.Session.CompleteText, "optimal execution") {
		t.Errorf("the Butler's run lines: %q / %q", s.Session.StartedText, s.Session.CompleteText)
	}
	if len(s.GoldenAnswers) != 2 || s.GoldenAnswers[0].Queries[0] != "what is ai wordplay" || !strings.HasPrefix(s.GoldenAnswers[0].Line, "AI wordplay: the fun challenges") {
		t.Errorf("golden answers: %+v", s.GoldenAnswers)
	}
	if len(s.Nicknames.Titles) != 2 || len(s.Nicknames.List) < 10 || s.Nicknames.List[0] != "gear-head" {
		t.Errorf("nicknames: %+v", s.Nicknames)
	}
	tp := s.TopicPictures
	if len(tp.EggsKeep) != 5 || tp.EggNames["roast_battle"] != "Roast Battle" || len(tp.NotListed) != 3 || len(tp.NotEggs) != 2 {
		t.Errorf("easter eggs: keep %v names %v", tp.EggsKeep, tp.EggNames)
	}
	if len(s.ValuePoints.ReminderEggs) != 2 || s.ValuePoints.ReminderEggs[1].Label != "Get Photo" || s.Photo.EggName != "Get Photo" || len(s.Photo.AskPhrases) != 7 {
		t.Errorf("phrase eggs: %+v, photo %q %v", s.ValuePoints.ReminderEggs, s.Photo.EggName, s.Photo.AskPhrases)
	}
	// The page gets them too.
	data, _ := gopodSettingsMerged()
	var page map[string]json.RawMessage
	if json.Unmarshal(gopodClustersOverlay(data), &page) != nil || !strings.Contains(string(page["session"]), "Right away, sir") || !strings.Contains(string(page["golden_answers"]), "AI wordplay") {
		t.Error("the page's settings don't carry the clusters")
	}
	// The settings files themselves don't hold them (one source).
	if strings.Contains(string(data), "Executing your commands through Goverlord") || strings.Contains(string(data), "squeaky treads") || strings.Contains(string(data), "\"egg_names\"") {
		t.Error("a cluster's value is still in the settings files")
	}
}

// 2026-11-03 (step 1 of the clusters plan: the shell window's words): the system
// texts and the help list come from system-texts.md and system-help.md - a
// status line keeps its leading spaces, "\n" is a break, " | " stays words, and
// the help keeps its columns.
func TestSystemTextsFromClusters(t *testing.T) {
	st := gopodBrobotsSettings.SystemText
	if len(st) < 150 {
		t.Fatalf("system texts: %d, want the cluster's ~168", len(st))
	}
	if v := st["system_status.page_here"]; !strings.HasPrefix(v, "  chat page") {
		t.Errorf("status line lost its column: %q", v)
	}
	if v := st["system_robots.system_brobot_beat_alias_1"]; !strings.Contains(v, "\n{out2}") {
		t.Errorf("line break lost: %q", v)
	}
	if v := st["system_edit.system_topic_topic_words_add"]; !strings.Contains(v, " | words ") {
		t.Errorf("\" | \" not kept as words: %q", v)
	}
	help := gopodBrobotsSettings.SystemHelp.Lines
	if len(help) < 50 || !strings.HasPrefix(help[1], "  system-set ") {
		t.Errorf("help list or its columns lost: %d lines, %q", len(help), help)
	}
}

// 2026-11-03 (operator: "one CHALK file"; step 2): each persona's settings come
// from its own profile's "## My settings (managers.personas.<HOST>)" - and the
// profile's own sections stay out of the settings.
func TestPersonaSettingsFromProfiles(t *testing.T) {
	per := gopodBrobotsSettings.Managers.Personas
	for _, host := range []string{"CHALK", "PLAYHEAD", "GIT REPO", "CACHE PYC", "INTEL 404 PNFH", "INTEL 404 PNFT", "CAMEO", "BUTLER"} {
		if p, ok := per[host]; !ok || p.Host != host || p.Prompt == "" || p.WhoLine == "" {
			t.Errorf("%s: settings not read from its profile: %+v", host, p)
		}
	}
	if len(per["PLAYHEAD"].WIP) != 1 || per["BUTLER"].VoiceOf != "Goverlord" {
		t.Errorf("PLAYHEAD's wip list or the Butler's voice_of lost: %q %q", per["PLAYHEAD"].WIP, per["BUTLER"].VoiceOf)
	}
	if v := gopodPersonaProfileText("CHALK"); strings.Contains(v, "who_line:") {
		t.Errorf("the settings section leaked into CHALK's profile lines: %q", v)
	}
}

// 2026-11-03 (step 3: the Brobots' brain and reply rules): brobot-brain.md,
// reply-rules.md and big-picture.md fill the brain - a backticked value keeps its
// leading space, the brain's lines are a line each, "\n" stays a break.
func TestBrainFromClusters(t *testing.T) {
	bc := gopodBrobotsSettings.BrainChain
	if bc.CastMe != " (me)" || !strings.HasPrefix(bc.RemindersPart, " Reminders:") {
		t.Errorf("backticked texts lost their space: %q %q", bc.CastMe, bc.RemindersPart)
	}
	if len(bc.Lines) < 3 || bc.Lines[0] != "{full_self}" {
		t.Errorf("brain lines: %q", bc.Lines)
	}
	if len(gopodBrobotsSettings.ReplyTighten.NotSentenceEnds) == 0 || !strings.Contains(gopodBrobotsSettings.Summary.FactsNote, "\n") {
		t.Errorf("reply rules or the summary's line break lost")
	}
}

// 2026-11-03 (step 4: spice, here and now, host pairs): spice.md,
// here-and-now.md and host-pairs.md fill them; each pair's dial stays in the
// settings, laid under its words; "; " keeps its space.
func TestCastWordsFromClusters(t *testing.T) {
	if len(gopodBrobotsSettings.Spice.Exclamations) == 0 || gopodBrobotsSettings.Spice.Text == "" {
		t.Error("spice not read from spice.md")
	}
	if gopodBrobotsSettings.HereNow.CastJoin != "; " || gopodBrobotsSettings.HereNow.Genders["Brobot 1"] == "" {
		t.Errorf("here and now: %q %q", gopodBrobotsSettings.HereNow.CastJoin, gopodBrobotsSettings.HereNow.Genders)
	}
	p := gopodBrobotsSettings.HostPairs["Doc & Pip"]
	if len(p.Hosts) != 2 || p.Banter == "" || p.Dial == 0 {
		t.Errorf("Doc & Pip's pair (words and dial): %+v", p)
	}
}

// 2026-11-03 (step 5: the Goverlord level): the reporters' opening and closing
// and the sign-off are "who: line" points in preshow.md, in order; the page's
// words and the Butler's own persona text come from their files.
func TestPreshowFromClusters(t *testing.T) {
	rv := gopodBrobotsSettings.ReporterVoices
	if len(rv.Opening) < 5 || rv.Opening[0].Host != "Brobot 3" || !strings.HasPrefix(rv.Opening[0].Text, "Okay. We're about to go live.") {
		t.Errorf("opening: %+v", rv.Opening)
	}
	if len(rv.Closing) == 0 || len(gopodBrobotsSettings.SignOff.Lines) == 0 || rv.SystemPersona == "" {
		t.Error("closing, sign-off or the Butler's persona not read from their files")
	}
}

// 2026-11-03: system-set on a value a cluster fills says which file to edit -
// it never writes the settings file the cluster would win over.
func TestSystemSetNamesTheCluster(t *testing.T) {
	if f := gopodClusterHolds("spice.exclamations"); f != "spice.md" {
		t.Errorf("spice.exclamations: %q, want spice.md", f)
	}
	if f := gopodClusterHolds("managers.personas.CHALK.prompt"); f != "chalk.md" {
		t.Errorf("CHALK's prompt: %q, want chalk.md", f)
	}
	if f := gopodClusterHolds("spice.pick"); f != "" {
		t.Errorf("spice.pick is in the settings, not %q", f)
	}
	if out := gopodSystemSetPath("spice.text", "x"); !strings.Contains(out, "spice.md") {
		t.Errorf("system-set on a cluster value: %q", out)
	}
}

// 2026-11-03 (step 6: the game, pictures, math, photo, notices): their words
// come from game-words.md, topic-pictures.md, math.md, photo.md and
// bubble-notices.md - a card keeps its line break, "• " its space.
func TestOtherWordsFromClusters(t *testing.T) {
	s := gopodBrobotsSettings
	if s.Math.Bullet != "• " || !strings.Contains(s.TopicPictures.Cards["roast_battle"], "\n") {
		t.Errorf("math bullet or a card's line break lost: %q %q", s.Math.Bullet, s.TopicPictures.Cards["roast_battle"])
	}
	if s.Game.WinText == "" || len(s.Photo.StallLines) == 0 || s.BubbleNotices.ValuePoints.Score == "" {
		t.Error("game, photo or the notice's score line not read from its file")
	}
}

// 2026-11-03 (step 7, the format step): a number in the place is a list's entry,
// "(path[])" a line list, a bare number/true/false/[] is typed, a backticked
// number stays text, and an entry out of order is a problem, not a guess.
func TestClusterFormatNestedAndTyped(t *testing.T) {
	all := map[string]json.RawMessage{"x": json.RawMessage(`{"keep": 1}`)}
	blocks := map[string]any{}
	var v any
	_ = json.Unmarshal(all["x"], &v)
	blocks["x"] = v
	node, set, bad := gopodClusterPlace(blocks, "x", []string{"items", "0"})
	if bad != "" || node != nil {
		t.Fatalf("first entry: %v %q", node, bad)
	}
	set(map[string]any{"cmd": "a"})
	if _, _, bad := gopodClusterPlace(blocks, "x", []string{"items", "2"}); bad == "" {
		t.Error("entry 2 before entry 1 wasn't a problem")
	}
	if n, ok := gopodClusterTyped("60"); !ok || n != float64(60) {
		t.Errorf("60: %v %v", n, ok)
	}
	if b, ok := gopodClusterTyped("true"); !ok || b != true {
		t.Errorf("true: %v", b)
	}
	if l, ok := gopodClusterTyped("[]"); !ok || len(l.([]any)) != 0 {
		t.Errorf("[]: %v", l)
	}
	if _, ok := gopodClusterTyped("29%"); ok {
		t.Error("29% read as a number")
	}
	if blocks["x"].(map[string]any)["keep"] != float64(1) {
		t.Error("the settings' own value was lost")
	}
	// The live files: the menu's scriptures, a tour's lines, a command's switch.
	merged, _ := gopodSettingsMerged()
	var page struct {
		SystemMenu struct {
			Groups []struct {
				Tour  []string `json:"tour"`
				Items []struct {
					Cmd string `json:"cmd"`
					Run bool   `json:"run"`
				} `json:"items"`
			} `json:"groups"`
		} `json:"system_menu"`
	}
	_ = json.Unmarshal(gopodClustersOverlay(merged), &page)
	menu := page.SystemMenu.Groups
	if len(menu) != 10 || len(menu[0].Tour) == 0 || menu[0].Items[2].Cmd != "system-status" || !menu[0].Items[2].Run {
		t.Errorf("system menu from system-menu.md: %d groups", len(menu))
	}
	if len(gopodBrobotsSettings.Managers.Takes) != 4 || len(gopodBrobotsSettings.HereNow.Cast) == 0 {
		t.Error("takes or the cast not read from their files")
	}
}

// 2026-11-03 (step 7b, the remaining groups): the run cards, the topics' map,
// DocBS, the roast dial, addressing and the brand words come from their files -
// an empty list or group ("[]", "{}") stays one.
func TestNestedGroupsFromClusters(t *testing.T) {
	merged, _ := gopodSettingsFull()
	var s struct {
		Game struct {
			Runs []map[string]any `json:"runs"`
		} `json:"game"`
		ValuePoints struct {
			Funnels    [][]string          `json:"funnels"`
			CoreTopics map[string][]string `json:"core_topics"`
		} `json:"value_points"`
	}
	if err := json.Unmarshal(merged, &s); err != nil {
		t.Fatal(err)
	}
	if len(s.Game.Runs) != 4 || len(s.ValuePoints.Funnels) != 3 || len(s.ValuePoints.CoreTopics) == 0 {
		t.Errorf("runs %d, funnels %d, core topics %d", len(s.Game.Runs), len(s.ValuePoints.Funnels), len(s.ValuePoints.CoreTopics))
	}
	if exp, _ := s.Game.Runs[0]["experience"].(map[string]any); exp == nil || exp["takes"] == nil {
		t.Errorf("an empty group was lost: %v", exp)
	}
	bs := gopodBrobotsSettings
	if len(bs.DocBS.Chains) == 0 || len(bs.RoastDial.Levels) != 4 || len(bs.Addressing.Names["Brobot 1"]) == 0 || len(bs.ReplyTighten.BrandWords) != 4 {
		t.Error("DocBS, the roast dial, addressing or the brand words not read from their files")
	}
}

// 2026-11-03 (step 8 of the clusters plan): the texts left in the settings
// files only come down - more than topic_clusters.words_left_allowed fails the
// deploy, so a new spoken line goes into a topic-cluster, not the settings.
func TestWordsLeftOnlyComeDown(t *testing.T) {
	left, allowed := gopodWordsLeft()
	if allowed <= 0 {
		t.Fatal("topic_clusters.words_left_allowed isn't set")
	}
	if len(left) > allowed {
		t.Errorf("%d texts in the settings files, %d allowed - put the new ones in a topic-cluster: %q", len(left), allowed, left)
	}
}
