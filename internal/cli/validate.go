package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/harish551/editpilot/internal/models"
	"github.com/harish551/editpilot/internal/validator"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	var planPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a plan JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			fmt.Println("plan is valid")
			return nil
		},
	}

	cmd.Flags().StringVar(&planPath, "plan", "", "Path to plan JSON")
	return cmd
}
