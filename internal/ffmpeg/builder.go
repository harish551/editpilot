package ffmpeg

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/harish551/editpilot/internal/models"
)

func BuildArgs(plan models.Plan) ([]string, error) {
	if len(plan.Inputs) == 0 {
		return nil, fmt.Errorf("no inputs provided")
	}
	if len(plan.Operations) == 0 {
		return nil, fmt.Errorf("no operations provided")
	}

	inputByID := make(map[string]models.InputAsset, len(plan.Inputs))
	for _, in := range plan.Inputs {
		inputByID[in.ID] = in
	}

	if len(plan.Operations) == 1 && plan.Operations[0].Type == "noop" {
		return []string{"-i", plan.Inputs[0].Path, "-c", "copy", "-y", plan.Output.Path}, nil
	}

	if concatOp := findOp(plan.Operations, "concat"); concatOp != nil {
		if hasPerInputTransforms(plan.Operations) {
			return buildFilteredConcatArgs(plan, inputByID, plan.Output.Path)
		}
		return buildConcatArgs(*concatOp, inputByID, plan.Output.Path)
	}

	return buildSingleInputArgs(plan, inputByID, plan.Output.Path)
}

func buildSingleInputArgs(plan models.Plan, inputByID map[string]models.InputAsset, output string) ([]string, error) {
	primary := plan.Inputs[0]
	args := []string{}
	videoFilters := make([]string, 0)
	audioFilters := make([]string, 0)
	mute := false

	for _, op := range plan.Operations {
		if op.Input != "" && op.Input != primary.ID {
			continue
		}
		switch op.Type {
		case "trim":
			start, _ := getFloat(op.Params, "start")
			end, _ := getFloat(op.Params, "end")
			args = append(args, "-ss", formatFloat(start), "-to", formatFloat(end))
		case "resize":
			width, _ := getInt(op.Params, "width")
			height, _ := getInt(op.Params, "height")
			videoFilters = append(videoFilters, fmt.Sprintf("scale=%d:%d", width, height))
		case "mute":
			mute = true
		case "text_overlay":
			text := escapeDrawText(getString(op.Params, "text", "EditPilot"))
			x := getString(op.Params, "x", "(w-text_w)/2")
			y := getString(op.Params, "y", "80")
			fontSize := getIntDefault(op.Params, "font_size", 48)
			videoFilters = append(videoFilters, fmt.Sprintf("drawtext=text='%s':x=%s:y=%s:fontsize=%d:fontcolor=white:box=1:boxcolor=black@0.45:boxborderw=12", text, x, y, fontSize))
		case "speed":
			factor, _ := getFloat(op.Params, "factor")
			if factor <= 0 {
				return nil, fmt.Errorf("speed factor must be > 0")
			}
			videoFilters = append(videoFilters, fmt.Sprintf("setpts=%f*PTS", 1.0/factor))
			if !mute {
				audioFilters = append(audioFilters, buildATempoChain(factor)...)
			}
		case "noop", "concat":
			// handled elsewhere
		default:
			return nil, fmt.Errorf("unsupported operation in builder: %s", op.Type)
		}
	}

	args = append(args, "-i", primary.Path)
	if len(videoFilters) > 0 {
		args = append(args, "-vf", strings.Join(videoFilters, ","))
	}
	if len(audioFilters) > 0 && !mute {
		args = append(args, "-af", strings.Join(audioFilters, ","))
	}
	args = append(args, "-c:v", "libx264")
	if mute {
		args = append(args, "-an")
	} else {
		args = append(args, "-c:a", "aac")
	}
	args = append(args, "-y", output)
	return normalizeArgs(args), nil
}

func buildConcatArgs(op models.Operation, inputByID map[string]models.InputAsset, output string) ([]string, error) {
	listFile, err := writeConcatList(op.Inputs, inputByID)
	if err != nil {
		return nil, err
	}
	return []string{"-f", "concat", "-safe", "0", "-i", listFile, "-c", "copy", "-y", output}, nil
}

