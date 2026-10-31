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
	"os"
	"path/filepath"
	"strings"
)

// gopodPersonaProfileFile: who's profile ("CACHE PYC" -> cache-pyc.md).
func gopodPersonaProfileFile(who string) string {
	folder := strings.TrimSpace(gopodBrobotsSettings.PersonaProfiles.Folder)
	if folder == "" {
		return ""
	}
	if !filepath.IsAbs(folder) {
		folder = filepath.Join(gopodChipperRoot(), folder)
	}
	slug := strings.Trim(gopodPersonaChainSlugRE.ReplaceAllString(strings.ToLower(who), "-"), "-")
	return filepath.Join(folder, slug+".md")
}

// gopodPersonaProfileSections: a profile's sections, each its bullet lines.
func gopodPersonaProfileSections(who string) map[string][]string {
	path := gopodPersonaProfileFile(who)
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	out := map[string][]string{}
	section := ""
	for _, line := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "## "):
			section = strings.TrimSpace(t[3:])
		case strings.HasPrefix(t, "- ") && section != "" && !strings.HasSuffix(t, "(to come)"):
			// A line still to come is the operator's to write - not said.
			out[section] = append(out[section], strings.TrimSpace(t[2:]))
		}
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

// ## EOF
