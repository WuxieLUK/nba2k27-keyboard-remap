// NBA 2K27 键盘自定义 —— 图形界面改键工具。
//
// 深色主题、全自绘界面：标题栏 / 圆角按钮 / 卡片式列表 / 复选框均自绘，
// 仅"游戏进程名"编辑框使用原生 EDIT（通过 WM_CTLCOLOREDIT 套深色主题）。
// 点击键位框后直接按键即可改键。
package main

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/WuxieLUK/nba2k27-keyboard-remap/internal/keymap"
)

// ---- Windows 常量 ----
const (
	wmCreate       = 0x0001
	wmDestroy      = 0x0002
	wmClose        = 0x0010
	wmEraseBkgnd   = 0x0014
	wmPaint        = 0x000F
	wmLButtonDown  = 0x0201
	wmLButtonUp    = 0x0202
	wmMouseMove    = 0x0200
	wmMouseLeave   = 0x02A3
	wmKeyDown      = 0x0100
	wmSysKeyDown   = 0x0104
	wmCommand      = 0x0111
	wmCtlColorEdit = 0x0137

	esAutohscroll = 0x0080
	wsChild       = 0x40000000
	wsVisible     = 0x10000000
	wsTabStop     = 0x00010000
	wsOverlapped  = 0x00CF0000

	wmSettext = 0x000C
	wmGettext = 0x000D
	enChange  = 0x0300

	swShownormal  = 1
	swShowdefault = 10

	vkEscape = 0x1B
	vkShift  = 0x10
	vkCtrl   = 0x11
	vkMenu   = 0x12

	tmeLeave = 0x0002

	dtSingleline = 0x0020
	dtCenter     = 0x0001
	dtVCenter    = 0x0004
	dtNoprefix   = 0x0800

	psSolid = 0

	transparent = 1
)

// remapWindowClass 与后台映射器保持一致。
const remapWindowClass = "NBA2K27RemapWindow"

// 控件 ID（编辑框与动作标识共用）。
const (
	idBtnApply = 1
	idBtnSave  = 2
	idBtnStop  = 3
	idBtnReset = 4
	idBtnOpen  = 5
	idBtnHelp  = 6

	idChkEnable = 11
	idChkInGame = 12
	idChkTest   = 13

	idEditProc = 21
)

// ---- 配色（深色主题）----
const (
	bgMain       uint32 = 0x16161F // 窗口背景
	bgCard       uint32 = 0x1F1F2E // 列表卡片
	rowHover     uint32 = 0x27273C // 行悬停
	bgKeyBox     uint32 = 0x2A2A44 // 键位框底
	borderKey    uint32 = 0x3D3D5C // 键位框边框
	borderKeyHot uint32 = 0x6B94FF // 键位框悬停边框
	textMain     uint32 = 0xE8E8F2 // 主文字
	textDim      uint32 = 0x8A8AA3 // 次要文字
	accent       uint32 = 0x4C7DFF // 品牌蓝
	accentHover  uint32 = 0x6B94FF
	accentDark   uint32 = 0x3A63D0 // 按下
	btnBg        uint32 = 0x2E2E48 // 普通按钮
	btnBgHover   uint32 = 0x3A3A5A
	btnBorder    uint32 = 0x43436A
	danger       uint32 = 0xFF6B6B // 错误 / 监听提示
	okColor      uint32 = 0x5BD08A // 成功
	white        uint32 = 0xFFFFFF
)

