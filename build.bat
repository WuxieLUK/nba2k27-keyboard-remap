@echo off
rem NBA 2K27 Keyboard Remap - build script
setlocal

where go >nul 2>nul
if errorlevel 1 (
    echo [ERROR] 未检测到 Go，请先安装：https://go.dev/dl/
    exit /b 1
)

if not exist release mkdir release

rem 可选：生成 PE 版本资源（需安装 goversioninfo）
where goversioninfo >nul 2>nul
if not errorlevel 1 (
    echo 正在生成版本资源...
    cd cmd\remap && goversioninfo -o resource.syso versioninfo.json && cd ..\..
    cd cmd\gui && goversioninfo -o resource.syso versioninfo.json && cd ..\..
)

echo 正在构建后台映射器...
go build -ldflags "-s -w -H=windowsgui" -o release\NBA2K27_KeyboardRemap.exe .\cmd\remap
if errorlevel 1 goto :fail

echo 正在构建图形改键工具...
go build -ldflags "-s -w -H=windowsgui" -o release\NBA2K27_Keyboard_GUI.exe .\cmd\gui
if errorlevel 1 goto :fail

echo.
echo 构建完成：release\
dir /b release\*.exe
exit /b 0

:fail
echo [ERROR] 构建失败
exit /b 1
