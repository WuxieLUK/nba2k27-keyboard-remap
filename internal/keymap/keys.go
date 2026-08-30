// Package keymap 提供 NBA 2K27 键盘自定义所需的键名解析、
// 配置文件读写与默认绑定定义。
package keymap

import (
	"fmt"
	"strings"
)

// Key 描述一个可绑定的物理键。
type Key struct {
	VK     uint16 // Windows 虚拟键码
	Scan   uint8  // 硬件扫描码（Set 1）
	Extend bool   // 是否为扩展键（方向键、Insert 等）
	Name   string // 配置文件中使用的键名（大写）
	Label  string // 界面显示名
}

// Windows 虚拟键码（仅本包用到的一部分）。
const (
	VK_BACK       = 0x08
	VK_TAB        = 0x09
	VK_RETURN     = 0x0D
	VK_SHIFT      = 0x10
	VK_CONTROL    = 0x11
	VK_MENU       = 0x12
	VK_PAUSE      = 0x13
	VK_CAPITAL    = 0x14
	VK_ESCAPE     = 0x1B
	VK_SPACE      = 0x20
	VK_PRIOR      = 0x21 // PageUp
	VK_NEXT       = 0x22 // PageDown
	VK_END        = 0x23
	VK_HOME       = 0x24
	VK_LEFT       = 0x25
	VK_UP         = 0x26
	VK_RIGHT      = 0x27
	VK_DOWN       = 0x28
	VK_INSERT     = 0x2D
	VK_DELETE     = 0x2E
	VK_0          = 0x30
	VK_1          = 0x31
	VK_2          = 0x32
	VK_3          = 0x33
	VK_4          = 0x34
	VK_5          = 0x35
	VK_6          = 0x36
	VK_7          = 0x37
	VK_8          = 0x38
	VK_9          = 0x39
	VK_A          = 0x41
	VK_B          = 0x42
	VK_C          = 0x43
	VK_D          = 0x44
	VK_E          = 0x45
	VK_F          = 0x46
	VK_G          = 0x47
	VK_H          = 0x48
	VK_I          = 0x49
	VK_J          = 0x4A
	VK_K          = 0x4B
	VK_L          = 0x4C
	VK_M          = 0x4D
	VK_N          = 0x4E
	VK_O          = 0x4F
	VK_P          = 0x50
	VK_Q          = 0x51
	VK_R          = 0x52
	VK_S          = 0x53
	VK_T          = 0x54
	VK_U          = 0x55
	VK_V          = 0x56
	VK_W          = 0x57
	VK_X          = 0x58
	VK_Y          = 0x59
	VK_Z          = 0x5A
	VK_NUMPAD0    = 0x60
	VK_NUMPAD1    = 0x61
	VK_NUMPAD2    = 0x62
	VK_NUMPAD3    = 0x63
	VK_NUMPAD4    = 0x64
	VK_NUMPAD5    = 0x65
	VK_NUMPAD6    = 0x66
	VK_NUMPAD7    = 0x67
	VK_NUMPAD8    = 0x68
	VK_NUMPAD9    = 0x69
	VK_MULTIPLY   = 0x6A
	VK_ADD        = 0x6B
	VK_SUBTRACT   = 0x6D
	VK_DIVIDE     = 0x6F
	VK_F1         = 0x70
	VK_F2         = 0x71
	VK_F3         = 0x72
	VK_F4         = 0x73
	VK_F5         = 0x74
	VK_F6         = 0x75
	VK_F7         = 0x76
	VK_F8         = 0x77
	VK_F9         = 0x78
	VK_F10        = 0x79
	VK_F11        = 0x7A
	VK_F12        = 0x7B
	VK_F13        = 0x7C
	VK_F14        = 0x7D
	VK_F15        = 0x7E
	VK_F16        = 0x7F
	VK_F17        = 0x80
	VK_F18        = 0x81
	VK_F19        = 0x82
	VK_F20        = 0x83
	VK_F21        = 0x84
	VK_F22        = 0x85
	VK_F23        = 0x86
	VK_F24        = 0x87
	VK_LSHIFT     = 0xA0
	VK_RSHIFT     = 0xA1
	VK_LCONTROL   = 0xA2
	VK_RCONTROL   = 0xA3
	VK_LMENU      = 0xA4
	VK_RMENU      = 0xA5
	VK_OEM_1      = 0xBA // ;:
	VK_OEM_PLUS   = 0xBB // =+
	VK_OEM_COMMA  = 0xBC // ,<
	VK_OEM_MINUS  = 0xBD // -_
	VK_OEM_PERIOD = 0xBE // .>
	VK_OEM_2      = 0xBF // /?
	VK_OEM_3      = 0xC0 // `~
	VK_OEM_4      = 0xDB // [{
	VK_OEM_5      = 0xDC // \|
	VK_OEM_6      = 0xDD // ]}
	VK_OEM_7      = 0xDE // '"
)

