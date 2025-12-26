package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jidckii/kde-keyboard-flag-qmk/internal/config"
	"github.com/jidckii/kde-keyboard-flag-qmk/internal/dbus"
	"github.com/jidckii/kde-keyboard-flag-qmk/internal/hid"
)

// App - главное приложение
type App struct {
	cfg     *config.Config
	watcher *dbus.KDELayoutWatcher
	device  *hid.VIARGBDevice
	logger  *slog.Logger
}

// New создаёт новое приложение
func New(configPath string, logger *slog.Logger) (*App, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	// Загрузка конфигурации
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	logger.Info("loaded config", "mode", cfg.Mode)

	// Инициализация D-Bus watcher
	watcher, err := dbus.NewKDELayoutWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create layout watcher: %w", err)
	}

	// Инициализация HID устройства
	device := hid.NewVIARGBDevice(
		cfg.Device.VendorID,
		cfg.Device.ProductID,
		cfg.Device.UsagePage,
		cfg.Device.Usage,
	)

	return &App{
		cfg:     cfg,
		watcher: watcher,
		device:  device,
		logger:  logger,
	}, nil
}

// Run запускает приложение
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		a.logger.Info("received signal, shutting down", "signal", sig)
		cancel()
	}()

	// Открытие устройства
	a.logger.Info("opening HID device",
		"vid", fmt.Sprintf("%04X", a.cfg.Device.VendorID),
		"pid", fmt.Sprintf("%04X", a.cfg.Device.ProductID))

	if err := a.device.Open(); err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}
	defer a.device.Close()

	// Инициализация режима в зависимости от конфигурации
	if err := a.initializeMode(); err != nil {
		return fmt.Errorf("failed to initialize mode: %w", err)
	}

	// Установка начального состояния
	layout, err := a.watcher.GetCurrentLayout()
	if err != nil {
		a.logger.Warn("failed to get initial layout", "error", err)
	} else {
		a.logger.Info("current layout", "layout", layout.Layout, "name", layout.Name)
		if err := a.applyLayout(layout.Layout); err != nil {
			a.logger.Error("failed to apply initial layout", "error", err)
		}
	}

	// Запуск отслеживания
	events, err := a.watcher.Watch(ctx)
	if err != nil {
		return fmt.Errorf("failed to start watching: %w", err)
	}

	a.logger.Info("watching for layout changes...")

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("shutting down")
			return nil
		case event, ok := <-events:
			if !ok {
				return nil
			}
			a.logger.Info("layout changed",
				"layout", event.Layout,
				"name", event.Name,
				"index", event.Index)

			if err := a.applyLayout(event.Layout); err != nil {
				a.logger.Error("failed to apply layout", "error", err)
			}
		}
	}
}

// initializeMode инициализирует режим RGB
func (a *App) initializeMode() error {
	a.logger.Info("initializing", "firmware", a.cfg.Firmware, "mode", a.cfg.Mode)

	switch a.cfg.Firmware {
	case config.FirmwareStock:
		// Stock прошивка - только VIA RGB Matrix команды
		a.logger.Info("using VIA RGB Matrix commands (stock firmware)")
		if err := a.device.EnableSolidColor(); err != nil {
			a.logger.Warn("failed to enable solid color mode", "error", err)
		}

	case config.FirmwareVial:
		// Vial прошивка - Vial RGB команды
		a.logger.Info("using Vial RGB commands")
		// Получаем количество LED для информации
		ledCount, err := a.device.GetLEDCount()
		if err != nil {
			a.logger.Warn("failed to get LED count", "error", err)
		} else {
			a.logger.Info("detected LED count", "count", ledCount)
		}
		// Включаем Vial Direct режим
		if err := a.device.EnableVialDirectMode(); err != nil {
			return fmt.Errorf("failed to enable Vial direct mode: %w", err)
		}
	}
	return nil
}

// applyLayout применяет цвет/флаг для указанной раскладки
func (a *App) applyLayout(layout string) error {
	switch a.cfg.Mode {
	case config.ModeMono:
		return a.applyMonoLayout(layout)
	case config.ModeFlags:
		return a.applyFlagLayout(layout)
	default:
		return fmt.Errorf("unknown mode: %s", a.cfg.Mode)
	}
}

// applyMonoLayout применяет глобальный цвет для раскладки
func (a *App) applyMonoLayout(layout string) error {
	color := a.cfg.GetColorForLayout(layout)
	if color == nil {
		a.logger.Warn("no color configured for layout", "layout", layout)
		return nil
	}

	switch a.cfg.Firmware {
	case config.FirmwareStock:
		return a.applyMonoStock(color)
	case config.FirmwareVial:
		return a.applyMonoVial(color)
	default:
		return fmt.Errorf("unknown firmware: %s", a.cfg.Firmware)
	}
}

