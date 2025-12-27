package discover

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jidckii/kolor-keyboard/pkg/hid"
	hidlib "github.com/sstallion/go-hid"
)

// VIAUsagePage - стандартный Usage Page для VIA/Vial клавиатур
const VIAUsagePage = 0xFF60

// VIAUsage - стандартный Usage для VIA/Vial
const VIAUsage = 0x61

// DeviceInfo содержит информацию об обнаруженном устройстве
type DeviceInfo struct {
	VendorID     uint16
	ProductID    uint16
	UsagePage    uint16
	Usage        uint16
	Manufacturer string
	Product      string
	Path         string
	IsVial       bool
	LEDCount     int
}

// DiscoveredConfig - сгенерированная конфигурация
type DiscoveredConfig struct {
	Device       DeviceInfo
	Firmware     string // "vial" или "stock"
	KeyboardRows [][]int
}

// FindVIADevices ищет все VIA/Vial совместимые клавиатуры
func FindVIADevices() ([]DeviceInfo, error) {
	if err := hidlib.Init(); err != nil {
		return nil, fmt.Errorf("failed to init HID: %w", err)
	}
	defer hidlib.Exit()

	var devices []DeviceInfo
	seen := make(map[string]bool)

	err := hidlib.Enumerate(0, 0, func(info *hidlib.DeviceInfo) error {
		// Ищем устройства с VIA Usage Page
		if info.UsagePage == VIAUsagePage && info.Usage == VIAUsage {
			key := fmt.Sprintf("%04x:%04x", info.VendorID, info.ProductID)
			if !seen[key] {
				seen[key] = true
				devices = append(devices, DeviceInfo{
					VendorID:     info.VendorID,
					ProductID:    info.ProductID,
					UsagePage:    info.UsagePage,
					Usage:        info.Usage,
					Manufacturer: info.MfrStr,
					Product:      info.ProductStr,
					Path:         info.Path,
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to enumerate devices: %w", err)
	}

	return devices, nil
}

// CheckVialSupport проверяет поддерживает ли устройство Vial и возвращает количество LED
func CheckVialSupport(dev *DeviceInfo) error {
	device := hid.NewVIARGBDevice(dev.VendorID, dev.ProductID, dev.UsagePage, dev.Usage)

	if err := device.Open(); err != nil {
		return fmt.Errorf("cannot open device: %w", err)
	}
	defer device.Close()

	// Пробуем получить количество LED через Vial команду
	ledCount, err := device.GetLEDCount()
	if err != nil {
		dev.IsVial = false
		dev.LEDCount = 0
		return nil // не ошибка, просто не Vial
	}

	if ledCount > 0 {
		dev.IsVial = true
		dev.LEDCount = ledCount
	}

	return nil
}

// RunLEDMappingTour запускает интерактивный тур для маппинга LED по рядам
func RunLEDMappingTour(dev *DeviceInfo) ([][]int, error) {
	device := hid.NewVIARGBDevice(dev.VendorID, dev.ProductID, dev.UsagePage, dev.Usage)

	if err := device.Open(); err != nil {
		return nil, fmt.Errorf("failed to open device: %w", err)
	}
	defer device.Close()

	// Включаем Vial Direct режим
	if err := device.EnableVialDirectMode(); err != nil {
		return nil, fmt.Errorf("failed to enable direct mode: %w", err)
	}

	ledCount := dev.LEDCount
	if ledCount == 0 {
		var err error
		ledCount, err = device.GetLEDCount()
		if err != nil {
			return nil, fmt.Errorf("failed to get LED count: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║              LED Mapping Tour                                ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Detected %d LEDs (indices 0-%d)                             \n", ledCount, ledCount-1)
	fmt.Println("║                                                              ║")
	fmt.Println("║  Commands:                                                   ║")
	fmt.Println("║    Enter/n  - next LED (add to current row)                  ║")
	fmt.Println("║    r        - end row here (add LED and start new row)       ║")
	fmt.Println("║    s        - skip this LED (don't add to any row)           ║")
	fmt.Println("║    b        - go back (undo last LED)                        ║")
	fmt.Println("║    q        - quit and save                                  ║")
	fmt.Println("║                                                              ║")
	fmt.Println("║  Colors:                                                     ║")
	fmt.Println("║    RED     - current LED                                     ║")
	fmt.Println("║    GREEN   - current row LEDs                                ║")
	fmt.Println("║    YELLOW  - saved rows (different shades per row)           ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	var rows [][]int
	currentRow := []int{}
	currentLED := 0

	// Выключаем все LED
	allOff := make([]hid.LEDUpdate, ledCount)
	for i := 0; i < ledCount; i++ {
		allOff[i] = hid.LEDUpdate{Index: i, Color: hid.HSVColor{H: 0, S: 0, V: 0}}
	}

	for currentLED < ledCount {
		// Обновляем дисплей
		device.EnableVialDirectMode()
		device.SetLEDs(allOff)

		// Показываем уже сохранённые ряды разными оттенками жёлтого
		for rowIdx, row := range rows {
			// Оттенки жёлтого: H=20-40 (оранжево-жёлтый спектр)
			hue := uint8(20 + (rowIdx%5)*4) // 20, 24, 28, 32, 36, циклично
			for _, idx := range row {
				device.SetLEDs([]hid.LEDUpdate{
					{Index: idx, Color: hid.HSVColor{H: hue, S: 255, V: 200}},
				})
			}
		}

		// Показываем предыдущие LED в текущем ряду зелёным
		for _, idx := range currentRow {
			device.SetLEDs([]hid.LEDUpdate{
				{Index: idx, Color: hid.HSVColor{H: 85, S: 255, V: 255}}, // зелёный
			})
		}

		// Текущий LED красным
		device.SetLEDs([]hid.LEDUpdate{
			{Index: currentLED, Color: hid.HSVColor{H: 0, S: 255, V: 255}}, // красный
		})

		// Показываем статус
		rowNum := len(rows)
		fmt.Printf("\r[Row %d] LED %d/%d (row has %d LEDs) > ",
			rowNum, currentLED, ledCount-1, len(currentRow))

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "q":
			// Сохраняем текущий ряд если не пустой
			if len(currentRow) > 0 {
				rows = append(rows, currentRow)
			}
			fmt.Println("\nMapping complete!")
			device.SetLEDs(allOff)
			return rows, nil

		case "r":
			// Добавить LED к ряду, сохранить ряд и начать новый
			currentRow = append(currentRow, currentLED)
			rows = append(rows, currentRow)
			fmt.Printf("\n  → Row %d saved: %v (%d LEDs)\n", len(rows)-1, currentRow, len(currentRow))
			currentRow = []int{}
			currentLED++

		case "s":
			// Пропустить LED
			fmt.Printf(" (skipped LED %d)\n", currentLED)
			currentLED++

		case "b":
			// Назад
			if len(currentRow) > 0 {
				// Убираем последний LED из текущего ряда
				currentLED = currentRow[len(currentRow)-1]
				currentRow = currentRow[:len(currentRow)-1]
				fmt.Printf(" (back to LED %d)\n", currentLED)
			} else if len(rows) > 0 {
				// Возвращаемся к предыдущему ряду
				currentRow = rows[len(rows)-1]
				rows = rows[:len(rows)-1]
				if len(currentRow) > 0 {
					currentLED = currentRow[len(currentRow)-1]
					currentRow = currentRow[:len(currentRow)-1]
				}
				fmt.Printf(" (back to previous row, LED %d)\n", currentLED)
			}

		case "", "n":
			// Добавить LED к текущему ряду и перейти к следующему
			currentRow = append(currentRow, currentLED)
			currentLED++
		}
	}

	// Сохраняем последний ряд
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
		fmt.Printf("\n  → Row %d saved: %v (%d LEDs)\n", len(rows)-1, currentRow, len(currentRow))
	}

	// Выключаем все LED
	device.SetLEDs(allOff)

	fmt.Println("\n✓ All LEDs mapped!")
	return rows, nil
}

// GenerateConfig генерирует YAML конфигурацию
func GenerateConfig(cfg *DiscoveredConfig) string {
	var sb strings.Builder

	vendor, model, variant := GetKeyboardInfo(&cfg.Device)

	sb.WriteString("# kolor-keyboard configuration\n")
	sb.WriteString(fmt.Sprintf("# Generated for: %s %s\n", cfg.Device.Manufacturer, cfg.Device.Product))
	sb.WriteString(fmt.Sprintf("# Keyboard: %s/%s/%s\n", vendor, model, variant))
	sb.WriteString("\n")

	sb.WriteString("device:\n")
	sb.WriteString(fmt.Sprintf("  vendor_id: 0x%04X\n", cfg.Device.VendorID))
	sb.WriteString(fmt.Sprintf("  product_id: 0x%04X\n", cfg.Device.ProductID))
	sb.WriteString(fmt.Sprintf("  usage_page: 0x%04X\n", cfg.Device.UsagePage))
	sb.WriteString(fmt.Sprintf("  usage: 0x%02X\n", cfg.Device.Usage))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("firmware: %s\n", cfg.Firmware))

	if cfg.Firmware == "vial" && len(cfg.KeyboardRows) > 0 {
		sb.WriteString("mode: draw\n")
		sb.WriteString("\n")
		sb.WriteString("# Global RGB settings\n")
		sb.WriteString("brightness: 200  # 0-255\n")
		sb.WriteString("speed: 128       # 0-255\n")
		sb.WriteString("\n")
		sb.WriteString("# LED layout (generated by discover)\n")
		sb.WriteString("keyboard:\n")
		sb.WriteString("  rows:\n")
		for i, row := range cfg.KeyboardRows {
			sb.WriteString(fmt.Sprintf("    # Row %d (%d LEDs)\n", i, len(row)))
			sb.WriteString(fmt.Sprintf("    - %v\n", row))
		}
		sb.WriteString("\n")
		sb.WriteString("draw:\n")
		sb.WriteString("  # Example: all rows same color\n")
		sb.WriteString("  - layout: \"*\"\n")
		sb.WriteString("    stripes:\n")

		// Генерируем все ряды
		rowIndices := make([]int, len(cfg.KeyboardRows))
		for i := range cfg.KeyboardRows {
			rowIndices[i] = i
		}
		sb.WriteString(fmt.Sprintf("      - rows: %v\n", rowIndices))
		sb.WriteString("        color: {rgb: {r: 0, g: 255, b: 0}}\n")
	} else {
		sb.WriteString("mode: mono\n")
		sb.WriteString("\n")
		sb.WriteString("# Global RGB settings\n")
		sb.WriteString("brightness: 200  # 0-255\n")
		sb.WriteString("\n")
		sb.WriteString("colors:\n")
		sb.WriteString("  - layout: \"*\"\n")
		sb.WriteString("    color: {rgb: {r: 0, g: 255, b: 0}}\n")
	}

	return sb.String()
}
