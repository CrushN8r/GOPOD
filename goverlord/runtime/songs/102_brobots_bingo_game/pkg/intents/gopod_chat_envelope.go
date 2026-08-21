package intents

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type ChatEnvelope struct {
	SpeakerID      string `json:"speaker_id"`
	DisplayName    string `json:"display_name"`
	Text           string `json:"text"`
	Timestamp      string `json:"timestamp"`
	SourceProducer string `json:"source_producer"`
}

func AppendBingoEnvelope(chatFilePath string, sourceProducer string, text string) error {
	env := ChatEnvelope{
		SpeakerID:      "bingo",
		DisplayName:    "Bingo",
		Text:           text,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		SourceProducer: sourceProducer,
	}

	raw, err := os.ReadFile(chatFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"GOPOD_BINGO_ENVELOPE_WRITE_SKIP reason=read_failed path=%s err=%v\n",
			chatFilePath, err)
		return nil
	}

	var chat struct {
		Event    string         `json:"event"`
		Status   string         `json:"status"`
		Messages []ChatEnvelope `json:"messages"`
	}
	if err := json.Unmarshal(raw, &chat); err != nil {
		fmt.Fprintf(os.Stderr,
			"GOPOD_BINGO_ENVELOPE_WRITE_SKIP reason=parse_failed err=%v\n", err)
		return nil
	}

	chat.Messages = append(chat.Messages, env)

	const messageLimit = 40
	if len(chat.Messages) > messageLimit {
		chat.Messages = chat.Messages[len(chat.Messages)-messageLimit:]
	}

	out, err := json.MarshalIndent(chat, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"GOPOD_BINGO_ENVELOPE_WRITE_SKIP reason=marshal_failed err=%v\n", err)
		return nil
	}

	tmpPath := chatFilePath + ".bingo.tmp"
	if err := os.WriteFile(tmpPath, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr,
			"GOPOD_BINGO_ENVELOPE_WRITE_SKIP reason=tmp_write_failed err=%v\n", err)
		return nil
	}
	if err := os.Rename(tmpPath, chatFilePath); err != nil {
		fmt.Fprintf(os.Stderr,
			"GOPOD_BINGO_ENVELOPE_WRITE_SKIP reason=rename_failed err=%v\n", err)
		return nil
	}

	fmt.Fprintf(os.Stderr,
		"GOPOD_BINGO_ENVELOPE_WRITTEN source_producer=%s text=%s\n",
		sourceProducer, text)
	return nil
}