// applyMonoStock применяет цвет через VIA RGB Matrix (stock прошивка)
func (a *App) applyMonoStock(color *config.RGBColor) error {
	// Используем VIA RGB Matrix команды
	if err := a.device.SetColorRGB(color.R, color.G, color.B); err != nil {
		return fmt.Errorf("failed to set color: %w", err)
	}

	a.logger.Debug("applied mono color (stock)", "r", color.R, "g", color.G, "b", color.B)
	return nil
}

// applyMonoVial применяет цвет через Vial Direct режим (vial прошивка)
func (a *App) applyMonoVial(color *config.RGBColor) error {
	// Включаем Vial Direct режим
	if err := a.device.EnableVialDirectMode(); err != nil {
		a.logger.Warn("failed to enable Vial direct mode", "error", err)
	}

	// Получаем количество LED
	ledCount, err := a.device.GetLEDCount()
	if err != nil {
		a.logger.Warn("failed to get LED count", "error", err)
		ledCount = 87 // fallback
	}

	// Все LED одного цвета
	hsvColor := hid.RGBToHSV(color.R, color.G, color.B)
	updates := make([]hid.LEDUpdate, ledCount)
	for i := 0; i < ledCount; i++ {
		updates[i] = hid.LEDUpdate{Index: i, Color: hsvColor}
	}

	if err := a.device.SetLEDs(updates); err != nil {
		return fmt.Errorf("failed to set LEDs: %w", err)
	}

	a.logger.Debug("applied mono color (vial)", "r", color.R, "g", color.G, "b", color.B)
	return nil
}

// applyFlagLayout применяет флаг (per-key RGB) для раскладки
func (a *App) applyFlagLayout(layout string) error {
	flag := a.cfg.GetFlagForLayout(layout)
	if flag == nil {
		a.logger.Warn("no flag configured for layout", "layout", layout)
		return nil
	}

	// Каждый раз включаем Vial Direct режим (на случай если пользователь переключил режим)
	if err := a.device.EnableVialDirectMode(); err != nil {
		a.logger.Warn("failed to re-enable Vial direct mode", "error", err)
	}

	// Получаем количество LED
	ledCount, err := a.device.GetLEDCount()
	if err != nil {
		a.logger.Warn("failed to get LED count", "error", err)
		ledCount = 87 // fallback для Keychron V3
	}

	// Создаём массив для ВСЕХ LED, инициализируем чёрным (выключено)
	// Это гарантирует что все LED будут обновлены и в правильном порядке
	ledColors := make([]hid.HSVColor, ledCount)
	for i := range ledColors {
		ledColors[i] = hid.HSVColor{H: 0, S: 0, V: 0} // чёрный
	}

	// Заполняем цветами из конфига флага
	for _, stripe := range flag.Stripes {
		hsvColor := hid.RGBToHSV(stripe.Color.R, stripe.Color.G, stripe.Color.B)

		// Если указаны конкретные LED - используем их
		if len(stripe.LEDs) > 0 {
			for _, ledIdx := range stripe.LEDs {
				if ledIdx >= 0 && ledIdx < ledCount {
					ledColors[ledIdx] = hsvColor
				}
			}
		} else {
			// Иначе используем ряды
			for _, rowIdx := range stripe.Rows {
				ledIndices := a.cfg.GetLEDsForRow(rowIdx)
				for _, ledIdx := range ledIndices {
					if ledIdx >= 0 && ledIdx < ledCount {
						ledColors[ledIdx] = hsvColor
					}
				}
			}
		}
	}

	// Формируем обновления в порядке индексов (0, 1, 2, ..., ledCount-1)
	updates := make([]hid.LEDUpdate, ledCount)
	for i := 0; i < ledCount; i++ {
		updates[i] = hid.LEDUpdate{
			Index: i,
			Color: ledColors[i],
		}
	}

	a.logger.Debug("applying flag", "layout", layout, "led_count", len(updates))

	if err := a.device.SetLEDs(updates); err != nil {
		return fmt.Errorf("failed to set LEDs: %w", err)
	}

	return nil
}

// Close закрывает все ресурсы
func (a *App) Close() error {
	if a.watcher != nil {
		a.watcher.Close()
	}
	if a.device != nil {
		a.device.Close()
	}
	return nil
}
