.PHONY: build install uninstall clean enable disable status test discover coverage-html

BINARY := kolor-keyboard
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

PREFIX := $(HOME)/.local
CONFIG_DIR := $(HOME)/.config/kolor-keyboard
SYSTEMD_USER_DIR := $(HOME)/.config/systemd/user

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/kolor-keyboard

install: build
	@echo "Installing $(BINARY)..."
	install -Dm755 $(BINARY) $(PREFIX)/bin/$(BINARY)
	@echo "Installing example config..."
	mkdir -p $(CONFIG_DIR)
	install -Dm644 keyboards/keychron/v3/ansi_encoder/vial_draw.yaml $(CONFIG_DIR)/config.yaml
	@echo "Config installed to $(CONFIG_DIR)/config.yaml"
	@echo "Installing systemd service..."
	mkdir -p $(SYSTEMD_USER_DIR)
	install -Dm644 scripts/kolor-keyboard.service $(SYSTEMD_USER_DIR)/kolor-keyboard.service
	systemctl --user daemon-reload
	@echo ""
	@echo "Installation complete!"
	@echo "To enable and start the service, run:"
	@echo "  make enable"

enable:
	systemctl --user enable --now kolor-keyboard.service
	@echo "Service enabled and started"

disable:
	systemctl --user disable --now kolor-keyboard.service || true
	@echo "Service disabled"

status:
	systemctl --user status kolor-keyboard.service

logs:
	journalctl --user -u kolor-keyboard.service -f

uninstall: disable
	rm -f $(PREFIX)/bin/$(BINARY)
	rm -f $(SYSTEMD_USER_DIR)/kolor-keyboard.service
	systemctl --user daemon-reload
	@echo "Uninstalled. Config left at $(CONFIG_DIR)"

clean:
	rm -f $(BINARY)
	go clean

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML coverage report: make coverage-html"

coverage-html: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Install udev rules (requires sudo)
install-udev:
	@echo "Installing udev rules..."
	sudo wget -O /etc/udev/rules.d/50-qmk.rules \
		https://raw.githubusercontent.com/qmk/qmk_firmware/master/util/udev/50-qmk.rules
	@echo "Adding Keychron V3 rule..."
	echo 'SUBSYSTEM=="hidraw", ATTRS{idVendor}=="3434", ATTRS{idProduct}=="0331", TAG+="uaccess"' | \
		sudo tee /etc/udev/rules.d/51-keychron.rules
	sudo udevadm control --reload-rules
	sudo udevadm trigger
	@echo "udev rules installed. You may need to replug the keyboard."

# Debug: show HID devices
list-hid:
	@echo "Looking for Keychron devices..."
	@lsusb | grep -i keychron || echo "No Keychron found in lsusb"
	@echo ""
	@echo "HID devices:"
	@ls -la /dev/hidraw* 2>/dev/null || echo "No hidraw devices"

# Debug: test D-Bus
test-dbus:
	@echo "Testing KDE keyboard D-Bus interface..."
	qdbus6 org.kde.keyboard /Layouts org.kde.KeyboardLayouts.getLayout
	@echo ""
	@echo "Layouts list:"
	qdbus6 org.kde.keyboard /Layouts org.kde.KeyboardLayouts.getLayoutsList

# Run in debug mode
run-debug: build
	./$(BINARY) run --debug -c keyboards/keychron/v3/ansi_encoder/vial_draw.yaml

# Discover keyboard and generate config
discover: build
	./$(BINARY) discover
