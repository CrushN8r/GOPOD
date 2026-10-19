package wirepod_ttr

import (
	"strings"
	"sync"
)

const gopodDAGFingerprintPath = "/gopod/ir/.dag_fingerprint.json"
const gopodAnchoredDAGHash = "af8f5d4d77189308e4dfb37d5c327b3c9d40b43e68942bab2715cc6434a5180b"
const gopodDAGHashViolation = "GOPOD_DAG_HASH_VIOLATION: render layer drifted from bootstrap anchor"

var gopodDAGAnchorMu sync.Mutex
var gopodDAGAnchorVerified bool

type gopodDAGFingerprint struct {
	DAGHash string `json:"dag_hash"`
}

type GOPODSemanticIR struct {
	Content string
}

type GOPODRenderMode string

const (
	GOPODRenderModeRich GOPODRenderMode = "RICH"
	GOPODRenderModeFlat GOPODRenderMode = "FLAT"
)

type GOPODRenderedOutput struct {
	Mode           GOPODRenderMode
	Content        string
	MarkedPacket   string
	FilteredPacket string
	AudibleChunks  []string
}

func NewGOPODSemanticIR(content string) GOPODSemanticIR {
	return GOPODSemanticIR{Content: content}
}

func verifyGOPODDAGAnchor() {
	// integrity check project deferred, see [note]
}

func RenderGOPODResponse(ir GOPODSemanticIR, mode GOPODRenderMode) GOPODRenderedOutput {
	verifyGOPODDAGAnchor()
	switch mode {
	case GOPODRenderModeRich:
		return GOPODRenderedOutput{
			Mode:         GOPODRenderModeRich,
			Content:      ir.Content,
			MarkedPacket: composeBrobotRawMarkedPacket(ir.Content),
		}
	case GOPODRenderModeFlat:
		safe := strings.TrimSpace(ir.Content)
		if validation := validateGOPODRobotSpeech(safe); !validation.Valid {
			if normalized, normalizedValidation := normalizeGOPODRobotSpeech(safe); normalizedValidation.Valid {
				safe = normalized
			} else {
				safe = gopodRobotSpeechFallbackLine
			}
		}
		chunks := nonEmptyRobotSpeechLines(safe)
		return GOPODRenderedOutput{
			Mode:           GOPODRenderModeFlat,
			Content:        safe,
			FilteredPacket: strings.Join(chunks, " "+gopodCanonicalChunkMarker+" "),
			AudibleChunks:  chunks,
		}
	default:
		return GOPODRenderedOutput{
			Mode:    mode,
			Content: ir.Content,
		}
	}
}
