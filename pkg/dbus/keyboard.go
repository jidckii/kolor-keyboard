package dbus

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	kdeKeyboardDest = "org.kde.keyboard"
	kdeLayoutsPath  = "/Layouts"
	kdeLayoutsIface = "org.kde.KeyboardLayouts"
)

// LayoutEvent - событие смены раскладки
type LayoutEvent struct {
	Index  uint32 // Индекс раскладки (0, 1, 2...)
	Layout string // Код раскладки ("ru", "us", "ua")
	Name   string // Полное имя ("Russian", "English (US)")
}

// LayoutWatcher - интерфейс для отслеживания раскладки
type LayoutWatcher interface {
	Watch(ctx context.Context) (<-chan LayoutEvent, error)
	GetCurrentLayout() (LayoutEvent, error)
	Close() error
}

// KDELayoutWatcher реализует LayoutWatcher для KDE Plasma 6
type KDELayoutWatcher struct {
	conn   *dbus.Conn
	cancel context.CancelFunc
}

// NewKDELayoutWatcher создаёт новый watcher для KDE
func NewKDELayoutWatcher() (*KDELayoutWatcher, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}
	return &KDELayoutWatcher{conn: conn}, nil
}

// Watch запускает отслеживание смены раскладки
func (w *KDELayoutWatcher) Watch(ctx context.Context) (<-chan LayoutEvent, error) {
	events := make(chan LayoutEvent, 10)

	// Подписка на сигнал layoutChanged
	matchRule := fmt.Sprintf(
		"type='signal',interface='%s',member='layoutChanged',path='%s'",
		kdeLayoutsIface, kdeLayoutsPath,
	)

	call := w.conn.BusObject().Call(
		"org.freedesktop.DBus.AddMatch", 0, matchRule,
	)
	if call.Err != nil {
		return nil, fmt.Errorf("failed to add match rule: %w", call.Err)
	}

	signals := make(chan *dbus.Signal, 10)
	w.conn.Signal(signals)

	ctx, w.cancel = context.WithCancel(ctx)

	go func() {
		defer close(events)
		for {
			select {
			case <-ctx.Done():
				return
			case sig, ok := <-signals:
				if !ok {
					return
				}
				if sig.Name == kdeLayoutsIface+".layoutChanged" {
					event, err := w.parseLayoutSignal(sig)
					if err == nil {
						select {
						case events <- event:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	return events, nil
}

// GetCurrentLayout возвращает текущую раскладку
func (w *KDELayoutWatcher) GetCurrentLayout() (LayoutEvent, error) {
	obj := w.conn.Object(kdeKeyboardDest, kdeLayoutsPath)

	var index uint32
	err := obj.Call(kdeLayoutsIface+".getLayout", 0).Store(&index)
	if err != nil {
		return LayoutEvent{}, fmt.Errorf("failed to get current layout: %w", err)
	}

	// Получаем список раскладок для имени
	layouts, err := w.getLayoutsList()
	if err != nil {
		return LayoutEvent{Index: index}, nil
	}

	layout := ""
	name := ""
	if int(index) < len(layouts) {
		info := layouts[index]
		layout = info.Code
		name = info.Name
	}

	return LayoutEvent{
		Index:  index,
		Layout: layout,
		Name:   name,
	}, nil
}

// LayoutInfo содержит информацию о раскладке из D-Bus
type LayoutInfo struct {
	Code    string // "us", "ru"
	Variant string // "", "phonetic"
	Name    string // "English (US)", "Russian"
}

// getLayoutsList получает список всех раскладок
func (w *KDELayoutWatcher) getLayoutsList() ([]LayoutInfo, error) {
	obj := w.conn.Object(kdeKeyboardDest, kdeLayoutsPath)

	// D-Bus возвращает a(sss) - массив структур из 3 строк
	call := obj.Call(kdeLayoutsIface+".getLayoutsList", 0)
	if call.Err != nil {
		return nil, fmt.Errorf("failed to get layouts list: %w", call.Err)
	}

	if len(call.Body) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	// Парсим как [][]interface{} так как godbus возвращает структуры как слайсы
	rawLayouts, ok := call.Body[0].([][]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T", call.Body[0])
	}

	result := make([]LayoutInfo, len(rawLayouts))
	for i, l := range rawLayouts {
		if len(l) >= 3 {
			code, _ := l[0].(string)
			variant, _ := l[1].(string)
			name, _ := l[2].(string)
			result[i] = LayoutInfo{
				Code:    code,
				Variant: variant,
				Name:    name,
			}
		}
	}

	return result, nil
}

// parseLayoutSignal парсит сигнал смены раскладки
func (w *KDELayoutWatcher) parseLayoutSignal(sig *dbus.Signal) (LayoutEvent, error) {
	if len(sig.Body) < 1 {
		return LayoutEvent{}, fmt.Errorf("invalid signal body")
	}

	index, ok := sig.Body[0].(uint32)
	if !ok {
		return LayoutEvent{}, fmt.Errorf("invalid index type")
	}

	// Получаем полную информацию о раскладке
	layouts, err := w.getLayoutsList()
	if err != nil {
		return LayoutEvent{Index: index}, nil
	}

	layout := ""
	name := ""
	if int(index) < len(layouts) {
		info := layouts[index]
		layout = info.Code
		name = info.Name
	}

	return LayoutEvent{
		Index:  index,
		Layout: layout,
		Name:   name,
	}, nil
}

// Close закрывает соединение
func (w *KDELayoutWatcher) Close() error {
	if w.cancel != nil {
		w.cancel()
	}
	return w.conn.Close()
}
