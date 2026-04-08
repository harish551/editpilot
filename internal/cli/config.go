package cli

import (
	"fmt"
	"strings"

	"github.com/harish551/editpilot/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage EditPilot configuration",
	}

	cmd.AddCommand(
		newConfigShowCmd(),
		newConfigInitCmd(),
		newConfigSetCmd(),
	)

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current config values",
		RunE: func(cmd *cobra.Command, args []string) error {
			if path == "" {
				path = config.DefaultEnvFilePath()
			}
			cfg := config.Load()
			fmt.Printf("Config file: %s\n", path)
			fmt.Printf("AI Provider: %s\n", cfg.AIProvider)
			fmt.Printf("AI Model: %s\n", cfg.AIModel)
			if strings.TrimSpace(cfg.APIKey) == "" {
				fmt.Println("AI API Key: (not set)")
			} else {
				fmt.Println("AI API Key: (set)")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Optional config env file path")
	return cmd
}

func newConfigInitCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize config env file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if path == "" {
				path = config.DefaultEnvFilePath()
			}
			if err := config.InitEnvFile(path); err != nil {
				return err
			}
			fmt.Printf("Initialized config at %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Optional config env file path")
	return cmd
}

func newConfigSetCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if path == "" {
				path = config.DefaultEnvFilePath()
			}
			_ = config.InitEnvFile(path)
			cfg, err := config.LoadFromEnvFile(path)
			if err != nil {
				return err
			}
			key := strings.ToLower(strings.TrimSpace(args[0]))
			value := args[1]
			switch key {
			case "provider":
				cfg.AIProvider = value
			case "model":
				cfg.AIModel = value
			case "api-key", "apikey", "key":
				cfg.APIKey = value
			default:
				return fmt.Errorf("unsupported config key: %s", args[0])
			}
			if err := config.SaveToEnvFile(path, cfg); err != nil {
				return err
			}
			fmt.Printf("Updated %s in %s\n", key, path)
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Optional config env file path")
	return cmd
}
