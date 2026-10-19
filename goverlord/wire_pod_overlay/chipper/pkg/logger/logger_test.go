package logger

import (
	"strings"
	"testing"
)

func resetLogsForTest() {
	LogList = ""
	LogArray = nil
	LogTrayList = ""
	LogTrayArray = nil
	LogTrayChan = make(chan string, 1)
}

func TestLogDebugUIRoutesOnlyToDebugLogs(t *testing.T) {
	resetLogsForTest()

	LogUI("normal operational line")
	LogDebugUI("FILTERED_SPLIT_READY:\nrobot chunk")

	if !strings.Contains(LogList, "normal operational line") {
		t.Fatalf("normal log missing operational line: %q", LogList)
	}
	if strings.Contains(LogList, "FILTERED_SPLIT_READY") {
		t.Fatalf("normal log contains debug diagnostic: %q", LogList)
	}
	if !strings.Contains(LogTrayList, "FILTERED_SPLIT_READY:\nrobot chunk") {
		t.Fatalf("debug log missing diagnostic: %q", LogTrayList)
	}
}

func TestOriginalWirePodFinalLLMResponseRoutesOnlyToNormalLogs(t *testing.T) {
	resetLogsForTest()

	LogUI("LLM response for 0dd1b9e9: full raw output")

	if !strings.Contains(LogList, "LLM response for 0dd1b9e9: full raw output") {
		t.Fatalf("normal log missing final LLM response: %q", LogList)
	}
	if strings.Contains(LogTrayList, "LLM response for 0dd1b9e9") {
		t.Fatalf("debug log unexpectedly contains normal final LLM response: %q", LogTrayList)
	}
}
