// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_persona_profiles.go
package wirepod_ttr

// Persona profiles (2026-10-30, operator: "fatten up all these personas for self
// & chat awareness * offload to persona profile md files for look ups"; "These
// will be your first workings of a brain bubble template for scaffolding for the
// other personas"). One md per persona (persona_profiles.folder/<name>.md): the
// same sections for everyone - Self (my informal name, my name, my job title, my
// role, who helps me), Voice, Partner and banter, The name. The sections named in
// persona_profiles.sections are looked up for the persona's prompt, one line per
// bullet, through persona_profiles.text.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// gopodPersonaProfileFile: who's profile ("CACHE PYC" -> cache-pyc.md).
func gopodPersonaProfileFile(who string) string {
	return gopodPersonaProfileFileIn(gopodBrobotsSettings.PersonaProfiles.Folder, who)
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
	return filepath.Join(folder, slug+".md")
}

// gopodPersonaProfileRead (2026-10-30, operator: "make the topic-cluster shape the
// one master format * YES!! PLEASE!!"): a profile is a topic-cluster md, read the
// format's one way (gopodTopicHeadKV): its head - "# <full name>" (the label),
// "- **nickname:**", "- **job:**", "- **words:**", "- **ask:**" - then "---", then
// its sections ("## Self", ...), each its points ("1. ...").
func gopodPersonaProfileRead(path string) (head map[string]string, sections map[string][]string) {
	head, sections = gopodPersonaProfileReadOne(path)
	// 2026-11-04 (operator: "Pull all Doc & Pip from brobots md files"; "Doc is
	// not Brobot 1. Doc is the persona hosted by Brobot 1"): a robot's file names
	// the persona it hosts ("- **hosts:** doc.md"); the persona's head lines win,
	// and its sections follow the robot's.
	if host := strings.TrimSpace(head["hosts"]); host != "" {
		ph, ps := gopodPersonaProfileReadOne(filepath.Join(filepath.Dir(path), host))
		for k, v := range ph {
			if k != "hosts" {
				head[k] = v
			}
		}
		for k, pts := range ps {
			sections[k] = append(sections[k], pts...)
		}
	}
	return head, sections
}

// gopodPersonaProfileReadOne: one profile file's head and sections.
func gopodPersonaProfileReadOne(path string) (head map[string]string, sections map[string][]string) {
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
		// A point still to come is the operator's to write - not said.
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

// gopodApplyPersonaProfiles (2026-10-30, operator: "All should branch from single
// master sources"): each persona's labels come from its profile's head - its full
// name (the "# " label), nickname, job name - into host_personas, so the profile
// is the one place a persona's name is written.
func gopodApplyPersonaProfiles(s *gopodBrobotsSettingsFile) {
	for _, who := range gopodPersonaProfileHosts(s) {
		head, sections := gopodPersonaProfileRead(gopodPersonaProfileFileIn(s.PersonaProfiles.Folder, who))
		if head == nil {
			continue
		}
		hp := s.HostPersonas[who]
		// 2026-10-30 (operator: "Commit & go on giving each a profile in the master
		// format * pull them out of the shared prompt * so all come from one
		// template"): a Brobot's own words - who I am, my role, my way through an
		// answer, my signature - from his profile too.
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

// gopodPersonaProfileLabelsJSON: the served settings with each persona's labels
// from its profile (the page reads host_personas from here).
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
	return out
}

// gopodPersonaProfileText: the profile lines for who's prompt, "" with none.
func gopodPersonaProfileText(who string) string {
	pp := gopodBrobotsSettings.PersonaProfiles
	sections := gopodPersonaProfileSections(who)
	if len(sections) == 0 || strings.TrimSpace(pp.Text) == "" {
		return ""
	}
	var lines []string
	for _, name := range pp.Sections {
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

// gopodGomadBrain (2026-10-30, operator: "NEW CORE TEMPLATE RULE: everybody gets a
// brain bubble"): a Gomad's brain bubble, at the top of its own bubble - the same
// title as a Brobot's ("{name} Brain") and the same lines, the one brain template
// (brain_chain.lines, gopodBrainLines).
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

// The two lanes of a persona's profile (2026-11-02, operator: "relationship &
// lingo are 2 separate lanes * lingo is how words are pronounced * personal/
// working relationship connections, past & present, hence persona history").

// gopodLingoWords: how words are said by the reporter voices - the shared list
// (reporter_voices.say_words) plus each persona's own "## Lingo" points
// ("CHALK = chock"), his own words winning. The voice only: never in a prompt.
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

// gopodProfileLingo: every persona's own "## Lingo" points ("CHALK = chock") -
// the one source for how a persona's names are said, by every voice (the
// reporter voices here, the Brobots' robots in gopodBrobotPronunciations).
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

// gopodBrainRelationships: who's relationships for his brain - his profile's
// "## Relationships" points; until those are written, what's there already:
// his "Who helps me" and his "Partner and banter".
func gopodBrainRelationships(who string) string {
	sections := gopodPersonaProfileSections(who)
	points := sections["Relationships"]
	if len(points) == 0 {
		if v := gopodPersonaProfileField(who, "Who helps me"); v != "" {
			points = append(points, v)
		}
		points = append(points, sections["Partner and banter"]...)
	}
	var out []string
	for _, p := range points {
		if p = strings.TrimSuffix(strings.TrimSpace(gopodPersonaFill(who, p)), "."); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "; ") + "."
}

// ## EOF
