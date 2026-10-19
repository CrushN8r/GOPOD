package wirepod_ttr

import (
	"context"
	"os"
	"strings"
	"testing"
)

const brobot2LLMProbeESN = "0dd1d8bf"

func TestBrobot2IndependentLLMPath(t *testing.T) {
	runLLM := strings.TrimSpace(os.Getenv("GOPOD_BROBOT2_LLM_PROBE")) == "1"
	if runLLM {
		loadAPIConfigForDemoProbeTest(t)
	}

	input := "I feel good. Elaborate."
	req := CreateAIReq(input, brobot2LLMProbeESN, false, true)
	if len(req.Messages) < 2 {
		t.Fatalf("CreateAIReq messages length = %d, want system and user messages", len(req.Messages))
	}
	systemPrompt := req.Messages[0].Content
	// 2026-10-05: the prompt names the brother by his persona line, not his ESN.
	if !strings.Contains(systemPrompt, "WHO I AM:\n"+gopodBrobotsSettings.Personas["Brobot 2"]) {
		t.Fatal("system prompt missing Brobot 2's own persona line")
	}

	raw := "happy... I feel good, so I am turning that little spark into a clear GOPOD moment."
	if runLLM {
		var err error
		raw, err = callGOPODDemoProbeLLM(context.Background(), req)
		if err != nil {
			t.Fatalf("Brobot 2 LLM call failed: %v", err)
		}
		if strings.TrimSpace(raw) == "" {
			t.Fatal("Brobot 2 LLM response is empty")
		}
	}

	result := buildGOPODStreamResultWithRepair(context.Background(), raw, req.Model, nil)
	t.Logf("Brobot 2 raw LLM response: %s", raw)
	t.Logf("Brobot 2 filtered robot-safe packet: %s", result.FilteredPacket)
	if strings.TrimSpace(result.Raw) == "" {
		t.Fatal("raw response is empty after validation path")
	}
	if len(result.AudibleChunks) == 0 {
		t.Fatalf("AudibleChunks empty for raw response %q", raw)
	}
	for _, chunk := range result.AudibleChunks {
		compiled := CompileEmotionalBeatActions(chunk)
		if compiled.FormatPass != true {
			t.Fatalf("compiled emotional beat failed format for chunk %q: %#v", chunk, compiled)
		}
		if len(compiled.Actions) != 2 {
			t.Fatalf("compiled actions = %#v, want animation + say text", compiled.Actions)
		}
	}
	plan := CompileRobotSafeEmotionalBeatsDry(strings.Join(result.AudibleChunks, "\n"))
	if plan.PerformActionsCalled || plan.RobotCalls || plan.CameraCalls || plan.VisionAPICalls {
		t.Fatalf("dry action plan crossed live boundary: %#v", plan)
	}
	if len(plan.ActionPlan) == 0 {
		t.Fatal("dry action plan did not produce actions")
	}
}
