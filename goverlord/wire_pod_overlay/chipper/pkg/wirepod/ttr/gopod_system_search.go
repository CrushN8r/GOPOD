// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_search.go
package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// system-search (2026-10-17, operator: "Expire go search, replace with
// \"system-search\" * \"system-search\" is used for live guided search responses *
// brobots use \"system-search\" for 'search [search-text]' like 'stat [run-name]'"):
// GOPOD's own SearXNG (the gopod-searxng container, wire_pod_overlay/searxng/,
// settings system_search.url) is asked, with no LLM in the search. A line to the
// Brobots that says "search <words>" gets the top results' short snippets as
// facts (gopodQuestionFactsFor), and the reply opens with the first one, said
// by the code word for word (gopodScriptedSnippet) - like the weather line.
// Typed "system-search <words>" lists the results in the terminal bubble.

type gopodSearchResult struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

var gopodSearchCache = struct {
	sync.Mutex
	got map[string][]gopodSearchResult
	at  map[string]time.Time
}{got: map[string][]gopodSearchResult{}, at: map[string]time.Time{}}

// gopodSearchWordsRE: "search", "look up" at the start of a line, and what's after.
// Like "start [run-name]": the line opens with it, after a Brobot's name at most
// ("Brobot 1, search Toronto Ohio").
var gopodSearchWordsRE = regexp.MustCompile(`(?i)^(?:(?:hey\s+)?(?:brobots?|brobot\s*[0-9]|doc|pip)\s*[,:]?\s*)?(?:search(?:\s+for)?|look\s+up)\s+(.+)$`)

// gopodSearchQuery: the words a line asks to search for, or "".
func gopodSearchQuery(message string) string {
	if strings.TrimSpace(gopodBrobotsSettings.SystemSearch.URL) == "" {
		return ""
	}
	m := gopodSearchWordsRE.FindStringSubmatch(strings.TrimSpace(message))
	if m == nil {
		return ""
	}
	q := strings.Trim(strings.TrimSpace(m[1]), " ?!.,")
	if len(strings.Fields(q)) == 0 {
		return ""
	}
	return q
}

