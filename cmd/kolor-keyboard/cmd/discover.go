package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jidckii/kolor-keyboard/pkg/discover"
	"github.com/spf13/cobra"
)

var (
	globalConfig bool
	outputDir    string
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover keyboard and generate config",
	Long: `Scan for VIA/Vial compatible keyboards and generate configuration files.

By default, generates config files in the current directory with names:
  - stock_mono.yaml   (for Stock QMK/VIA firmware)
  - vial_mono.yaml    (for Vial firmware, mono color mode)
  - vial_draw.yaml    (for Vial firmware, per-key RGB mode)

With --global flag, generates a single config.yaml in:
  ~/.config/kolor-keyboard/keyboards/<vendor>/<model>/<variant>/config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDiscover()
	},
}

func init() {
	rootCmd.AddCommand(discoverCmd)
	discoverCmd.Flags().BoolVarP(&globalConfig, "global", "g", false, "save to global config directory (~/.config/kolor-keyboard/)")
	discoverCmd.Flags().StringVarP(&outputDir, "output", "o", "", "output directory (default: current directory)")
}

func runDiscover() error {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           kolor-keyboard - Keyboard Discovery                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Find devices
	fmt.Println("Scanning for VIA/Vial compatible keyboards...")
	devices, err := discover.FindVIADevices()
	if err != nil {
		return fmt.Errorf("failed to scan devices: %w", err)
	}

	if len(devices) == 0 {
		fmt.Println()
		fmt.Println("No VIA/Vial keyboards found!")
		fmt.Println()
		fmt.Println("Possible reasons:")
		fmt.Println("  - Keyboard is not connected")
		fmt.Println("  - Keyboard doesn't have VIA/Vial firmware")
		fmt.Println("  - Missing udev rules (run: make install-udev)")
		fmt.Println("  - Need to run as root or add user to input group")
		return nil
	}

	fmt.Printf("\nFound %d device(s):\n\n", len(devices))

	for i, dev := range devices {
		fmt.Printf("  [%d] %s %s\n", i+1, dev.Manufacturer, dev.Product)
		fmt.Printf("      VID: 0x%04X  PID: 0x%04X\n", dev.VendorID, dev.ProductID)
	}

	// Select device
	var selectedDev *discover.DeviceInfo
	if len(devices) == 1 {
		selectedDev = &devices[0]
		fmt.Printf("\nUsing: %s %s\n", selectedDev.Manufacturer, selectedDev.Product)
	} else {
		fmt.Print("\nSelect device [1]: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		idx := 0
		if input != "" {
			fmt.Sscanf(input, "%d", &idx)
			idx--
		}
		if idx < 0 || idx >= len(devices) {
			idx = 0
		}
		selectedDev = &devices[idx]
	}

	// Check Vial support
	fmt.Println("\nChecking Vial support...")
	if err := discover.CheckVialSupport(selectedDev); err != nil {
		return fmt.Errorf("failed to check device: %w", err)
	}

	cfg := &discover.DiscoveredConfig{
		Device: *selectedDev,
	}

	if selectedDev.IsVial {
		fmt.Printf("✓ Vial firmware detected! LED count: %d\n", selectedDev.LEDCount)
		cfg.Firmware = "vial"

		fmt.Print("\nDo you want to map LED rows for per-key RGB (draw mode)? [Y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" || input == "y" || input == "yes" {
			rows, err := discover.RunLEDMappingTour(selectedDev)
			if err != nil {
				return fmt.Errorf("LED mapping failed: %w", err)
			}
			cfg.KeyboardRows = rows

			fmt.Println("\nMapped rows:")
			for i, row := range rows {
				fmt.Printf("  Row %d: %v (%d LEDs)\n", i, row, len(row))
			}
		}
	} else {
		fmt.Println("  Vial not detected, using stock VIA mode")
		cfg.Firmware = "stock"
	}

	// Get keyboard info
	vendor, model, variant := discover.GetKeyboardInfo(selectedDev)
	fmt.Printf("\nKeyboard identified as: %s/%s/%s\n", vendor, model, variant)

	// Determine output path
	var outDir string
	if globalConfig {
		home, _ := os.UserHomeDir()
		outDir = filepath.Join(home, ".config/kolor-keyboard/keyboards", vendor, model, variant)
	} else if outputDir != "" {
		outDir = filepath.Join(outputDir, vendor, model, variant)
	} else {
		outDir = filepath.Join("keyboards", vendor, model, variant)
	}

	// Create directory
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if globalConfig {
		// Single config.yaml for global mode
		return saveGlobalConfig(outDir, cfg)
	}

	// Multiple config files for local mode
	return saveLocalConfigs(outDir, cfg)
}

func saveGlobalConfig(outDir string, cfg *discover.DiscoveredConfig) error {
	outPath := filepath.Join(outDir, "config.yaml")

	// Check if exists
	if _, err := os.Stat(outPath); err == nil {
		fmt.Printf("\nConfig file already exists: %s\n", outPath)
		fmt.Print("Overwrite? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	content := discover.GenerateConfig(cfg)
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("\n✓ Config saved to: %s\n", outPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit the config to customize colors for different layouts")
	fmt.Println("  2. Run: kolor-keyboard run")
	fmt.Println("  3. Or install as service: make install && make enable")

	return nil
}

func saveLocalConfigs(outDir string, cfg *discover.DiscoveredConfig) error {
	var savedFiles []string

	// Always generate stock_mono.yaml
	stockCfg := *cfg
	stockCfg.Firmware = "stock"
	stockCfg.KeyboardRows = nil

	stockPath := filepath.Join(outDir, "stock_mono.yaml")
	stockContent := discover.GenerateConfig(&stockCfg)
	if err := os.WriteFile(stockPath, []byte(stockContent), 0644); err != nil {
		return fmt.Errorf("failed to write stock config: %w", err)
	}
	savedFiles = append(savedFiles, stockPath)

	// Generate vial_mono.yaml
	vialMonoCfg := *cfg
	vialMonoCfg.Firmware = "vial"
	vialMonoCfg.KeyboardRows = nil

	vialMonoPath := filepath.Join(outDir, "vial_mono.yaml")
	vialMonoContent := discover.GenerateConfig(&vialMonoCfg)
	if err := os.WriteFile(vialMonoPath, []byte(vialMonoContent), 0644); err != nil {
		return fmt.Errorf("failed to write vial mono config: %w", err)
	}
	savedFiles = append(savedFiles, vialMonoPath)

	// Generate vial_draw.yaml if we have LED rows
	if len(cfg.KeyboardRows) > 0 {
		vialDrawCfg := *cfg
		vialDrawCfg.Firmware = "vial"

		vialDrawPath := filepath.Join(outDir, "vial_draw.yaml")
		vialDrawContent := discover.GenerateConfig(&vialDrawCfg)
		if err := os.WriteFile(vialDrawPath, []byte(vialDrawContent), 0644); err != nil {
			return fmt.Errorf("failed to write vial draw config: %w", err)
		}
		savedFiles = append(savedFiles, vialDrawPath)
	}

	fmt.Printf("\n✓ Configs saved to: %s/\n", outDir)
	for _, f := range savedFiles {
		fmt.Printf("  - %s\n", filepath.Base(f))
	}

	fmt.Println("\nNext steps:")
	fmt.Println("  1. Copy desired config to ~/.config/kolor-keyboard/config.yaml")
	fmt.Printf("     cp %s ~/.config/kolor-keyboard/config.yaml\n", savedFiles[0])
	fmt.Println("  2. Or run with specific config:")
	fmt.Printf("     kolor-keyboard run -c %s\n", savedFiles[0])

	return nil
}
