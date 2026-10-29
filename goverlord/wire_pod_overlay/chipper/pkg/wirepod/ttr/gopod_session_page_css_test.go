package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 2026-10-25 (operator: "move all css, js, etc... into files"): a saved session
// page carries its look from webroot/css/gopod_session_page.css.
func TestSessionPageCarriesItsCSS(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "chat.jsonl"), []byte(`{"time":"1","who":"Operator","text":"hi"}`+"\n"), 0644)
	gopodRecordWriteHTML(dir)
	matches, _ := filepath.Glob(filepath.Join(dir, "*.html"))
	if len(matches) == 0 {
		t.Fatal("no page written")
	}
	page, _ := os.ReadFile(matches[0])
	if !strings.Contains(string(page), ".b2{background:var(--brobot-2-bg)}") || !strings.Contains(string(page), "--brobot-2-bg:#101a4a") || !strings.Contains(string(page), "Operator") {
		t.Fatalf("page lacks its css or line:\n%s", page)
	}
}
