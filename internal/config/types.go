package config

import "fmt"

// Mode - режим работы
type Mode string

const (
	ModeMono Mode = "mono" // Глобальный цвет для всех клавиш
	ModeDraw Mode = "draw" // Per-key RGB с флагами
)

// Firmware - тип прошивки
type Firmware string

const (
	FirmwareStock Firmware = "stock" // Стоковая QMK/VIA прошивка
	FirmwareVial  Firmware = "vial"  // Vial прошивка
)

// Config - корневая структура конфигурации
type Config struct {
	Device   DeviceConfig `yaml:"device"`
	Firmware Firmware     `yaml:"firmware"` // stock или vial
	Mode     Mode         `yaml:"mode"`

	// Глобальные настройки RGB
	Brightness *uint8 `yaml:"brightness,omitempty"` // 0-255 (nil = не менять)
	Speed      *uint8 `yaml:"speed,omitempty"`      // 0-255 (nil = 128 по умолчанию)

	// Для mono режима - один цвет на всю клавиатуру
	Colors []ColorMapping `yaml:"colors,omitempty"`

	// Для draw режима - per-key RGB
	Keyboard KeyboardConfig `yaml:"keyboard,omitempty"`
	Drawings []FlagMapping  `yaml:"draw,omitempty"`
}

// DeviceConfig - параметры HID устройства
type DeviceConfig struct {
	VendorID  uint16 `yaml:"vendor_id"`
	ProductID uint16 `yaml:"product_id"`
	UsagePage uint16 `yaml:"usage_page"`
	Usage     uint16 `yaml:"usage"`
}

// ColorMapping - маппинг раскладки на цвет (для mono режима)
type ColorMapping struct {
	Layout string   `yaml:"layout"`
	Color  RGBColor `yaml:"color"`
}

// RGBColor - цвет в RGB или HSV формате
// При использовании HSV формата значения автоматически конвертируются в RGB
type RGBColor struct {
	R uint8 `yaml:"r"`
	G uint8 `yaml:"g"`
	B uint8 `yaml:"b"`
}

// rgbValues - значения RGB
type rgbValues struct {
	R uint8 `yaml:"r"`
	G uint8 `yaml:"g"`
	B uint8 `yaml:"b"`
}

// hsvValues - значения HSV (0-255, как в QMK/Vial)
type hsvValues struct {
	H uint8 `yaml:"h"`
	S uint8 `yaml:"s"`
	V uint8 `yaml:"v"`
}

// colorRaw - промежуточная структура для парсинга цвета из YAML
type colorRaw struct {
	RGB *rgbValues `yaml:"rgb"`
	HSV *hsvValues `yaml:"hsv"`
}

// UnmarshalYAML реализует кастомный парсинг цвета из YAML
// Поддерживает два формата:
//   - color: {rgb: {r: 255, g: 0, b: 0}}
//   - color: {hsv: {h: 0, s: 255, v: 255}}
func (c *RGBColor) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw colorRaw
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if raw.RGB != nil && raw.HSV != nil {
		return fmt.Errorf("cannot specify both rgb and hsv in color definition")
	}

	if raw.HSV != nil {
		// Конвертируем HSV (0-255) в RGB
		c.R, c.G, c.B = hsv255ToRGB(raw.HSV.H, raw.HSV.S, raw.HSV.V)
	} else if raw.RGB != nil {
		c.R = raw.RGB.R
		c.G = raw.RGB.G
		c.B = raw.RGB.B
	} else {
		return fmt.Errorf("color must have either 'rgb' or 'hsv' key")
	}

	return nil
}

// hsv255ToRGB конвертирует HSV (0-255) в RGB
// Формат как в QMK/Vial: H, S, V все в диапазоне 0-255
func hsv255ToRGB(h, s, v uint8) (r, g, b uint8) {
	if s == 0 {
		return v, v, v
	}

	// Нормализуем к float для вычислений
	hf := float64(h) / 255.0 * 6.0 // 0-6
	sf := float64(s) / 255.0
	vf := float64(v) / 255.0

	i := int(hf)
	f := hf - float64(i)

	p := vf * (1 - sf)
	q := vf * (1 - sf*f)
	t := vf * (1 - sf*(1-f))

	var rf, gf, bf float64
	switch i % 6 {
	case 0:
		rf, gf, bf = vf, t, p
	case 1:
		rf, gf, bf = q, vf, p
	case 2:
		rf, gf, bf = p, vf, t
	case 3:
		rf, gf, bf = p, q, vf
	case 4:
		rf, gf, bf = t, p, vf
	case 5:
		rf, gf, bf = vf, p, q
	}

	return uint8(rf * 255), uint8(gf * 255), uint8(bf * 255)
}

// KeyboardConfig - конфигурация рядов клавиатуры для draw режима
type KeyboardConfig struct {
	// Rows - LED индексы для каждого ряда клавиатуры
	// Пример: [[0,1,2,3,...], [15,16,17,...], ...]
	Rows [][]int `yaml:"rows"`
}

// FlagMapping - маппинг раскладки на флаг (для draw режима)
type FlagMapping struct {
	Layout  string       `yaml:"layout"`
	Stripes []FlagStripe `yaml:"stripes"`
}

// FlagStripe - горизонтальная полоса флага
type FlagStripe struct {
	// Rows - какие ряды клавиатуры занимает эта полоса (0-indexed)
	Rows []int `yaml:"rows,omitempty"`
	// LEDs - конкретные индексы LED (альтернатива Rows для сложных флагов)
	LEDs []int `yaml:"leds,omitempty"`
	// Color - цвет полосы
	Color RGBColor `yaml:"color"`
}
