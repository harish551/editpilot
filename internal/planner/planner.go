package planner

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/harish551/editpilot/internal/models"
)

func BuildPromptPlan(inputs []string, prompt string, output string) (models.Plan, error) {
	if len(inputs) == 0 {
		return models.Plan{}, fmt.Errorf("at least one input is required")
	}

	assets := make([]models.InputAsset, 0, len(inputs))
	assetIDs := make([]string, 0, len(inputs))
	for i, in := range inputs {
		id := fmt.Sprintf("input%d", i+1)
		assets = append(assets, models.InputAsset{ID: id, Path: in})
		assetIDs = append(assetIDs, id)
	}

	ops := inferOperations(prompt, assetIDs)

	return models.Plan{
		Inputs:     assets,
		Operations: ops,
		Output:     models.OutputSpec{Path: output},
	}, nil
}

func inferOperations(prompt string, assetIDs []string) []models.Operation {
	lower := strings.ToLower(prompt)
	ops := make([]models.Operation, 0)

	if trimOp, ok := parseTrim(lower, assetIDs[0]); ok {
		ops = append(ops, trimOp)
	}

	if resizeOp, ok := parseResize(lower, assetIDs[0]); ok {
		ops = append(ops, resizeOp)
	}

	if len(assetIDs) > 1 && (strings.Contains(lower, "merge") || strings.Contains(lower, "concat") || strings.Contains(lower, "combine")) {
		ops = append(ops, models.Operation{Type: "concat", Inputs: assetIDs})
	}

	if strings.Contains(lower, "mute") {
		ops = append(ops, models.Operation{Type: "mute", Input: assetIDs[0]})
	}
	if textOp, ok := parseTextOverlay(prompt, assetIDs[0]); ok {
		ops = append(ops, textOp)
	}
	if speedOp, ok := parseSpeed(lower, assetIDs[0]); ok {
		ops = append(ops, speedOp)
	}

	if len(ops) == 0 {
		ops = append(ops, models.Operation{Type: "noop", Input: assetIDs[0], Params: map[string]any{"note": "no supported operations inferred yet"}})
	}

	return ops
}

func parseTrim(prompt string, assetID string) (models.Operation, bool) {
	rangeRe := regexp.MustCompile(`trim(?:[^0-9]{0,20})?(\d+)(?:\s*to\s*|\s*-\s*|\s*until\s*)(\d+)`)
	matches := rangeRe.FindStringSubmatch(prompt)
	if len(matches) == 3 {
		start, _ := strconv.ParseFloat(matches[1], 64)
		end, _ := strconv.ParseFloat(matches[2], 64)
		return models.Operation{Type: "trim", Input: assetID, Params: map[string]any{"start": start, "end": end}}, true
	}

	firstRe := regexp.MustCompile(`first\s+(\d+)\s*seconds?`)
	firstMatches := firstRe.FindStringSubmatch(prompt)
	if len(firstMatches) == 2 {
		end, _ := strconv.ParseFloat(firstMatches[1], 64)
		return models.Operation{Type: "trim", Input: assetID, Params: map[string]any{"start": 0.0, "end": end}}, true
	}

	return models.Operation{}, false
}

func parseResize(prompt string, assetID string) (models.Operation, bool) {
	if strings.Contains(prompt, "vertical") || strings.Contains(prompt, "9:16") {
		return models.Operation{Type: "resize", Input: assetID, Params: map[string]any{"width": 1080, "height": 1920}}, true
	}

	dimRe := regexp.MustCompile(`(\d{3,5})x(\d{3,5})`)
	matches := dimRe.FindStringSubmatch(prompt)
	if len(matches) == 3 {
		width, _ := strconv.Atoi(matches[1])
		height, _ := strconv.Atoi(matches[2])
		return models.Operation{Type: "resize", Input: assetID, Params: map[string]any{"width": width, "height": height}}, true
	}

	return models.Operation{}, false
}

func parseTextOverlay(prompt string, assetID string) (models.Operation, bool) {
	quotedRe := regexp.MustCompile(`(?i)(?:title|text|caption)(?:[^"\n]{0,20})"([^"]+)"`)
	matches := quotedRe.FindStringSubmatch(prompt)
	if len(matches) == 2 {
		return models.Operation{Type: "text_overlay", Input: assetID, Params: map[string]any{"text": matches[1], "x": "(w-text_w)/2", "y": "80", "font_size": 48}}, true
	}
	if strings.Contains(strings.ToLower(prompt), "title") || strings.Contains(strings.ToLower(prompt), "text") || strings.Contains(strings.ToLower(prompt), "caption") {
		return models.Operation{Type: "text_overlay", Input: assetID, Params: map[string]any{"text": "EditPilot", "x": "(w-text_w)/2", "y": "80", "font_size": 48}}, true
	}
	return models.Operation{}, false
}

func parseSpeed(prompt string, assetID string) (models.Operation, bool) {
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)x`)
	matches := re.FindStringSubmatch(prompt)
	if len(matches) == 2 && strings.Contains(prompt, "speed") {
		rate, _ := strconv.ParseFloat(matches[1], 64)
		return models.Operation{Type: "speed", Input: assetID, Params: map[string]any{"factor": rate}}, true
	}
	if strings.Contains(prompt, "faster") {
		return models.Operation{Type: "speed", Input: assetID, Params: map[string]any{"factor": 1.25}}, true
	}
	if strings.Contains(prompt, "slower") {
		return models.Operation{Type: "speed", Input: assetID, Params: map[string]any{"factor": 0.8}}, true
	}
	return models.Operation{}, false
}
