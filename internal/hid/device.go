package hid

import (
	"fmt"
	"sync"
	"time"

	"github.com/sstallion/go-hid"
)

// LEDUpdate - обновление одного LED
type LEDUpdate struct {
	Index int
	Color HSVColor
}

// VIARGBDevice реализует управление RGB для VIA клавиатур
type VIARGBDevice struct {
	vendorID  uint16
	productID uint16
	usagePage uint16
	usage     uint16

	device   *hid.Device
	mu       sync.Mutex
	ledCount int
}

// NewVIARGBDevice создаёт новое устройство
func NewVIARGBDevice(vendorID, productID, usagePage, usage uint16) *VIARGBDevice {
	return &VIARGBDevice{
		vendorID:  vendorID,
		productID: productID,
		usagePage: usagePage,
		usage:     usage,
	}
}

// Open открывает HID устройство
func (d *VIARGBDevice) Open() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := hid.Init(); err != nil {
		return fmt.Errorf("failed to init HID: %w", err)
	}

	var targetDevice *hid.DeviceInfo
	err := hid.Enumerate(d.vendorID, d.productID, func(info *hid.DeviceInfo) error {
		if info.UsagePage == d.usagePage && info.Usage == d.usage {
			targetDevice = info
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to enumerate devices: %w", err)
	}

	if targetDevice == nil {
		return fmt.Errorf("device not found: VID=%04X PID=%04X", d.vendorID, d.productID)
	}

	dev, err := hid.OpenPath(targetDevice.Path)
	if err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}

	d.device = dev
	return nil
}

// Close закрывает устройство
func (d *VIARGBDevice) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.device != nil {
		err := d.device.Close()
		d.device = nil
		hid.Exit()
		return err
	}
	return nil
}

// write отправляет пакет и читает ответ
func (d *VIARGBDevice) write(packet []byte) error {
	if d.device == nil {
		return fmt.Errorf("device not opened")
	}

	_, err := d.device.Write(packet)
	if err != nil {
		return err
	}

	// Читаем ответ (VIA требует это)
	response := make([]byte, PacketSize)
	_, _ = d.device.ReadWithTimeout(response, 100*time.Millisecond)
	return nil
}

// writeWithResponse отправляет пакет и возвращает ответ
func (d *VIARGBDevice) writeWithResponse(packet []byte) ([]byte, error) {
	if d.device == nil {
		return nil, fmt.Errorf("device not opened")
	}

	_, err := d.device.Write(packet)
	if err != nil {
		return nil, err
	}

	response := make([]byte, PacketSize)
	_, err = d.device.ReadWithTimeout(response, 100*time.Millisecond)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// SetColor устанавливает глобальный цвет (HSV)
func (d *VIARGBDevice) SetColor(color HSVColor) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	packet := BuildSetColorPacket(color.H, color.S)
	return d.write(packet)
}

// SetColorRGB устанавливает глобальный цвет (RGB)
func (d *VIARGBDevice) SetColorRGB(r, g, b uint8) error {
	color := RGBToHSV(r, g, b)
	return d.SetColor(color)
}

// SetEffect устанавливает эффект RGB
func (d *VIARGBDevice) SetEffect(effect uint8) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	packet := BuildSetEffectPacket(effect)
	return d.write(packet)
}

// SetBrightness устанавливает яркость
func (d *VIARGBDevice) SetBrightness(brightness uint8) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	packet := BuildSetBrightnessPacket(brightness)
	return d.write(packet)
}

// EnableSolidColor включает режим solid color
func (d *VIARGBDevice) EnableSolidColor() error {
	return d.SetEffect(EffectSolidColor)
}

// EnableVialDirectMode включает режим прямого управления LED через Vial RGB
// Это нужно для per-key RGB. Использует команду 0x41 (vialrgb_set_mode)
func (d *VIARGBDevice) EnableVialDirectMode() error {
	return d.EnableVialDirectModeWithSpeed(128)
}

// EnableVialDirectModeWithSpeed включает режим прямого управления LED с указанной скоростью
func (d *VIARGBDevice) EnableVialDirectModeWithSpeed(speed uint8) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// mode=1 (VIALRGB_EFFECT_DIRECT), HSV=0,255,255
	packet := BuildVialSetModePacket(VialEffectDirect, speed, 0, 255, 255)
	return d.write(packet)
}

// GetLEDCount возвращает количество LED
func (d *VIARGBDevice) GetLEDCount() (int, error) {
	if d.ledCount > 0 {
		return d.ledCount, nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	packet := BuildGetLEDCountPacket()
	response, err := d.writeWithResponse(packet)
	if err != nil {
		return 0, fmt.Errorf("failed to get LED count: %w", err)
	}

	d.ledCount = ParseLEDCountResponse(response)
	return d.ledCount, nil
}

// SetLEDs устанавливает цвета для группы LED (per-key RGB)
func (d *VIARGBDevice) SetLEDs(updates []LEDUpdate) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Группируем последовательные LED для оптимальной отправки
	for i := 0; i < len(updates); i += MaxLEDsPerPacket {
		end := i + MaxLEDsPerPacket
		if end > len(updates) {
			end = len(updates)
		}

		batch := updates[i:end]
		if len(batch) == 0 {
			continue
		}

		// Конвертируем в формат для пакета
		colors := make([]HSVColor, len(batch))
		startIndex := batch[0].Index

		for j, u := range batch {
			colors[j] = u.Color
		}

		packet := BuildDirectSetPacket(startIndex, colors)
		if err := d.write(packet); err != nil {
			return fmt.Errorf("failed to set LEDs at index %d: %w", startIndex, err)
		}

		// Небольшая задержка между пакетами
		time.Sleep(1 * time.Millisecond)
	}

	return nil
}

// SetAllLEDs устанавливает один цвет для всех LED
func (d *VIARGBDevice) SetAllLEDs(color HSVColor, ledCount int) error {
	updates := make([]LEDUpdate, ledCount)
	for i := 0; i < ledCount; i++ {
		updates[i] = LEDUpdate{Index: i, Color: color}
	}
	return d.SetLEDs(updates)
}

// SetLEDsByIndices устанавливает цвет для указанных индексов LED
func (d *VIARGBDevice) SetLEDsByIndices(indices []int, color HSVColor) error {
	updates := make([]LEDUpdate, len(indices))
	for i, idx := range indices {
		updates[i] = LEDUpdate{Index: idx, Color: color}
	}
	return d.SetLEDs(updates)
}
