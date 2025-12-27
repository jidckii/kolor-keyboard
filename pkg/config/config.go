package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load загружает конфигурацию из YAML файла
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Если firmware не указан, используем vial
	if cfg.Firmware == "" {
		cfg.Firmware = FirmwareVial
	}

	// Если mode не указан, используем mono
	if cfg.Mode == "" {
		cfg.Mode = ModeMono
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.Device.VendorID == 0 || c.Device.ProductID == 0 {
		return fmt.Errorf("device vendor_id and product_id are required")
	}

	// Проверяем firmware
	switch c.Firmware {
	case FirmwareStock, FirmwareVial:
		// ok
	default:
		return fmt.Errorf("unknown firmware: %s (expected 'stock' or 'vial')", c.Firmware)
	}

	// draw режим доступен только для vial
	if c.Mode == ModeDraw && c.Firmware == FirmwareStock {
		return fmt.Errorf("draw mode requires vial firmware (stock firmware only supports mono mode)")
	}

	switch c.Mode {
	case ModeMono:
		return c.validateMono()
	case ModeDraw:
		return c.validateDraw()
	default:
		return fmt.Errorf("unknown mode: %s (expected 'mono' or 'draw')", c.Mode)
	}
}

func (c *Config) validateMono() error {
	if len(c.Colors) == 0 {
		return fmt.Errorf("at least one color mapping is required for mono mode")
	}
	return nil
}

func (c *Config) validateDraw() error {
	if len(c.Keyboard.Rows) == 0 {
		return fmt.Errorf("keyboard.rows is required for draw mode")
	}

	if len(c.Drawings) == 0 {
		return fmt.Errorf("at least one drawing mapping is required for draw mode")
	}

	// Проверяем что все stripes ссылаются на существующие ряды
	numRows := len(c.Keyboard.Rows)
	for i, flag := range c.Drawings {
		if len(flag.Stripes) == 0 {
			return fmt.Errorf("flag[%d] (%s): at least one stripe is required", i, flag.Layout)
		}
		for j, stripe := range flag.Stripes {
			for _, row := range stripe.Rows {
				if row < 0 || row >= numRows {
					return fmt.Errorf("flag[%d] (%s) stripe[%d]: invalid row %d (keyboard has %d rows)",
						i, flag.Layout, j, row, numRows)
				}
			}
		}
	}

	return nil
}

// GetColorForLayout возвращает цвет для указанной раскладки (mono mode)
func (c *Config) GetColorForLayout(layout string) *RGBColor {
	for i := range c.Colors {
		if c.Colors[i].Layout == layout {
			return &c.Colors[i].Color
		}
	}
	// Fallback на wildcard
	for i := range c.Colors {
		if c.Colors[i].Layout == "*" {
			return &c.Colors[i].Color
		}
	}
	return nil
}

// GetFlagForLayout возвращает флаг для указанной раскладки (draw mode)
func (c *Config) GetFlagForLayout(layout string) *FlagMapping {
	for i := range c.Drawings {
		if c.Drawings[i].Layout == layout {
			return &c.Drawings[i]
		}
	}
	// Fallback на wildcard
	for i := range c.Drawings {
		if c.Drawings[i].Layout == "*" {
			return &c.Drawings[i]
		}
	}
	return nil
}

// GetLEDsForRow возвращает индексы LED для указанного ряда клавиатуры
func (c *Config) GetLEDsForRow(row int) []int {
	if row < 0 || row >= len(c.Keyboard.Rows) {
		return nil
	}
	return c.Keyboard.Rows[row]
}

// GetAllLEDIndices возвращает все индексы LED из конфигурации клавиатуры
func (c *Config) GetAllLEDIndices() []int {
	var indices []int
	for _, row := range c.Keyboard.Rows {
		indices = append(indices, row...)
	}
	return indices
}

// GetSpeed возвращает скорость эффектов (или 128 по умолчанию)
func (c *Config) GetSpeed() uint8 {
	if c.Speed != nil {
		return *c.Speed
	}
	return 128 // по умолчанию
}
