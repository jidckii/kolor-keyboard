#!/bin/sh
# Post-installation script for kolor-keyboard

# Reload udev rules
if command -v udevadm >/dev/null 2>&1; then
    udevadm control --reload-rules 2>/dev/null || true
    udevadm trigger 2>/dev/null || true
fi

echo ""
echo "kolor-keyboard installed successfully!"
echo ""
echo "Quick start:"
echo "  1. Generate config: kolor-keyboard discover --global"
echo "  2. Run manually:    kolor-keyboard run"
echo "  3. Or as service:   kolor-keyboard service install"
echo ""
echo "Note: You may need to replug your keyboard for udev rules to take effect."
echo ""