// ---- 布局 ----
const (
	winW    = 680
	winH    = 706
	rowH    = 23
	titleH  = 44
	listTop = 72
	keyBoxX = 450
	keyBoxW = 200
	keyBoxH = 19

	editY = 586
	editX = 96
	editW = 180
	editH = 22

	chkY = 614
	btnY = 642
	btnW = 92
	btnH = 28

	statusY = 678
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	procRegisterClassExW    = user32.NewProc("RegisterClassExW")
	procCreateWindowExW     = user32.NewProc("CreateWindowExW")
	procDefWindowProcW      = user32.NewProc("DefWindowProcW")
	procDestroyWindow       = user32.NewProc("DestroyWindow")
	procPostQuitMessage     = user32.NewProc("PostQuitMessage")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procDispatchMessageW    = user32.NewProc("DispatchMessageW")
	procShowWindow          = user32.NewProc("ShowWindow")
	procUpdateWindow        = user32.NewProc("UpdateWindow")
	procBeginPaint          = user32.NewProc("BeginPaint")
	procEndPaint            = user32.NewProc("EndPaint")
	procInvalidateRect      = user32.NewProc("InvalidateRect")
	procFillRect            = user32.NewProc("FillRect")
	procDrawTextW           = user32.NewProc("DrawTextW")
	procSetWindowTextW      = user32.NewProc("SetWindowTextW")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	procSetFocus            = user32.NewProc("SetFocus")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
	procSendMessageW        = user32.NewProc("SendMessageW")
	procPostMessageW        = user32.NewProc("PostMessageW")
	procFindWindowW         = user32.NewProc("FindWindowW")
	procMessageBoxW         = user32.NewProc("MessageBoxW")
	procGetDlgItem          = user32.NewProc("GetDlgItem")
	procSetCapture          = user32.NewProc("SetCapture")
	procReleaseCapture      = user32.NewProc("ReleaseCapture")
	procTrackMouseEvent     = user32.NewProc("TrackMouseEvent")
	procSetBkMode           = gdi32.NewProc("SetBkMode")
	procSetTextColor        = gdi32.NewProc("SetTextColor")
	procSetBkColor          = gdi32.NewProc("SetBkColor")
	procCreateFontIndirectW = gdi32.NewProc("CreateFontIndirectW")
	procCreateSolidBrush    = gdi32.NewProc("CreateSolidBrush")
	procCreateRoundRectRgn  = gdi32.NewProc("CreateRoundRectRgn")
	procCreatePen           = gdi32.NewProc("CreatePen")
	procSelectObject        = gdi32.NewProc("SelectObject")
	procDeleteObject        = gdi32.NewProc("DeleteObject")
	procFillRgn             = gdi32.NewProc("FillRgn")
	procFrameRgn            = gdi32.NewProc("FrameRgn")
	procMoveToEx            = gdi32.NewProc("MoveToEx")
	procLineTo              = gdi32.NewProc("LineTo")
	procGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
	procShellExecuteW       = shell32.NewProc("ShellExecuteW")
)

type rect struct {
	Left, Top, Right, Bottom int32
}

type paintstruct struct {
	Hdc         uintptr
	FErase      int32
	RcPaint     rect
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

type wndclassexw struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
}

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type logfontw struct {
	LfHeight         int32
	LfWidth          int32
	LfEscapement     int32
	LfOrientation    int32
	LfWeight         int32
	LfItalic         byte
	LfUnderline      byte
	LfStrikeOut      byte
	LfCharSet        byte
	LfOutPrecision   byte
	LfClipPrecision  byte
	LfQuality        byte
	LfPitchAndFamily byte
	LfFaceName       [32]uint16
}

type trackMouseEvent struct {
	CbSize    uint32
	DwFlags   uint32
	HwndTrack uintptr
	DwHover   uint32
}

// ---- 全局状态 ----
var (
	cfg         *keymap.Config
	iniPath     string
	remapPath   string
	listening   = -1 // 正在监听改键的行号
	dirty       = false
	hFont       uintptr
	hTitleFont  uintptr
	editBrush   uintptr
	mainHwnd    uintptr
	statusText         = "就绪"
	statusColor uint32 = textDim

	// syscall.NewCallback 生成的函数指针必须保持全局引用，否则可能被 GC 回收。
	wndProcCallback = syscall.NewCallback(wndProc)

	// 交互状态
	hoverRow   = -1
	hoverKey   = -1
	hoverBtn   = -1
	hoverChk   = -1
	pressedBtn = -1
)

type buttonDef struct {
	id      uintptr
	text    string
	primary bool
}

var buttons = []buttonDef{
	{idBtnApply, "保存并应用", true},
	{idBtnSave, "仅保存", false},
	{idBtnStop, "停止映射", false},
	{idBtnReset, "恢复默认", false},
	{idBtnOpen, "打开配置", false},
	{idBtnHelp, "使用说明", false},
}

