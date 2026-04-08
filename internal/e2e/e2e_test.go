package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditPilotRunDryRun(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not available")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not available")
	}

	input := makeFixtureVideo(t, "clip.mp4", "320x240", 2)
	planPath := filepath.Join(t.TempDir(), "plan.json")
	cmdPath := filepath.Join(t.TempDir(), "command.txt")

	cmd := exec.Command(goBin,
		"run", "./cmd/editpilot",
		"run",
		"--input", input,
		"--prompt", `trim first 1 seconds and add title "Demo"`,
		"--output", filepath.Join(t.TempDir(), "out.mp4"),
		"--dry-run",
		"--save-plan", planPath,
		"--save-command", cmdPath,
	)
	cmd.Dir = filepath.Join("..", "..")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("editpilot dry run failed: %v\n%s", err, string(out))
	}
	if _, err := os.Stat(planPath); err != nil {
		t.Fatalf("expected plan file: %v", err)
	}
	data, err := os.ReadFile(cmdPath)
	if err != nil {
		t.Fatalf("expected command file: %v", err)
	}
	if !strings.Contains(string(data), "ffmpeg ") {
		t.Fatalf("expected ffmpeg command, got %s", string(data))
	}
}

func makeFixtureVideo(t *testing.T, name, size string, duration int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", fmt.Sprintf("color=c=red:s=%s:d=%d", size, duration),
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
	return path
}