// keyTable 为全部可识别键。字母/数字/符号键的 Name 即其字符本身
// （逗号、句点除外，与原版一致使用 COMMA / PERIOD）。
var keyTable = []Key{
	// ---- 字母 A-Z ----
	{VK_A, 0x1E, false, "A", "A"}, {VK_B, 0x30, false, "B", "B"},
	{VK_C, 0x2E, false, "C", "C"}, {VK_D, 0x20, false, "D", "D"},
	{VK_E, 0x12, false, "E", "E"}, {VK_F, 0x21, false, "F", "F"},
	{VK_G, 0x22, false, "G", "G"}, {VK_H, 0x23, false, "H", "H"},
	{VK_I, 0x17, false, "I", "I"}, {VK_J, 0x24, false, "J", "J"},
	{VK_K, 0x25, false, "K", "K"}, {VK_L, 0x26, false, "L", "L"},
	{VK_M, 0x32, false, "M", "M"}, {VK_N, 0x31, false, "N", "N"},
	{VK_O, 0x18, false, "O", "O"}, {VK_P, 0x19, false, "P", "P"},
	{VK_Q, 0x10, false, "Q", "Q"}, {VK_R, 0x13, false, "R", "R"},
	{VK_S, 0x1F, false, "S", "S"}, {VK_T, 0x14, false, "T", "T"},
	{VK_U, 0x16, false, "U", "U"}, {VK_V, 0x2F, false, "V", "V"},
	{VK_W, 0x11, false, "W", "W"}, {VK_X, 0x2D, false, "X", "X"},
	{VK_Y, 0x15, false, "Y", "Y"}, {VK_Z, 0x2C, false, "Z", "Z"},

	// ---- 数字 0-9 ----
	{VK_0, 0x0B, false, "0", "0"}, {VK_1, 0x02, false, "1", "1"},
	{VK_2, 0x03, false, "2", "2"}, {VK_3, 0x04, false, "3", "3"},
	{VK_4, 0x05, false, "4", "4"}, {VK_5, 0x06, false, "5", "5"},
	{VK_6, 0x07, false, "6", "6"}, {VK_7, 0x08, false, "7", "7"},
	{VK_8, 0x09, false, "8", "8"}, {VK_9, 0x0A, false, "9", "9"},

	// ---- 主键盘符号键 ----
	{VK_OEM_MINUS, 0x0C, false, "-", "-"}, {VK_OEM_PLUS, 0x0D, false, "=", "="},
	{VK_OEM_4, 0x1A, false, "[", "["}, {VK_OEM_6, 0x1B, false, "]", "]"},
	{VK_OEM_5, 0x2B, false, "\\", "\\"}, {VK_OEM_1, 0x27, false, ";", ";"},
	{VK_OEM_7, 0x28, false, "'", "'"}, {VK_OEM_COMMA, 0x33, false, "COMMA", ","},
	{VK_OEM_PERIOD, 0x34, false, "PERIOD", "."}, {VK_OEM_2, 0x35, false, "/", "/"},
	{VK_OEM_3, 0x29, false, "`", "`"},

	// ---- 常用控制键 ----
	{VK_SPACE, 0x39, false, "SPACE", "Space"},
	{VK_TAB, 0x0F, false, "TAB", "Tab"},
	{VK_RETURN, 0x1C, false, "ENTER", "Enter"},
	{VK_BACK, 0x0E, false, "BACKSPACE", "Backspace"},
	{VK_ESCAPE, 0x01, false, "ESC", "Esc"},
	{VK_CAPITAL, 0x3A, false, "CAPSLOCK", "Caps Lock"},

	// ---- 修饰键 ----
	{VK_LSHIFT, 0x2A, false, "LSHIFT", "Left Shift"},
	{VK_RSHIFT, 0x36, false, "RSHIFT", "Right Shift"},
	{VK_LCONTROL, 0x1D, true, "LCTRL", "Left Ctrl"},
	{VK_RCONTROL, 0x1D, true, "RCTRL", "Right Ctrl"},
	{VK_LMENU, 0x38, true, "LALT", "Left Alt"},
	{VK_RMENU, 0x38, true, "RALT", "Right Alt"},

	// ---- 方向键与编辑键 ----
	{VK_UP, 0x48, true, "UP", "↑"},
	{VK_DOWN, 0x50, true, "DOWN", "↓"},
	{VK_LEFT, 0x4B, true, "LEFT", "←"},
	{VK_RIGHT, 0x4D, true, "RIGHT", "→"},
	{VK_PRIOR, 0x49, true, "PGUP", "Page Up"},
	{VK_NEXT, 0x51, true, "PGDN", "Page Down"},
	{VK_HOME, 0x47, true, "HOME", "Home"},
	{VK_END, 0x4F, true, "END", "End"},
	{VK_INSERT, 0x52, true, "INSERT", "Insert"},
	{VK_DELETE, 0x53, true, "DELETE", "Delete"},

	// ---- 小键盘 ----
	{VK_NUMPAD0, 0x52, false, "NUMPAD0", "Num 0"},
	{VK_NUMPAD1, 0x4F, false, "NUMPAD1", "Num 1"},
	{VK_NUMPAD2, 0x50, false, "NUMPAD2", "Num 2"},
	{VK_NUMPAD3, 0x51, false, "NUMPAD3", "Num 3"},
	{VK_NUMPAD4, 0x4B, false, "NUMPAD4", "Num 4"},
	{VK_NUMPAD5, 0x4C, false, "NUMPAD5", "Num 5"},
	{VK_NUMPAD6, 0x4D, false, "NUMPAD6", "Num 6"},
	{VK_NUMPAD7, 0x47, false, "NUMPAD7", "Num 7"},
	{VK_NUMPAD8, 0x48, false, "NUMPAD8", "Num 8"},
	{VK_NUMPAD9, 0x49, false, "NUMPAD9", "Num 9"},
	{VK_ADD, 0x4E, false, "NUMPAD+", "Num +"},
	{VK_SUBTRACT, 0x4A, false, "NUMPAD-", "Num -"},
	{VK_MULTIPLY, 0x37, false, "NUMPAD*", "Num *"},
	{VK_DIVIDE, 0x35, true, "NUMPAD/", "Num /"},
	{VK_RETURN, 0x1C, true, "NUMPADENTER", "Num Enter"},

	// ---- 功能键 F1-F24（F8/F9/F10 为保留键，不参与绑定）----
	{VK_F1, 0x3B, false, "F1", "F1"}, {VK_F2, 0x3C, false, "F2", "F2"},
	{VK_F3, 0x3D, false, "F3", "F3"}, {VK_F4, 0x3E, false, "F4", "F4"},
	{VK_F5, 0x3F, false, "F5", "F5"}, {VK_F6, 0x40, false, "F6", "F6"},
	{VK_F7, 0x41, false, "F7", "F7"}, {VK_F8, 0x42, false, "F8", "F8"},
	{VK_F9, 0x43, false, "F9", "F9"}, {VK_F10, 0x44, false, "F10", "F10"},
	{VK_F11, 0x57, false, "F11", "F11"}, {VK_F12, 0x58, false, "F12", "F12"},
	{VK_F13, 0x64, false, "F13", "F13"}, {VK_F14, 0x65, false, "F14", "F14"},
	{VK_F15, 0x66, false, "F15", "F15"}, {VK_F16, 0x67, false, "F16", "F16"},
	{VK_F17, 0x68, false, "F17", "F17"}, {VK_F18, 0x69, false, "F18", "F18"},
	{VK_F19, 0x6A, false, "F19", "F19"}, {VK_F20, 0x6B, false, "F20", "F20"},
	{VK_F21, 0x6C, false, "F21", "F21"}, {VK_F22, 0x6D, false, "F22", "F22"},
	{VK_F23, 0x6E, false, "F23", "F23"}, {VK_F24, 0x6F, false, "F24", "F24"},
}

