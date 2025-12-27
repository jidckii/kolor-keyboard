# kolor-keyboard: Индикация раскладки через RGB

Утилита на Go для индикации раскладки клавиатуры через RGB подсветку.

<p align="center">
  <img src="img/example_mono.gif" alt="Режим Mono" width="45%">
  <img src="img/example_flag.gif" alt="Режим Draw" width="45%">
</p>
<p align="center">
  <b>Mono</b> — глобальный цвет &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
  <b>Draw</b> — per-key RGB (флаги, рисунки)
</p>

> ⚠️ **Дисклеймер**
>
> Этот проект полностью сгенерирован с помощью [Claude Code](https://claude.com/claude-code).
> Использование на свой страх и риск. Автор проекта не несёт ответственности за любые
> последствия, включая возможные проблемы с вашим аппаратным или программным обеспечением.

## Поддерживаемые клавиатуры

| Клавиатура | Раскладка | Энкодер | Статус |
|------------|-----------|---------|--------|
| Keychron V3 | ANSI | ✅ | Протестировано |

## Поддерживаемые платформы

| ОС | DE | Статус |
|----|-----|--------|
| Linux | KDE Plasma 6 | Протестировано (openSUSE Tumbleweed) |
| Linux | KDE Plasma 5 | Должно работать |
| Linux | GNOME, Sway и др. | Не поддерживается (планируется) |

## Поддерживаемые прошивки

| Прошивка | Режим Mono | Режим Draw |
|----------|------------|------------|
| **Stock** (QMK/VIA) | ✅ | ❌ |
| **Vial** | ✅ | ✅ |

### Режимы работы:
- **Mono** — глобальный цвет для всей клавиатуры
- **Draw** — per-key RGB для отрисовки флагов стран (только Vial)

---

## Быстрый старт

### 1. Сборка

```bash
git clone https://github.com/jidckii/kolor-keyboard.git
cd kolor-keyboard
make build
```

### 2. Генерация конфигурации для вашей клавиатуры

```bash
./kolor-keyboard discover
```

Команда `discover` автоматически:
- Найдёт подключённые VIA/Vial клавиатуры
- Определит поддержку Vial и количество LED
- Предложит интерактивный маппинг LED по рядам (для draw режима)
- Сгенерирует конфигурационные файлы

### 3. Запуск

```bash
# С конфигом из keyboards/
./kolor-keyboard run -c keyboards/keychron/v3/ansi_encoder/vial_draw.yaml

# Или скопируйте конфиг в стандартное место
cp keyboards/keychron/v3/ansi_encoder/vial_draw.yaml ~/.config/kolor-keyboard/config.yaml
./kolor-keyboard run
```

---

## Команда discover

Команда `discover` — самый простой способ настроить kolor-keyboard для новой клавиатуры.

### Использование

```bash
# Генерация в текущую директорию (keyboards/<vendor>/<model>/<variant>/)
./kolor-keyboard discover

# Генерация в глобальный конфиг (~/.config/kolor-keyboard/)
./kolor-keyboard discover --global

# Указать output директорию
./kolor-keyboard discover -o /path/to/output
```

### Что генерируется

При запуске без `--global` создаются три файла:

```
keyboards/<vendor>/<model>/<variant>/
├── stock_mono.yaml   # Stock QMK/VIA прошивка, mono режим
├── vial_mono.yaml    # Vial прошивка, mono режим
└── vial_draw.yaml    # Vial прошивка, draw режим (если сделан LED маппинг)
```

### LED Mapping Tour

Если клавиатура поддерживает Vial, discover предложит интерактивный маппинг LED:

```
╔══════════════════════════════════════════════════════════════╗
║              LED Mapping Tour                                ║
╠══════════════════════════════════════════════════════════════╣
║  Detected 87 LEDs (indices 0-86)
║                                                              ║
║  Commands:                                                   ║
║    Enter/n  - next LED (add to current row)                  ║
║    r        - end row here (add LED and start new row)       ║
║    s        - skip this LED (don't add to any row)           ║
║    b        - go back (undo last LED)                        ║
║    q        - quit and save                                  ║
║                                                              ║
║  Colors:                                                     ║
║    RED     - current LED                                     ║
║    GREEN   - current row LEDs                                ║
║    YELLOW  - saved rows (different shades per row)           ║
╚══════════════════════════════════════════════════════════════╝
```

Процесс:
1. Текущий LED подсвечивается **красным**
2. Нажимайте `Enter` чтобы добавить LED в текущий ряд (LED станет **зелёным**)
3. На последней кнопке ряда нажмите `r` и затем `Enter` — ряд сохранится (**жёлтый**), начнётся новый
4. `s` — пропустить LED (например, энкодер или индикатор)
5. `b` — вернуться назад
6. `q` — завершить и сохранить
7. По достижении последней кнопки, просто нажмите `Enter`, конфигурация автоматически сохранится и работа discover завершится.

---

## Команды

```bash
# Показать справку
./kolor-keyboard --help

# Запустить демон
./kolor-keyboard run -c config.yaml
./kolor-keyboard run --debug -c config.yaml

# Обнаружение клавиатуры
./kolor-keyboard discover
./kolor-keyboard discover --global

# Показать версию
./kolor-keyboard version
```

### Поиск конфигурации

Команда `run` ищет конфиг в следующем порядке:
1. Путь указанный через `-c/--config`
2. `./kolor-keyboard.yaml` (текущая директория)
3. `~/.config/kolor-keyboard/config.yaml`
4. Авто-сгенерированные конфиги в `~/.config/kolor-keyboard/keyboards/`

---

## Конфигурация

### Формат цвета

Поддерживаются два формата:

```yaml
# RGB (0-255)
color: {rgb: {r: 255, g: 0, b: 0}}

# HSV (0-255, как в QMK/Vial)
color: {hsv: {h: 0, s: 255, v: 255}}
```

### Глобальные настройки

```yaml
brightness: 200  # 0-255, яркость подсветки
speed: 128       # 0-255, скорость эффектов (только Vial)
```

### Режим Mono

```yaml
device:
  vendor_id: 0x3434
  product_id: 0x0331
  usage_page: 0xFF60
  usage: 0x61

firmware: vial  # или stock
mode: mono

brightness: 200

colors:
  - layout: ru
    color: {rgb: {r: 255, g: 0, b: 0}}    # Красный
  - layout: us
    color: {rgb: {r: 0, g: 100, b: 255}}  # Синий
  - layout: "*"
    color: {rgb: {r: 0, g: 255, b: 0}}    # Fallback
```

### Режим Draw (per-key RGB)

```yaml
firmware: vial
mode: draw

brightness: 200
speed: 128

keyboard:
  rows:
    - [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15]     # Row 0
    - [16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, ...]  # Row 1
    # ...

draw:
  # Флаг России: триколор
  - layout: ru
    stripes:
      - rows: [0, 1]
        color: {rgb: {r: 255, g: 255, b: 255}}  # Белый
      - rows: [2, 3]
        color: {rgb: {r: 0, g: 50, b: 255}}     # Синий
      - rows: [4, 5]
        color: {rgb: {r: 255, g: 0, b: 0}}      # Красный

  # Английская раскладка: синий mono
  - layout: us
    stripes:
      - rows: [0, 1, 2, 3, 4, 5]
        color: {rgb: {r: 0, g: 100, b: 255}}
```

---

## Структура проекта

```
kolor-keyboard/
├── cmd/kolor-keyboard/
│   ├── main.go                    # Точка входа
│   └── cmd/                       # Cobra команды
│       ├── root.go
│       ├── run.go
│       ├── discover.go
│       └── version.go
├── pkg/
│   ├── app/app.go                 # Главное приложение
│   ├── config/                    # Загрузка и валидация конфига
│   ├── dbus/keyboard.go           # KDE D-Bus watcher
│   ├── hid/                       # HID устройство и протокол
│   └── discover/                  # Обнаружение клавиатур
├── keyboards/                     # Конфиги для известных клавиатур
│   └── keychron/v3/ansi_encoder/
│       ├── stock_mono.yaml
│       ├── vial_mono.yaml
│       └── vial_draw.yaml
├── examples/                      # Примеры конфигов
├── docs/
│   ├── FIRMWARE.md                # Инструкция по прошивке Vial
│   └── LED_MAP.md                 # Карта LED для клавиатур
├── scripts/
│   └── kolor-keyboard.service     # systemd unit
├── Makefile
└── README.md
```

---

## Установка

### Простая установка

```bash
make install
make enable
```

Это:
- Соберёт бинарник
- Установит в `~/.local/bin/`
- Скопирует пример конфига в `~/.config/kolor-keyboard/`
- Установит и запустит systemd user service

### Ручная установка

```bash
# Сборка
make build

# Копирование
sudo cp kolor-keyboard /usr/local/bin/

# Конфигурация (выберите подходящий конфиг)
mkdir -p ~/.config/kolor-keyboard
cp keyboards/keychron/v3/ansi_encoder/vial_draw.yaml ~/.config/kolor-keyboard/config.yaml

# Systemd service
cp scripts/kolor-keyboard.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now kolor-keyboard
```

### udev правила

Для доступа к HID устройствам без root:

```bash
make install-udev
```

---

## Зависимости

- `github.com/godbus/dbus/v5` — D-Bus для KDE
- `github.com/sstallion/go-hid` — HID устройства (требует CGO)
- `github.com/spf13/cobra` — CLI
- `gopkg.in/yaml.v3` — конфигурация

### Для сборки

```bash
# openSUSE
sudo zypper install systemd-devel

# Debian/Ubuntu
sudo apt install libudev-dev
```

---

## Известные ограничения

1. **Per-key RGB требует прошивку Vial** — стоковая прошивка поддерживает только глобальный цвет
2. **Только прямое подключение** — через USB-хаб может не работать
3. **Только KDE Plasma** — используется KDE D-Bus API

## Лицензия

MIT License. См. файл [LICENSE](LICENSE).
