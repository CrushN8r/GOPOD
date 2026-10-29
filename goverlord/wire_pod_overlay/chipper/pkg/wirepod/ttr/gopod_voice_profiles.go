// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_voice_profiles.go
package wirepod_ttr

// GOLDEN, one place (2026-10-26, operator: "expand that tag into a
// per-audio-file-profile-tag * per-audio-file-profile-tag carries all
// per-audio-file tagged-values * also carries these usage-details: generation
// date, # of times used, time last used * these usage-details will help
// determine what to keep over the course of a week/month"; "In GOPOD sysytem,
// add a report-run * report-run past 3-days/week/month/year * this will help me
// determine what is \"popular\" for GOPOD"). Every voice file beside the voices
// (gopod_reporter_audio) has one profile in voice_profiles.json: its kind
// (system, scripted, live, mic), voice, who it's for, its words, the leveler's
// values, when it was made, how many times it played (per day too), when it
// last played, and how many times it was made again (a newer mic recording of
// the same words takes it over). system-voices-report reads them - report only,
// nothing moves. voice_kinds.json (kinds only) was the start of this.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type gopodVoiceProfile struct {
	Kind          string         `json:"kind"`
	Voice         string         `json:"voice,omitempty"`
	Host          string         `json:"host,omitempty"`
	Text          string         `json:"text,omitempty"`
	LoudnessLUFS  float64        `json:"loudness_lufs,omitempty"`
	CeilingDB     float64        `json:"ceiling_db,omitempty"`
	Made          string         `json:"made"`
	Uses          int            `json:"uses"`
	LastUsed      string         `json:"last_used,omitempty"`
	UsedDays      map[string]int `json:"used_days,omitempty"`
	Regenerations int            `json:"regenerations,omitempty"`
}

var gopodVoiceProfiles = struct {
	sync.Mutex
	loaded   bool
	since    string
	files    map[string]*gopodVoiceProfile
	saveWait bool
}{}

func gopodVoiceProfilesPath() string {
	return filepath.Join(gopodReporterAudioDir(), "voice_profiles.json")
}

// gopodVoiceProfilesLoadLocked: read once; the first time, every voice file on
// disk gets a profile (its kind from voice_kinds.json, made = its file time,
// its words from its .words.json), and counting starts now.
func gopodVoiceProfilesLoadLocked() {
	p := &gopodVoiceProfiles
	if p.loaded {
		return
	}
	p.loaded = true
	p.files = map[string]*gopodVoiceProfile{}
	var disk struct {
		CountingSince string                        `json:"counting_since"`
		Files         map[string]*gopodVoiceProfile `json:"files"`
	}
	if data, err := os.ReadFile(gopodVoiceProfilesPath()); err == nil && json.Unmarshal(data, &disk) == nil && disk.Files != nil {
		p.since, p.files = disk.CountingSince, disk.Files
		return
	}
	p.since = time.Now().Format(time.RFC3339)
	kinds := map[string]string{}
	if data, err := os.ReadFile(gopodVoiceKindsPath()); err == nil {
		json.Unmarshal(data, &kinds)
	}
	entries, _ := os.ReadDir(gopodReporterAudioDir())
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".wav") || !gopodVoiceAnyFileRE.MatchString(name) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		prof := &gopodVoiceProfile{Kind: kinds[name], Voice: gopodVoiceOfFile(name), Made: info.ModTime().Format(time.RFC3339), Text: gopodWordsText(filepath.Join(gopodReporterAudioDir(), name))}
		if prof.Kind == "" {
			prof.Kind = "scripted"
		}
		if strings.HasPrefix(name, "op_mic_") {
			prof.Kind = "mic"
		}
		p.files[name] = prof
	}
	gopodVoiceProfilesSaveLocked()
}

// gopodVoiceProfilesSaveLocked: written now (safe: a temp file, then renamed).
func gopodVoiceProfilesSaveLocked() {
	p := &gopodVoiceProfiles
	data, err := json.MarshalIndent(map[string]any{"counting_since": p.since, "files": p.files}, "", " ")
	if err != nil {
		return
	}
	tmp := gopodVoiceProfilesPath() + ".tmp"
	if gopodWriteFile(tmp, data, 0o644) == nil {
		os.Rename(tmp, gopodVoiceProfilesPath())
	}
}

// gopodVoiceProfilesSaveSoon: one write for a burst of plays.
func gopodVoiceProfilesSaveSoon() {
	p := &gopodVoiceProfiles
	if p.saveWait {
		return
	}
	p.saveWait = true
	time.AfterFunc(5*time.Second, func() {
		p.Lock()
		p.saveWait = false
		gopodVoiceProfilesSaveLocked()
		p.Unlock()
	})
}

// gopodVoiceOfFile: the Kokoro voice a file is named for ("" for a mic recording).
func gopodVoiceOfFile(name string) string {
	f := strings.SplitN(name, "_", 3)
	if len(f) < 3 || name == "" || strings.HasPrefix(name, "op_mic_") {
		return ""
	}
	return f[0] + "_" + f[1]
}

