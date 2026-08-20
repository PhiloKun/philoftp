//go:build !windows

package main

import "fmt"

// winNotify 在非 Windows 平台为空实现（控制台模式下信息直接输出到 stderr）。
func winNotify(title, msg string) {}

// formatSingleInstanceMsg 生成已存在实例时的提示文案。
func formatSingleInstanceMsg() string {
	return "PhiloFTP 已在运行中，请勿重复启动。"
}

// formatPortInUseMsg 生成端口被占用时的中文提示文案。
func formatPortInUseMsg(svc string, port int, rawErr error) string {
	return fmt.Sprintf("%s端口 %d 已被占用: %v", svc, port, rawErr)
}
