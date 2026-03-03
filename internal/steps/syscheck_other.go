//go:build !windows

package steps

import "fmt"

func getFreeDiskMB(_ string) int64 {
	return 9999 // 非 Windows 平台跳过磁盘检测
}

func getOSVersion() string {
	return "非 Windows 平台（仅供开发调试）"
}

func getPortProcess(port int) string {
	return fmt.Sprintf("端口 %d 进程（需在 Windows 上查询）", port)
}