var checkboxDefs = []struct {
	id   uintptr
	text string
}{
	{idChkEnable, "启用映射"},
	{idChkInGame, "仅游戏前台生效"},
	{idChkTest, "全局测试模式"},
}

func main() {
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	guiPath := filepath.Clean(exe)
	iniPath = keymap.ConfigPath(guiPath)
	remapPath = filepath.Join(filepath.Dir(guiPath), "NBA2K27_KeyboardRemap.exe")

	cfg, err = keymap.LoadConfig(iniPath)
	if err != nil {
		fatal("读取配置失败: " + err.Error())
	}

	hInst, _, _ := procGetModuleHandleW.Call(0)
	createMainWindow(hInst)

	var m msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if ret == 0 {
			break
		}
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func createMainWindow(hInst uintptr) {
	className, _ := syscall.UTF16PtrFromString("NBA2K27KeyboardGUI")
	wc := wndclassexw{
		CbSize:        uint32(unsafe.Sizeof(wndclassexw{})),
		LpfnWndProc:   wndProcCallback,
		HInstance:     hInst,
		LpszClassName: className,
	}
	if ret, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); ret == 0 {
		fatal("窗口类注册失败")
	}

	title, _ := syscall.UTF16PtrFromString("NBA 2K27 键盘自定义 v5.0 — 直接按键绑定")
	hwnd, _, _ := procCreateWindowExW.Call(0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		wsOverlapped,
		100, 30, winW, winH,
		0, 0, hInst, 0)
	if hwnd == 0 {
		fatal("主窗口创建失败")
	}
	procShowWindow.Call(hwnd, swShownormal)
	procUpdateWindow.Call(hwnd)
}

func wndProc(hwnd uintptr, message uintptr, wParam, lParam uintptr) uintptr {
	switch message {
	case wmCreate:
		mainHwnd = hwnd
		hFont = createFont(-16, 400)
		hTitleFont = createFont(-22, 700)
		editBrush = newBrush(bgCard)
		createEditControl(hwnd)
		armMouseLeave(hwnd)
		return 0
	case wmPaint:
		onPaint(hwnd)
		return 0
	case wmEraseBkgnd:
		return 1 // 全量自绘，避免闪烁
	case wmLButtonDown:
		onLButtonDown(lParam)
		return 0
	case wmLButtonUp:
		onLButtonUp(lParam)
		return 0
	case wmMouseMove:
		onMouseMove(lParam)
		return 0
	case wmMouseLeave:
		hoverRow, hoverKey, hoverBtn, hoverChk = -1, -1, -1, -1
		armMouseLeave(mainHwnd)
		redraw()
		return 0
	case wmKeyDown, wmSysKeyDown:
		if listening >= 0 {
			onCaptureKey(wParam, lParam)
			return 0
		}
	case wmCommand:
		if (wParam&0xFFFF) == idEditProc && ((wParam>>16)&0xFFFF) == enChange {
			dirty = true
		}
		return 0
	case wmCtlColorEdit:
		procSetTextColor.Call(wParam, uintptr(textMain))
		procSetBkColor.Call(wParam, uintptr(bgCard))
		procSetBkMode.Call(wParam, transparent)
		return editBrush
	case wmClose:
		procDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		if hFont != 0 {
			procDeleteObject.Call(hFont)
		}
		if hTitleFont != 0 {
			procDeleteObject.Call(hTitleFont)
		}
		if editBrush != 0 {
			procDeleteObject.Call(editBrush)
		}
		procPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, message, wParam, lParam)
	return ret
}

// ---- 区域计算 ----

func listRect() rect {
	return rect{6, listTop - 6, winW - 6, listTop + int32(len(keymap.Bindings))*rowH + 6}
}

func keyBoxRect(i int) rect {
	y := listTop + int32(i)*rowH
	return rect{keyBoxX, y + 2, keyBoxX + keyBoxW, y + 2 + keyBoxH}
}

func editRect() rect {
	return rect{editX - 1, editY - 1, editX + editW + 1, editY + editH + 1}
}

func buttonRect(i int) rect {
	x := int32(14 + i*(btnW+8))
	return rect{x, btnY, x + btnW, btnY + btnH}
}

func checkboxRect(i int) rect {
	xs := [...]int32{14, 148, 318}
	return rect{xs[i], chkY, xs[i] + 130, chkY + 20}
}

