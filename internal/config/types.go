package config

// Mode - режим работы
type Mode string

const (
	ModeMono  Mode = "mono"  // Глобальный цвет для всех клавиш
	ModeFlags Mode = "flags" // Per-key RGB с флагами
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

	// Для mono режима - один цвет на всю клавиатуру
	Colors []ColorMapping `yaml:"colors,omitempty"`

	// Для flags режима - per-key RGB
	Keyboard KeyboardConfig `yaml:"keyboard,omitempty"`
	Flags    []FlagMapping  `yaml:"flags,omitempty"`
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

// RGBColor - цвет в RGB формате
type RGBColor struct {
	R uint8 `yaml:"r"`
	G uint8 `yaml:"g"`
	B uint8 `yaml:"b"`
}

// KeyboardConfig - конфигурация рядов клавиатуры для flags режима
type KeyboardConfig struct {
	// Rows - LED индексы для каждого ряда клавиатуры
	// Пример: [[0,1,2,3,...], [15,16,17,...], ...]
	Rows [][]int `yaml:"rows"`
}

// FlagMapping - маппинг раскладки на флаг (для flags режима)
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
