// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_persona_profiles.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// gopodPersonaProfileFile: who's profile ("CACHE PYC" -> cache-pyc.md).
func gopodPersonaProfileFile(who string) string {
	return gopodPersonaProfileFileIn(gopodBrobotsSettings.PersonaProfiles.Folder, gopodAnswerer(who))
}

func gopodPersonaProfileFileIn(folder, who string) string {
	folder = strings.TrimSpace(folder)
	if folder == "" {
		return ""
	}
	if !filepath.IsAbs(folder) {
		folder = filepath.Join(gopodChipperRoot(), folder)
	}
	slug := strings.Trim(gopodPersonaChainSlugRE.ReplaceAllString(strings.ToLower(who), "-"), "-")
	p := filepath.Join(folder, slug+".md")
	if _, err := os.Stat(p); err != nil {
		if f := gopodTopicTreeFind(slug + ".md"); f != "" {
			return f
		}
		if f := gopodPersonaProfileByWords(folder, who); f != "" {
			return f
		}
	}
	return p
}

// gopodPersonaProfileByWords: the profile under folder whose head's words list has who ("CHALK" ->
// the file with "- **words:** chalk, chalk bored, ...").
func gopodPersonaProfileByWords(folder, who string) string {
	want := strings.ToLower(strings.TrimSpace(who))
	found := ""
	_ = filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
		if err != nil || found != "" || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		head, _ := gopodPersonaProfileReadOne(path)
		for _, w := range strings.Split(head["words"], ",") {
			if strings.ToLower(strings.TrimSpace(w)) == want {
				found = path
				break
			}
		}
		return nil
	})
	return found
}

// gopodPersonaProfileRead: a profile is a topic-cluster md, read the format's one way
// (gopodTopicHeadKV): its head - "# <full name>" (the label), "- **nickname:**", "- **job:**", "-
// **words:**", "- **ask:**" - then "---", then its sections ("## Self", ...), each its points ("1.
func gopodPersonaProfileRead(path string) (head map[string]string, sections map[string][]string) {
	return gopodPersonaProfileReadRooted(path, 0)
}

// gopodPersonaProfileReadRooted: a file's own head and sections, with its root's sections first ("-
// **root:** groups/group-gomads.md", found beside it or anywhere under topic_clusters/) - the
// root's own root before that (company.md), at most three deep.
func gopodPersonaProfileReadRooted(path string, depth int) (head map[string]string, sections map[string][]string) {
	head, sections = gopodPersonaProfileReadOne(path)
	root := strings.TrimSpace(head["root"])
	if root == "" || depth >= 3 {
		return head, sections
	}
	rootPath := filepath.Join(filepath.Dir(path), root)
	if _, err := os.Stat(rootPath); err != nil {
		if f := gopodTopicTreeFind(filepath.Base(root)); f != "" {
			rootPath = f
		}
	}
	_, rs := gopodPersonaProfileReadRooted(rootPath, depth+1)
	if rs == nil {
		return head, sections
	}
	for k, pts := range sections {
		rs[k] = append(rs[k], pts...)
	}
	return head, rs
}

// gopodPersonaProfileReadOne: one profile file's head and sections.
func gopodPersonaProfileReadOne(path string) (head map[string]string, sections map[string][]string) {
	head, sections = gopodPersonaProfileReadFile(path)
	if head != nil {
		if pc, ok := gopodPersonaCards[filepath.Base(path)]; ok {
			if head["root"] == "" && pc.Group != "" {
				head["root"] = pc.Group
			}
			if head["body"] == "" && pc.Body != "" {
				head["body"] = pc.Body
			}
		}
	}
	return head, sections
}

// gopodPersonaCards: persona_cards, set before the cards are read (settings load).
var gopodPersonaCards map[string]struct {
	Group string `json:"group"`
	Body  string `json:"body"`
}

