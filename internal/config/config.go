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

	switch c.Mode {
	case ModeMono:
		return c.validateMono()
	case ModeFlags:
		return c.validateFlags()
	default:
		return fmt.Errorf("unknown mode: %s (expected 'mono' or 'flags')", c.Mode)
	}
}

func (c *Config) validateMono() error {
	if len(c.Colors) == 0 {
		return fmt.Errorf("at least one color mapping is required for mono mode")
	}
	return nil
}

func (c *Config) validateFlags() error {
	if len(c.Keyboard.Rows) == 0 {
		return fmt.Errorf("keyboard.rows is required for flags mode")
	}

	if len(c.Flags) == 0 {
		return fmt.Errorf("at least one flag mapping is required for flags mode")
	}

	// Проверяем что все stripes ссылаются на существующие ряды
	numRows := len(c.Keyboard.Rows)
	for i, flag := range c.Flags {
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

// GetFlagForLayout возвращает флаг для указанной раскладки (flags mode)
func (c *Config) GetFlagForLayout(layout string) *FlagMapping {
	for i := range c.Flags {
		if c.Flags[i].Layout == layout {
			return &c.Flags[i]
		}
	}
	// Fallback на wildcard
	for i := range c.Flags {
		if c.Flags[i].Layout == "*" {
			return &c.Flags[i]
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
