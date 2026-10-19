package wirepod_ttr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/sashabaranov/go-openai"
)

type gopodDemoProbeSet struct {
	Name   string           `json:"name"`
	Probes []gopodDemoProbe `json:"probes"`
}

type gopodDemoProbe struct {
	ID                     int    `json:"id"`
	Question               string `json:"question"`
	FollowUp               string `json:"follow_up,omitempty"`
	ExpectedScenarioPacket string `json:"expected_scenario_packet"`
	ExpectedValueTarget    string `json:"expected_value_target"`
	ExpectedCTA            string `json:"expected_cta"`
}

type gopodDemoProbeRunReport struct {
	Name            string                 `json:"name"`
	GeneratedAtUTC  string                 `json:"generated_at_utc"`
	LLMMode         bool                   `json:"llm_mode"`
	LiveRobotSpeech bool                   `json:"live_robot_speech"`
	ProbeFile       string                 `json:"probe_file"`
	Results         []gopodDemoProbeResult `json:"results"`
	Notes           []string               `json:"notes,omitempty"`
}

type gopodDemoProbeResult struct {
	ID                       int    `json:"id"`
	Question                 string `json:"question"`
	FollowUp                 string `json:"follow_up,omitempty"`
	ScoredQuestion           string `json:"scored_question,omitempty"`
	ExpectedScenario         string `json:"expected_scenario_packet"`
	MatchedScenario          string `json:"matched_scenario_packet"`
	PacketPass               bool   `json:"packet_pass"`
	PromptOrderPass          bool   `json:"prompt_order_pass"`
	WrongScenarioBleed       string `json:"wrong_scenario_bleed"`
	FormatPass               bool   `json:"format_pass,omitempty"`
	PersonaScore             string `json:"persona_score,omitempty"`
	CTAScore                 string `json:"cta_score,omitempty"`
	ValueScore               string `json:"value_score,omitempty"`
	RepairUsed               string `json:"repair_used,omitempty"`
	DemoReady                string `json:"demo_ready"`
	QuestionRawOutput        string `json:"question_raw_output,omitempty"`
	QuestionNormalizedOutput string `json:"question_normalized_output,omitempty"`
	QuestionFinalOutput      string `json:"question_final_robot_safe_output,omitempty"`
	QuestionRepairUsed       string `json:"question_repair_used,omitempty"`
	RawOutput                string `json:"raw_output,omitempty"`
	NormalizedOutput         string `json:"normalized_output,omitempty"`
	FinalRobotSafeOutput     string `json:"final_robot_safe_output,omitempty"`
	FirstFailingLine         string `json:"first_failing_line,omitempty"`
	RepairPathTaken          string `json:"repair_path_taken,omitempty"`
}

func TestGOPODDemoProbeFollowUpChoosesFollowUpAsScoredQuestion(t *testing.T) {
	probe := gopodDemoProbe{
		ID:       1001,
		Question: "General probe stays first.",
		FollowUp: "Targeted follow-up gets scored.",
	}
	if got := scoredQuestionForGOPODDemoProbe(probe); got != probe.FollowUp {
		t.Fatalf("scored question = %q, want follow_up %q", got, probe.FollowUp)
	}
	probe.FollowUp = ""
	if got := scoredQuestionForGOPODDemoProbe(probe); got != probe.Question {
		t.Fatalf("scored question without follow_up = %q, want question %q", got, probe.Question)
	}
}

func scoredQuestionForGOPODDemoProbe(probe gopodDemoProbe) string {
	if strings.TrimSpace(probe.FollowUp) != "" {
		return probe.FollowUp
	}
	return probe.Question
}

func loadAPIConfigForDemoProbeTest(t *testing.T) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "apiConfig.json"))
	if err != nil {
		t.Fatalf("ReadFile(apiConfig.json) error = %v", err)
	}
	if err := json.Unmarshal(data, &vars.APIConfig); err != nil {
		t.Fatalf("Unmarshal(apiConfig.json) error = %v", err)
	}
}

func callGOPODDemoProbeLLM(ctx context.Context, req openai.ChatCompletionRequest) (string, error) {
	req.Stream = false
	client, err := openAIClientForDemoProbe()
	if err != nil {
		return "", err
	}
	response, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("empty LLM response")
	}
	return response.Choices[0].Message.Content, nil
}

func openAIClientForDemoProbe() (*openai.Client, error) {
	switch vars.APIConfig.Knowledge.Provider {
	case "together":
		conf := openai.DefaultConfig(vars.APIConfig.Knowledge.Key)
		conf.BaseURL = "https://api.together.xyz/v1"
		return openai.NewClientWithConfig(conf), nil
	case "custom":
		conf := openai.DefaultConfig(vars.APIConfig.Knowledge.Key)
		conf.BaseURL = vars.APIConfig.Knowledge.Endpoint
		return openai.NewClientWithConfig(conf), nil
	case "openai":
		return openai.NewClient(vars.APIConfig.Knowledge.Key), nil
	default:
		return nil, fmt.Errorf("unsupported or empty knowledge provider %q", vars.APIConfig.Knowledge.Provider)
	}
}

func enforceGOPODDemoProbeRobotSpeech(ctx context.Context, model, raw string) (final, normalized, repairUsed, firstFailingLine string) {
	validation := validateGOPODRobotSpeech(raw)
	if validation.Valid {
		return strings.TrimSpace(raw), strings.TrimSpace(raw), "none", ""
	}
	firstFailingLine = validation.FirstFailingLine
	if normalizedOutput, normalizedValidation := normalizeGOPODRobotSpeech(raw); normalizedValidation.Valid {
		return normalizedOutput, normalizedOutput, "deterministic", firstFailingLine
	}
	client, err := openAIClientForDemoProbe()
	if err == nil {
		repairer := openAIGOPODRobotSpeechRepairer(client)
		repaired, repairErr := repairer(ctx, createGOPODRobotSpeechRepairRequest(model, raw))
		if repairErr == nil && validateGOPODRobotSpeech(repaired).Valid {
			return strings.TrimSpace(repaired), normalized, "llm", firstFailingLine
		}
	}
	return gopodRobotSpeechFallbackLine, normalized, "fallback", firstFailingLine
}
