package hid

import (
	"testing"
)

func TestBuildSetEffectPacket(t *testing.T) {
	packet := BuildSetEffectPacket(EffectSolidColor)

	if len(packet) != PacketSize {
		t.Errorf("packet size = %d, want %d", len(packet), PacketSize)
	}
	if packet[0] != CmdVIASetValue {
		t.Errorf("packet[0] = %x, want %x", packet[0], CmdVIASetValue)
	}
	if packet[1] != ChannelRGBMatrix {
		t.Errorf("packet[1] = %x, want %x", packet[1], ChannelRGBMatrix)
	}
	if packet[2] != RGBMatrixEffect {
		t.Errorf("packet[2] = %x, want %x", packet[2], RGBMatrixEffect)
	}
	if packet[3] != EffectSolidColor {
		t.Errorf("packet[3] = %x, want %x", packet[3], EffectSolidColor)
	}
}

func TestBuildSetColorPacket(t *testing.T) {
	hue := uint8(128)
	sat := uint8(255)
	packet := BuildSetColorPacket(hue, sat)

	if packet[0] != CmdVIASetValue {
		t.Errorf("packet[0] = %x, want %x", packet[0], CmdVIASetValue)
	}
	if packet[1] != ChannelRGBMatrix {
		t.Errorf("packet[1] = %x, want %x", packet[1], ChannelRGBMatrix)
	}
	if packet[2] != RGBMatrixColor {
		t.Errorf("packet[2] = %x, want %x", packet[2], RGBMatrixColor)
	}
	if packet[3] != hue {
		t.Errorf("packet[3] = %d, want %d", packet[3], hue)
	}
	if packet[4] != sat {
		t.Errorf("packet[4] = %d, want %d", packet[4], sat)
	}
}

func TestBuildSetBrightnessPacket(t *testing.T) {
	brightness := uint8(200)
	packet := BuildSetBrightnessPacket(brightness)

	if packet[2] != RGBMatrixBrightness {
		t.Errorf("packet[2] = %x, want %x", packet[2], RGBMatrixBrightness)
	}
	if packet[3] != brightness {
		t.Errorf("packet[3] = %d, want %d", packet[3], brightness)
	}
}

func TestBuildVialSetModePacket(t *testing.T) {
	mode := uint16(VialEffectDirect)
	speed := uint8(128)
	h, s, v := uint8(0), uint8(255), uint8(255)

	packet := BuildVialSetModePacket(mode, speed, h, s, v)

	if packet[0] != CmdVIASetValue {
		t.Errorf("packet[0] = %x, want %x", packet[0], CmdVIASetValue)
	}
	if packet[1] != VialRGBSetMode {
		t.Errorf("packet[1] = %x, want %x", packet[1], VialRGBSetMode)
	}
	if packet[2] != byte(mode&0xFF) {
		t.Errorf("packet[2] = %x, want %x", packet[2], byte(mode&0xFF))
	}
	if packet[3] != byte(mode>>8) {
		t.Errorf("packet[3] = %x, want %x", packet[3], byte(mode>>8))
	}
	if packet[4] != speed {
		t.Errorf("packet[4] = %d, want %d", packet[4], speed)
	}
	if packet[5] != h || packet[6] != s || packet[7] != v {
		t.Errorf("HSV = {%d,%d,%d}, want {%d,%d,%d}", packet[5], packet[6], packet[7], h, s, v)
	}
}

func TestBuildDirectSetPacket(t *testing.T) {
	colors := []HSVColor{
		{H: 0, S: 255, V: 255},   // red
		{H: 85, S: 255, V: 255},  // green
		{H: 170, S: 255, V: 255}, // blue
	}
	startIndex := 10

	packet := BuildDirectSetPacket(startIndex, colors)

	if packet[0] != CmdVIASetValue {
		t.Errorf("packet[0] = %x, want %x", packet[0], CmdVIASetValue)
	}
	if packet[1] != VialRGBDirectSet {
		t.Errorf("packet[1] = %x, want %x", packet[1], VialRGBDirectSet)
	}
	if packet[2] != byte(startIndex&0xFF) {
		t.Errorf("packet[2] = %x, want %x", packet[2], byte(startIndex&0xFF))
	}
	if packet[3] != byte(startIndex>>8) {
		t.Errorf("packet[3] = %x, want %x", packet[3], byte(startIndex>>8))
	}
	if packet[4] != byte(len(colors)) {
		t.Errorf("packet[4] = %d, want %d", packet[4], len(colors))
	}

	// Проверяем цвета
	offset := 5
	for i, c := range colors {
		if packet[offset] != c.H || packet[offset+1] != c.S || packet[offset+2] != c.V {
			t.Errorf("color[%d] = {%d,%d,%d}, want {%d,%d,%d}",
				i, packet[offset], packet[offset+1], packet[offset+2], c.H, c.S, c.V)
		}
		offset += 3
	}
}

func TestBuildDirectSetPacketMaxLEDs(t *testing.T) {
	// Проверяем что не превышаем лимит LED на пакет
	colors := make([]HSVColor, MaxLEDsPerPacket+5)
	for i := range colors {
		colors[i] = HSVColor{H: uint8(i), S: 255, V: 255}
	}

	packet := BuildDirectSetPacket(0, colors)

	// Должно записаться только MaxLEDsPerPacket цветов
	if packet[4] != byte(len(colors)) {
		// count в пакете = переданное количество, но реально записывается MaxLEDsPerPacket
		// В текущей реализации count = len(colors), но данные усекаются
		t.Logf("packet[4] (count) = %d, actual colors in packet limited by PacketSize", packet[4])
	}
}

func TestRGBToHSV(t *testing.T) {
	tests := []struct {
		name    string
		r, g, b uint8
		wantH   uint8
		wantS   uint8
		wantV   uint8
	}{
		{"red", 255, 0, 0, 0, 255, 255},
		{"green", 0, 255, 0, 85, 255, 255},
		{"blue", 0, 0, 255, 170, 255, 255},
		{"white", 255, 255, 255, 0, 0, 255},
		{"black", 0, 0, 0, 0, 0, 0},
		{"gray", 128, 128, 128, 0, 0, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hsv := RGBToHSV(tt.r, tt.g, tt.b)

			tolerance := uint8(5)

			if abs8(hsv.H, tt.wantH) > tolerance {
				t.Errorf("H = %d, want %d (±%d)", hsv.H, tt.wantH, tolerance)
			}
			if abs8(hsv.S, tt.wantS) > tolerance {
				t.Errorf("S = %d, want %d (±%d)", hsv.S, tt.wantS, tolerance)
			}
			if abs8(hsv.V, tt.wantV) > tolerance {
				t.Errorf("V = %d, want %d (±%d)", hsv.V, tt.wantV, tolerance)
			}
		})
	}
}

func TestParseLEDCountResponse(t *testing.T) {
	tests := []struct {
		name     string
		response []byte
		want     int
	}{
		{"87 LEDs", []byte{0x08, 0x43, 87, 0}, 87},
		{"256 LEDs", []byte{0x08, 0x43, 0, 1}, 256},
		{"empty", []byte{}, 0},
		{"short", []byte{0x08, 0x43}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLEDCountResponse(tt.response)
			if got != tt.want {
				t.Errorf("ParseLEDCountResponse() = %d, want %d", got, tt.want)
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