var (
	byName  = map[string]Key{}
	byVK    = map[uint16][]Key{}
	aliases = map[string]string{
		"PAGEUP": "PGUP", "PAGEDOWN": "PGDN",
		"INS": "INSERT", "DEL": "DELETE",
		"NUMPADADD": "NUMPAD+", "NUMPADSUBTRACT": "NUMPAD-",
		"NUMPADMULTIPLY": "NUMPAD*", "NUMPADDIVIDE": "NUMPAD/",
		"NUMENTER": "NUMPADENTER", "RETURN": "ENTER",
		"BKSPACE": "BACKSPACE", "ESC": "ESC",
	}
)

func init() {
	for _, k := range keyTable {
		if _, dup := byName[k.Name]; !dup {
			byName[k.Name] = k
		}
		byVK[k.VK] = append(byVK[k.VK], k)
	}
}

// Parse 将配置中的键名解析为 Key。
// 支持 A-Z、0-9、常用符号字符以及 SPACE/TAB/ENTER/LSHIFT/PGUP/NUMPAD+ 等名称。
func Parse(name string) (Key, error) {
	s := strings.TrimSpace(strings.ToUpper(name))
	if s == "" {
		return Key{}, fmt.Errorf("empty key name")
	}
	if alias, ok := aliases[s]; ok {
		s = alias
	}
	if k, ok := byName[s]; ok {
		return k, nil
	}
	return Key{}, fmt.Errorf("unknown key name: %s", name)
}

