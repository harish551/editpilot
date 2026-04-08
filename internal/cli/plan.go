package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/harish551/editpilot/internal/ai"
	"github.com/harish551/editpilot/internal/config"
	"github.com/spf13/cobra"
)

func newPlanCmd() *cobra.Command {
	var inputs []string
	var prompt string
	var output string

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Generate a structured edit plan from prompt + inputs",
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
			encoded, err := json.MarshalIndent(plan, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(encoded))
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&inputs, "input", nil, "Input media files (repeat flag for multiple files)")
	cmd.Flags().StringVar(&prompt, "prompt", "", "Natural language edit request")
	cmd.Flags().StringVar(&output, "output", "output.mp4", "Output media path")
	return cmd
}