func inRect(r rect, x, y int32) bool {
	return x >= r.Left && x < r.Right && y >= r.Top && y < r.Bottom
}

// ---- 事件处理 ----

func onLButtonDown(lParam uintptr) {
	x := int32(lParam & 0xFFFF)
	y := int32((lParam >> 16) & 0xFFFF)

	// 键位框：进入改键监听
	for i := range keymap.Bindings {
		if inRect(keyBoxRect(i), x, y) {
			listening = i
			procSetFocus.Call(mainHwnd)
			redraw()
			return
		}
	}
	// 复选框
	for i := range checkboxDefs {
		if inRect(checkboxRect(i), x, y) {
			toggleCheckbox(i)
			redraw()
			return
		}
	}
	// 按钮
	for i := range buttons {
		if inRect(buttonRect(i), x, y) {
			pressedBtn = i
			procSetCapture.Call(mainHwnd)
			redraw()
			return
		}
	}
}

func onLButtonUp(lParam uintptr) {
	if pressedBtn < 0 {
		return
	}
	procReleaseCapture.Call()
	i := pressedBtn
	pressedBtn = -1
	x := int32(lParam & 0xFFFF)
	y := int32((lParam >> 16) & 0xFFFF)
	if inRect(buttonRect(i), x, y) {
		doAction(buttons[i].id)
	}
	redraw()
}

func onMouseMove(lParam uintptr) {
	x := int32(lParam & 0xFFFF)
	y := int32((lParam >> 16) & 0xFFFF)

	row, key, btn, chk := -1, -1, -1, -1
	for i := range keymap.Bindings {
		if inRect(keyBoxRect(i), x, y) {
			key = i
			row = i
			break
		}
	}
	for i := range buttons {
		if inRect(buttonRect(i), x, y) {
			btn = i
			break
		}
	}
	for i := range checkboxDefs {
		if inRect(checkboxRect(i), x, y) {
			chk = i
			break
		}
	}
	if row != hoverRow || key != hoverKey || btn != hoverBtn || chk != hoverChk {
		hoverRow, hoverKey, hoverBtn, hoverChk = row, key, btn, chk
		redraw()
	}
}

func armMouseLeave(hwnd uintptr) {
	tme := trackMouseEvent{
		CbSize:    uint32(unsafe.Sizeof(trackMouseEvent{})),
		DwFlags:   tmeLeave,
		HwndTrack: hwnd,
	}
	procTrackMouseEvent.Call(uintptr(unsafe.Pointer(&tme)))
}

func toggleCheckbox(i int) {
	switch i {
	case 0:
		cfg.StartEnabled = !cfg.StartEnabled
	case 1, 2:
		cfg.OnlyInGame = !cfg.OnlyInGame // test = !OnlyInGame，两点互斥联动
	}
	dirty = true
}

func onCaptureKey(wParam, lParam uintptr) {
	vk := uint16(wParam)
	if vk == vkEscape {
		listening = -1
		redraw()
		return
	}
	scan := uint8((lParam >> 16) & 0xFF)
	ext := (lParam>>24)&1 != 0
	k, ok := keymap.FromCode(vk, scan, ext)
	if !ok {
		setStatus("无法绑定该按键", danger)
		listening = -1
		redraw()
		return
	}
	if keymap.IsReserved(k) {
		setStatus("F8 / F9 / F10 是工具保留键，不能绑定", danger)
		listening = -1
		redraw()
		return
	}
	if !keymap.IsModifier(k) && modifierDown() {
		setStatus("暂不支持 Ctrl / Alt / Shift 组合键", danger)
		listening = -1
		redraw()
		return
	}
	def := keymap.Bindings[listening]
	cfg.Bindings[def.ID] = k.Name
	dirty = true
	setStatus("键位已更新："+def.Label+" → "+k.Label, okColor)
	listening = -1
	redraw()
}

func modifierDown() bool {
	return getAsync(vkShift) || getAsync(vkCtrl) || getAsync(vkMenu)
}

func getAsync(vk uintptr) bool {
	state, _, _ := procGetAsyncKeyState.Call(vk)
	return state&0x8000 != 0
}

