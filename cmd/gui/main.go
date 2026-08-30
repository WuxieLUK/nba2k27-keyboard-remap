// NBA 2K27 键盘自定义 —— 图形界面改键工具。
//
// 自绘 22 行操作列表，点击右侧键位框后直接按键即可改键；
// 底部提供 保存并应用 / 仅保存 / 停止映射 / 恢复默认 / 打开配置 / 使用说明。
package main

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/yourname/nba2k27-keyboard-remap/internal/keymap"
)

// ---- Windows 常量 ----
const (
	wmCreate      = 0x0001
	wmDestroy     = 0x0002
	wmClose       = 0x0010
	wmEraseBkgnd  = 0x0014
	wmPaint       = 0x000F
	wmLButtonDown = 0x0201
	wmKeyDown     = 0x0100
	wmSysKeyDown  = 0x0104
	wmCommand     = 0x0111

	bsPushbutton   = 0x0000
	bsAutoCheckbox = 0x0003
	esAutohscroll  = 0x0080
	wsChild        = 0x40000000
	wsVisible      = 0x10000000
	wsTabStop      = 0x00010000
	wsGroup        = 0x00020000
	wsBorder       = 0x00800000
	ssLeft         = 0x0000
	wsOverlapped   = 0x00CF0000

	bmSetcheck    = 0x00F1
	bmGetcheck    = 0x00F0
	wmSettext     = 0x000C
	wmGettext     = 0x000D
	enChange      = 0x0300
	swShownormal  = 1
	swShowdefault = 10

	vkEscape = 0x1B
	vkShift  = 0x10
	vkCtrl   = 0x11
	vkMenu   = 0x12

	dtSingleline = 0x0020
	dtCenter     = 0x0001
	dtVCenter    = 0x0004
	dtNoprefix   = 0x0800

	colBtnFace = 15

	transparent = 1
)

// remapWindowClass 与后台映射器保持一致。
const remapWindowClass = "NBA2K27RemapWindow"

// 控件 ID。
const (
	idBtnApply     = 1
	idBtnSave      = 2
	idBtnStop      = 3
	idBtnReset     = 4
	idBtnOpen      = 5
	idBtnHelp      = 6
	idChkEnable    = 11
	idChkInGame    = 12
	idChkTest      = 13
	idEditProc     = 21
	idStaticStatus = 23
)

