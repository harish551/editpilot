package ffmpeg

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/harish551/editpilot/internal/models"
)

func TestBuildArgsResize(t *testing.T) {
	plan := models.Plan{
		Inputs: []models.InputAsset{{ID: "input1", Path: "clip.mp4"}},
		Operations: []models.Operation{{
			Type:  "resize",
			Input: "input1",
			Params: map[string]any{
				"width":  1080,
				"height": 1920,
			},
		}},
		Output: models.OutputSpec{Path: "out.mp4"},
	}

	args, err := BuildArgs(plan)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "scale=1080:1920") {
		t.Fatalf("expected resize filter in args, got %s", joined)
	}
}

func TestBuildArgsTextOverlayAndSpeed(t *testing.T) {
	plan := models.Plan{
		Inputs: []models.InputAsset{{ID: "input1", Path: "clip.mp4"}},
		Operations: []models.Operation{
			{Type: "text_overlay", Input: "input1", Params: map[string]any{"text": "Hello", "font_size": 42}},
			{Type: "speed", Input: "input1", Params: map[string]any{"factor": 1.5}},
		},
		Output: models.OutputSpec{Path: "out.mp4"},
	}

	args, err := BuildArgs(plan)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "drawtext=") {
		t.Fatalf("expected drawtext filter, got %s", joined)
	}
	if !strings.Contains(joined, "setpts=") {
		t.Fatalf("expected setpts filter, got %s", joined)
	}
	if !strings.Contains(joined, "atempo=") {
		t.Fatalf("expected atempo filter, got %s", joined)
	}
}

func TestBuildArgsConcat(t *testing.T) {
	plan := models.Plan{
		Inputs: []models.InputAsset{
			{ID: "input1", Path: "a.mp4"},
			{ID: "input2", Path: "b.mp4"},
		},
		Operations: []models.Operation{{Type: "concat", Inputs: []string{"input1", "input2"}}},
		Output: models.OutputSpec{Path: "out.mp4"},
	}

	args, err := BuildArgs(plan)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-f concat") {
		t.Fatalf("expected concat mode args, got %s", joined)
	}
	if filepath.Base(args[len(args)-1]) != "out.mp4" {
		t.Fatalf("expected output at end, got %v", args[len(args)-1])
	}
}

func TestBuildArgsFilteredConcatWithResize(t *testing.T) {
	plan := models.Plan{
		Inputs: []models.InputAsset{
			{ID: "input1", Path: "a.mp4"},
			{ID: "input2", Path: "b.mp4"},
		},
		Operations: []models.Operation{
			{Type: "resize", Input: "input1", Params: map[string]any{"width": 1080, "height": 1920}},
			{Type: "resize", Input: "input2", Params: map[string]any{"width": 1080, "height": 1920}},
			{Type: "concat", Inputs: []string{"input1", "input2"}},
		},
		Output: models.OutputSpec{Path: "out.mp4"},
	}

	args, err := BuildArgs(plan)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-filter_complex") {
		t.Fatalf("expected filter_complex in args, got %s", joined)
	}
	if !strings.Contains(joined, "concat=n=2:v=1:a=1") {
		t.Fatalf("expected filtered concat graph, got %s", joined)
	}
	if !strings.Contains(joined, "scale=1080:1920") {
		t.Fatalf("expected resize in filter graph, got %s", joined)
	}
}