// FromCode 根据虚拟键码、扫描码与扩展标志反查 Key。
// 用于 GUI 捕获按键与后台映射器匹配源键。
func FromCode(vk uint16, scan uint8, extend bool) (Key, bool) {
	keys := byVK[vk]
	if len(keys) == 0 {
		return Key{}, false
	}
	// 优先精确匹配（扫描码 + 扩展标志都一致）。
	for _, k := range keys {
		if k.Extend == extend && (k.Scan == scan || scan == 0) {
			return k, true
		}
	}
	// 退而求其次：仅按扩展标志匹配（区分主回车 / 小键盘回车）。
	for _, k := range keys {
		if k.Extend == extend {
			return k, true
		}
	}
	return keys[0], true
}

// IsReserved 报告该键是否为后台映射器的保留键（F8/F9/F10），
// 这些键不能用于绑定游戏操作。
func IsReserved(k Key) bool {
	return k.VK == VK_F8 || k.VK == VK_F9 || k.VK == VK_F10
}

// IsModifier 报告该键是否为修饰键（Shift / Ctrl / Alt）。
func IsModifier(k Key) bool {
	switch k.VK {
	case VK_LSHIFT, VK_RSHIFT, VK_LCONTROL, VK_RCONTROL, VK_LMENU, VK_RMENU:
		return true
	}
	return false
}

// Same 判断两个键是否相同（虚拟键码与扩展标志一致）。
func Same(a, b Key) bool {
	return a.VK == b.VK && a.Extend == b.Extend
}