func doAction(id uintptr) {
	switch id {
	case idBtnApply:
		syncFromControls()
		if err := keymap.SaveConfig(iniPath, cfg); err != nil {
			setStatus("保存失败："+err.Error(), danger)
			return
		}
		stopRemap()
		if cfg.StartEnabled {
			if startRemap() {
				setStatus("已保存并启动映射器", okColor)
			} else {
				setStatus("配置已保存，但映射器启动失败", danger)
			}
		} else {
			setStatus("配置已保存", okColor)
		}
		dirty = false
	case idBtnSave:
		syncFromControls()
		if err := keymap.SaveConfig(iniPath, cfg); err != nil {
			setStatus("保存失败："+err.Error(), danger)
			return
		}
		dirty = false
		setStatus("配置已保存", okColor)
	case idBtnStop:
		stopRemap()
		setStatus("已停止后台映射器", textDim)
	case idBtnReset:
		cfg = keymap.DefaultConfig()
		syncToControls()
		dirty = true
		redraw()
		setStatus("已恢复默认键位，点击「保存并应用」生效", textDim)
	case idBtnOpen:
		shellOpen(iniPath)
	case idBtnHelp:
		showHelp()
	}
}

func syncFromControls() {
	buf := make([]uint16, 260)
	n, _, _ := procGetWindowTextW.Call(childOf(mainHwnd, idEditProc),
		uintptr(unsafe.Pointer(&buf[0])), 260)
	if n > 0 {
		cfg.GameProcess = syscall.UTF16ToString(buf[:n])
	}
	// OnlyInGame / StartEnabled 已由复选框点击实时维护。
}

func syncToControls() {
	procSetWindowTextW.Call(childOf(mainHwnd, idEditProc), u16ptr(cfg.GameProcess))
}

// ---- 映射器控制 ----

