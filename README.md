# NBA 2K27 键盘自定义（Keyboard Remap）

## 为什么做这个？

NBA 2K27 发布后，**游戏内部的键位设置存在 bug，无法正常自定义按键**。
这对键盘玩家极不友好——默认按键错乱、改键功能不可用，几乎没法正常游玩。

所以就有了这款工具：通过**系统级按键重映射**，把任意键盘按键映射为游戏操作键，
例如把 `小键盘 5` 改成 `J`、把方向键改为 `WASD`，让键盘玩家也能顺畅游玩 2K27。

> 本项目与 NBA 2K 官方及 2K Games 无任何关联，仅供个人学习与游戏辅助使用。

## 功能特性

- **后台映射器**（`NBA2K27_KeyboardRemap.exe`）：低级键盘钩子，仅游戏前台生效，
  拦截源键并注入目标键，支持高吊 / 空接（单击 = 高吊，快速双击 = 空接）。
- **图形改键工具**（`NBA2K27_Keyboard_GUI.exe`）：点击键位框 → 直接按键 → 保存并应用，
  无需手动编辑配置。
- **全局测试模式**：可在记事本等任意窗口验证按键映射效果。
- **兼容原版 v5 配置**：可直接读取旧版 `bindings.ini`（含 `PGUP` / `PGDN` /
  `NUMPADENTER` 等键名）。

## 使用方法

1. 以管理员身份运行 `NBA2K27_Keyboard_GUI.exe`（保存并应用时会自动提权启动映射器）。
2. 点击你要修改的操作**右侧键位框**，键位框显示 `请按一个键…`。
3. 直接按键盘上的目标键（`Esc` 取消）。
4. 点击 **保存并应用**。

示例：点击「投篮 / 假投」的 `NUMPAD5` → 按 `J` → 保存并应用，
以后按 `J` 游戏就会收到小键盘 `5`。

### 支持的按键

- `A-Z`、`0-9`
- 方向键 `↑ ↓ ← →`
- `Space`、`Tab`、`Enter`
- 左 / 右 `Shift`、`Ctrl`、`Alt`（`LSHIFT` / `RSHIFT` / `LCTRL` / `RCTRL` / `LALT` / `RALT`）
- `Page Up` / `Page Down`（`PGUP` / `PGDN`）、`Home` / `End`、`Insert` / `Delete`
- 小键盘 `0-9`、`+`、`-`、`*`、`/`、`Enter`
- `F1`-`F24`（`F8` / `F9` / `F10` 除外，它们是后台映射器的保留键）
- 常用符号键：`- = [ ] \ ; ' , . / \``

### 注意

- 每个操作绑定一个单键，**不支持** `Ctrl+J` 这类组合键，也不支持鼠标键。
- 在记事本测试：勾选 **全局测试模式** → 保存并应用；测完取消勾选再保存。
- 建议开启 `Num Lock` 后使用小键盘绑定，以保证游戏与 Windows 对小键盘键位的识别一致。

## 配置文件

`bindings.ini` 与可执行文件位于同一目录：

```ini
[General]
GameProcess=NBA2K27.exe   ; 游戏进程名
OnlyInGame=1              ; 仅在游戏前台时生效
StartEnabled=1            ; 保存后自动启动映射

[Bindings]
MOVE_UP=W                 ; 操作ID = 键名
SHOOT=SPACE
; ...共 22 项
```

可通过 GUI 的「打开配置」按钮用记事本查看。

## 构建

需要 [Go 1.21+](https://go.dev/dl/)（Windows）：

```bat
build.bat
```

或手动：

```bat
go build -ldflags "-s -w -H=windowsgui" -o release\NBA2K27_KeyboardRemap.exe .\cmd\remap
go build -ldflags "-s -w -H=windowsgui" -o release\NBA2K27_Keyboard_GUI.exe .\cmd\gui
```

## 目录结构

```
cmd/
  remap/            后台映射器（键盘钩子 + 按键注入）
  gui/              图形改键工具
internal/
  keymap/           键名解析、配置读写、默认绑定（含单元测试）
```

## 测试

```bat
go test ./...
```

## 许可

[MIT](LICENSE)
