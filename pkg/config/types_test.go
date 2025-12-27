package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRGBColorUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantR   uint8
		wantG   uint8
		wantB   uint8
		wantErr bool
	}{
		{
			name:  "RGB format",
			yaml:  `{rgb: {r: 255, g: 128, b: 64}}`,
			wantR: 255,
			wantG: 128,
			wantB: 64,
		},
		{
			name:  "HSV red (h=0)",
			yaml:  `{hsv: {h: 0, s: 255, v: 255}}`,
			wantR: 255,
			wantG: 0,
			wantB: 0,
		},
		{
			name:  "HSV green (h≈85)",
			yaml:  `{hsv: {h: 85, s: 255, v: 255}}`,
			wantR: 0,
			wantG: 255,
			wantB: 0,
		},
		{
			name:  "HSV blue (h≈170)",
			yaml:  `{hsv: {h: 170, s: 255, v: 255}}`,
			wantR: 0,
			wantG: 0,
			wantB: 255,
		},
		{
			name:  "HSV white (s=0)",
			yaml:  `{hsv: {h: 0, s: 0, v: 255}}`,
			wantR: 255,
			wantG: 255,
			wantB: 255,
		},
		{
			name:  "HSV black (v=0)",
			yaml:  `{hsv: {h: 0, s: 255, v: 0}}`,
			wantR: 0,
			wantG: 0,
			wantB: 0,
		},
		{
			name:  "HSV half brightness",
			yaml:  `{hsv: {h: 0, s: 255, v: 128}}`,
			wantR: 128,
			wantG: 0,
			wantB: 0,
		},
		{
			name:    "mixed RGB and HSV",
			yaml:    `{rgb: {r: 255}, hsv: {h: 0}}`,
			wantErr: true,
		},
		{
			name:    "empty color",
			yaml:    `{}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var color RGBColor
			err := yaml.Unmarshal([]byte(tt.yaml), &color)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if color.R != tt.wantR {
				t.Errorf("R = %d, want %d", color.R, tt.wantR)
			}
			if color.G != tt.wantG {
				t.Errorf("G = %d, want %d", color.G, tt.wantG)
			}
			if color.B != tt.wantB {
				t.Errorf("B = %d, want %d", color.B, tt.wantB)
			}
		})
	}
}

func TestHSV255ToRGB(t *testing.T) {
	tests := []struct {
		name                string
		h, s, v             uint8
		wantR, wantG, wantB uint8
	}{
		{"red", 0, 255, 255, 255, 0, 0},
		{"green", 85, 255, 255, 0, 255, 0},
		{"blue", 170, 255, 255, 0, 0, 255},
		{"white", 0, 0, 255, 255, 255, 255},
		{"black", 0, 255, 0, 0, 0, 0},
		{"gray", 0, 0, 128, 128, 128, 128},
		{"yellow", 42, 255, 255, 255, 255, 0},
		{"cyan", 128, 255, 255, 0, 255, 255},
		{"magenta", 213, 255, 255, 255, 0, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := hsv255ToRGB(tt.h, tt.s, tt.v)

			// Допускаем погрешность ±5 из-за округления
			tolerance := uint8(5)

			if abs8(r, tt.wantR) > tolerance {
				t.Errorf("R = %d, want %d (±%d)", r, tt.wantR, tolerance)
			}
			if abs8(g, tt.wantG) > tolerance {
				t.Errorf("G = %d, want %d (±%d)", g, tt.wantG, tolerance)
			}
			if abs8(b, tt.wantB) > tolerance {
				t.Errorf("B = %d, want %d (±%d)", b, tt.wantB, tolerance)
			}
		})
	}
}

func abs8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}
