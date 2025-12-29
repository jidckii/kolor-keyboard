# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Язык общения

Всегда отвечай на русском языке.

## Безопасность

- **ЗАПРЕЩЕНО** читать файлы `.env`, `.env.*`, `*.env`
- **ЗАПРЕЩЕНО** читать переменные окружения содержащие `TOKEN`, `SECRET`, `KEY`, `PASSWORD`
- **ЗАПРЕЩЕНО** выводить или логировать значения этих переменных
- **ЗАПРЕЩЕНО** выводить или использовать переменные окружения содержащие `TOKEN`, `SECRET`, `KEY`, `PASSWORD`

## Команды сборки и разработки

```bash
make build          # Собрать бинарник ./kolor-keyboard
make test           # Запустить тесты с race detector и coverage
make lint           # Запустить golangci-lint
make install        # Установить в ~/.local/bin/
make run            # Запуск в debug режиме
make discover       # Интерактивное обнаружение клавиатур
make install-udev   # Установить udev правила для HID доступа (требует sudo)
make release-snapshot  # GoReleaser snapshot сборка
```

Запуск одного теста: `go test -v ./pkg/config -run TestConfigLoad`

## Архитектура

**CLI слой** (`cmd/kolor-keyboard/`)
- CLI на базе Cobra с командами: `run`, `discover`, `service`, `version`
- Точка входа инжектит version/commit через ldflags при сборке

**Основное приложение** (`pkg/app/`)
- Главный цикл демона: отслеживает смену раскладки через D-Bus, обновляет LED клавиатуры
- Два режима: `mono` (глобальный цвет) и `draw` (per-key RGB паттерны)
- Graceful shutdown через SIGINT/SIGTERM

**Конфигурация** (`pkg/config/`)
- YAML конфигурация с автоопределением устройства и прошивки
- Формат цвета: RGB (0-255) или HSV (0-255, стиль QMK/Vial)
- Примеры конфигов в `keyboards/keychron/v3/ansi_encoder/`

**D-Bus интеграция** (`pkg/dbus/`)
- Только KDE Plasma 6: интерфейс `org.kde.keyboard`
- Слушает сигнал `layoutChanged` при смене раскладки

**HID протокол** (`pkg/hid/`)
- VIA/Vial совместимые клавиатуры (Usage Page `0xFF60`, Usage `0x61`)
- Stock прошивка: VIA RGB Matrix команды (Report ID `0x04`)
- Vial прошивка: Direct Control режим (Report ID `0x07`)
- Размер пакета: 32 байта

**Обнаружение** (`pkg/discover/`)
- Интерактивное обнаружение клавиатур и маппинг LED по рядам
- Генерирует стартовые конфиги для stock/vial mono/draw режимов

## Ключевые детали реализации

- **Автоопределение прошивки**: отправляет Vial magic bytes, по ответу определяет stock или vial
- **Vial Direct Mode**: должен быть включён перед per-key RGB управлением, отключается при выходе
- **LED маппинг**: клавиатуры определяют индексы LED по рядам в конфиге (`keyboard.rows`)
- **CGO обязателен**: библиотека `go-hid` требует `libhidapi-dev` и `CGO_ENABLED=1`

## Рабочий процесс

1. Загрузка конфига (или дефолты с автоопределением)
2. Открытие HID устройства, определение типа прошивки
3. Инициализация режима (яркость, включение Vial Direct если нужно)
4. Получение текущей раскладки из D-Bus
5. Применение цвета/паттерна на LED
6. Цикл: ожидание `layoutChanged`, обновление LED
7. Очистка при завершении (отключение Vial Direct Mode)
