package media

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

type ProbeStream struct {
	Index     int    `json:"index"`
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

type ProbeFormat struct {
	Filename   string `json:"filename"`
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	Size       string `json:"size,omitempty"`
}

type ProbeResult struct {
	Format  ProbeFormat   `json:"format"`
	Streams []ProbeStream `json:"streams"`
}

func Probe(path string) (ProbeResult, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return ProbeResult{}, fmt.Errorf("ffprobe failed: %w", err)
	}

	var result ProbeResult
	if err := json.Unmarshal(out, &result); err != nil {
		return ProbeResult{}, fmt.Errorf("decode ffprobe output: %w", err)
	}
	return result, nil
}

func (r ProbeResult) DurationSeconds() (float64, error) {
	if r.Format.Duration == "" {
		return 0, fmt.Errorf("duration missing from ffprobe result")
	}
	value, err := strconv.ParseFloat(r.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration: %w", err)
	}
	return value, nil
}

func (r ProbeResult) HasVideoStream() bool {
	for _, stream := range r.Streams {
		if stream.CodecType == "video" {
			return true
		}
	}
	return false
}

func (r ProbeResult) PrimaryVideoDimensions() (int, int, bool) {
	for _, stream := range r.Streams {
		if stream.CodecType == "video" {
			return stream.Width, stream.Height, true
		}
	}
	return 0, 0, false
}

func (r ProbeResult) HasAudioStream() bool {
	for _, stream := range r.Streams {
		if stream.CodecType == "audio" {
			return true
		}
	}
	return false
}
