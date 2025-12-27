package discover

import (
	"strings"
	"testing"
)

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *DiscoveredConfig
		wantContains []string
	}{
		{
			name: "vial with rows",
			cfg: &DiscoveredConfig{
				Device: DeviceInfo{
					VendorID:     0x3434,
					ProductID:    0x0331,
					Manufacturer: "Keychron",
					Product:      "V3",
					UsagePage:    0xFF60,
					Usage:        0x61,
				},
				Firmware:     "vial",
				KeyboardRows: [][]int{{0, 1, 2}, {3, 4, 5}},
			},
			wantContains: []string{
				"vendor_id: 0x3434",
				"product_id: 0x0331",
				"usage_page: 0xFF60",
				"usage: 0x61",
				"firmware: vial",
				"mode: draw",
				"keyboard:",
				"rows:",
				"- [0 1 2]",
				"- [3 4 5]",
				"draw:",
				"stripes:",
			},
		},
		{
			name: "vial without rows (mono mode)",
			cfg: &DiscoveredConfig{
				Device: DeviceInfo{
					VendorID:     0x3434,
					ProductID:    0x0331,
					Manufacturer: "Keychron",
					Product:      "V3",
					UsagePage:    0xFF60,
					Usage:        0x61,
				},
				Firmware:     "vial",
				KeyboardRows: nil,
			},
			wantContains: []string{
				"firmware: vial",
				"mode: mono",
				"colors:",
			},
		},
		{
			name: "stock firmware",
			cfg: &DiscoveredConfig{
				Device: DeviceInfo{
					VendorID:     0x1234,
					ProductID:    0x5678,
					Manufacturer: "Generic",
					Product:      "Keyboard",
					UsagePage:    0xFF60,
					Usage:        0x61,
				},
				Firmware: "stock",
			},
			wantContains: []string{
				"vendor_id: 0x1234",
				"product_id: 0x5678",
				"firmware: stock",
				"mode: mono",
				"colors:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateConfig(tt.cfg)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("GenerateConfig() missing %q\nGot:\n%s", want, got)
				}
			}
		})
	}
}

func TestGenerateConfigVialWithRows(t *testing.T) {
	cfg := &DiscoveredConfig{
		Device: DeviceInfo{
			VendorID:     0x3434,
			ProductID:    0x0331,
			Manufacturer: "Keychron",
			Product:      "V3",
			UsagePage:    0xFF60,
			Usage:        0x61,
		},
		Firmware:     "vial",
		KeyboardRows: [][]int{{0, 1, 2, 3}, {4, 5, 6, 7}, {8, 9, 10, 11}},
	}

	config := GenerateConfig(cfg)

	// Check header
	if !strings.Contains(config, "kolor-keyboard configuration") {
		t.Error("Missing header comment")
	}

	// Check device section
	if !strings.Contains(config, "device:") {
		t.Error("Missing device section")
	}

	// Check rows are present
	if !strings.Contains(config, "Row 0 (4 LEDs)") {
		t.Error("Missing row 0 comment")
	}
	if !strings.Contains(config, "Row 1 (4 LEDs)") {
		t.Error("Missing row 1 comment")
	}
	if !strings.Contains(config, "Row 2 (4 LEDs)") {
		t.Error("Missing row 2 comment")
	}

	// Check draw section
	if !strings.Contains(config, "draw:") {
		t.Error("Missing draw section")
	}

	// Check all rows index
	if !strings.Contains(config, "- rows: [0 1 2]") {
		t.Error("Missing rows index")
	}

	// Check default color
	if !strings.Contains(config, "color: {rgb: {r: 0, g: 255, b: 0}}") {
		t.Error("Missing default color")
	}
}

func TestGenerateConfigStockFirmware(t *testing.T) {
	cfg := &DiscoveredConfig{
		Device: DeviceInfo{
			VendorID:     0x1234,
			ProductID:    0x5678,
			Manufacturer: "TestVendor",
			Product:      "TestKeyboard",
			UsagePage:    0xFF60,
			Usage:        0x61,
		},
		Firmware: "stock",
	}

	config := GenerateConfig(cfg)

	// Should be mono mode
	if !strings.Contains(config, "mode: mono") {
		t.Error("Stock firmware should use mono mode")
	}

	// Should have colors section, not draw
	if !strings.Contains(config, "colors:") {
		t.Error("Mono mode should have colors section")
	}

	// Should NOT have draw or keyboard sections
	if strings.Contains(config, "draw:") {
		t.Error("Mono mode should not have draw section")
	}
	if strings.Contains(config, "keyboard:") {
		t.Error("Mono mode should not have keyboard section")
	}
}
