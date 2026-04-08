package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/harish551/editpilot/internal/ffmpeg"
	"github.com/harish551/editpilot/internal/models"
	"github.com/harish551/editpilot/internal/runner"
	"github.com/harish551/editpilot/internal/validator"
	"github.com/spf13/cobra"
)

func newRenderCmd() *cobra.Command {
	var planPath string
	var dryRun bool
	var saveCommand string

	cmd := &cobra.Command{
		Use:   "render",
		Short: "Render output from a plan JSON file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if planPath == "" {
				return fmt.Errorf("--plan is required")
			}
			data, err := os.ReadFile(planPath)
			if err != nil {
				return err
			}
			var plan models.Plan
			if err := json.Unmarshal(data, &plan); err != nil {
				return err
			}
			if err := validator.ValidatePlan(plan); err != nil {
				return err
			}
			ffmpegArgs, err := ffmpeg.BuildArgs(plan)
			if err != nil {
				return err
			}
			commandLine := "ffmpeg " + strings.Join(ffmpegArgs, " ")
			fmt.Println(commandLine)
			if saveCommand != "" {
				if err := os.WriteFile(saveCommand, []byte(commandLine+"\n"), 0o644); err != nil {
					return err
				}
			}
			if dryRun {
				return nil
			}
			return runner.RunFFmpeg(ffmpegArgs)
		},
	}

	cmd.Flags().StringVar(&planPath, "plan", "", "Path to plan JSON")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print ffmpeg args without running")
	cmd.Flags().StringVar(&saveCommand, "save-command", "", "Optional file path to save rendered ffmpeg command")
	return cmd
}