// gopodWordsText: a voice's words, from its .words.json.
func gopodWordsText(wav string) string {
	data, err := os.ReadFile(strings.TrimSuffix(wav, ".wav") + ".words.json")
	if err != nil {
		return ""
	}
	type word struct {
		T string `json:"t"`
	}
	var words []word
	if json.Unmarshal(data, &words) != nil {
		var wrapped struct {
			Words []word `json:"words"`
		}
		if json.Unmarshal(data, &wrapped) != nil {
			return ""
		}
		words = wrapped.Words
	}
	var out []string
	for _, w := range words {
		out = append(out, w.T)
	}
	return strings.TrimSpace(strings.Join(out, " "))
}

// gopodVoiceProfileMade: a voice file just made (or made again) - its profile.
func gopodVoiceProfileMade(name, kind, host, voice, text string) {
	ca := gopodBrobotsSettings.ChatAudio
	p := &gopodVoiceProfiles
	p.Lock()
	defer p.Unlock()
	gopodVoiceProfilesLoadLocked()
	prof := p.files[name]
	if prof == nil {
		prof = &gopodVoiceProfile{}
		p.files[name] = prof
	} else {
		prof.Regenerations++
	}
	prof.Kind, prof.Host, prof.Voice, prof.Text = kind, host, voice, strings.TrimSpace(text)
	prof.LoudnessLUFS, prof.CeilingDB = ca.LoudnessLUFS, ca.NormalizeDB
	prof.Made = time.Now().Format(time.RFC3339)
	gopodVoiceProfilesSaveLocked()
}

// gopodVoiceProfileTakeOver (operator: "popular repeating live questions,
// 'regenerate', so archive generated live voice & new raw voice audio file goes
// through the same leveler, then takes over the per-audio-file-profile-tag, now
// with a # of regenerations"): a new mic recording of the same words takes over
// the older ones' profile (its uses carried, made again +1) and they go to the
// archive.
func gopodVoiceProfileTakeOver(name, text string) {
	key := strings.ToLower(strings.Join(strings.Fields(text), " "))
	p := &gopodVoiceProfiles
	p.Lock()
	defer p.Unlock()
	gopodVoiceProfilesLoadLocked()
	prof := p.files[name]
	if prof == nil || key == "" {
		return
	}
	archive := filepath.Join(gopodArchiveDir(), "voices")
	for old, op := range p.files {
		if old == name || op.Kind != "mic" || strings.ToLower(strings.Join(strings.Fields(op.Text), " ")) != key {
			continue
		}
		prof.Uses += op.Uses
		prof.Regenerations += op.Regenerations + 1
		if op.LastUsed > prof.LastUsed {
			prof.LastUsed = op.LastUsed
		}
		for d, n := range op.UsedDays {
			if prof.UsedDays == nil {
				prof.UsedDays = map[string]int{}
			}
			prof.UsedDays[d] += n
		}
		if gopodMkdirAll(archive, 0o755) == nil {
			dir := gopodReporterAudioDir()
			os.Rename(filepath.Join(dir, old), filepath.Join(archive, old))
			words := strings.TrimSuffix(old, ".wav") + ".words.json"
			os.Rename(filepath.Join(dir, words), filepath.Join(archive, words))
		}
		delete(p.files, old)
	}
	gopodVoiceProfilesSaveLocked()
}

// gopodVoiceUsed: a voice file played (its line went to the page).
func gopodVoiceUsed(name string) {
	p := &gopodVoiceProfiles
	p.Lock()
	defer p.Unlock()
	gopodVoiceProfilesLoadLocked()
	prof := p.files[name]
	if prof == nil {
		return
	}
	now := time.Now()
	prof.Uses++
	prof.LastUsed = now.Format(time.RFC3339)
	if prof.UsedDays == nil {
		prof.UsedDays = map[string]int{}
	}
	prof.UsedDays[now.Format("2006-01-02")]++
	gopodVoiceProfilesSaveSoon()
}

// gopodVoiceProfileUses: each file's plays, for the archive's popular rule.
func gopodVoiceProfileUses() map[string]int {
	p := &gopodVoiceProfiles
	p.Lock()
	defer p.Unlock()
	gopodVoiceProfilesLoadLocked()
	out := map[string]int{}
	for name, prof := range p.files {
		out[name] = prof.Uses
	}
	return out
}

// gopodVoiceProfileKinds: each file's kind, for the archive's keep days.
func gopodVoiceProfileKinds() map[string]string {
	p := &gopodVoiceProfiles
	p.Lock()
	defer p.Unlock()
	gopodVoiceProfilesLoadLocked()
	out := map[string]string{}
	for name, prof := range p.files {
		kind := prof.Kind
		if kind == "mic" {
			kind = "live"
		}
		out[name] = kind
	}
	return out
}

