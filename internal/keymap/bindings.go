package keymap

// BindingDef 描述一个可绑定操作及其默认键位。
type BindingDef struct {
	ID      string // 配置文件中的操作ID
	Label   string // 界面显示名（中文）
	Default string // 默认键名
}

// Bindings 为全部可绑定操作，顺序与界面展示一致。
var Bindings = []BindingDef{
	{"MOVE_UP", "移动 - 上", "W"},
	{"MOVE_LEFT", "移动 - 左", "A"},
	{"MOVE_DOWN", "移动 - 下", "S"},
	{"MOVE_RIGHT", "移动 - 右", "D"},
	{"SHOOT", "投篮 / 假投", "SPACE"},
	{"PASS", "传球", "Q"},
	{"BOUNCE_PASS", "击地传球", "E"},
	{"LOB_ALLEY", "高吊 / 空接", "R"},
	{"SPRINT", "冲刺", "O"},
	{"POST_UP", "背身", "LSHIFT"},
	{"PICK_TACTIC", "呼叫挡拆", "TAB"},
	{"ICON_PASS", "图标传球", "NUMPAD+"},
	{"PRO_STICK_UP", "专家摇杆 - 上", "0"},
	{"PRO_STICK_DOWN", "专家摇杆 - 下", "COMMA"},
	{"PRO_STICK_LEFT", "专家摇杆 - 左", "I"},
	{"PRO_STICK_RIGHT", "专家摇杆 - 右", "P"},
	{"PAUSE_MENU", "暂停比赛", "ENTER"},
	{"TIMEOUT", "叫暂停", "RSHIFT"},
	{"ARROW_UP", "临场换人方向键 - 上", "UP"},
	{"ARROW_DOWN", "临场换人方向键 - 下", "DOWN"},
	{"ARROW_LEFT", "临场换人方向键 - 左", "LEFT"},
	{"ARROW_RIGHT", "临场换人方向键 - 右", "RIGHT"},
}

// BindingIndex 为操作ID到定义的索引。
var BindingIndex = func() map[string]BindingDef {
	m := make(map[string]BindingDef, len(Bindings))
	for _, b := range Bindings {
		m[b.ID] = b
	}
	return m
}()
