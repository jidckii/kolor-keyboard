package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const serviceName = "kolor-keyboard.service"

// serviceTemplate - шаблон systemd user service
const serviceTemplate = `[Unit]
Description=Kolor Keyboard - RGB backlight based on keyboard layout
Documentation=https://github.com/jidckii/kolor-keyboard
After=graphical-session.target dbus.service
Wants=dbus.service

[Service]
Type=simple
ExecStart=%s run
Restart=on-failure
RestartSec=5

# D-Bus session environment
Environment=DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/%%U/bus

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=read-only
PrivateTmp=true

[Install]
WantedBy=default.target
`

var serviceCmd = &cobra.Command{
	Use:     "service",
	Aliases: []string{"svc"},
	Short:   "Manage kolor-keyboard systemd user service",
	Long: `Manage kolor-keyboard as a systemd user service (no sudo required).

The service runs in user mode and starts automatically on login.`,
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install and enable the user service",
	Long: `Create systemd user service file and enable it.

The service file is created at:
  ~/.config/systemd/user/kolor-keyboard.service

After installation, the service will start automatically on login.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return installService()
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Stop, disable and remove the user service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return uninstallService()
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return systemctl("start", serviceName)
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return systemctl("stop", serviceName)
	},
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return systemctl("restart", serviceName)
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return systemctl("status", serviceName)
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceInstallCmd)
	serviceCmd.AddCommand(serviceUninstallCmd)
	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceStopCmd)
	serviceCmd.AddCommand(serviceRestartCmd)
	serviceCmd.AddCommand(serviceStatusCmd)
}

func getServicePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config/systemd/user", serviceName)
}

func getExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	// Resolve symlinks
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}
	return exe, nil
}

func installService() error {
	servicePath := getServicePath()
	serviceDir := filepath.Dir(servicePath)

	// Get executable path
	exePath, err := getExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Check if config exists
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config/kolor-keyboard/config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("Warning: Config file not found at", configPath)
		fmt.Println("Run 'kolor-keyboard discover -g' first to generate config.")
		fmt.Println()
	}

	// Create directory
	if err := os.MkdirAll(serviceDir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate service content
	serviceContent := fmt.Sprintf(serviceTemplate, exePath)

	// Write service file
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}
	fmt.Printf("Created: %s\n", servicePath)

	// Reload systemd
	if err := systemctlQuiet("daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}
	fmt.Println("Reloaded systemd user daemon")

	// Enable service
	if err := systemctlQuiet("enable", serviceName); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}
	fmt.Println("Enabled service for autostart")

	// Start service
	if err := systemctlQuiet("start", serviceName); err != nil {
		fmt.Printf("Warning: failed to start service: %v\n", err)
		fmt.Println("You can start it manually with: kolor-keyboard service start")
	} else {
		fmt.Println("Started service")
	}

	fmt.Println()
	fmt.Println("Service installed successfully!")
	fmt.Println("Check status with: kolor-keyboard service status")

	return nil
}

func uninstallService() error {
	servicePath := getServicePath()

	// Check if service exists
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		fmt.Println("Service is not installed")
		return nil
	}

	// Stop service (ignore errors)
	_ = systemctlQuiet("stop", serviceName)
	fmt.Println("Stopped service")

	// Disable service (ignore errors)
	_ = systemctlQuiet("disable", serviceName)
	fmt.Println("Disabled service")

	// Remove service file
	if err := os.Remove(servicePath); err != nil {
		return fmt.Errorf("failed to remove service file: %w", err)
	}
	fmt.Printf("Removed: %s\n", servicePath)

	// Reload systemd
	if err := systemctlQuiet("daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}
	fmt.Println("Reloaded systemd user daemon")

	fmt.Println()
	fmt.Println("Service uninstalled successfully!")

	return nil
}

func systemctl(args ...string) error {
	allArgs := append([]string{"--user"}, args...)
	cmd := exec.Command("systemctl", allArgs...) //nolint:gosec // systemctl is safe
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func systemctlQuiet(args ...string) error {
	allArgs := append([]string{"--user"}, args...)
	cmd := exec.Command("systemctl", allArgs...) //nolint:gosec // systemctl is safe
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
