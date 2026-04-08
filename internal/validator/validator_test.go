package validator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/harish551/editpilot/internal/models"
)

func TestValidatePlanAcceptsValidResizePlan(t *testing.T) {
	input := makeFixtureVideo(t, "clip.mp4", "320x240", 2)
	plan := models.Plan{
		Inputs: []models.InputAsset{{ID: "input1", Path: input}},
		Operations: []models.Operation{{
			Type:  "resize",
			Input: "input1",
			Params: map[string]any{"width": 1080, "height": 1920},
		}},
		Output: models.OutputSpec{Path: filepath.Join(t.TempDir(), "out.mp4")},
	}

	if err := ValidatePlan(plan); err != nil {
		t.Fatalf("expected valid plan, got %v", err)
	}
}

func TestValidatePlanRejectsBrokenTrim(t *testing.T) {
	input := makeFixtureVideo(t, "clip.mp4", "320x240", 2)
	plan := models.Plan{
		Inputs: []models.InputAsset{{ID: "input1", Path: input}},
		Operations: []models.Operation{{
			Type:  "trim",
			Input: "input1",
			Params: map[string]any{"start": 10.0, "end": 2.0},
		}},
		Output: models.OutputSpec{Path: filepath.Join(t.TempDir(), "out.mp4")},
	}

	if err := ValidatePlan(plan); err == nil {
		t.Fatalf("expected validation error for invalid trim")
	}
}

func TestValidatePlanRejectsTrimPastDuration(t *testing.T) {
	input := makeFixtureVideo(t, "clip.mp4", "320x240", 1)
	plan := models.Plan{
		Inputs: []models.InputAsset{{ID: "input1", Path: input}},
		Operations: []models.Operation{{
			Type:  "trim",
			Input: "input1",
			Params: map[string]any{"start": 0.0, "end": 5.0},
		}},
		Output: models.OutputSpec{Path: filepath.Join(t.TempDir(), "out.mp4")},
	}

	if err := ValidatePlan(plan); err == nil {
		t.Fatalf("expected validation error for trim beyond duration")
	}
}

func TestValidatePlanRejectsEmptyTextOverlay(t *testing.T) {
	input := makeFixtureVideo(t, "clip.mp4", "320x240", 2)
	plan := models.Plan{
		Inputs: []models.InputAsset{{ID: "input1", Path: input}},
		Operations: []models.Operation{{
			Type:  "text_overlay",
			Input: "input1",
			Params: map[string]any{"text": "   "},
		}},
		Output: models.OutputSpec{Path: filepath.Join(t.TempDir(), "out.mp4")},
	}

	if err := ValidatePlan(plan); err == nil {
		t.Fatalf("expected validation error for empty text")
	}
}

func TestValidatePlanRejectsConcatDimensionMismatch(t *testing.T) {
	input1 := makeFixtureVideo(t, "a.mp4", "320x240", 1)
	input2 := makeFixtureVideo(t, "b.mp4", "640x360", 1)
	plan := models.Plan{
		Inputs: []models.InputAsset{{ID: "input1", Path: input1}, {ID: "input2", Path: input2}},
		Operations: []models.Operation{{Type: "concat", Inputs: []string{"input1", "input2"}}},
		Output: models.OutputSpec{Path: filepath.Join(t.TempDir(), "out.mp4")},
	}

	if err := ValidatePlan(plan); err == nil {
		t.Fatalf("expected validation error for mismatched concat dimensions")
	}
}

func makeFixtureVideo(t *testing.T, name, size string, duration int) string {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available")
	}
	path := filepath.Join(t.TempDir(), name)
	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", fmt.Sprintf("color=c=blue:s=%s:d=%d", size, duration),
		"-f", "lavfi",
		"-i", fmt.Sprintf("anullsrc=r=44100:cl=stereo:d=%d", duration),
		"-shortest",
		"-c:v", "libx264",
		"-c:a", "aac",
		path,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create fixture video: %v\n%s", err, string(out))
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected fixture file: %v", err)
	}
	return path
}
