// NBA 2K27 后台按键映射器。
//
// 安装低级键盘钩子（WH_KEYBOARD_LL），当用户在游戏前台按下
// 绑定的源键时，将其拦截并改写为目标键，实现 NBA 2K27 的自定义按键。
package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/WuxieLUK/nba2k27-keyboard-remap/internal/keymap"
)

// ---- Windows 常量 ----
const (
	wmClose      = 0x0010
	wmDestroy    = 0x0002
	wmKeyDown    = 0x0100
	wmKeyUp      = 0x0101
	wmSysKeyDown = 0x0104
	wmSysKeyUp   = 0x0105

	whKeyboardLL    = 13
	hcAction        = 0
	llkhfExtended   = 0x01
	keyeventfKeyUp  = 0x0002
	keyeventfExtKey = 0x0001
	inputKeyboard   = 1

	swShownormal      = 1
	tokenQuery        = 0x0008
	tokenElevation    = 20
	processQueryLimit = 0x1000
)

// windowClass 为隐藏控制窗口的类名，GUI 借此定位并停止本进程。
const windowClass = "NBA2K27RemapWindow"

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	procSetWindowsHookExW    = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx       = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHookEx  = user32.NewProc("UnhookWindowsHookEx")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procDestroyWindow        = user32.NewProc("DestroyWindow")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procFindWindowW          = user32.NewProc("FindWindowW")
	procGetForegroundWindow  = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProc  = user32.NewProc("GetWindowThreadProcessId")
	procSendInput            = user32.NewProc("SendInput")
	procMessageBoxW          = user32.NewProc("MessageBoxW")
	procGetModuleHandleW     = kernel32.NewProc("GetModuleHandleW")
	procOpenProcess          = kernel32.NewProc("OpenProcess")
	procCloseHandle          = kernel32.NewProc("CloseHandle")
	procQueryFullProcessName = kernel32.NewProc("QueryFullProcessImageNameW")
	procGetCurrentProcess    = kernel32.NewProc("GetCurrentProcess")
	procOpenProcessToken     = advapi32.NewProc("OpenProcessToken")
	procGetTokenInformation  = advapi32.NewProc("GetTokenInformation")
	procShellExecuteW        = shell32.NewProc("ShellExecuteW")
)

