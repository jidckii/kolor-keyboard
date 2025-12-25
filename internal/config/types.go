package config

// Config - корневая структура конфигурации
type Config struct {
	Device DeviceConfig   `yaml:"device"`
	Colors []ColorMapping `yaml:"colors"`
}

// DeviceConfig - параметры HID устройства
type DeviceConfig struct {
	VendorID  uint16 `yaml:"vendor_id"`
	ProductID uint16 `yaml:"product_id"`
	UsagePage uint16 `yaml:"usage_page"`
	Usage     uint16 `yaml:"usage"`
}

// ColorMapping - маппинг раскладки на цвет
type ColorMapping struct {
	Layout string   `yaml:"layout"`
	Color  RGBColor `yaml:"color"`
}

// RGBColor - цвет в RGB
type RGBColor struct {
	R uint8 `yaml:"r"`
	G uint8 `yaml:"g"`
	B uint8 `yaml:"b"`
}
