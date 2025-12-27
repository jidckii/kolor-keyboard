package discover

import (
	"testing"
)

func TestLookupKeyboard(t *testing.T) {
	tests := []struct {
		name      string
		vid       uint16
		pid       uint16
		wantFound bool
		wantModel string
	}{
		{
			name:      "Keychron V3 ANSI encoder",
			vid:       0x3434,
			pid:       0x0331,
			wantFound: true,
			wantModel: "v3",
		},
		{
			name:      "Keychron V3 ISO",
			vid:       0x3434,
			pid:       0x0321,
			wantFound: true,
			wantModel: "v3",
		},
		{
			name:      "Keychron Q1 ANSI",
			vid:       0x3434,
			pid:       0x0100,
			wantFound: true,
			wantModel: "q1",
		},
		{
			name:      "Unknown keyboard",
			vid:       0x1234,
			pid:       0x5678,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb := LookupKeyboard(tt.vid, tt.pid)
			if tt.wantFound {
				if kb == nil {
					t.Errorf("LookupKeyboard() = nil, want keyboard")
					return
				}
				if kb.Model != tt.wantModel {
					t.Errorf("LookupKeyboard().Model = %v, want %v", kb.Model, tt.wantModel)
				}
			} else {
				if kb != nil {
					t.Errorf("LookupKeyboard() = %v, want nil", kb)
				}
			}
		})
	}
}

func TestDetectVariant(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		wantLayout  string
		wantEncoder bool
	}{
		{
			name:        "ANSI default",
			productName: "Keychron V3",
			wantLayout:  "ansi",
			wantEncoder: false,
		},
		{
			name:        "ISO explicit",
			productName: "Keychron V3 ISO",
			wantLayout:  "iso",
			wantEncoder: false,
		},
		{
			name:        "JIS explicit",
			productName: "Keychron V3 JIS",
			wantLayout:  "jis",
			wantEncoder: false,
		},
		{
			name:        "with encoder",
			productName: "Keychron V3 with Encoder",
			wantLayout:  "ansi",
			wantEncoder: true,
		},
		{
			name:        "with knob",
			productName: "Keychron Q1 Knob",
			wantLayout:  "ansi",
			wantEncoder: true,
		},
		{
			name:        "ISO with rotary",
			productName: "Some Keyboard ISO rotary",
			wantLayout:  "iso",
			wantEncoder: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variant := DetectVariant(tt.productName)
			if variant.Layout != tt.wantLayout {
				t.Errorf("DetectVariant().Layout = %v, want %v", variant.Layout, tt.wantLayout)
			}
			if variant.HasEncoder != tt.wantEncoder {
				t.Errorf("DetectVariant().HasEncoder = %v, want %v", variant.HasEncoder, tt.wantEncoder)
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "Keychron", "keychron"},
		{"with spaces", "My Keyboard", "my-keyboard"},
		{"with special chars", "Keyboard™ Pro!", "keyboard-pro"},
		{"multiple spaces", "Some  Keyboard", "some-keyboard"},
		{"leading/trailing", "  test  ", "test"},
		{"empty", "", "unknown"},
		{"only special", "!!!@@@", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractModel(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		want        string
	}{
		{"V3", "Keychron V3", "v3"},
		{"V10", "Keychron V10 Max", "v10"},
		{"Q1", "Keychron Q1 Pro", "q1"},
		{"Q2", "Q2 HE", "q2"},
		{"K2", "Keychron K2 Pro", "k2"},
		{"K8", "K8 RGB", "k8"},
		{"One2", "Ducky One2 Mini", "one2"},
		{"no model", "Generic Keyboard", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractModel(tt.productName)
			if got != tt.want {
				t.Errorf("extractModel(%q) = %q, want %q", tt.productName, got, tt.want)
			}
		})
	}
}

func TestGenerateConfigPath(t *testing.T) {
	tests := []struct {
		name string
		dev  DeviceInfo
		want string
	}{
		{
			name: "known keyboard",
			dev: DeviceInfo{
				VendorID:     0x3434,
				ProductID:    0x0331,
				Manufacturer: "Keychron",
				Product:      "V3",
			},
			want: "keychron/v3/ansi_encoder/config.yaml",
		},
		{
			name: "known keyboard ISO",
			dev: DeviceInfo{
				VendorID:     0x3434,
				ProductID:    0x0321,
				Manufacturer: "Keychron",
				Product:      "V3 ISO",
			},
			want: "keychron/v3/iso/config.yaml",
		},
		{
			name: "unknown with model pattern",
			dev: DeviceInfo{
				VendorID:     0x1234,
				ProductID:    0x5678,
				Manufacturer: "SomeVendor",
				Product:      "SomeBoard V5",
			},
			want: "somevendor/v5/ansi/config.yaml",
		},
		{
			name: "unknown no model",
			dev: DeviceInfo{
				VendorID:     0x1234,
				ProductID:    0x5678,
				Manufacturer: "Vendor",
				Product:      "Keyboard",
			},
			want: "vendor/keyboard/config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateConfigPath(&tt.dev)
			if got != tt.want {
				t.Errorf("GenerateConfigPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateConfigDir(t *testing.T) {
	dev := DeviceInfo{
		VendorID:     0x3434,
		ProductID:    0x0331,
		Manufacturer: "Keychron",
		Product:      "V3",
	}

	got := GenerateConfigDir(&dev)
	want := "keychron/v3/ansi_encoder"

	if got != want {
		t.Errorf("GenerateConfigDir() = %q, want %q", got, want)
	}
}

func TestGetKeyboardInfo(t *testing.T) {
	tests := []struct {
		name        string
		dev         DeviceInfo
		wantVendor  string
		wantModel   string
		wantVariant string
	}{
		{
			name: "known keyboard",
			dev: DeviceInfo{
				VendorID:     0x3434,
				ProductID:    0x0331,
				Manufacturer: "Keychron",
				Product:      "V3",
			},
			wantVendor:  "keychron",
			wantModel:   "v3",
			wantVariant: "ansi_encoder",
		},
		{
			name: "unknown keyboard with model",
			dev: DeviceInfo{
				VendorID:     0x1234,
				ProductID:    0x5678,
				Manufacturer: "CustomVendor",
				Product:      "K5 Pro ISO Encoder",
			},
			wantVendor:  "customvendor",
			wantModel:   "k5",
			wantVariant: "iso_encoder",
		},
		{
			name: "unknown keyboard no model",
			dev: DeviceInfo{
				VendorID:     0x1234,
				ProductID:    0x5678,
				Manufacturer: "TestCo",
				Product:      "MyBoard",
			},
			wantVendor:  "testco",
			wantModel:   "myboard",
			wantVariant: "ansi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vendor, model, variant := GetKeyboardInfo(&tt.dev)
			if vendor != tt.wantVendor {
				t.Errorf("GetKeyboardInfo() vendor = %q, want %q", vendor, tt.wantVendor)
			}
			if model != tt.wantModel {
				t.Errorf("GetKeyboardInfo() model = %q, want %q", model, tt.wantModel)
			}
			if variant != tt.wantVariant {
				t.Errorf("GetKeyboardInfo() variant = %q, want %q", variant, tt.wantVariant)
			}
		})
	}
}
