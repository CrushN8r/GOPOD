// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_typed_input_test.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

func TestGopodTypedInputRejects(t *testing.T) {
	if err := GopodTypedInput(gopodBrobot1ESN, "   "); err == nil {
		t.Errorf("empty text accepted")
	}
	if err := GopodTypedInput("00000000", "hello"); err == nil {
		t.Errorf("unknown robot accepted")
	}
}

// A typed voice-command phrase runs that command's script, as a heard one does.
func TestGopodTypedCommandRuns(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "ran")
	script := filepath.Join(dir, "cmd.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\ntouch "+marker+"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	oldIntents, oldExist := vars.CustomIntents, vars.CustomIntentsExist
	defer func() { vars.CustomIntents, vars.CustomIntentsExist = oldIntents, oldExist }()
	vars.CustomIntents = []vars.CustomIntent{{Name: "KILL_BINGO_TEST", Utterances: []string{"kill bingo test"}, Exec: script, ExecArgs: []string{""}}}
	vars.CustomIntentsExist = true

	if gopodRunTypedCommand(gopodBrobot2ESN, "what is gopod") {
		t.Fatalf("a question matched a command")
	}
	if !gopodRunTypedCommand(gopodBrobot2ESN, "Kill Bingo Test") {
		t.Fatalf("the command phrase didn't match")
	}
	for i := 0; i < 50; i++ {
		if _, err := os.Stat(marker); err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("the command's script didn't run")
}

// ## EOF
