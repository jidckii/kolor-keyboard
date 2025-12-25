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

	if len(c.Colors) == 0 {
		return fmt.Errorf("at least one color mapping is required")
	}

	return nil
}

// GetColorForLayout возвращает цвет для указанной раскладки
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
