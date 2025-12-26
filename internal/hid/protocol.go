package hid

// VIA/Vial RGB протокол
const (
	CmdVIASetValue = 0x07 // id_lighting_set_value
	CmdVIAGetValue = 0x08 // id_lighting_get_value

	// VIA RGB Matrix channel (для глобального цвета)
	ChannelRGBMatrix = 0x03

	// RGB Matrix value IDs
	RGBMatrixBrightness = 0x01
	RGBMatrixEffect     = 0x02
	RGBMatrixSpeed      = 0x03
	RGBMatrixColor      = 0x04

	// Vial RGB команды (для per-key RGB)
	VialRGBSetMode   = 0x41 // vialrgb_set_mode
	VialRGBDirectSet = 0x42 // vialrgb_direct_fastset
	VialRGBGetLEDs   = 0x43 // vialrgb_get_number_leds

	// VIA RGB Matrix эффекты (для mono режима)
	EffectDisable    = 0x00
	EffectSolidColor = 0x02

	// Vial RGB эффекты (для flags режима)
	VialEffectOff    = 0x0000
	VialEffectDirect = 0x0001
)

// Размер пакета (Vial использует 32 байта)
const PacketSize = 32

// Максимум LED на пакет: (32 - 5) / 3 = 9
const MaxLEDsPerPacket = 9

// HSVColor - цвет в формате HSV (0-255)
type HSVColor struct {
	H uint8
	S uint8
	V uint8
}

// RGBToHSV конвертирует RGB в HSV
func RGBToHSV(r, g, b uint8) HSVColor {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := rf
	if gf > max {
		max = gf
	}
	if bf > max {
		max = bf
	}

	min := rf
	if gf < min {
		min = gf
	}
	if bf < min {
		min = bf
	}

	delta := max - min

	var h, s, v float64
	v = max

	if max == 0 {
		s = 0
	} else {
		s = delta / max
	}

	if delta == 0 {
		h = 0
	} else {
		switch max {
		case rf:
			h = (gf - bf) / delta
			if gf < bf {
				h += 6
			}
		case gf:
			h = 2 + (bf-rf)/delta
		case bf:
			h = 4 + (rf-gf)/delta
		}
		h /= 6
	}

	return HSVColor{
		H: uint8(h * 255),
		S: uint8(s * 255),
		V: uint8(v * 255),
	}
}

// BuildSetEffectPacket устанавливает эффект VIA RGB Matrix (для mono режима)
func BuildSetEffectPacket(effect uint8) []byte {
	packet := make([]byte, PacketSize)
	packet[0] = CmdVIASetValue
	packet[1] = ChannelRGBMatrix
	packet[2] = RGBMatrixEffect
	packet[3] = effect
	return packet
}

// BuildSetColorPacket устанавливает глобальный цвет (hue, saturation)
func BuildSetColorPacket(hue, sat uint8) []byte {
	packet := make([]byte, PacketSize)
	packet[0] = CmdVIASetValue
	packet[1] = ChannelRGBMatrix
	packet[2] = RGBMatrixColor
	packet[3] = hue
	packet[4] = sat
	return packet
}

// BuildSetBrightnessPacket устанавливает яркость
func BuildSetBrightnessPacket(brightness uint8) []byte {
	packet := make([]byte, PacketSize)
	packet[0] = CmdVIASetValue
	packet[1] = ChannelRGBMatrix
	packet[2] = RGBMatrixBrightness
	packet[3] = brightness
	return packet
}

// BuildVialSetModePacket устанавливает режим Vial RGB (для flags режима)
// Формат: [0x07, 0x41, mode_lo, mode_hi, speed, H, S, V]
func BuildVialSetModePacket(mode uint16, speed, h, s, v uint8) []byte {
	packet := make([]byte, PacketSize)
	packet[0] = CmdVIASetValue
	packet[1] = VialRGBSetMode
	packet[2] = byte(mode & 0xFF)
	packet[3] = byte(mode >> 8)
	packet[4] = speed
	packet[5] = h
	packet[6] = s
	packet[7] = v
	return packet
}

// BuildDirectSetPacket создаёт пакет для установки LED (per-key RGB)
// Формат: [0x07, 0x42, start_lo, start_hi, count, H, S, V, H, S, V, ...]
func BuildDirectSetPacket(startIndex int, colors []HSVColor) []byte {
	packet := make([]byte, PacketSize)
	packet[0] = CmdVIASetValue
	packet[1] = VialRGBDirectSet
	packet[2] = byte(startIndex & 0xFF)
	packet[3] = byte(startIndex >> 8)
	packet[4] = byte(len(colors))

	offset := 5
	for _, c := range colors {
		if offset+3 > PacketSize {
			break
		}
		packet[offset] = c.H
		packet[offset+1] = c.S
		packet[offset+2] = c.V
		offset += 3
	}

	return packet
}

// BuildGetLEDCountPacket запрашивает количество LED
func BuildGetLEDCountPacket() []byte {
	packet := make([]byte, PacketSize)
	packet[0] = CmdVIAGetValue
	packet[1] = VialRGBGetLEDs
	return packet
}

// ParseLEDCountResponse парсит ответ на запрос количества LED
func ParseLEDCountResponse(response []byte) int {
	if len(response) < 4 {
		return 0
	}
	return int(response[2]) | (int(response[3]) << 8)
}

// BuildGetColorPacket запрашивает текущий цвет
func BuildGetColorPacket() []byte {
	packet := make([]byte, PacketSize)
	packet[0] = CmdVIAGetValue
	packet[1] = ChannelRGBMatrix
	packet[2] = RGBMatrixColor
	return packet
}

// BuildGetEffectPacket запрашивает текущий эффект
func BuildGetEffectPacket() []byte {
	packet := make([]byte, PacketSize)
	packet[0] = CmdVIAGetValue
	packet[1] = ChannelRGBMatrix
	packet[2] = RGBMatrixEffect
	return packet
}
