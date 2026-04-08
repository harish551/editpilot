package planner

import "testing"

func TestBuildPromptPlanInfersTrimAndResize(t *testing.T) {
	plan, err := BuildPromptPlan([]string{"clip.mp4"}, "trim first 10 seconds and make it vertical", "out.mp4")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(plan.Operations) < 2 {
		t.Fatalf("expected at least 2 operations, got %d", len(plan.Operations))
	}
	if plan.Operations[0].Type != "trim" {
		t.Fatalf("expected first op trim, got %s", plan.Operations[0].Type)
	}
	if plan.Operations[1].Type != "resize" {
		t.Fatalf("expected second op resize, got %s", plan.Operations[1].Type)
	}
}

func TestBuildPromptPlanInfersConcat(t *testing.T) {
	plan, err := BuildPromptPlan([]string{"a.mp4", "b.mp4"}, "merge these clips", "out.mp4")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	found := false
	for _, op := range plan.Operations {
		if op.Type == "concat" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected concat operation")
	}
}

func TestBuildPromptPlanInfersTextAndSpeed(t *testing.T) {
	plan, err := BuildPromptPlan([]string{"clip.mp4"}, "add title \"Hello World\" and make it 1.5x speed", "out.mp4")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	foundText := false
	foundSpeed := false
	for _, op := range plan.Operations {
		if op.Type == "text_overlay" {
			foundText = true
		}
		if op.Type == "speed" {
			foundSpeed = true
		}
	}
	if !foundText || !foundSpeed {
		t.Fatalf("expected text_overlay and speed operations, got %#v", plan.Operations)
	}
}
