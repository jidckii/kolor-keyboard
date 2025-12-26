# kbd-flag: Индикация раскладки через RGB

## Статус: РАБОТАЕТ

Утилита на Go для индикации раскладки клавиатуры через RGB подсветку Keychron V3.

---

## Что работает

- Отслеживание смены раскладки через KDE D-Bus
- Изменение глобального цвета клавиатуры через VIA RGB Matrix протокол
- Конфигурация цветов для разных раскладок

## Структура проекта

```
kbd-flag/
├── cmd/kbd-flag/main.go           # Точка входа
├── internal/
│   ├── config/
│   │   ├── config.go              # Загрузка конфига
│   │   └── types.go               # Типы конфигурации
│   ├── dbus/
│   │   └── keyboard.go            # KDE D-Bus watcher
│   ├── hid/
│   │   ├── device.go              # HID устройство
│   │   └── protocol.go            # VIA RGB Matrix протокол
│   └── app/app.go                 # Главное приложение
├── configs/
│   └── keychron_v3_ansi.yaml      # Конфигурация
├── scripts/
│   └── kbd-flag.service           # systemd unit
├── go.mod
└── Makefile
```

## Использование

```bash
# Сборка
go build -o kbd-flag ./cmd/kbd-flag

# Запуск
./kbd-flag -config configs/keychron_v3_ansi.yaml

# С отладкой
./kbd-flag -debug -config configs/keychron_v3_ansi.yaml
```

## Конфигурация

```yaml
device:
  vendor_id: 0x3434
  product_id: 0x0331
  usage_page: 0xFF60
  usage: 0x61

colors:
  - layout: ru
    color: {r: 255, g: 0, b: 0}    # Красный
  - layout: us
    color: {r: 0, g: 0, b: 255}    # Синий
  - layout: "*"
    color: {r: 255, g: 255, b: 255} # Белый (fallback)
```

## Зависимости

- `github.com/godbus/dbus/v5` — D-Bus для KDE
- `github.com/sstallion/go-hid` — HID устройства (требует CGO)
- `gopkg.in/yaml.v3` — конфигурация

### Для сборки

```bash
# openSUSE
sudo zypper install systemd-devel

# Debian/Ubuntu
sudo apt install libudev-dev
```

## VIA RGB Matrix протокол

Keychron V3 поддерживает стандартный VIA RGB Matrix:

```
# Set effect (1 = solid color)
0x07, 0x03, 0x02, effect

# Set color (hue, saturation)
0x07, 0x03, 0x04, hue, sat
```

**Важно:** Vial RGB (0x42 для индивидуальных LED) НЕ поддерживается.
Работает только глобальный цвет.

## Установка как systemd service

```bash
# Скопировать бинарник
sudo cp kbd-flag /usr/local/bin/

# Скопировать конфиг
mkdir -p ~/.config/kbd-flag
cp configs/keychron_v3_ansi.yaml ~/.config/kbd-flag/config.yaml

# Скопировать и включить service
cp scripts/kbd-flag.service ~/.config/systemd/user/
systemctl --user enable kbd-flag
systemctl --user start kbd-flag
```

## Известные ограничения

1. **Только глобальный цвет** — нельзя раскрасить отдельные клавиши (Vial RGB не поддерживается прошивкой)
2. **Только прямое подключение** — через USB-хаб может не работать
3. **Только KDE Plasma** — используется KDE D-Bus API

## Возможные улучшения

### Функциональность
- [ ] Поддержка GNOME через `org.gnome.Shell` D-Bus API
- [ ] Поддержка Sway/wlroots через `wlr-input-method`
- [ ] Автоматическое определение устройства по VID/PID
- [ ] Плавные переходы между цветами (fade animation)
- [ ] Поддержка нескольких клавиатур одновременно
- [ ] Горячая перезагрузка конфигурации (SIGHUP)
- [ ] Tray иконка с текущей раскладкой

### Per-key RGB (флаги на клавиатуре)

Текущая прошивка Keychron V3 **не поддерживает** per-key RGB control через HID.
Keychron Launcher управляет только предустановленными эффектами и глобальным цветом.

Чтобы раскрашивать отдельные клавиши разными цветами, нужно прошить QMK с Vial.

**Подробная инструкция:** [docs/FIRMWARE.md](docs/FIRMWARE.md)

После прошивки Vial станут доступны команды:
```
0x07, 0x42, start_lo, start_hi, count, H, S, V, ...  # Direct LED control
0x08, 0x43                                            # Get LED count
```

### Другие клавиатуры

- [ ] Поддержка других VIA-совместимых клавиатур
- [ ] Поддержка Razer через OpenRazer
- [ ] Поддержка Logitech через libratbag
- [ ] Поддержка SteelSeries через rivalcfg
