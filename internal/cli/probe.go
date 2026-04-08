package cli

import (
	"encoding/json"
	"fmt"

	"github.com/harish551/editpilot/internal/media"
	"github.com/spf13/cobra"
)

func newProbeCmd() *cobra.Command {
	var input string

	cmd := &cobra.Command{
		Use:   "probe",
		Short: "Inspect media metadata using ffprobe",
		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" {
				return fmt.Errorf("--input is required")
			}
			result, err := media.Probe(input)
			if err != nil {
				return err
			}
			encoded, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(encoded))
			return nil
		},
	}

	cmd.Flags().StringVar(&input, "input", "", "Path to input media file")
	return cmd
}