// gopodVoiceArchived: a file the archive moved has no profile here any more.
func gopodVoiceArchived(name string) {
	p := &gopodVoiceProfiles
	p.Lock()
	defer p.Unlock()
	gopodVoiceProfilesLoadLocked()
	if _, ok := p.files[name]; ok {
		delete(p.files, name)
		gopodVoiceProfilesSaveSoon()
	}
}

// gopodVoiceReportWindows: system-voices-report's windows.
var gopodVoiceReportWindows = map[string]int{"3d": 3, "3days": 3, "3-days": 3, "week": 7, "month": 30, "year": 365}

// gopodSystemVoicesReport: "system-voices-report [3d | week | month | year]" -
// report only, nothing moves.
func gopodSystemVoicesReport(text string) string {
	f := strings.Fields(strings.ToLower(text))
	window, label := 7, "week"
	if len(f) > 1 {
		days, ok := gopodVoiceReportWindows[f[1]]
		if !ok {
			return gopodSystemText("voice_profiles.system_voices_report_3d_week")
		}
		window, label = days, f[1]
	}
	from := time.Now().AddDate(0, 0, -window)
	fromDay := from.Format("2006-01-02")
	p := &gopodVoiceProfiles
	p.Lock()
	gopodVoiceProfilesLoadLocked()
	since := p.since
	type line struct {
		name string
		uses int
		prof gopodVoiceProfile
	}
	var all []line
	for name, prof := range p.files {
		n := 0
		for d, c := range prof.UsedDays {
			if d >= fromDay {
				n += c
			}
		}
		all = append(all, line{name, n, *prof})
	}
	p.Unlock()

	// Voices someone speaks with now.
	rv := gopodBrobotsSettings.ReporterVoices
	assigned := map[string]bool{rv.SystemVoice: true, gopodReporterVoice(gopodOperatorVoiceHost): true}
	for _, v := range rv.Hosts {
		assigned[v] = true
	}
	size := func(name string) int64 {
		if info, err := os.Stat(filepath.Join(gopodReporterAudioDir(), name)); err == nil {
			return info.Size()
		}
		return 0
	}
	mb := func(b int64) string { return fmt.Sprintf("%.1f MB", float64(b)/1e6) }
	kinds := map[string][2]int{}
	buckets := [4]int{}
	var made, used, idle, orphans int
	var total, idleSize, orphanSize int64
	for _, l := range all {
		s := size(l.name)
		total += s
		k := kinds[l.prof.Kind]
		k[0]++
		k[1] += l.uses
		kinds[l.prof.Kind] = k
		madeAt, _ := time.Parse(time.RFC3339, l.prof.Made)
		if madeAt.After(from) {
			made++
		}
		switch {
		case l.uses == 0:
		case l.uses == 1:
			buckets[0]++
		case l.uses <= 3:
			buckets[1]++
		case l.uses <= 9:
			buckets[2]++
		default:
			buckets[3]++
		}
		if l.uses > 0 {
			used++
		} else if madeAt.Before(from) {
			idle++
			idleSize += s
		}
		if l.prof.Voice != "" && !assigned[l.prof.Voice] {
			orphans++
			orphanSize += s
		}
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Voice files, past %s (report only - nothing moves; plays counted since %s):", label, strings.SplitN(since, "T", 2)[0]))
	b.WriteString(fmt.Sprintf("\n  on disk: %d files, %s; made in the %s: %d; played in the %s: %d", len(all), mb(total), label, made, label, used))
	var kn []string
	for k := range kinds {
		kn = append(kn, k)
	}
	sort.Strings(kn)
	for _, k := range kn {
		b.WriteString(fmt.Sprintf("\n  %s: %d files, %d plays", k, kinds[k][0], kinds[k][1]))
	}
	b.WriteString(fmt.Sprintf("\n  played 1x: %d, 2-3x: %d, 4-9x: %d, 10x+: %d", buckets[0], buckets[1], buckets[2], buckets[3]))
	for _, kind := range []string{"live", "mic"} {
		var top []line
		for _, l := range all {
			if l.prof.Kind == kind && l.uses > 0 {
				top = append(top, l)
			}
		}
		sort.Slice(top, func(i, j int) bool { return top[i].uses > top[j].uses })
		if len(top) == 0 {
			continue
		}
		b.WriteString("\n  most played (" + kind + "):")
		for i, l := range top {
			if i == 10 {
				break
			}
			t := l.prof.Text
			if r := []rune(t); len(r) > 60 {
				t = string(r[:60]) + "..."
			}
			b.WriteString(fmt.Sprintf("\n    %dx  %s", l.uses, t))
		}
	}
	b.WriteString(fmt.Sprintf("\n  not played in the %s and older: %d files, %s", label, idle, mb(idleSize)))
	b.WriteString(fmt.Sprintf("\n  in a voice no one speaks with now: %d files, %s", orphans, mb(orphanSize)))
	return b.String()
}

// ## EOF
