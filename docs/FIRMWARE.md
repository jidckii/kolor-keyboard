
# Прошивка Keychron V3 с Vial RGB

Инструкция по прошивке клавиатуры Keychron V3 прошивкой QMK с поддержкой Vial RGB для per-key RGB control.

## Зачем нужна прошивка Vial

Стоковая прошивка Keychron поддерживает только:
- Предустановленные RGB эффекты (22 штуки)
- Глобальный цвет для всех клавиш

Прошивка Vial добавляет:
- **Per-key RGB control** — управление цветом каждой клавиши отдельно
- Полная совместимость с VIA/Vial приложениями
- Открытый исходный код

## Требования

- Keychron V3 ANSI с энкодером (Knob)
- Linux (openSUSE Tumbleweed или другой дистрибутив)
- USB кабель
- ~1GB свободного места

## 1. Установка зависимостей

### QMK CLI

```bash
# Через pip
pip install qmk

# Или через pipx (рекомендуется)
pipx install qmk

# Проверка
qmk --version
```

### ARM GNU Toolchain 14.2.rel1

QMK не поддерживает автоустановку для openSUSE, поэтому устанавливаем вручную:

```bash
# Скачать ARM GNU Toolchain
cd /tmp
wget https://developer.arm.com/-/media/Files/downloads/gnu/14.2.rel1/binrel/arm-gnu-toolchain-14.2.rel1-x86_64-arm-none-eabi.tar.xz

# Распаковать
tar xf arm-gnu-toolchain-14.2.rel1-x86_64-arm-none-eabi.tar.xz

# Добавить в PATH для текущей сессии
export PATH="/tmp/arm-gnu-toolchain-14.2.rel1-x86_64-arm-none-eabi/bin:$PATH"

# Проверить
arm-none-eabi-gcc --version
# Должно показать: arm-none-eabi-gcc (Arm GNU Toolchain 14.2.Rel1) 14.2.1
```

#### Постоянная установка (опционально)

```bash
# Переместить в /opt
sudo mv /tmp/arm-gnu-toolchain-14.2.rel1-x86_64-arm-none-eabi /opt/

# Добавить в ~/.bashrc или ~/.zshrc
echo 'export PATH="/opt/arm-gnu-toolchain-14.2.rel1-x86_64-arm-none-eabi/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### dfu-util

```bash
# openSUSE
sudo zypper install dfu-util

# Debian/Ubuntu
sudo apt install dfu-util

# Проверка
dfu-util --version
```

## 2. Клонирование Vial QMK

```bash
# Клонировать репозиторий
cd /tmp
git clone --depth 1 https://github.com/vial-kb/vial-qmk.git
cd vial-qmk

# Инициализировать submodules (обязательно!)
make git-submodule
# Или вручную:
# git submodule update --init --recursive
```

## 3. Проверка наличия клавиатуры

```bash
# Проверить что Keychron V3 есть в списке
ls keyboards/keychron/v3/

# Должны быть папки:
# ansi/  ansi_encoder/  iso/  iso_encoder/  jis/  jis_encoder/

# Проверить keymap vial
ls keyboards/keychron/v3/ansi_encoder/keymaps/vial/
# config.h  keymap.c  rules.mk  vial.json
```

## 4. Компиляция прошивки

```bash
cd /tmp/vial-qmk

# Компиляция для Keychron V3 ANSI с энкодером
qmk compile -kb keychron/v3/ansi_encoder -km vial

# Результат:
# keychron_v3_ansi_encoder_vial.bin (~64KB)
```

### Варианты клавиатур

| Модель | Команда |
|--------|---------|
| V3 ANSI | `qmk compile -kb keychron/v3/ansi -km vial` |
| V3 ANSI Knob | `qmk compile -kb keychron/v3/ansi_encoder -km vial` |
| V3 ISO | `qmk compile -kb keychron/v3/iso -km vial` |
| V3 ISO Knob | `qmk compile -kb keychron/v3/iso_encoder -km vial` |
| V3 JIS | `qmk compile -kb keychron/v3/jis -km vial` |
| V3 JIS Knob | `qmk compile -kb keychron/v3/jis_encoder -km vial` |

## 5. Перевод клавиатуры в DFU режим

1. **Отключи** клавиатуру от USB
2. **Зажми** клавишу **Esc**
3. **Подключи** USB кабель (держи Esc)
4. **Отпусти** Esc через 1-2 секунды

### Проверка DFU режима

```bash
lsusb | grep -i "stm\|dfu"
# Должно показать:
# Bus 001 Device 042: ID 0483:df11 STMicroelectronics STM Device in DFU Mode
```

Если не видно — попробуй ещё раз или проверь права:

```bash
# Добавить udev правило для STM32 DFU
sudo tee /etc/udev/rules.d/50-stm32-dfu.rules << 'EOF'
# STM32 DFU
SUBSYSTEMS=="usb", ATTRS{idVendor}=="0483", ATTRS{idProduct}=="df11", MODE="0666"
EOF

sudo udevadm control --reload-rules
sudo udevadm trigger
```

## 6. Прошивка

```bash
cd /tmp/vial-qmk

# Прошить
qmk flash -kb keychron/v3/ansi_encoder -km vial

# Или напрямую через dfu-util:
dfu-util -a 0 -d 0483:df11 -s 0x08000000:leave -D keychron_v3_ansi_encoder_vial.bin
```

**Важно:** Не отключай клавиатуру во время прошивки!

После прошивки клавиатура автоматически перезагрузится.

## 7. Проверка

После прошивки:

1. Клавиатура должна работать как обычно
2. Keychron Launcher больше **не будет работать**
3. Используй [Vial](https://get.vial.today/) для настройки

### Проверка Vial RGB

```bash
# Проверить что Vial RGB команды работают
hidapitester --vidpid 3434:0331 --usagePage 0xFF60 --usage 0x61 --open \
  --send-output 0x08,0x43 --read-input

# Должен вернуть количество LED (не 0x00)
```

## Откат на стоковую прошивку

Если что-то пошло не так:

1. Скачай стоковую прошивку: https://www.keychron.com/pages/firmware-and-json-files-of-the-keychron-v3
2. Переведи клавиатуру в DFU режим (Esc + подключение)
3. Прошей:

```bash
dfu-util -a 0 -d 0483:df11 -s 0x08000000:leave -D keychron_v3_ansi_encoder_default.bin
```

## Риски и ограничения

### Риски
- Можно временно "окирпичить" клавиатуру (восстанавливается через DFU)
- Потеря гарантии Keychron

### Ограничения Vial прошивки
- **Bluetooth не работает** — только USB
- Keychron Launcher не работает — используй Vial

### Что сохраняется
- DFU bootloader (всегда доступен через Esc + подключение)
- Физический энкодер работает

## Ссылки

- [ARM GNU Toolchain Downloads](https://developer.arm.com/downloads/-/arm-gnu-toolchain-downloads)
- [Vial QMK](https://github.com/vial-kb/vial-qmk)
- [QMK Documentation](https://docs.qmk.fm/)
- [Vial](https://get.vial.today/)
- [Keychron Firmware](https://www.keychron.com/pages/firmware-and-json-files-of-the-keychron-v3)