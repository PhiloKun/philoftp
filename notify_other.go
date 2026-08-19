//go:build !windows

package main

// winNotify 在非 Windows 平台为空实现（控制台模式下信息直接输出到 stderr）。
func winNotify(title, msg string) {}
