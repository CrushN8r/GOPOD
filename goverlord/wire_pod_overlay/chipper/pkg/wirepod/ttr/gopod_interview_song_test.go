// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_interview_song_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The record cut's draft_script.md, written as a song folder and read back, is
// the same run - every question, line, cue and move.
func TestInterviewSongPlaysTheSameAsTheDraft(t *testing.T) {
	home, _ := os.UserHomeDir()
	script := filepath.Join(home, "gopod_tts/sessions/harvest/2026-10-08/1300_record_cut_best_takes", gopodDraftScriptName)
	data, err := os.ReadFile(script)
	if err != nil {
		t.Skip("no record-cut draft on this machine")
	}
	want, err := gopodParseDraftScript(string(data))
	if err != nil {
		t.Fatal(err)
	}
	story, knobs, err := gopodInterviewSongFiles(want)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "story.md"), []byte(story), 0o644)
	os.WriteFile(filepath.Join(dir, "knobs.json"), knobs, 0o644)
	got, err := gopodInterviewSongDraft(dir)
	if err != nil {
		t.Fatal(err)
	}
	got.Run, want.Run = "", ""
	if !reflect.DeepEqual(got, want) {
		t.Errorf("the song folder plays differently:\n%#v\nwant\n%#v", got, want)
	}
}

// The repo's own interview song, found by its song_id, reads, and written back
// as a song folder it plays the same - chapters, pictures and closer too.
func TestRepoInterviewSongRoundTrip(t *testing.T) {
	repo, err := gopodFindSong(gopodBrobotsSettings.InterviewScript.SongID)
	if err != nil {
		t.Skip("no interview song")
	}
	want, err := gopodInterviewSongDraft(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(want.Closer) == 0 || want.Questions[0].Chapter == "" {
		t.Error("the interview song has no closer or no first chapter")
	}
	story, knobs, err := gopodInterviewSongFiles(want)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "story.md"), []byte(story), 0o644)
	os.WriteFile(filepath.Join(dir, "knobs.json"), knobs, 0o644)
	got, err := gopodInterviewSongDraft(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("round trip differs:\n%#v\nwant\n%#v", got, want)
	}
	for _, q := range want.Questions {
		t.Logf("%-28s | %-22s | %s -> %s: %s %v", q.Chapter, q.Topic, q.ReadBy, q.To, q.Ask, q.Pictures)
	}
}

// GOPOD_WRITE_INTERVIEW_SONG=<folder> writes the newest draft there as a song
// folder (story.md, knobs.json); an existing knobs.json is left alone.
func TestWriteInterviewSong(t *testing.T) {
	dir := os.Getenv("GOPOD_WRITE_INTERVIEW_SONG")
	if dir == "" {
		t.Skip("GOPOD_WRITE_INTERVIEW_SONG not set")
	}
	if _, err := os.Stat(filepath.Join(dir, "knobs.json")); err == nil {
		t.Fatal(dir + " already has a knobs.json")
	}
	_, d, err := gopodLoadDraft()
	if err != nil {
		t.Fatal(err)
	}
	story, knobs, err := gopodInterviewSongFiles(d)
	if err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(dir, 0o755)
	if err := os.WriteFile(filepath.Join(dir, "story.md"), []byte(story), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "knobs.json"), knobs, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Log("wrote " + dir)
}

// Songs are found by song_id, never by folder name (gopod_song_folders.go).
func TestSongsFoundBySongID(t *testing.T) {
	songs, err := gopodAllSongs()
	if err != nil {
		t.Fatal(err)
	}
	if len(songs) == 0 {
		t.Skip("no songs folder")
	}
	for _, id := range []string{"brobots_bingo", "brobots_interview", "brobots_baby_robots_sleep"} {
		if _, ok := songs[id]; !ok {
			t.Errorf("song %s not found (songs: %v)", id, songs)
		}
	}
	if gopodBrobotsSettings.InterviewScript.SongID == "" {
		t.Error("interview_script.song_id is empty")
	}
}
