// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_prober_invite_test.go
package wirepod_ttr

import (
	"os"
	"strings"
	"testing"
)

// TestReminderInvitesVary: a reminder's "Try asking" invite is shuffled, never the same twice
// running, lands on its topic, and comes in each Brobot's own shapes too.
func TestReminderInvitesVary(t *testing.T) {
	if _, err := os.Stat(gopodChaptersDir()); err != nil {
		t.Skip("no chapters here")
	}
	gopodCoverage.Lock()
	kept := gopodCoverageSaid
	gopodCoverageSaid = map[string]map[int]bool{} // a fresh game: every point still to say
	gopodCoverage.Unlock()
	defer func() {
		gopodCoverage.Lock()
		gopodCoverageSaid = kept
		gopodCoverage.Unlock()
	}()
	chapters, _, _ := gopodReadChaptersFolder()
	topics, own := 0, map[string]int{}
	for _, c := range chapters {
		for _, tp := range c.Topics {
			for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
				who := gopodBrobotName(esn)
				prev := ""
				for i := 0; i < 3; i++ {
					q := gopodProberFor(tp.Name, esn)
					if q == "" {
						break
					}
					topics++
					if l := loadScenarioPacketsForMessage(q); len(l) == 0 || l[0].Name != tp.Name {
						t.Errorf("%s invite %q doesn't land on %s", who, q, tp.Name)
					}
					if q == prev && len(gopodProbesLive(tp.Name, esn)) > 1 {
						t.Errorf("%s: %q twice running", tp.Name, q)
					}
					prev = q
					for _, shape := range gopodBrobotsSettings.ProbersMake.PersonaShapes[who] {
						if lead := strings.SplitN(shape, "{", 2)[0]; lead != "" && strings.HasPrefix(q, lead) {
							own[who]++
						}
					}
				}
			}
		}
	}
	if topics == 0 {
		t.Fatal("no invites made")
	}
	for who := range gopodBrobotsSettings.ProbersMake.PersonaShapes {
		if own[who] == 0 {
			t.Errorf("no invite in %s's own shapes", who)
		}
	}
	t.Logf("invites %d, own shapes %v", topics, own)
}

// ## EOF
