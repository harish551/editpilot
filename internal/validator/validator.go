package validator

import (
	"fmt"
	"os"
	"strings"

	"github.com/harish551/editpilot/internal/media"
	"github.com/harish551/editpilot/internal/models"
)

var supportedOps = map[string]struct{}{
	"trim":         {},
	"concat":       {},
	"resize":       {},
	"mute":         {},
	"text_overlay": {},
	"speed":        {},
	"noop":         {},
}

func ValidatePlan(plan models.Plan) error {
	if len(plan.Inputs) == 0 {
		return fmt.Errorf("plan must include at least one input")
	}

	inputSet := make(map[string]struct{}, len(plan.Inputs))
	probes := make(map[string]media.ProbeResult, len(plan.Inputs))
	for _, input := range plan.Inputs {
		if input.ID == "" {
			return fmt.Errorf("input id cannot be empty")
		}
		if input.Path == "" {
			return fmt.Errorf("input path cannot be empty")
		}
		if _, err := os.Stat(input.Path); err != nil {
			return fmt.Errorf("input %s not accessible: %w", input.Path, err)
		}
		probe, err := media.Probe(input.Path)
		if err != nil {
			return fmt.Errorf("probe input %s: %w", input.Path, err)
		}
		if !probe.HasVideoStream() {
			return fmt.Errorf("input %s does not contain a video stream", input.Path)
		}
		inputSet[input.ID] = struct{}{}
		probes[input.ID] = probe
	}

	if plan.Output.Path == "" {
		return fmt.Errorf("output path cannot be empty")
	}
	for _, input := range plan.Inputs {
		if input.Path == plan.Output.Path {
			return fmt.Errorf("output path must differ from input path")
		}
	}

	for _, op := range plan.Operations {
		if _, ok := supportedOps[op.Type]; !ok {
			return fmt.Errorf("unsupported operation: %s", op.Type)
		}
		if op.Input != "" {
			if _, ok := inputSet[op.Input]; !ok {
				return fmt.Errorf("operation %s references unknown input %s", op.Type, op.Input)
			}
		}
		for _, in := range op.Inputs {
			if _, ok := inputSet[in]; !ok {
				return fmt.Errorf("operation %s references unknown input %s", op.Type, in)
			}
		}
		if err := validateOperation(op, probes); err != nil {
			return err
		}
	}

	if concatOp := findConcat(plan.Operations); concatOp != nil {
		if err := validateConcatCompatibility(concatOp.Inputs, probes); err != nil {
			return err
		}
	}

	return nil
}

func validateOperation(op models.Operation, probes map[string]media.ProbeResult) error {
	switch op.Type {
	case "trim":
		start, ok := getFloat(op.Params, "start")
		if !ok {
			return fmt.Errorf("trim operation requires numeric start")
		}
		end, ok := getFloat(op.Params, "end")
		if !ok {
			return fmt.Errorf("trim operation requires numeric end")
		}
		if start < 0 || end <= start {
			return fmt.Errorf("trim operation requires end > start >= 0")
		}
		if op.Input != "" {
			duration, err := probes[op.Input].DurationSeconds()
			if err == nil && end > duration {
				return fmt.Errorf("trim end %.2f exceeds input duration %.2f for %s", end, duration, op.Input)
			}
		}
	case "resize":
		width, ok := getInt(op.Params, "width")
		if !ok || width <= 0 {
			return fmt.Errorf("resize operation requires positive width")
		}
		height, ok := getInt(op.Params, "height")
		if !ok || height <= 0 {
			return fmt.Errorf("resize operation requires positive height")
		}
	case "concat":
		if len(op.Inputs) < 2 {
			return fmt.Errorf("concat operation requires at least two inputs")
		}
	case "text_overlay":
		text, ok := getString(op.Params, "text")
		if !ok || strings.TrimSpace(text) == "" {
			return fmt.Errorf("text_overlay operation requires non-empty text")
		}
		fontSize, ok := getInt(op.Params, "font_size")
		if ok && fontSize <= 0 {
			return fmt.Errorf("text_overlay font_size must be positive")
		}
	case "speed":
		factor, ok := getFloat(op.Params, "factor")
		if !ok || factor <= 0 {
			return fmt.Errorf("speed operation requires factor > 0")
		}
	}
	return nil
}

func validateConcatCompatibility(inputs []string, probes map[string]media.ProbeResult) error {
	if len(inputs) < 2 {
		return nil
	}
	base := probes[inputs[0]]
	baseWidth, baseHeight, _ := base.PrimaryVideoDimensions()
	for _, input := range inputs[1:] {
		probe := probes[input]
		w, h, _ := probe.PrimaryVideoDimensions()
		if w != baseWidth || h != baseHeight {
			return fmt.Errorf("concat inputs have mismatched video dimensions: %s is %dx%d, expected %dx%d", input, w, h, baseWidth, baseHeight)
		}
	}
	return nil
}

func findConcat(ops []models.Operation) *models.Operation {
	for i := range ops {
		if ops[i].Type == "concat" {
			return &ops[i]
		}
	}
	return nil
}

func getFloat(params map[string]any, key string) (float64, bool) {
	if params == nil {
		return 0, false
	}
	value, ok := params[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func getInt(params map[string]any, key string) (int, bool) {
	if params == nil {
		return 0, false
	}
	value, ok := params[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func getString(params map[string]any, key string) (string, bool) {
	if params == nil {
		return "", false
	}
	value, ok := params[key]
	if !ok {
		return "", false
	}
	s, ok := value.(string)
	return s, ok
}