// 窗口布局（96 DPI 下按像素计算）。
const (
	winW    = 660
	winH    = 672
	rowH    = 22
	listTop = 30
	keyBoxX = 430
	keyBoxW = 200
	keyBoxH = 18

	editY   = 562
	editX   = 96
	editW   = 180
	chkY    = 590
	btnY    = 614
	btnW    = 92
	btnH    = 26
	statusY = 646
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
	procFrameRect           = user32.NewProc("FrameRect")
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
	procSetBkMode           = gdi32.NewProc("SetBkMode")
	procSetTextColor        = gdi32.NewProc("SetTextColor")
	procCreateFontIndirectW = gdi32.NewProc("CreateFontIndirectW")
	procCreateSolidBrush    = gdi32.NewProc("CreateSolidBrush")
	procSelectObject        = gdi32.NewProc("SelectObject")
	procDeleteObject        = gdi32.NewProc("DeleteObject")
	procGetStockObject      = gdi32.NewProc("GetStockObject")
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

// ---- 全局状态 ----
var (
	cfg        *keymap.Config
	iniPath    string
	remapPath  string
	listening  = -1 // 正在监听改键的行号，-1 表示无
	dirty      = false
	hFont      uintptr
	mainHwnd   uintptr
	statusText = "就绪"
)

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
	hbr, _, _ := procGetStockObject.Call(colBtnFace)
	wc := wndclassexw{
		CbSize:        uint32(unsafe.Sizeof(wndclassexw{})),
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInst,
		HbrBackground: hbr,
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	title, _ := syscall.UTF16PtrFromString("NBA 2K27 键盘自定义 v5.0 — 直接按键绑定")
	hwnd, _, _ := procCreateWindowExW.Call(0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		wsOverlapped,
		100, 40, winW, winH,
		0, 0, hInst, 0)
	procShowWindow.Call(hwnd, swShownormal)
	procUpdateWindow.Call(hwnd)
}

func wndProc(hwnd uintptr, message uintptr, wParam, lParam uintptr) uintptr {
	switch message {
	case wmCreate:
		mainHwnd = hwnd
		createControls(hwnd)
		return 0
	case wmPaint:
		onPaint(hwnd)
		return 0
	case wmEraseBkgnd:
		return 1 // 由 WM_PAINT 全量重绘，避免闪烁
	case wmLButtonDown:
		onLButtonDown(hwnd, lParam)
		return 0
	case wmKeyDown, wmSysKeyDown:
		if listening >= 0 {
			onCaptureKey(hwnd, wParam, lParam)
			return 0
		}
	case wmCommand:
		onCommand(wParam)
		return 0
	case wmClose:
		procDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		if hFont != 0 {
			procDeleteObject.Call(hFont)
		}
		procPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, message, wParam, lParam)
	return ret
}

// ---- 控件创建 ----

func createControls(hwnd uintptr) {
	hInst, _, _ := procGetModuleHandleW.Call(0)

	// 游戏进程名
	procCreateWindowExW.Call(0, u16ptr("STATIC"), u16ptr("游戏进程名"),
		wsChild|wsVisible|ssLeft, 12, editY+3, 80, 20, hwnd, 0, hInst, 0)

	procCreateWindowExW.Call(0, u16ptr("EDIT"), 0,
		wsChild|wsVisible|wsBorder|wsTabStop|esAutohscroll,
		editX, editY, editW, 22, hwnd, idEditProc, hInst, 0)
	procSetWindowTextW.Call(childOf(hwnd, idEditProc), u16ptr(cfg.GameProcess))

	// 复选框
	procCreateWindowExW.Call(0, u16ptr("BUTTON"), u16ptr("启用映射"),
		wsChild|wsVisible|wsTabStop|bsAutoCheckbox,
		296, chkY, 92, 20, hwnd, idChkEnable, hInst, 0)
	procCreateWindowExW.Call(0, u16ptr("BUTTON"), u16ptr("仅游戏前台生效"),
		wsChild|wsVisible|wsTabStop|bsAutoCheckbox|wsGroup,
		392, chkY, 130, 20, hwnd, idChkInGame, hInst, 0)
	procCreateWindowExW.Call(0, u16ptr("BUTTON"), u16ptr("全局测试模式"),
		wsChild|wsVisible|wsTabStop|bsAutoCheckbox|wsGroup,
		12, chkY, 120, 20, hwnd, idChkTest, hInst, 0)

	setCheck(idChkEnable, cfg.StartEnabled)
	setCheck(idChkInGame, cfg.OnlyInGame)
	setCheck(idChkTest, !cfg.OnlyInGame)

	// 按钮
	buttons := []struct {
		id   uintptr
		text string
	}{
		{idBtnApply, "保存并应用"},
		{idBtnSave, "仅保存"},
		{idBtnStop, "停止映射"},
		{idBtnReset, "恢复默认"},
		{idBtnOpen, "打开配置"},
		{idBtnHelp, "使用说明"},
	}
	for i, b := range buttons {
		procCreateWindowExW.Call(0, u16ptr("BUTTON"), u16ptr(b.text),
			wsChild|wsVisible|wsTabStop|bsPushbutton,
			12+uintptr(i)*(btnW+8), btnY, btnW, btnH, hwnd, b.id, hInst, 0)
	}

	// 状态栏
	procCreateWindowExW.Call(0, u16ptr("STATIC"), 0,
		wsChild|wsVisible|ssLeft, 12, statusY, winW-24, 18, hwnd, idStaticStatus, hInst, 0)
	setStatus("就绪")
}

// ---- 事件处理 ----

func onLButtonDown(hwnd uintptr, lParam uintptr) {
	x := int32(lParam & 0xFFFF)
	y := int32((lParam >> 16) & 0xFFFF)
	idx := int(y-listTop) / rowH
	if idx < 0 || idx >= len(keymap.Bindings) {
		return
	}
	if x >= keyBoxX && x <= keyBoxX+keyBoxW {
		listening = idx
		procSetFocus.Call(hwnd)
		redraw(hwnd)
	}
}

func onCaptureKey(hwnd uintptr, wParam, lParam uintptr) {
	vk := uint16(wParam)
	if vk == vkEscape {
		listening = -1
		redraw(hwnd)
		return
	}
	scan := uint8((lParam >> 16) & 0xFF)
	ext := (lParam>>24)&1 != 0
	k, ok := keymap.FromCode(vk, scan, ext)
	if !ok {
		setStatus("无法绑定该按键")
		listening = -1
		redraw(hwnd)
		return
	}
	if keymap.IsReserved(k) {
		setStatus("F8 / F9 / F10 是工具保留键，不能绑定")
		listening = -1
		redraw(hwnd)
		return
	}
	// 组合键检测：按下非修饰键时若存在修饰键按下则拒绝。
	if !keymap.IsModifier(k) && modifierDown() {
		setStatus("暂不支持 Ctrl / Alt / Shift 组合键")
		listening = -1
		redraw(hwnd)
		return
	}
	def := keymap.Bindings[listening]
	cfg.Bindings[def.ID] = k.Name
	dirty = true
	setStatus("键位已更新：" + def.Label + " → " + k.Label)
	listening = -1
	redraw(hwnd)
}

func modifierDown() bool {
	return getAsync(vkShift) || getAsync(vkCtrl) || getAsync(vkMenu)
}

func getAsync(vk uintptr) bool {
	state, _, _ := procGetAsyncKeyState.Call(vk)
	return state&0x8000 != 0
}

func onCommand(wParam uintptr) {
	id := wParam & 0xFFFF
	code := (wParam >> 16) & 0xFFFF
	switch id {
	case idBtnApply:
		syncFromControls()
		if err := keymap.SaveConfig(iniPath, cfg); err != nil {
			setStatus("保存失败：" + err.Error())
			return
		}
		stopRemap()
		if cfg.StartEnabled {
			if startRemap() {
				setStatus("已保存并启动映射器")
			} else {
				setStatus("配置已保存，但映射器启动失败")
			}
		} else {
			setStatus("配置已保存")
		}
		dirty = false
	case idBtnSave:
		syncFromControls()
		if err := keymap.SaveConfig(iniPath, cfg); err != nil {
			setStatus("保存失败：" + err.Error())
			return
		}
		dirty = false
		setStatus("配置已保存")
	case idBtnStop:
		stopRemap()
		setStatus("已停止后台映射器")
	case idBtnReset:
		cfg = keymap.DefaultConfig()
		syncToControls()
		dirty = true
		redraw(mainHwnd)
		setStatus("已恢复默认键位，点击「保存并应用」生效")
	case idBtnOpen:
		shellOpen(iniPath)
	case idBtnHelp:
		showHelp()
	case idChkEnable:
		cfg.StartEnabled = getCheck(idChkEnable)
		dirty = true
	case idChkInGame:
		cfg.OnlyInGame = getCheck(idChkInGame)
		setCheck(idChkTest, !cfg.OnlyInGame)
		dirty = true
	case idChkTest:
		cfg.OnlyInGame = !getCheck(idChkTest)
		setCheck(idChkInGame, cfg.OnlyInGame)
		dirty = true
	case idEditProc:
		if code == enChange {
			dirty = true
		}
	}
}

func syncFromControls() {
	buf := make([]uint16, 260)
	n, _, _ := procGetWindowTextW.Call(childOf(mainHwnd, idEditProc),
		uintptr(unsafe.Pointer(&buf[0])), 260)
	if n > 0 {
		cfg.GameProcess = syscall.UTF16ToString(buf[:n])
	}
	cfg.OnlyInGame = getCheck(idChkInGame)
	cfg.StartEnabled = getCheck(idChkEnable)
}

func syncToControls() {
	procSetWindowTextW.Call(childOf(mainHwnd, idEditProc), u16ptr(cfg.GameProcess))
	setCheck(idChkEnable, cfg.StartEnabled)
	setCheck(idChkInGame, cfg.OnlyInGame)
	setCheck(idChkTest, !cfg.OnlyInGame)
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
	return ret > 32 // SE_ERR_* 错误码 <= 32
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
	procMessageBoxW.Call(0, u16ptr(text), u16ptr("使用说明"), 0x40 /* MB_ICONINFORMATION */)
}

// ---- 自绘列表 ----

func onPaint(hwnd uintptr) {
	var ps paintstruct
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

	procSetBkMode.Call(hdc, transparent)
	if hFont == 0 {
		hFont = createFont()
	}
	procSelectObject.Call(hdc, hFont)

	// 说明文字
	drawText(hdc, 12, 6, winW-24, 20,
		"点击右侧键位框，然后直接按键盘上的键（Esc 取消）", dtSingleline, 0x000000)

	// 列表背景（白色）
	listRect := rect{2, listTop - 2, winW - 2, listTop + int32(len(keymap.Bindings))*rowH + 2}
	fillColor(hdc, &listRect, 0xFFFFFF)

	for i, b := range keymap.Bindings {
		y := listTop + int32(i)*rowH
		drawText(hdc, 12, y+2, 410, rowH-2, b.Label, dtSingleline, 0x1A1A1A)

		box := rect{keyBoxX, y + 2, keyBoxX + keyBoxW, y + 2 + keyBoxH}
		fillColor(hdc, &box, 0xF5F5F5)
		frameColor(hdc, &box, 0x999999)

		if listening == i {
			drawText(hdc, keyBoxX+2, y+2, keyBoxW-4, keyBoxH,
				"请按一个键…", dtCenter|dtVCenter|dtSingleline|dtNoprefix, 0xC00000)
		} else {
			label := keyLabel(cfg.Bindings[b.ID])
			drawText(hdc, keyBoxX+2, y+2, keyBoxW-4, keyBoxH,
				label, dtCenter|dtVCenter|dtSingleline|dtNoprefix, 0x003366)
		}
	}

	// 状态栏文本
	drawText(hdc, 12, statusY+1, winW-24, 16, statusText, dtSingleline, 0x404040)
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
	procDrawTextW.Call(hdc, u16ptr(text), ^uintptr(0), /* 以 NUL 结尾自动计算长度 */
		uintptr(unsafe.Pointer(&r)), uintptr(flags))
}

func fillColor(hdc uintptr, r *rect, rgb uint32) {
	brush, _, _ := procCreateSolidBrush.Call(uintptr(rgb))
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(r)), brush)
	procDeleteObject.Call(brush)
}

