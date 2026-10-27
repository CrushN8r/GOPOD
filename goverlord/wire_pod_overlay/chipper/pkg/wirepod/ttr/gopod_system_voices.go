// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_voices.go
package wirepod_ttr

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// system-voices-rebuild (2026-10-18, operator: "requires the wpr 2 & F5 *
// temporarily back up all current used generated voice audio files * system
// message, ready for a restart * upon restart, GOPOD automatically does a voice
// files generation check * rebuild/generate all new voice audio files"; first:
// "the files match check for story.md, zKnobs.json, knob.json ... do a files
// diff check for updates & other edits * anything questionable? pause & ask").
//
// 1. system-check (story.md, knobs.json, zKnobs.json and every file named); a
//    problem pauses it - "system-voices-rebuild go" rebuilds anyway.
// 2. The scripted lines that need a voice (new, edited, or not timed by Vosk).
// 3. Every generated voice (and its word times) moves to
//    webroot/gopod_reporter_audio_backup/<time>/; sounds stay.
// 4. At the next start the generation check makes every scripted line again;
//    each voice's backup copy goes once its new voice is made (gopodVoiceMade).

var gopodVoiceFileRE = regexp.MustCompile(`^[a-z]{2}_[a-z]+_(s\d+_)?[0-9a-f]{16}\.(wav|words\.json)$`)

// gopodVoiceAnyFileRE: every voice file the profiles and the archive look after
// - a made voice (pitched too, voice_tweaks) or a mic recording (op_mic_).
var gopodVoiceAnyFileRE = regexp.MustCompile(`^([a-z]{2}_[a-z]+_(s\d+_)?(p[+-]\d+_)?|op_mic_)[0-9a-f]{16}\.(wav|words\.json)$`)

func gopodVoicesBackupDir() string {
	return filepath.Join(gopodChipperRoot(), "webroot", "gopod_reporter_audio_backup")
}

// gopodVoicesNeeded: the scripted lines whose voice isn't made (or not timed by Vosk).
func gopodVoicesNeeded() []string {
	var out []string
	check := func(host, text string) {
		says := gopodReporterSays(text)
		if sr := gopodBrobotsSettings.SignOff.SystemRole; sr != "" && host == sr {
			says = gopodSaySwap(says, gopodBrobotsSettings.ReporterVoices.SystemSayWords)
		}
		if _, _, ok := gopodReporterCached(host, says); !ok {
			out = append(out, host+": "+gopodHeardShort(text))
		}
	}
	for _, l := range gopodReporterAllLines() {
		check(l.Host, l.Text)
	}
	for _, text := range gopodSystemLines() {
		if strings.TrimSpace(text) != "" {
			check(gopodBrobotsSettings.SignOff.SystemRole, text)
		}
	}
	return out
}

// gopodSystemVoicesRebuild: the answer to "system-voices-rebuild [go]".
func gopodSystemVoicesRebuild(text string) string {
	force := len(strings.Fields(text)) > 1 && strings.EqualFold(strings.Fields(text)[1], "go")
	var out []string
	problems, checked := gopodSystemCheckProblems()
	needed := gopodVoicesNeeded()
	if len(needed) > 0 {
		out = append(out, fmt.Sprintf("%d scripted lines need a voice (new, edited, or re-timing):", len(needed)))
		for i, n := range needed {
			if i == 8 {
				out = append(out, fmt.Sprintf("  ... %d more", len(needed)-8))
				break
			}
			out = append(out, "  "+n)
		}
	} else {
		out = append(out, "Every scripted line has its voice.")
	}
	if len(problems) > 0 && !force {
		out = append(out, fmt.Sprintf("Paused - system-check found %d to look at (%d checked):", len(problems), checked))
		for i, p := range problems {
			if i == 8 {
				out = append(out, fmt.Sprintf("  ... %d more (system-check)", len(problems)-8))
				break
			}
			out = append(out, "  "+p)
		}
		out = append(out, "Fix them, or type \"system-voices-rebuild go\" to rebuild anyway.")
		return gopodJoinLines(out)
	}
	dir := gopodReporterAudioDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "system-voices-rebuild: the voice folder can't be read - " + err.Error()
	}
	backup := filepath.Join(gopodVoicesBackupDir(), time.Now().Format("20060102_150405"))
	if err := gopodMkdirAll(backup, 0o755); err != nil {
		return "system-voices-rebuild: no backup folder - " + err.Error()
	}
	moved := 0
	for _, e := range entries {
		if e.IsDir() || !gopodVoiceFileRE.MatchString(e.Name()) {
			continue
		}
		if os.Rename(filepath.Join(dir, e.Name()), filepath.Join(backup, e.Name())) == nil && strings.HasSuffix(e.Name(), ".wav") {
			moved++
		}
	}
	logger.Println(fmt.Sprintf("GOPOD_VOICES_REBUILD backed up %d voices to %s", moved, backup))
	out = append(out, fmt.Sprintf("Backed up %d voices. Ready for a restart: wpr 2, then F5 and click Log - every voice is made again at the start.", moved))
	return gopodJoinLines(out)
}

