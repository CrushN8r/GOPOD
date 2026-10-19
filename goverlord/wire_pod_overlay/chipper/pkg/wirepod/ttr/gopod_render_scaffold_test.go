package wirepod_ttr

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGOPODRenderProjectsRichAndFlatFromOneSemanticIR(t *testing.T) {
	ir := NewGOPODSemanticIR("happy... Brobot answer for the room.")

	rich := RenderGOPODResponse(ir, GOPODRenderModeRich)
	flat := RenderGOPODResponse(ir, GOPODRenderModeFlat)

	if rich.Mode != GOPODRenderModeRich {
		t.Fatalf("rich mode = %q", rich.Mode)
	}
	if rich.Content != ir.Content {
		t.Fatalf("rich content = %q, want IR content %q", rich.Content, ir.Content)
	}
	if rich.MarkedPacket != "|||0||| happy... Brobot answer for the room. |||1|||" {
		t.Fatalf("rich marked packet = %q", rich.MarkedPacket)
	}
	if flat.Mode != GOPODRenderModeFlat {
		t.Fatalf("flat mode = %q", flat.Mode)
	}
	if flat.Content != "happy... Brobot answer for the room." {
		t.Fatalf("flat content = %q", flat.Content)
	}
	if flat.FilteredPacket != "happy... Brobot answer for the room." {
		t.Fatalf("flat filtered packet = %q", flat.FilteredPacket)
	}
	if len(flat.AudibleChunks) != 1 || flat.AudibleChunks[0] != "happy... Brobot answer for the room." {
		t.Fatalf("flat audible chunks = %#v", flat.AudibleChunks)
	}
}

func TestGOPODRenderAPIKeepsESNBeforeIRAndRuntimeOutsideRender(t *testing.T) {
	irType := reflect.TypeOf(GOPODSemanticIR{})
	renderType := reflect.TypeOf(RenderGOPODResponse)

	if renderType.NumIn() != 2 || renderType.In(0) != irType || renderType.In(1) != reflect.TypeOf(GOPODRenderMode("")) {
		t.Fatalf("RenderGOPODResponse inputs = %v, want only SemanticIR and render mode", renderType)
	}

	ir := NewGOPODSemanticIR("happy... Same answer, same projections.")
	if !reflect.DeepEqual(RenderGOPODResponse(ir, GOPODRenderModeRich), RenderGOPODResponse(ir, GOPODRenderModeRich)) {
		t.Fatalf("rich render is not deterministic for identical IR")
	}
	if !reflect.DeepEqual(RenderGOPODResponse(ir, GOPODRenderModeFlat), RenderGOPODResponse(ir, GOPODRenderModeFlat)) {
		t.Fatalf("flat render is not deterministic for identical IR")
	}
}

func TestGOPODRenderScaffoldContainsNoForbiddenRuntimeOrPersonaCoupling(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(gopodChipperRoot(), "pkg", "wirepod", "ttr", "gopod_render_scaffold.go"))
	if err != nil {
		t.Fatalf("ReadFile(gopod_render_scaffold.go) error = %v", err)
	}
	source := string(data)
	for _, forbidden := range []string{
		"CHALK",
		"FLATHEAD",
		"ESN",
		"esn",
		"CreateAIReq",
		"CreateChatCompletion",
		"CreateChatCompletionStream",
		"GOPOD_STREAM_MARKER_0",
		"GOPOD_STREAM_MARKER_1",
		"BROBOT_RELEASE_READY",
		"Repair",
		"repair",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("render scaffold contains forbidden coupling %q", forbidden)
		}
	}
}

func TestGOPODStreamResultUsesRenderScaffoldWithoutChangingValidOutput(t *testing.T) {
	raw := "happy... Brobot answer for the room.\nthinking... Operator sees the same source."
	result := buildGOPODStreamResultWithRepair(context.Background(), raw, "ignored-model", nil)

	if result.Raw != raw {
		t.Fatalf("result raw = %q, want %q", result.Raw, raw)
	}
	if result.MarkedPacket != "|||0||| "+raw+" |||1|||" {
		t.Fatalf("result marked packet = %q", result.MarkedPacket)
	}
	if result.FilteredPacket != "happy... Brobot answer for the room. |||C||| thinking... Operator sees the same source." {
		t.Fatalf("result filtered packet = %q", result.FilteredPacket)
	}
	wantChunks := []string{"happy... Brobot answer for the room.", "thinking... Operator sees the same source."}
	if len(result.AudibleChunks) != len(wantChunks) {
		t.Fatalf("audible chunks = %#v, want %#v", result.AudibleChunks, wantChunks)
	}
	for i := range wantChunks {
		if result.AudibleChunks[i] != wantChunks[i] {
			t.Fatalf("audible chunk %d = %q, want %q", i, result.AudibleChunks[i], wantChunks[i])
		}
	}
}
