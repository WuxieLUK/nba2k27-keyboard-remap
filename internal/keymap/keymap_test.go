package keymap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseBasic(t *testing.T) {
	cases := map[string]uint16{
		"A": VK_A, "Z": VK_Z, "0": VK_0, "9": VK_9,
		"SPACE": VK_SPACE, "TAB": VK_TAB, "ENTER": VK_RETURN,
		"LSHIFT": VK_LSHIFT, "RSHIFT": VK_RSHIFT,
		"LCTRL": VK_LCONTROL, "RCTRL": VK_RCONTROL,
		"LALT": VK_LMENU, "RALT": VK_RMENU,
		"UP": VK_UP, "DOWN": VK_DOWN, "LEFT": VK_LEFT, "RIGHT": VK_RIGHT,
		"PGUP": VK_PRIOR, "PGDN": VK_NEXT,
		"HOME": VK_HOME, "END": VK_END, "INSERT": VK_INSERT, "DELETE": VK_DELETE,
		"NUMPAD0": VK_NUMPAD0, "NUMPAD9": VK_NUMPAD9,
		"NUMPAD+": VK_ADD, "NUMPAD-": VK_SUBTRACT,
		"NUMPAD*": VK_MULTIPLY, "NUMPAD/": VK_DIVIDE,
		"NUMPADENTER": VK_RETURN,
		"F1": VK_F1, "F12": VK_F12, "F24": VK_F24,
		"COMMA": VK_OEM_COMMA, "PERIOD": VK_OEM_PERIOD,
		"-": VK_OEM_MINUS, "=": VK_OEM_PLUS, "[": VK_OEM_4,
		"]": VK_OEM_6, "\\": VK_OEM_5, ";": VK_OEM_1,
		"'": VK_OEM_7, "/": VK_OEM_2, "`": VK_OEM_3,
	}
	for name, want := range cases {
		k, err := Parse(name)
		if err != nil {
			t.Errorf("Parse(%q) unexpected error: %v", name, err)
			continue
		}
		if k.VK != want {
			t.Errorf("Parse(%q).VK = %#x, want %#x", name, k.VK, want)
		}
	}
}

func TestParseAliases(t *testing.T) {
	aliases := map[string]string{
		"PAGEUP": "PGUP", "PAGEDOWN": "PGDN",
		"INS": "INSERT", "DEL": "DELETE",
		"NUMPADADD": "NUMPAD+", "NUMPADMULTIPLY": "NUMPAD*",
	}
	for alias, canonical := range aliases {
		a, err := Parse(alias)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", alias, err)
			continue
		}
		b, err := Parse(canonical)
		if err != nil {
			continue
		}
		if !Same(a, b) {
			t.Errorf("alias %q should match %q", alias, canonical)
		}
	}
}

func TestParseUnknown(t *testing.T) {
	if _, err := Parse("NOT_A_KEY"); err == nil {
		t.Error("expected error for unknown key")
	}
	if _, err := Parse(""); err == nil {
		t.Error("expected error for empty key")
	}
}

func TestFromCode(t *testing.T) {
	cases := []struct {
		vk     uint16
		scan   uint8
		extend bool
		name   string
	}{
		{VK_A, 0x1E, false, "A"},
		{VK_SPACE, 0x39, false, "SPACE"},
		{VK_RETURN, 0x1C, false, "ENTER"},
		{VK_RETURN, 0x1C, true, "NUMPADENTER"},
		{VK_LSHIFT, 0x2A, false, "LSHIFT"},
		{VK_RSHIFT, 0x36, false, "RSHIFT"},
		{VK_UP, 0x48, true, "UP"},
		{VK_ADD, 0x4E, false, "NUMPAD+"},
		{VK_OEM_COMMA, 0x33, false, "COMMA"},
		{VK_NUMPAD0, 0x52, false, "NUMPAD0"},
		{VK_F8, 0x42, false, "F8"},
	}
	for _, c := range cases {
		k, ok := FromCode(c.vk, c.scan, c.extend)
		if !ok {
			t.Errorf("FromCode(%#x,%#x,%v) not found", c.vk, c.scan, c.extend)
			continue
		}
		if k.Name != c.name {
			t.Errorf("FromCode(%#x,%#x,%v).Name = %q, want %q",
				c.vk, c.scan, c.extend, k.Name, c.name)
		}
	}
	// 无法识别的 VK（如鼠标键）返回 false。
	if _, ok := FromCode(0x01, 0, false); ok {
		t.Error("mouse key should not be recognized")
	}
}

