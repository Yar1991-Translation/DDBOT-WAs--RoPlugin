//go:build windows
// +build windows

package main

import "syscall"

func init() {
    const CP_UTF8 = 65001
    // 将 Windows 控制台输入/输出码页设置为 UTF-8，避免中文乱码
    syscall.SetConsoleOutputCP(CP_UTF8)
    syscall.SetConsoleCP(CP_UTF8)
} 