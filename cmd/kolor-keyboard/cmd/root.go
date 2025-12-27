package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version info (set at build time)
	Version = "dev"
	Commit  = "none"

	// Global flags
	debug  bool
	logger *slog.Logger
)

var rootCmd = &cobra.Command{
	Use:   "kolor-keyboard",
	Short: "Change keyboard RGB color based on keyboard layout",
	Long: `kolor-keyboard is a tool for controlling RGB backlight
on VIA/Vial compatible keyboards based on the current keyboard layout.

Supports:
  - Stock QMK/VIA firmware (mono color mode)
  - Vial firmware (mono and per-key RGB "draw" mode)

Examples:
  kolor-keyboard run -c config.yaml
  kolor-keyboard discover
  kolor-keyboard discover --global`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logLevel := slog.LevelInfo
		if debug {
			logLevel = slog.LevelDebug
		}
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		}))
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug logging")
}

// GetLogger returns the configured logger
func GetLogger() *slog.Logger {
	return logger
}
