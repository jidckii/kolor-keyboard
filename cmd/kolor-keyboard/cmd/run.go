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
  2. ./kolor-keyboard.yaml (current directory)
  3. ~/.config/kolor-keyboard/config.yaml
  4. Auto-generated configs in ~/.config/kolor-keyboard/keyboards/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := configPath
		if cfg == "" {
			cfg = findConfig()
		}

		if cfg == "" {
			fmt.Fprintln(os.Stderr, "Config file not found. Searched locations:")
			fmt.Fprintln(os.Stderr, "  - ./kolor-keyboard.yaml")
			fmt.Fprintln(os.Stderr, "  - ~/.config/kolor-keyboard/config.yaml")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "Tip: Run 'kolor-keyboard discover' to detect your keyboard and generate a config")
			return fmt.Errorf("config file not found")
		}

		logger := GetLogger()
		logger.Info("using config", "path", cfg)

		application, err := app.New(cfg, logger)
		if err != nil {
			return fmt.Errorf("failed to initialize app: %w", err)
		}
		defer application.Close()

		return application.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file")
}

func findConfig() string {
	home, _ := os.UserHomeDir()

	// Explicit paths first
	paths := []string{
		"kolor-keyboard.yaml",
		filepath.Join(home, ".config/kolor-keyboard/config.yaml"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Search in keyboards directory (auto-generated configs)
	keyboardsDir := filepath.Join(home, ".config/kolor-keyboard/keyboards")
	if entries, err := os.ReadDir(keyboardsDir); err == nil {
		for _, vendor := range entries {
			if !vendor.IsDir() {
				continue
			}
			vendorPath := filepath.Join(keyboardsDir, vendor.Name())
			if models, err := os.ReadDir(vendorPath); err == nil {
				for _, model := range models {
					if !model.IsDir() {
						continue
					}
					modelPath := filepath.Join(vendorPath, model.Name())
					if variants, err := os.ReadDir(modelPath); err == nil {
						for _, variant := range variants {
							if !variant.IsDir() {
								continue
							}
							configPath := filepath.Join(modelPath, variant.Name(), "config.yaml")
							if _, err := os.Stat(configPath); err == nil {
								return configPath
							}
						}
					}
				}
			}
		}
	}

	return ""
}