func gopodPersonaProfileReadFile(path string) (head map[string]string, sections map[string][]string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	head, sections = map[string]string{}, map[string][]string{}
	inBody, section := false, ""
	for _, line := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(line)
		if !inBody {
			if t == "---" {
				inBody = true
				continue
			}
			if k, v, ok := gopodTopicHeadKV(t); ok {
				switch k {
				case "nickname history":
					k = "why"
				case "people call me":
					k = "words"
					v = strings.ReplaceAll(v, " or ", ", ")
				}
				head[k] = v
			}
			continue
		}
		if strings.HasPrefix(t, "## ") {
			section = strings.TrimSpace(t[3:])
			continue
		}
		point := ""
		if m := gopodProfilePointRE.FindStringSubmatch(t); m != nil {
			point = m[1]
		} else if strings.HasPrefix(t, "- ") {
			point = strings.TrimSpace(t[2:])
		}
		if point != "" && section != "" && !strings.HasSuffix(point, "(to come)") {
			sections[section] = append(sections[section], point)
		}
	}
	return head, sections
}

var gopodProfilePointRE = regexp.MustCompile(`^\d+\.\s+(.+)$`)

// gopodPersonaProfileSections: a profile's sections, each its points.
func gopodPersonaProfileSections(who string) map[string][]string {
	_, sections := gopodPersonaProfileRead(gopodPersonaProfileFile(who))
	return sections
}

// gopodApplyPersonaProfiles: each persona's labels come from its profile's head - its full name
// (the "# " label), nickname, job name - into host_personas, so the profile is the one place a
// persona's name is written.
func gopodApplyPersonaProfiles(s *gopodBrobotsSettingsFile) {
	gopodPersonaCards = s.PersonaCards
	for _, who := range gopodPersonaProfileHosts(s) {
		head, sections := gopodPersonaProfileRead(gopodPersonaProfileFileIn(s.PersonaProfiles.Folder, who))
		if head == nil {
			continue
		}
		hp := s.HostPersonas[who]
		if pts := sections["Who I am"]; len(pts) > 0 {
			if s.Personas == nil {
				s.Personas = map[string]string{}
			}
			s.Personas[who] = strings.Join(pts, " ")
		}
		if pts := sections["Role"]; len(pts) > 0 {
			hp.Role = strings.Join(pts, " ")
		}
		if pts := sections["My way through an answer"]; len(pts) > 0 {
			hp.Steps = pts
		}
		if pts := sections["Signature"]; len(pts) > 0 {
			hp.Signature = pts[0]
		}
		if v := head["label"]; v != "" {
			hp.Name = v
		}
		if v := head["nickname"]; v != "" {
			hp.Nickname = v
		}
		if v := head["job"]; v != "" {
			hp.JobName = v
		}
		if v := strings.TrimSpace(head["body"]); v != "" && !strings.EqualFold(v, "none") {
			hp.Body = v
		}
		if s.HostPersonas == nil {
			s.HostPersonas = map[string]gopodHostPersona{}
		}
		s.HostPersonas[who] = hp
	}
}

// gopodPersonaProfileHosts: every persona a profile can belong to.
func gopodPersonaProfileHosts(s *gopodBrobotsSettingsFile) []string {
	seen := map[string]bool{}
	var out []string
	for k := range s.HostPersonas {
		seen[k] = true
		out = append(out, k)
	}
	for k := range s.Managers.Personas {
		if !seen[k] {
			out = append(out, k)
		}
	}
	return out
}

// gopodPersonaProfileLabelsJSON: the served settings with each persona's labels from its profile
// (the page reads host_personas from here).
func gopodPersonaProfileLabelsJSON(data []byte) []byte {
	var all map[string]json.RawMessage
	if json.Unmarshal(data, &all) != nil {
		return data
	}
	var hps map[string]map[string]any
	if json.Unmarshal(all["host_personas"], &hps) != nil {
		return data
	}
	for who, hp := range gopodBrobotsSettings.HostPersonas {
		if hps[who] == nil {
			hps[who] = map[string]any{}
		}
		hps[who]["name"], hps[who]["nickname"], hps[who]["job_name"] = hp.Name, hp.Nickname, hp.JobName
		if hp.CallName != "" {
			hps[who]["call_name"] = hp.CallName
		}
	}
	raw, err := json.Marshal(hps)
	if err != nil {
		return data
	}
	all["host_personas"] = raw
	out, err := json.Marshal(all)
	if err != nil {
		return data
	}
	return []byte(gopodNameVarsFill(string(out), gopodBrobotsSettings.HostPersonas))
}

