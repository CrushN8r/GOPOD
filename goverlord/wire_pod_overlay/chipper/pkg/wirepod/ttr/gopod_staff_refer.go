// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_staff_refer.go
package wirepod_ttr

import (
	"regexp"
	"strings"
	"sync/atomic"
)

// Staff refer the room's questions to the GOPOD reps (the company's rule 10): a staff answer that
// ends on an open question to the room ends instead on managers.staff_refer.line. One question in
// keep_once ("What do you think?") stays, once a session.

var gopodStaffReferKept atomic.Bool // keep_once used this session (gopodGameReset clears it)

var gopodStaffSentenceRE = regexp.MustCompile(`[^.!?]*[.!?]+[^\pL\pN]*`)

// gopodStaffRefer: the answer, its trailing questions to the room swapped for the refer line.
func gopodStaffRefer(reply string) string {
	sr := gopodBrobotsSettings.Managers.StaffRefer
	line := strings.TrimSpace(sr.Line)
	if line == "" {
		return reply
	}
	parts := gopodStaffSentenceRE.FindAllString(reply, -1)
	if len(parts) == 0 || strings.TrimSpace(strings.Join(parts, "")) != strings.TrimSpace(reply) {
		parts = []string{reply}
	}
	end := len(parts)
	for end > 0 && gopodStaffOpenAsk(parts[end-1]) {
		end--
	}
	if end == len(parts) {
		return reply
	}
	kept := ""
	for _, q := range parts[end:] {
		for _, k := range sr.KeepOnce {
			if k = strings.TrimSpace(k); k != "" && strings.Contains(strings.ToLower(q), strings.ToLower(k)) && kept == "" && !gopodStaffReferKept.Load() {
				kept = strings.TrimSpace(q)
			}
		}
	}
	out := strings.TrimSpace(strings.Join(parts[:end], ""))
	if kept != "" {
		gopodStaffReferKept.Store(true)
		return strings.TrimSpace(out + " " + kept)
	}
	return strings.TrimSpace(out + " " + line)
}

// gopodStaffOpenAsk: a sentence asking the room - a question, or one of open_asks ("let me know").
func gopodStaffOpenAsk(sentence string) bool {
	if strings.Contains(sentence, "?") {
		return true
	}
	lower := strings.ToLower(sentence)
	for _, a := range gopodBrobotsSettings.Managers.StaffRefer.OpenAsks {
		if a = strings.ToLower(strings.TrimSpace(a)); a != "" && strings.Contains(lower, a) {
			return true
		}
	}
	return false
}

// ## EOF
