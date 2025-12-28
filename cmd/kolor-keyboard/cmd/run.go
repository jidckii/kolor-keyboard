package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jidckii/kolor-keyboard/pkg/app"
	"github.com/spf13/cobra"
)

var (
	configPath string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the keyboard color daemon",
	Long: `Start the main loop that monitors keyboard layout changes
and updates the RGB backlight accordingly.

The config file is searched in the following order:
  1. Path specified with -c/--config flag
  2. ~/.config/kolor-keyboard/config.yaml

Device and firmware are auto-detected if not specified in config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := configPath
		if cfg == "" {
			cfg = findConfig()
		}

		if cfg == "" {
			fmt.Fprintln(os.Stderr, "Config file not found.")
			fmt.Fprintln(os.Stderr, "Expected: ~/.config/kolor-keyboard/config.yaml")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "Tip: Run 'kolor-keyboard discover -g' to detect your keyboard and generate a config")
			return fmt.Errorf("config file not found")
		}

		logger := GetLogger()
		logger.Info("using config", "path", cfg)

		application, err := app.New(cfg, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize app: %w", err)
		}
		defer func() { _ = application.Close() }()

		return application.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file")
}

func findConfig() string {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config/kolor-keyboard/config.yaml")

	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	return ""
}