// gopodVoicesAtStart: the start's generation check is making the scripted lines.
var gopodVoicesAtStart atomic.Bool

// Voice kinds and keeping (2026-10-18, operator: "no auto-deletes * delete
// backup upon successful new playable voice wav file * operator_keep_days for
// system lines * scripted_keep_days for scripted lines (reporter/narrator/etc) *
// live_keep_days for live lines * I think we already have an archiving process,
// use that to start from"): each voice is noted as it's made - system
// (Goverlord's voice), scripted (made by the start's check or a playing script)
// or live (the Operator's lines, the managers' answers) - in voice_kinds.json
// beside the voices. At each start, a voice not in use and not said for its
// kind's keep days moves to the sessions' archives (gopodArchiveDir()/voices),
// the same archive the session chats go to. Nothing is deleted.
// gopodVoiceKindsPath: the kinds-only list the profiles started from (read once,
// when voice_profiles.json is first made).
func gopodVoiceKindsPath() string {
	return filepath.Join(gopodReporterAudioDir(), "voice_kinds.json")
}

// gopodVoiceMade: a new playable voice - its profile (gopod_voice_profiles.go,
// its kind among the rest), its backup copy gone.
func gopodVoiceMade(host, file, text string) {
	name := filepath.Base(file)
	kind := "live"
	switch {
	case host == gopodBrobotsSettings.SignOff.SystemRole && host != "":
		kind = "system"
	case host == gopodOperatorVoiceHost:
		kind = "live"
	case gopodVoicesAtStart.Load() || gopodScriptPlaying.Load():
		kind = "scripted"
	}
	gopodVoiceProfileMade(name, kind, host, gopodReporterVoice(host), text)
	// The rebuild's backup of this voice (system-voices-rebuild): the new one plays.
	backups, _ := filepath.Glob(filepath.Join(gopodVoicesBackupDir(), "*", name))
	for _, b := range backups {
		os.Remove(b)
		os.Remove(strings.TrimSuffix(b, ".wav") + ".words.json")
		if left, _ := os.ReadDir(filepath.Dir(b)); len(left) == 0 {
			os.Remove(filepath.Dir(b))
		}
	}
}

// gopodVoicesArchive: at the start, every voice not in use and not said for its
// kind's keep days moves to the archive with its word times.
func gopodVoicesArchive() {
	rv := gopodBrobotsSettings.ReporterVoices
	dir := gopodReporterAudioDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	gopodReporterInUse.Lock()
	inUse := map[string]bool{}
	for k := range gopodReporterInUse.files {
		inUse[k] = true
	}
	gopodReporterInUse.Unlock()
	kinds := gopodVoiceProfileKinds()
	uses := gopodVoiceProfileUses()
	keep := map[string]float64{"system": rv.OperatorKeepDays, "scripted": rv.ScriptedKeepDays, "live": rv.LiveKeepDays}
	archive := filepath.Join(gopodArchiveDir(), "voices")
	moved := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".wav") || !gopodVoiceAnyFileRE.MatchString(name) || inUse[name] {
			continue
		}
		kind := kinds[name]
		if kind == "" {
			// Made before kinds were noted: Goverlord's voice is system, the
			// Operator's live, the rest scripted (the longest keep).
			switch {
			case strings.HasPrefix(name, rv.SystemVoice+"_") && rv.SystemVoice != "":
				kind = "system"
			case strings.HasPrefix(name, rv.OperatorVoice+"_") && rv.OperatorVoice != "":
				kind = "live"
			default:
				kind = "scripted"
			}
		}
		days := keep[kind]
		// 2026-10-27 (operator: "Finish the voice archiver limits"; the voice
		// report: how many plays make a file "popular"): a popular file
		// (popular_uses plays or more) keeps popular_keep_days, if longer.
		if rv.PopularUses > 0 && uses[name] >= rv.PopularUses && rv.PopularKeepDays > days {
			days = rv.PopularKeepDays
		}
		info, err := e.Info()
		if days <= 0 || err != nil || time.Since(info.ModTime()) < gopodSecondsToDuration(days*86400) {
			continue
		}
		if gopodMkdirAll(archive, 0o755) != nil {
			return
		}
		if os.Rename(filepath.Join(dir, name), filepath.Join(archive, name)) == nil {
			moved++
			gopodVoiceArchived(name)
			words := strings.TrimSuffix(name, ".wav") + ".words.json"
			os.Rename(filepath.Join(dir, words), filepath.Join(archive, words))
		}
	}
	logger.Println(fmt.Sprintf("GOPOD_REPORTER_VOICES archived %d voices past their keep days (%s), kept %d in use", moved, archive, len(inUse)))
}

// ## EOF
