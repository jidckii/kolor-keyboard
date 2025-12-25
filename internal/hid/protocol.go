package hid

// VIA RGB Matrix протокол
// Channel 0x03 = RGB Matrix
const (
	CmdVIASetValue = 0x07 // id_lighting_set_value
	CmdVIAGetValue = 0x08 // id_lighting_get_value

	// RGB Matrix channel
	ChannelRGBMatrix = 0x03

	// RGB Matrix value IDs
	RGBMatrixBrightness = 0x01
	RGBMatrixEffect     = 0x02
	RGBMatrixSpeed      = 0x03
	RGBMatrixColor      = 0x04

	// Эффекты
	EffectOff        = 0x00
	EffectSolidColor = 0x01
)

// Размер пакета
const PacketSize = 64

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

// BuildSetEffectPacket устанавливает эффект RGB
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