func stopRemap() {
	hwnd := findWindow(remapWindowClass)
	if hwnd == 0 {
		return
	}
	procPostMessageW.Call(hwnd, wmClose, 0, 0)
	for i := 0; i < 20; i++ {
		if findWindow(remapWindowClass) == 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func startRemap() bool {
	if _, err := os.Stat(remapPath); err != nil {
		return false
	}
	return shellOpen(remapPath)
}

func shellOpen(target string) bool {
	ret, _, _ := procShellExecuteW.Call(0, u16ptr("open"), u16ptr(target), 0, 0, swShowdefault)
	return ret > 32
}

func showHelp() {
	text := `NBA 2K27 键盘自定义 v5.0 — 直接按键绑定

使用方法：
1. 点击你要修改的操作右侧键位框
2. 键位框显示"请按一个键…"
3. 直接按键盘上的目标键（Esc 取消）
4. 点击「保存并应用」

说明：
- 每个操作绑定一个单键，不支持 Ctrl+J 这类组合键，也不支持鼠标键
- F8 / F9 / F10 是后台映射器的保留键，不能绑定
- 高吊 / 空接：单击自定义键 = 高吊，快速双击 = 空接
- 在记事本测试：勾选「全局测试模式」→ 保存并应用，测完取消勾选再保存
- 建议开启 Num Lock 后使用小键盘绑定`
	procMessageBoxW.Call(0, u16ptr(text), u16ptr("使用说明"), 0x40)
}

// ---- 自绘 ----

func onPaint(hwnd uintptr) {
	var ps paintstruct
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

	procSetBkMode.Call(hdc, transparent)

	// 窗口背景
	fillRect(hdc, rect{0, 0, winW, winH}, bgMain)

	// 标题区
	fillRect(hdc, rect{0, 0, 5, winH}, accent)
	procSelectObject.Call(hdc, hTitleFont)
	drawText(hdc, 24, 6, winW-40, 26, "NBA 2K27 键盘自定义", dtSingleline, white)
	procSelectObject.Call(hdc, hFont)
	drawText(hdc, 24, 32, winW-40, 14, "系统级按键重映射 · 修复游戏内无法改键的问题", dtSingleline, textDim)

	// 说明
	drawText(hdc, 14, 50, winW-28, 18,
		"点击右侧键位框，然后直接按键盘上的键（Esc 取消）", dtSingleline, textDim)

	// 列表卡片
	card := listRect()
	fillRound(hdc, &card, bgCard, 10)

	for i, b := range keymap.Bindings {
		y := listTop + int32(i)*rowH
		// 行悬停高亮
		if hoverRow == i {
			row := rect{card.Left + 2, y - 2, card.Right - 2, y + rowH + 1}
			fillRound(hdc, &row, rowHover, 8)
		}
		// 操作名
		drawText(hdc, 22, y+3, 400, rowH-2, b.Label, dtSingleline, textMain)

		// 键位框
		box := keyBoxRect(i)
		fillRound(hdc, &box, bgKeyBox, 5)
		bd := borderKey
		if listening == i || hoverKey == i {
			bd = accent
		}
		if listening == i {
			frameRound(hdc, &box, accent, 5, 2, 2)
		} else {
			frameRound(hdc, &box, bd, 5, 1, 1)
		}
		if listening == i {
			drawText(hdc, box.Left+4, box.Top, box.Right-box.Left-8, box.Bottom-box.Top,
				"请按一个键…", dtCenter|dtVCenter|dtSingleline|dtNoprefix, danger)
		} else {
			label := keyLabel(cfg.Bindings[b.ID])
			drawText(hdc, box.Left+4, box.Top, box.Right-box.Left-8, box.Bottom-box.Top,
				label, dtCenter|dtVCenter|dtSingleline|dtNoprefix, accent)
		}
	}

	// 游戏进程名
	drawText(hdc, 14, editY+3, 80, 18, "游戏进程名", dtSingleline, textMain)
	drawText(hdc, editX+editW+12, editY+3, 320, 18,
		"仅在该进程为前台时生效", dtSingleline, textDim)
	er := editRect()
	frameRound(hdc, &er, borderKey, 5, 1, 1)

	// 复选框
	for i := range checkboxDefs {
		drawCheckbox(hdc, i)
	}

	// 按钮
	for i := range buttons {
		drawButton(hdc, i)
	}

	// 状态栏
	drawText(hdc, 14, statusY+1, winW-28, 16, statusText, dtSingleline, statusColor)
}

func drawButton(hdc uintptr, i int) {
	r := buttonRect(i)
	primary := buttons[i].primary
	bg := btnBg
	switch {
	case primary && pressedBtn == i:
		bg = accentDark
	case primary && hoverBtn == i:
		bg = accentHover
	case primary:
		bg = accent
	case pressedBtn == i:
		bg = btnBgHover
	case hoverBtn == i:
		bg = btnBgHover
	}
	fillRound(hdc, &r, bg, 6)
	frameRound(hdc, &r, btnBorder, 6, 1, 1)
	color := textMain
	if primary {
		color = white
	}
	drawText(hdc, r.Left, r.Top, r.Right-r.Left, r.Bottom-r.Top,
		buttons[i].text, dtCenter|dtVCenter|dtSingleline|dtNoprefix, color)
}

func drawCheckbox(hdc uintptr, i int) {
	r := checkboxRect(i)
	checked := checkboxChecked(i)
	box := rect{r.Left, r.Top + 3, r.Left + 14, r.Top + 17}
	if checked {
		fillRound(hdc, &box, accent, 3)
		// 白色对勾（折线）
		pen, _, _ := procCreatePen.Call(psSolid, 2, uintptr(white))
		oldPen, _, _ := procSelectObject.Call(hdc, pen)
		procMoveToEx.Call(hdc, uintptr(box.Left+3), uintptr(box.Top+7), 0)
		procLineTo.Call(hdc, uintptr(box.Left+6), uintptr(box.Top+10))
		procLineTo.Call(hdc, uintptr(box.Left+11), uintptr(box.Top+3))
		procSelectObject.Call(hdc, oldPen)
		procDeleteObject.Call(pen)
	} else {
		fillRound(hdc, &box, bgKeyBox, 3)
		frameRound(hdc, &box, borderKey, 3, 1, 1)
	}
	color := textMain
	if hoverChk == i {
		color = white
	}
	drawText(hdc, r.Left+22, r.Top, r.Right-r.Left-22, 20,
		checkboxDefs[i].text, dtSingleline|dtVCenter, color)
}

func checkboxChecked(i int) bool {
	switch i {
	case 0:
		return cfg.StartEnabled
	case 1:
		return cfg.OnlyInGame
	default:
		return !cfg.OnlyInGame
	}
}

func keyLabel(name string) string {
	if name == "" {
		return "-"
	}
	k, err := keymap.Parse(name)
	if err != nil {
		return name
	}
	return k.Label
}

func drawText(hdc uintptr, x, y, w, h int32, text string, flags uint32, color uint32) {
	procSetTextColor.Call(hdc, uintptr(color))
	r := rect{x, y, x + w, y + h}
	procDrawTextW.Call(hdc, u16ptr(text), ^uintptr(0),
		uintptr(unsafe.Pointer(&r)), uintptr(flags))
}

func fillRect(hdc uintptr, r rect, rgb uint32) {
	brush, _, _ := procCreateSolidBrush.Call(uintptr(rgb))
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), brush)
	procDeleteObject.Call(brush)
}