// gopodPersonaProfileText: the profile lines for who's prompt, "" with none.
func gopodPersonaProfileText(who string) string {
	return gopodPersonaProfileTextOf(who, gopodBrobotsSettings.PersonaProfiles.Sections)
}

// gopodPersonaProfileTextOf: the profile lines from these sections.
func gopodPersonaProfileTextOf(who string, names []string) string {
	pp := gopodBrobotsSettings.PersonaProfiles
	sections := gopodPersonaProfileSections(who)
	if len(sections) == 0 || strings.TrimSpace(pp.Text) == "" {
		return ""
	}
	var lines []string
	for _, name := range names {
		lines = append(lines, sections[name]...)
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.ReplaceAll(pp.Text, "{lines}", gopodPersonaFill(who, strings.Join(lines, " ")))
}

// gopodPersonaProfileField: a Self line's value by its label ("My job title").
func gopodPersonaProfileField(who, label string) string {
	for _, line := range gopodPersonaProfileSections(who)["Self"] {
		if strings.HasPrefix(strings.ToLower(line), strings.ToLower(label)+":") {
			return strings.TrimSpace(gopodPersonaFill(who, line[len(label)+1:]))
		}
	}
	return ""
}

// gopodGomadBrain: a Gomad's brain bubble, at the top of its own bubble - the same title as a
// Brobot's ("{name} Brain") and the same lines, the one brain template (brain_chain.lines,
// gopodBrainLines).
func gopodGomadBrain(host, to, question string) {
	title := strings.TrimSpace(gopodBrobotsSettings.BubbleNotices.ValuePoints.Title)
	if title == "" {
		return
	}
	title = strings.ReplaceAll(title, "{name}", gopodCallName(host))
	for _, line := range append(gopodBrainBubbleLines(host, question), gopodBrainScoreRows()...) {
		gopodBubbleNotice(to, title, line)
	}
}

// gopodLingoWords: how words are said by the reporter voices - the shared list
// (reporter_voices.say_words) plus each persona's own "## Lingo" points ("CHALK = chock"), his own
// words winning.
func gopodLingoWords() map[string]string {
	out := map[string]string{}
	for k, v := range gopodBrobotsSettings.ReporterVoices.SayWords {
		out[k] = v
	}
	for k, v := range gopodProfileLingo() {
		out[k] = v
	}
	return out
}

// gopodProfileLingo: every persona's own "## Lingo" points ("CHALK = chock") - the one source for
// how a persona's names are said, by every voice (the reporter voices here, the Brobots' robots in
// gopodBrobotPronunciations).
func gopodProfileLingo() map[string]string {
	out := map[string]string{}
	for _, c := range gopodBrobotsSettings.HereNow.Cast {
		for _, point := range gopodPersonaProfileSections(c.Who)["Lingo"] {
			if word, said, ok := strings.Cut(point, " = "); ok && strings.TrimSpace(word) != "" {
				out[strings.TrimSpace(word)] = strings.TrimSuffix(strings.TrimSpace(said), ".")
			}
		}
	}
	return out
}

// gopodBrainRelationships: who's relationships for his brain - his card's "## Who I answer to" and
// "## Who works with me".
func gopodBrainRelationships(who string) string {
	sections := gopodPersonaProfileSections(who)
	var points []string
	points = append(points, sections["Who I answer to"]...)
	points = append(points, sections["Who works with me"]...)
	var out []string
	for _, p := range points {
		if p = strings.TrimSuffix(strings.TrimSpace(gopodPersonaFill(who, p)), "."); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, ". ") + "."
}

// ## EOF
