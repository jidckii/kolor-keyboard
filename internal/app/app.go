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
	switch a.cfg.Mode {
	case config.ModeMono:
		a.logger.Info("enabling solid color mode")
		if err := a.device.EnableSolidColor(); err != nil {
			a.logger.Warn("failed to enable solid color mode", "error", err)
		}
	case config.ModeFlags:
		a.logger.Info("enabling Vial direct mode for per-key RGB")
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

	if err := a.device.SetColorRGB(color.R, color.G, color.B); err != nil {
		return fmt.Errorf("failed to set color: %w", err)
	}

	a.logger.Debug("applied mono color", "layout", layout, "r", color.R, "g", color.G, "b", color.B)
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

	// Собираем все LED обновления для флага
	var updates []hid.LEDUpdate

	for _, stripe := range flag.Stripes {
		hsvColor := hid.RGBToHSV(stripe.Color.R, stripe.Color.G, stripe.Color.B)

		// Если указаны конкретные LED - используем их
		if len(stripe.LEDs) > 0 {
			for _, ledIdx := range stripe.LEDs {
				updates = append(updates, hid.LEDUpdate{
					Index: ledIdx,
					Color: hsvColor,
				})
			}
		} else {
			// Иначе используем ряды
			for _, rowIdx := range stripe.Rows {
				ledIndices := a.cfg.GetLEDsForRow(rowIdx)
				for _, ledIdx := range ledIndices {
					updates = append(updates, hid.LEDUpdate{
						Index: ledIdx,
						Color: hsvColor,
					})
				}
			}
		}
	}

	if len(updates) == 0 {
		a.logger.Warn("no LED updates for flag", "layout", layout)
		return nil
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