func fillRound(hdc uintptr, r *rect, rgb uint32, radius int32) {
	brush, _, _ := procCreateSolidBrush.Call(uintptr(rgb))
	rgn, _, _ := procCreateRoundRectRgn.Call(
		uintptr(r.Left), uintptr(r.Top), uintptr(r.Right+1), uintptr(r.Bottom+1),
		uintptr(radius*2), uintptr(radius*2))
	procFillRgn.Call(hdc, rgn, brush)
	procDeleteObject.Call(rgn)
	procDeleteObject.Call(brush)
}

func frameRound(hdc uintptr, r *rect, rgb uint32, radius int32, w, h uintptr) {
	brush, _, _ := procCreateSolidBrush.Call(uintptr(rgb))
	rgn, _, _ := procCreateRoundRectRgn.Call(
		uintptr(r.Left), uintptr(r.Top), uintptr(r.Right+1), uintptr(r.Bottom+1),
		uintptr(radius*2), uintptr(radius*2))
	procFrameRgn.Call(hdc, rgn, brush, w, h)
	procDeleteObject.Call(rgn)
	procDeleteObject.Call(brush)
}

func newBrush(rgb uint32) uintptr {
	brush, _, _ := procCreateSolidBrush.Call(uintptr(rgb))
	return brush
}

func createFont(height, weight int32) uintptr {
	var lf logfontw
	lf.LfHeight = height
	lf.LfWeight = weight
	lf.LfCharSet = 0x86 // DEFAULT_CHARSET
	face := "Microsoft YaHei"
	for i := 0; i < len(face) && i < 31; i++ {
		lf.LfFaceName[i] = uint16(face[i])
	}
	h, _, _ := procCreateFontIndirectW.Call(uintptr(unsafe.Pointer(&lf)))
	return h
}

func redraw() {
	if mainHwnd == 0 {
		return
	}
	procInvalidateRect.Call(mainHwnd, 0, 1)
	procUpdateWindow.Call(mainHwnd)
}

// ---- 控件辅助 ----

func createEditControl(hwnd uintptr) {
	hInst, _, _ := procGetModuleHandleW.Call(0)
	procCreateWindowExW.Call(0, u16ptr("EDIT"), 0,
		wsChild|wsVisible|wsTabStop|esAutohscroll,
		editX, editY, editW, editH, hwnd, idEditProc, hInst, 0)
	procSetWindowTextW.Call(childOf(hwnd, idEditProc), u16ptr(cfg.GameProcess))
}

func childOf(parent, id uintptr) uintptr {
	hwnd, _, _ := procGetDlgItem.Call(parent, id)
	return hwnd
}

func setStatus(text string, color uint32) {
	statusText = text
	statusColor = color
	if mainHwnd != 0 {
		redraw()
	}
}

func findWindow(className string) uintptr {
	name, _ := syscall.UTF16PtrFromString(className)
	hwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(name)), 0)
	return hwnd
}

func u16ptr(s string) uintptr {
	p, _ := syscall.UTF16PtrFromString(s)
	return uintptr(unsafe.Pointer(p))
}

func fatal(text string) {
	procMessageBoxW.Call(0, u16ptr(text), u16ptr("NBA 2K27 键盘自定义"), 0x10)
	os.Exit(1)
}
