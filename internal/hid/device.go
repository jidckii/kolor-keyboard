package hid

import (
	"fmt"
	"sync"
	"time"

	"github.com/sstallion/go-hid"
)

// RGBController - интерфейс управления RGB
type RGBController interface {
	Open() error
	Close() error
	SetColor(color HSVColor) error
	SetColorRGB(r, g, b uint8) error
	SetEffect(effect uint8) error
	SetBrightness(brightness uint8) error
}

// VIARGBDevice реализует RGBController для VIA клавиатур
type VIARGBDevice struct {
	vendorID  uint16
	productID uint16
	usagePage uint16
	usage     uint16

	device *hid.Device
	mu     sync.Mutex
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
	_, err = d.device.ReadWithTimeout(response, 100*time.Millisecond)
	// Игнорируем ошибку таймаута
	return nil
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
