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

	// Включаем режим solid color
	a.logger.Info("enabling solid color mode")
	if err := a.device.EnableSolidColor(); err != nil {
		a.logger.Warn("failed to enable solid color mode", "error", err)
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

// applyLayout применяет цвет для указанной раскладки
func (a *App) applyLayout(layout string) error {
	color := a.cfg.GetColorForLayout(layout)
	if color == nil {
		a.logger.Warn("no color configured for layout", "layout", layout)
		return nil
	}

	if err := a.device.SetColorRGB(color.R, color.G, color.B); err != nil {
		return fmt.Errorf("failed to set color: %w", err)
	}

	a.logger.Debug("applied color", "layout", layout, "r", color.R, "g", color.G, "b", color.B)
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
