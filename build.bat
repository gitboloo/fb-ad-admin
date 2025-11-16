@echo off
echo Building Ad Platform Backend...
go build -o ad-platform.exe cmd\main.go
if %errorlevel% equ 0 (
    echo Build successful!
    echo Run ad-platform.exe to start the server.
) else (
    echo Build failed!
)
pause