func TestReserved(t *testing.T) {
	for _, name := range []string{"F8", "F9", "F10"} {
		k, _ := Parse(name)
		if !IsReserved(k) {
			t.Errorf("%s should be reserved", name)
		}
	}
	k, _ := Parse("F7")
	if IsReserved(k) {
		t.Error("F7 should not be reserved")
	}
}

func TestModifier(t *testing.T) {
	for _, name := range []string{"LSHIFT", "RSHIFT", "LCTRL", "RCTRL", "LALT", "RALT"} {
		k, _ := Parse(name)
		if !IsModifier(k) {
			t.Errorf("%s should be modifier", name)
		}
	}
	k, _ := Parse("A")
	if IsModifier(k) {
		t.Error("A should not be modifier")
	}
}

func TestDefaultBindings(t *testing.T) {
	cfg := DefaultConfig()
	if len(cfg.Bindings) != len(Bindings) {
		t.Fatalf("default bindings = %d, want %d", len(cfg.Bindings), len(Bindings))
	}
	for _, b := range Bindings {
		name, ok := cfg.Bindings[b.ID]
		if !ok {
			t.Errorf("missing default for %s", b.ID)
			continue
		}
		if _, err := Parse(name); err != nil {
			t.Errorf("default for %s invalid: %s (%v)", b.ID, name, err)
		}
	}
}

func TestConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bindings.ini")

	cfg := DefaultConfig()
	cfg.GameProcess = "NBA2K27.exe"
	cfg.OnlyInGame = true
	cfg.StartEnabled = false
	cfg.Bindings["SHOOT"] = "J"
	cfg.Bindings["PAUSE_MENU"] = "PGDN"

	if err := SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	got, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got.GameProcess != cfg.GameProcess {
		t.Errorf("GameProcess = %q, want %q", got.GameProcess, cfg.GameProcess)
	}
	if got.OnlyInGame != cfg.OnlyInGame {
		t.Errorf("OnlyInGame = %v, want %v", got.OnlyInGame, cfg.OnlyInGame)
	}
	if got.StartEnabled != cfg.StartEnabled {
		t.Errorf("StartEnabled = %v, want %v", got.StartEnabled, cfg.StartEnabled)
	}
	for id, name := range cfg.Bindings {
		if got.Bindings[id] != name {
			t.Errorf("Bindings[%s] = %q, want %q", id, got.Bindings[id], name)
		}
	}
}

func TestLoadMissingConfig(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "nope.ini"))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Bindings) != len(Bindings) {
		t.Errorf("missing config should fall back to defaults")
	}
}

func TestLoadLegacyConfig(t *testing.T) {
	// 模拟原版 v5 生成的配置文件（含旧版键名 PGUP/PGDN/NUMPADENTER）。
	content := `; NBA 2K27 Keyboard Remap v5.0 - press-to-bind GUI
; 本文件由图形界面自动生成

[General]
GameProcess=NBA2K27.exe
OnlyInGame=1
StartEnabled=1

[Bindings]
MOVE_UP=W
MOVE_LEFT=A
MOVE_DOWN=S
MOVE_RIGHT=D
SHOOT=SPACE
PASS=E
BOUNCE_PASS=Q
LOB_ALLEY=R
SPRINT=NUMPADENTER
POST_UP=LSHIFT
PICK_TACTIC=TAB
ICON_PASS=NUMPAD+
PRO_STICK_UP=0
PRO_STICK_DOWN=COMMA
PRO_STICK_LEFT=I
PRO_STICK_RIGHT=P
PAUSE_MENU=PGDN
TIMEOUT=PGUP
ARROW_UP=UP
ARROW_DOWN=DOWN
ARROW_LEFT=LEFT
ARROW_RIGHT=RIGHT
`
	path := filepath.Join(t.TempDir(), "bindings.ini")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	want := map[string]string{
		"SPRINT":     "NUMPADENTER",
		"PAUSE_MENU": "PGDN",
		"TIMEOUT":    "PGUP",
		"ICON_PASS":  "NUMPAD+",
		"PRO_STICK_DOWN": "COMMA",
	}
	for id, name := range want {
		if cfg.Bindings[id] != name {
			t.Errorf("legacy Bindings[%s] = %q, want %q", id, cfg.Bindings[id], name)
		}
	}
	if !cfg.OnlyInGame || !cfg.StartEnabled {
		t.Errorf("legacy general flags not parsed: %+v", cfg)
	}
	if cfg.GameProcess != "NBA2K27.exe" {
		t.Errorf("GameProcess = %q", cfg.GameProcess)
	}
}