func frameColor(hdc uintptr, r *rect, rgb uint32) {
	brush, _, _ := procCreateSolidBrush.Call(uintptr(rgb))
	procFrameRect.Call(hdc, uintptr(unsafe.Pointer(r)), brush)
	procDeleteObject.Call(brush)
}

func createFont() uintptr {
	var lf logfontw
	lf.LfHeight = -16
	lf.LfCharSet = 0x86 // DEFAULT_CHARSET
	face := "Microsoft YaHei"
	for i := 0; i < len(face) && i < 31; i++ {
		lf.LfFaceName[i] = uint16(face[i])
	}
	h, _, _ := procCreateFontIndirectW.Call(uintptr(unsafe.Pointer(&lf)))
	return h
}

func redraw(hwnd uintptr) {
	procInvalidateRect.Call(hwnd, 0, 1)
	procUpdateWindow.Call(hwnd)
}

// ---- 控件辅助 ----

func childOf(parent, id uintptr) uintptr {
	hwnd, _, _ := procGetDlgItem.Call(parent, id)
	return hwnd
}

func setCheck(id uintptr, on bool) {
	var v uintptr
	if on {
		v = 1
	}
	procSendMessageW.Call(childOf(mainHwnd, id), bmSetcheck, v, 0)
}

func getCheck(id uintptr) bool {
	ret, _, _ := procSendMessageW.Call(childOf(mainHwnd, id), bmGetcheck, 0, 0)
	return ret != 0
}

func setStatus(text string) {
	statusText = text
	if mainHwnd != 0 {
		procSetWindowTextW.Call(childOf(mainHwnd, idStaticStatus), u16ptr(text))
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
