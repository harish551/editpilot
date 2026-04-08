package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/harish551/editpilot/internal/ai"
	"github.com/harish551/editpilot/internal/config"
	"github.com/harish551/editpilot/internal/ffmpeg"
	"github.com/harish551/editpilot/internal/runner"
	"github.com/harish551/editpilot/internal/validator"
	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	var inputs []string
	var prompt string
	var output string
	var dryRun bool
	var savePlan string
	var saveCommand string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Plan, validate, and render in a single command",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(inputs) == 0 {
				return fmt.Errorf("at least one --input is required")
			}
			if prompt == "" {
				return fmt.Errorf("--prompt is required")
			}
			if output == "" {
				output = "output.mp4"
			}

			cfg := config.Load()
			plannerSvc := ai.NewService(cfg)
			plan, err := plannerSvc.BuildPlan(context.Background(), inputs, prompt, output)
			if err != nil {
				return err
			}
			if err := validator.ValidatePlan(plan); err != nil {
				return err
			}

			encoded, err := json.MarshalIndent(plan, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(encoded))
			if savePlan != "" {
				if err := os.WriteFile(savePlan, encoded, 0o644); err != nil {
					return err
				}
			}

			argsOut, err := ffmpeg.BuildArgs(plan)
			if err != nil {
				return err
			}
			commandLine := "ffmpeg " + strings.Join(argsOut, " ")
			fmt.Println(commandLine)
			if saveCommand != "" {
				if err := os.WriteFile(saveCommand, []byte(commandLine+"\n"), 0o644); err != nil {
					return err
				}
			}

			if dryRun {
				return nil
			}
			return runner.RunFFmpeg(argsOut)
		},
	}

	cmd.Flags().StringSliceVar(&inputs, "input", nil, "Input media files (repeat flag for multiple files)")
	cmd.Flags().StringVar(&prompt, "prompt", "", "Natural language edit request")
	cmd.Flags().StringVar(&output, "output", "output.mp4", "Output media path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print plan and ffmpeg args without running")
	cmd.Flags().StringVar(&savePlan, "save-plan", "", "Optional file path to save generated plan JSON")
	cmd.Flags().StringVar(&saveCommand, "save-command", "", "Optional file path to save rendered ffmpeg command")
	return cmd
}
