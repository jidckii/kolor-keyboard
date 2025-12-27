package discover

import (
	"regexp"
	"strings"
)

// KeyboardVariant - вариант клавиатуры
type KeyboardVariant struct {
	Layout     string // ansi, iso, jis
	HasEncoder bool
}

// KnownKeyboard - известная клавиатура
type KnownKeyboard struct {
	VendorID  uint16
	ProductID uint16
	Vendor    string // keychron, ducky, etc
	Model     string // v3, k2, one2, etc
	Variant   KeyboardVariant
	LEDCount  int // ожидаемое количество LED (0 = неизвестно)
}

// knownKeyboards - база известных клавиатур
var knownKeyboards = []KnownKeyboard{
	// Keychron V3
	{0x3434, 0x0320, "keychron", "v3", KeyboardVariant{"ansi", false}, 87},
	{0x3434, 0x0321, "keychron", "v3", KeyboardVariant{"iso", false}, 88},
	{0x3434, 0x0322, "keychron", "v3", KeyboardVariant{"jis", false}, 89},
	{0x3434, 0x0330, "keychron", "v3", KeyboardVariant{"ansi", true}, 87},
	{0x3434, 0x0331, "keychron", "v3", KeyboardVariant{"ansi", true}, 87},
	{0x3434, 0x0332, "keychron", "v3", KeyboardVariant{"iso", true}, 88},
	{0x3434, 0x0333, "keychron", "v3", KeyboardVariant{"jis", true}, 89},

	// Keychron V1
	{0x3434, 0x0310, "keychron", "v1", KeyboardVariant{"ansi", false}, 87},
	{0x3434, 0x0311, "keychron", "v1", KeyboardVariant{"iso", false}, 88},
	{0x3434, 0x0312, "keychron", "v1", KeyboardVariant{"jis", false}, 0},

	// Keychron Q1
	{0x3434, 0x0100, "keychron", "q1", KeyboardVariant{"ansi", true}, 82},
	{0x3434, 0x0101, "keychron", "q1", KeyboardVariant{"iso", true}, 83},

	// Keychron Q2
	{0x3434, 0x0200, "keychron", "q2", KeyboardVariant{"ansi", true}, 67},
	{0x3434, 0x0201, "keychron", "q2", KeyboardVariant{"iso", true}, 68},

	// Keychron K2
	{0x3434, 0x0220, "keychron", "k2", KeyboardVariant{"ansi", false}, 84},

	// TODO: добавить больше клавиатур
}

// LookupKeyboard ищет клавиатуру в базе по VID/PID
func LookupKeyboard(vid, pid uint16) *KnownKeyboard {
	for _, kb := range knownKeyboards {
		if kb.VendorID == vid && kb.ProductID == pid {
			return &kb
		}
	}
	return nil
}

// DetectVariant пытается определить вариант клавиатуры по названию продукта
func DetectVariant(productName string) KeyboardVariant {
	name := strings.ToLower(productName)
	var variant KeyboardVariant

	// Определяем раскладку
	switch {
	case strings.Contains(name, "iso"):
		variant.Layout = "iso"
	case strings.Contains(name, "jis"):
		variant.Layout = "jis"
	default:
		variant.Layout = "ansi" // по умолчанию
	}

	// Определяем наличие энкодера
	variant.HasEncoder = strings.Contains(name, "encoder") ||
		strings.Contains(name, "knob") ||
		strings.Contains(name, "rotary")

	return variant
}

// GenerateConfigPath генерирует путь для конфига в стиле Vial
// Формат: keyboards/vendor/model/variant/config.yaml
func GenerateConfigPath(dev *DeviceInfo) string {
	vendor := sanitizeName(dev.Manufacturer)
	product := sanitizeName(dev.Product)

	// Ищем в базе известных клавиатур
	known := LookupKeyboard(dev.VendorID, dev.ProductID)
	if known != nil {
		variant := known.Variant.Layout
		if known.Variant.HasEncoder {
			variant += "_encoder"
		}
		return vendor + "/" + known.Model + "/" + variant + "/config.yaml"
	}

	// Пытаемся извлечь модель из названия продукта
	model := extractModel(product)
	variant := DetectVariant(dev.Product)

	variantStr := variant.Layout
	if variant.HasEncoder {
		variantStr += "_encoder"
	}

	if model != "" {
		return vendor + "/" + model + "/" + variantStr + "/config.yaml"
	}

	// Fallback: просто vendor/product
	return vendor + "/" + product + "/config.yaml"
}

// GenerateConfigDir генерирует только директорию (без config.yaml)
func GenerateConfigDir(dev *DeviceInfo) string {
	path := GenerateConfigPath(dev)
	// Убираем /config.yaml
	if strings.HasSuffix(path, "/config.yaml") {
		return path[:len(path)-12]
	}
	return path
}

// sanitizeName очищает название для использования в пути
func sanitizeName(name string) string {
	// Убираем лишние пробелы
	name = strings.TrimSpace(name)
	// Приводим к нижнему регистру
	name = strings.ToLower(name)
	// Заменяем пробелы на дефисы
	name = strings.ReplaceAll(name, " ", "-")
	// Убираем специальные символы
	name = regexp.MustCompile(`[^a-z0-9\-_]`).ReplaceAllString(name, "")
	// Убираем множественные дефисы
	name = regexp.MustCompile(`-+`).ReplaceAllString(name, "-")
	// Убираем дефисы в начале и конце
	name = strings.Trim(name, "-")

	if name == "" {
		name = "unknown"
	}

	return name
}

// extractModel пытается извлечь модель из названия продукта
func extractModel(productName string) string {
	name := strings.ToLower(productName)

	// Паттерны для извлечения модели
	patterns := []string{
		`(v\d+)`,      // V1, V3, V10
		`(q\d+)`,      // Q1, Q2
		`(k\d+)`,      // K2, K8
		`(s\d+)`,      // S1
		`(c\d+)`,      // C1, C2
		`(one\s*\d*)`, // One, One2
		`(pro\s*\d*)`, // Pro, Pro2
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(name)
		if len(match) > 1 {
			model := strings.ReplaceAll(match[1], " ", "")
			return model
		}
	}

	return ""
}

// GetKeyboardInfo возвращает расширенную информацию о клавиатуре
func GetKeyboardInfo(dev *DeviceInfo) (vendor, model, variant string) {
	// Сначала пробуем найти в базе
	known := LookupKeyboard(dev.VendorID, dev.ProductID)
	if known != nil {
		vendor = known.Vendor
		model = known.Model
		variant = known.Variant.Layout
		if known.Variant.HasEncoder {
			variant += "_encoder"
		}
		return
	}

	// Иначе пытаемся определить по названию
	vendor = sanitizeName(dev.Manufacturer)
	model = extractModel(dev.Product)
	if model == "" {
		model = sanitizeName(dev.Product)
	}

	v := DetectVariant(dev.Product)
	variant = v.Layout
	if v.HasEncoder {
		variant += "_encoder"
	}

	return
}