func buildFilteredConcatArgs(plan models.Plan, inputByID map[string]models.InputAsset, output string) ([]string, error) {
	concatOp := findOp(plan.Operations, "concat")
	if concatOp == nil {
		return nil, fmt.Errorf("concat operation not found")
	}

	args := make([]string, 0, len(concatOp.Inputs)*2)
	filterParts := make([]string, 0)
	videoRefs := make([]string, 0, len(concatOp.Inputs))
	audioRefs := make([]string, 0, len(concatOp.Inputs))
	mute := hasMute(plan.Operations)

	for idx, inputID := range concatOp.Inputs {
		input, ok := inputByID[inputID]
		if !ok {
			return nil, fmt.Errorf("concat input not found: %s", inputID)
		}
		args = append(args, "-i", input.Path)

		vLabel := fmt.Sprintf("v%d", idx)
		aLabel := fmt.Sprintf("a%d", idx)
		videoChain := []string{fmt.Sprintf("[%d:v]", idx)}
		audioChain := []string{fmt.Sprintf("[%d:a]", idx)}

		for _, op := range plan.Operations {
			if op.Input != "" && op.Input != inputID {
				continue
			}
			switch op.Type {
			case "trim":
				start, _ := getFloat(op.Params, "start")
				end, _ := getFloat(op.Params, "end")
				videoChain = append(videoChain, fmt.Sprintf("trim=start=%s:end=%s", formatFloat(start), formatFloat(end)), "setpts=PTS-STARTPTS")
				if !mute {
					audioChain = append(audioChain, fmt.Sprintf("atrim=start=%s:end=%s", formatFloat(start), formatFloat(end)), "asetpts=PTS-STARTPTS")
				}
			case "resize":
				width, _ := getInt(op.Params, "width")
				height, _ := getInt(op.Params, "height")
				videoChain = append(videoChain, fmt.Sprintf("scale=%d:%d", width, height))
			case "text_overlay":
				text := escapeDrawText(getString(op.Params, "text", "EditPilot"))
				x := getString(op.Params, "x", "(w-text_w)/2")
				y := getString(op.Params, "y", "80")
				fontSize := getIntDefault(op.Params, "font_size", 48)
				videoChain = append(videoChain, fmt.Sprintf("drawtext=text='%s':x=%s:y=%s:fontsize=%d:fontcolor=white:box=1:boxcolor=black@0.45:boxborderw=12", text, x, y, fontSize))
			case "speed":
				factor, _ := getFloat(op.Params, "factor")
				if factor <= 0 {
					return nil, fmt.Errorf("speed factor must be > 0")
				}
				videoChain = append(videoChain, fmt.Sprintf("setpts=%f*PTS", 1.0/factor))
				if !mute {
					audioChain = append(audioChain, buildATempoChain(factor)...)
				}
			case "mute", "concat", "noop":
			case "crop":
				return nil, fmt.Errorf("crop not implemented yet")
			default:
				return nil, fmt.Errorf("unsupported operation in filtered concat builder: %s", op.Type)
			}
		}

		filterParts = append(filterParts, strings.Join(videoChain, ",")+fmt.Sprintf("[%s]", vLabel))
		videoRefs = append(videoRefs, fmt.Sprintf("[%s]", vLabel))
		if !mute {
			filterParts = append(filterParts, strings.Join(audioChain, ",")+fmt.Sprintf("[%s]", aLabel))
			audioRefs = append(audioRefs, fmt.Sprintf("[%s]", aLabel))
		}
	}

	concatVideoCount := len(videoRefs)
	if mute {
		filterParts = append(filterParts, fmt.Sprintf("%sconcat=n=%d:v=1:a=0[outv]", strings.Join(videoRefs, ""), concatVideoCount))
	} else {
		filterParts = append(filterParts, fmt.Sprintf("%s%sconcat=n=%d:v=1:a=1[outv][outa]", strings.Join(videoRefs, ""), strings.Join(audioRefs, ""), concatVideoCount))
	}

	args = append(args, "-filter_complex", strings.Join(filterParts, ";"), "-map", "[outv]")
	if mute {
		args = append(args, "-an")
	} else {
		args = append(args, "-map", "[outa]", "-c:a", "aac")
	}
	args = append(args, "-c:v", "libx264", "-y", output)
	return normalizeArgs(args), nil
}

func writeConcatList(ids []string, inputByID map[string]models.InputAsset) (string, error) {
	file, err := os.CreateTemp("", "editpilot-concat-*.txt")
	if err != nil {
		return "", fmt.Errorf("create concat list: %w", err)
	}
	defer file.Close()

	for _, id := range ids {
		input, ok := inputByID[id]
		if !ok {
			return "", fmt.Errorf("concat input not found: %s", id)
		}
		abs, err := filepath.Abs(input.Path)
		if err != nil {
			return "", fmt.Errorf("resolve input path: %w", err)
		}
		if _, err := fmt.Fprintf(file, "file '%s'\n", escapeConcatPath(abs)); err != nil {
			return "", fmt.Errorf("write concat list: %w", err)
		}
	}
	return file.Name(), nil
}

func escapeConcatPath(path string) string {
	return strings.ReplaceAll(path, "'", "'\\''")
}

func escapeDrawText(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ":", "\\:")
	text = strings.ReplaceAll(text, "'", "\\'")
	text = strings.ReplaceAll(text, "%", "\\%")
	return text
}

func buildATempoChain(factor float64) []string {
	remaining := factor
	parts := make([]string, 0)
	for remaining > 2.0 {
		parts = append(parts, "atempo=2.0")
		remaining /= 2.0
	}
	for remaining < 0.5 {
		parts = append(parts, "atempo=0.5")
		remaining /= 0.5
	}
	parts = append(parts, fmt.Sprintf("atempo=%s", formatFloat(remaining)))
	return parts
}

func hasPerInputTransforms(ops []models.Operation) bool {
	for _, op := range ops {
		switch op.Type {
		case "trim", "resize", "mute", "text_overlay", "speed":
			return true
		}
	}
	return false
}

func hasMute(ops []models.Operation) bool {
	for _, op := range ops {
		if op.Type == "mute" {
			return true
		}
	}
	return false
}

func findOp(ops []models.Operation, kind string) *models.Operation {
	for i := range ops {
		if ops[i].Type == kind {
			return &ops[i]
		}
	}
	return nil
}

func normalizeArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		trimmed := strings.TrimSpace(arg)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
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

func getIntDefault(params map[string]any, key string, fallback int) int {
	if value, ok := getInt(params, key); ok {
		return value
	}
	return fallback
}

func getString(params map[string]any, key, fallback string) string {
	if params == nil {
		return fallback
	}
	value, ok := params[key]
	if !ok {
		return fallback
	}
	if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
		return s
	}
	return fallback
}
