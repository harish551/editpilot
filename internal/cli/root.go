package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "editpilot",
		Short: "EditPilot is an AI-assisted FFmpeg video editing CLI",
	}

	cmd.AddCommand(
		newProbeCmd(),
		newPlanCmd(),
		newValidateCmd(),
		newRenderCmd(),
		newRunCmd(),
		newConfigCmd(),
	)

	return cmd
}
