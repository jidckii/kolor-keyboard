package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Создаём временный конфиг
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")

	configContent := `
device:
  vendor_id: 0x3434
  product_id: 0x0331
  usage_page: 0xFF60
  usage: 0x61

firmware: vial
mode: mono

brightness: 200
speed: 64

colors:
  - layout: ru
    color: {rgb: {r: 255, g: 0, b: 0}}
  - layout: us
    color: {hsv: {h: 170, s: 255, v: 255}}
  - layout: "*"
    color: {rgb: {r: 255, g: 255, b: 255}}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Проверяем device
	if cfg.Device.VendorID != 0x3434 {
		t.Errorf("VendorID = %x, want 0x3434", cfg.Device.VendorID)
	}
	if cfg.Device.ProductID != 0x0331 {
		t.Errorf("ProductID = %x, want 0x0331", cfg.Device.ProductID)
	}

	// Проверяем firmware и mode
	if cfg.Firmware != FirmwareVial {
		t.Errorf("Firmware = %s, want vial", cfg.Firmware)
	}
	if cfg.Mode != ModeMono {
		t.Errorf("Mode = %s, want mono", cfg.Mode)
	}

	// Проверяем brightness и speed
	if cfg.Brightness == nil || *cfg.Brightness != 200 {
		t.Errorf("Brightness = %v, want 200", cfg.Brightness)
	}
	if cfg.Speed == nil || *cfg.Speed != 64 {
		t.Errorf("Speed = %v, want 64", cfg.Speed)
	}

	// Проверяем colors
	if len(cfg.Colors) != 3 {
		t.Fatalf("len(Colors) = %d, want 3", len(cfg.Colors))
	}

	// RGB цвет
	ruColor := cfg.GetColorForLayout("ru")
	if ruColor == nil {
		t.Fatal("GetColorForLayout(ru) returned nil")
	}
	if ruColor.R != 255 || ruColor.G != 0 || ruColor.B != 0 {
		t.Errorf("ru color = {%d,%d,%d}, want {255,0,0}", ruColor.R, ruColor.G, ruColor.B)
	}

	// HSV цвет (h=170 ≈ синий)
	usColor := cfg.GetColorForLayout("us")
	if usColor == nil {
		t.Fatal("GetColorForLayout(us) returned nil")
	}
	// HSV h=170 (синий), s=255, v=255 → RGB примерно (0, 85, 255)
	if usColor.B < 200 {
		t.Errorf("us color B = %d, want > 200 (blue from HSV)", usColor.B)
	}
}

func TestLoadDrawMode(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "draw.yaml")

	configContent := `
device:
  vendor_id: 0x3434
  product_id: 0x0331
  usage_page: 0xFF60
  usage: 0x61

firmware: vial
mode: draw

keyboard:
  rows:
    - [0, 1, 2, 3]
    - [4, 5, 6, 7]

draw:
  - layout: ru
    stripes:
      - rows: [0]
        color: {rgb: {r: 255, g: 255, b: 255}}
      - rows: [1]
        color: {rgb: {r: 0, g: 0, b: 255}}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Mode != ModeDraw {
		t.Errorf("Mode = %s, want draw", cfg.Mode)
	}

	if len(cfg.Keyboard.Rows) != 2 {
		t.Errorf("len(Rows) = %d, want 2", len(cfg.Keyboard.Rows))
	}

	flag := cfg.GetFlagForLayout("ru")
	if flag == nil {
		t.Fatal("GetFlagForLayout(ru) returned nil")
	}
	if len(flag.Stripes) != 2 {
		t.Errorf("len(Stripes) = %d, want 2", len(flag.Stripes))
	}
}

func TestGetSpeed(t *testing.T) {
	cfg := &Config{}

	// Без указания speed - должно быть 128
	if cfg.GetSpeed() != 128 {
		t.Errorf("GetSpeed() = %d, want 128 (default)", cfg.GetSpeed())
	}

	// С указанием speed
	speed := uint8(64)
	cfg.Speed = &speed
	if cfg.GetSpeed() != 64 {
		t.Errorf("GetSpeed() = %d, want 64", cfg.GetSpeed())
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{
			name: "valid mono config",
			config: `
device:
  vendor_id: 0x1234
  product_id: 0x5678
  usage_page: 0xFF60
  usage: 0x61
firmware: vial
mode: mono
colors:
  - layout: "*"
    color: {rgb: {r: 255, g: 0, b: 0}}
`,
			wantErr: false,
		},
		{
			name: "missing device",
			config: `
firmware: vial
mode: mono
colors:
  - layout: "*"
    color: {rgb: {r: 255, g: 0, b: 0}}
`,
			wantErr: true,
		},
		{
			name: "draw mode with stock firmware",
			config: `
device:
  vendor_id: 0x1234
  product_id: 0x5678
  usage_page: 0xFF60
  usage: 0x61
firmware: stock
mode: draw
keyboard:
  rows: [[0,1,2]]
draw:
  - layout: "*"
    stripes:
      - rows: [0]
        color: {rgb: {r: 255, g: 0, b: 0}}
`,
			wantErr: true, // draw требует vial
		},
		{
			name: "invalid row reference",
			config: `
device:
  vendor_id: 0x1234
  product_id: 0x5678
  usage_page: 0xFF60
  usage: 0x61
firmware: vial
mode: draw
keyboard:
  rows: [[0,1,2]]
draw:
  - layout: "*"
    stripes:
      - rows: [5]
        color: {rgb: {r: 255, g: 0, b: 0}}
`,
			wantErr: true, // ряд 5 не существует
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "test.yaml")
			if err := os.WriteFile(configPath, []byte(tt.config), 0644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			_, err := Load(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
