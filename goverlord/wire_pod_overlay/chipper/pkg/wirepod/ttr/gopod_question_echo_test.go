// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_question_echo_test.go
package wirepod_ttr

import "testing"

// TestDropQuestionEcho: the question read back is dropped, the rest stays.
func TestDropQuestionEcho(t *testing.T) {
	q := "I want to visit my granny today. She is so gangsta! What's the weather in Jerusalem, Israel?"
	reply := "Granny is gonna catch some sunshine. I want to visit my granny today. She is gangsta! What's the weather in Jerusalem, Israel? Brobots got this!"
	if got, want := gopodDropQuestionEcho(reply, q), "Granny is gonna catch some sunshine. Brobots got this!"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := gopodDropQuestionEcho("Jerusalem is sunny!", q); got != "Jerusalem is sunny!" {
		t.Errorf("a short sentence went: %q", got)
	}
}

// ## EOF