// ---- Windows 结构 ----
type kbdllhookstruct struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type keybdinput struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type input struct {
	Type uint32
	Ki   keybdinput
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

type tokenElevationInfo struct {
	TokenIsElevated uint32
}

// ---- 全局状态 ----
var (
	cfg      *keymap.Config
	bindings []*binding
	sending  bool // 发送目标键期间置位，防止钩子递归

	// syscall.NewCallback 生成的函数指针必须保持全局引用，否则可能被 GC 回收。
	keyboardProcCallback = syscall.NewCallback(keyboardProc)
	wndProcCallback      = syscall.NewCallback(wndProc)
)

type binding struct {
	def keymap.BindingDef
	src keymap.Key
	dst keymap.Key
}

func main() {
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	exe = filepath.Clean(exe)

	// 需要管理员权限安装全局键盘钩子。
	if !isAdmin() {
		ensureAdmin(exe)
		return
	}

	// 单实例：已有映射器在运行时直接退出。
	if findWindow(windowClass) != 0 {
		return
	}

	cfg, err = keymap.LoadConfig(keymap.ConfigPath(exe))
	if err != nil {
		fatal("读取配置失败: " + err.Error())
	}
	bindings = buildBindings(cfg)

	hInst, _, _ := procGetModuleHandleW.Call(0)
	createHiddenWindow(hInst)

	hook, _, _ := procSetWindowsHookExW.Call(
		whKeyboardLL, keyboardProcCallback, hInst, 0)
	if hook == 0 {
		fatal("键盘钩子安装失败")
	}
	defer procUnhookWindowsHookEx.Call(hook)

	// 消息循环：低级键盘钩子依赖其所在线程的消息泵。
	var m msg
	for {
		ret, _, _ := procGetMessageW.Call(
			uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if ret == 0 { // WM_QUIT
			break
		}
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func buildBindings(c *keymap.Config) []*binding {
	var out []*binding
	for _, def := range keymap.Bindings {
		name := c.Bindings[def.ID]
		src, err := keymap.Parse(name)
		if err != nil {
			src, _ = keymap.Parse(def.Default)
		}
		dst, err := keymap.Parse(def.Default)
		if err != nil {
			continue
		}
		out = append(out, &binding{def: def, src: src, dst: dst})
	}
	return out
}

// keyboardProc 为低级键盘钩子回调：匹配源键则拦截并发送目标键。
func keyboardProc(nCode uintptr, wParam, lParam uintptr) uintptr {
	if nCode == hcAction {
		switch wParam {
		case wmKeyDown, wmSysKeyDown, wmKeyUp, wmSysKeyUp:
			if sending {
				break
			}
			kbd := (*kbdllhookstruct)(unsafe.Pointer(lParam))
			ext := kbd.Flags&llkhfExtended != 0
			k, ok := keymap.FromCode(uint16(kbd.VkCode), uint8(kbd.ScanCode), ext)
			if !ok {
				break
			}
			if !active() {
				break
			}
			for _, b := range bindings {
				if keymap.Same(k, b.src) {
					down := wParam == wmKeyDown || wParam == wmSysKeyDown
					sending = true
					sendKey(b.dst, !down)
					sending = false
					return 1 // 拦截原按键
				}
			}
		}
	}
	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return ret
}

// active 判断映射是否应当在当前时刻生效。
func active() bool {
	if !cfg.StartEnabled {
		return false
	}
	if !cfg.OnlyInGame {
		return true
	}
	return gameActive()
}

// sendKey 通过 SendInput 注入目标键的按下 / 抬起事件。
func sendKey(k keymap.Key, up bool) {
	flags := uint32(0)
	if up {
		flags |= keyeventfKeyUp
	}
	if k.Extend {
		flags |= keyeventfExtKey
	}
	in := input{
		Type: inputKeyboard,
		Ki: keybdinput{
			WVk:     k.VK,
			DwFlags: flags,
		},
	}
	procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
}

// gameActive 判断前台窗口是否属于配置的游戏进程。
func gameActive() bool {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return false
	}
	var pid uint32
	procGetWindowThreadProc.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return false
	}
	h, _, _ := procOpenProcess.Call(processQueryLimit, 0, uintptr(pid))
	if h == 0 {
		return false
	}
	defer procCloseHandle.Call(h)
	var buf [512]uint16
	size := uint32(len(buf))
	ret, _, _ := procQueryFullProcessName.Call(h, 0,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if ret == 0 {
		return false
	}
	name := syscall.UTF16ToString(buf[:size])
	base := filepath.Base(name)
	return strings.EqualFold(base, cfg.GameProcess)
}

// ---- 管理员权限 ----

func isAdmin() bool {
	var token uintptr
	cur, _, _ := procGetCurrentProcess.Call()
	ok, _, _ := procOpenProcessToken.Call(cur, tokenQuery, uintptr(unsafe.Pointer(&token)))
	if ok == 0 {
		return false
	}
	defer procCloseHandle.Call(token)
	var elev tokenElevationInfo
	var retLen uint32
	ok, _, _ = procGetTokenInformation.Call(token, tokenElevation,
		uintptr(unsafe.Pointer(&elev)), unsafe.Sizeof(elev),
		uintptr(unsafe.Pointer(&retLen)))
	return ok != 0 && elev.TokenIsElevated != 0
}

// ensureAdmin 以管理员权限重新启动自身，然后退出当前进程。
func ensureAdmin(exe string) {
	verb, _ := syscall.UTF16PtrFromString("runas")
	path, _ := syscall.UTF16PtrFromString(exe)
	procShellExecuteW.Call(0, uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(path)), 0, 0, swShownormal)
	os.Exit(0)
}

// ---- 隐藏窗口 ----

func createHiddenWindow(hInst uintptr) {
	className, _ := syscall.UTF16PtrFromString(windowClass)
	wc := wndclassexw{
		CbSize:        uint32(unsafe.Sizeof(wndclassexw{})),
		LpfnWndProc:   wndProcCallback,
		HInstance:     hInst,
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	procCreateWindowExW.Call(0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(className)),
		0, 0, 0, 0, 0, 0, 0, hInst, 0)
}

func wndProc(hwnd uintptr, message uintptr, wParam, lParam uintptr) uintptr {
	switch message {
	case wmClose:
		procDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		procPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, message, wParam, lParam)
	return ret
}

func findWindow(className string) uintptr {
	name, _ := syscall.UTF16PtrFromString(className)
	hwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(name)), 0)
	return hwnd
}

// fatal 弹出错误提示框并退出。
func fatal(text string) {
	title, _ := syscall.UTF16PtrFromString("NBA 2K27 Keyboard Remap")
	body, _ := syscall.UTF16PtrFromString(text)
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(body)),
		uintptr(unsafe.Pointer(title)), 0x10 /* MB_ICONERROR */)
	os.Exit(1)
}
