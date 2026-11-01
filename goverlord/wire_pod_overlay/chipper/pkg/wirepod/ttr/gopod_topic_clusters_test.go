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