// gopodSearch asks GOPOD's SearXNG; the same query within two minutes is the
// same answer (both Brobots say the same facts).
func gopodSearch(query string) ([]gopodSearchResult, error) {
	ss := gopodBrobotsSettings.SystemSearch
	key := strings.ToLower(strings.Join(strings.Fields(query), " "))
	gopodSearchCache.Lock()
	if got, ok := gopodSearchCache.got[key]; ok && time.Since(gopodSearchCache.at[key]) < 2*time.Minute {
		gopodSearchCache.Unlock()
		return got, nil
	}
	gopodSearchCache.Unlock()
	timeout := gopodSecondsToDuration(ss.TimeoutSeconds)
	started := time.Now()
	resp, err := (&http.Client{Timeout: timeout}).Get(strings.TrimRight(ss.URL, "/") + "?format=json&q=" + url.QueryEscape(query))
	if err != nil {
		logger.Println("GOPOD_SEARCH " + query + ": no answer - " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Println(fmt.Sprintf("GOPOD_SEARCH %s: answered %d", query, resp.StatusCode))
		return nil, fmt.Errorf("search answered %d", resp.StatusCode)
	}
	var body struct {
		Results []gopodSearchResult `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	most := ss.Results
	if most <= 0 {
		most = 3
	}
	var got []gopodSearchResult
	for _, r := range body.Results {
		r.Title, r.Content = strings.TrimSpace(r.Title), strings.Join(strings.Fields(r.Content), " ")
		if r.Content == "" || r.Title == "" {
			continue
		}
		got = append(got, r)
		if len(got) == most {
			break
		}
	}
	logger.Println(fmt.Sprintf("GOPOD_SEARCH %s: %d result(s) in %dms", query, len(got), time.Since(started).Milliseconds()))
	gopodSearchCache.Lock()
	gopodSearchCache.got[key], gopodSearchCache.at[key] = got, time.Now()
	gopodSearchCache.Unlock()
	return got, nil
}

// gopodSearchFacts: the facts a "search ..." line gets, or "".
func gopodSearchFacts(message string) string {
	q := gopodSearchQuery(message)
	if q == "" {
		return ""
	}
	got, err := gopodSearch(q)
	if err != nil || len(got) == 0 {
		return strings.ReplaceAll(gopodBrobotsSettings.SystemSearch.NoneNote, "{search}", q)
	}
	var b strings.Builder
	b.WriteString(strings.ReplaceAll(gopodBrobotsSettings.SystemSearch.Note, "{search}", q))
	for i, r := range got {
		b.WriteString(fmt.Sprintf("\n%d. %s: %s", i+1, r.Title, r.Content))
	}
	gopodSearchShow(q, got)
	return b.String()
}

// gopodSearchShown: when each search's results bubble last showed.
var gopodSearchShown = struct {
	sync.Mutex
	at map[string]time.Time
}{at: map[string]time.Time{}}

// gopodSearchShow (2026-10-17, operator: "Don't speak or highlight search
// results * put a centered rounded-rectangle bubble around the search results
// * search results font @ 50% size * do a fast character by character reveal *
// then brobot responses"): the results go on the feed in their own bubble
// (system_search.results_role), once for a line to both Brobots.
func gopodSearchShow(q string, got []gopodSearchResult) {
	role := gopodBrobotsSettings.SystemSearch.ResultsRole
	if role == "" {
		return
	}
	key := strings.ToLower(q)
	gopodSearchShown.Lock()
	if time.Since(gopodSearchShown.at[key]) < 2*time.Minute {
		gopodSearchShown.Unlock()
		return
	}
	gopodSearchShown.at[key] = time.Now()
	gopodSearchShown.Unlock()
	var lines []string
	for _, r := range got {
		lines = append(lines, strings.TrimSpace(r.Title+": "+r.Content))
	}
	// 2026-10-21 (operator: "'search ...' results * PLAYHEAD's bubble in brobot
	// bubble * since not a direct call to the gomad"): to the Brobot answering.
	// 2026-10-26: the results are Cache PYC's, the intern - PLAYHEAD
	// (system_search.handoff_host) hands them over first.
	to := gopodManagerTo()
	gopodHandoff(gopodBrobotsSettings.SystemSearch.HandoffHost, gopodBrobotsSettings.SystemSearch.ResultsHost, to)
	GopodChatFeedAddTo(role, to, strings.Join(lines, "\n"))
}

// gopodSearchLine: the reply's opening line - the first result's first sentence.
func gopodSearchLine(message string) string {
	q := gopodSearchQuery(message)
	if q == "" {
		return ""
	}
	got, err := gopodSearch(q)
	if err != nil || len(got) == 0 {
		return ""
	}
	sentence := got[0].Content
	if i := strings.IndexAny(sentence, ".!?"); i > 20 {
		sentence = sentence[:i+1]
	}
	line := gopodBrobotsSettings.SystemSearch.Line
	if line == "" {
		line = "{sentence}"
	}
	return strings.NewReplacer("{search}", q, "{title}", got[0].Title, "{sentence}", sentence).Replace(line)
}

// gopodSystemSearchList: typed "system-search <words>" - the results, for the terminal bubble.
func gopodSystemSearchList(text string) string {
	q := strings.TrimSpace(strings.Join(strings.Fields(text)[1:], " "))
	if q == "" {
		return gopodSystemText("system_search.system_search_words_gopod_s")
	}
	got, err := gopodSearch(q)
	if err != nil {
		return gopodSystemText("system_search.system_search_no_answer_from", "q", q, "error", err.Error())
	}
	if len(got) == 0 {
		return gopodSystemText("system_search.system_search_nothing_found", "q", q)
	}
	var b strings.Builder
	b.WriteString("system-search \"" + q + "\":")
	for i, r := range got {
		content := r.Content
		if len(content) > 180 {
			content = gopodCut(content, 180)
		}
		b.WriteString(fmt.Sprintf("\n%d. %s\n   %s\n   %s", i+1, r.Title, content, r.URL))
	}
	return b.String()
}

// ## EOF
