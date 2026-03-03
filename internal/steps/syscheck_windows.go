//go:build windows

package steps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func getFreeDiskMB(installDir string) int64 {
	// 向上遍历，直到找到真实存在的路径（用于在目录未创建时仍能检测所在盘）
	checkPath := installDir
	for {
		if _, err := os.Stat(checkPath); err == nil {
			break
		}
		parent := filepath.Dir(checkPath)
		if parent == checkPath {
			break
		}
		checkPath = parent
	}

	pathPtr, err := syscall.UTF16PtrFromString(checkPath)
	if err != nil {
		return 9999
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	if err := windows.GetDiskFreeSpaceEx(pathPtr, &freeBytesAvailable, &totalBytes, &totalFreeBytes); err != nil {
		return 9999
	}
	return int64(freeBytesAvailable / (1024 * 1024))
}

func getOSVersion() string {
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return "Windows（版本未知）"
	}
	defer k.Close()

	buildStr, _, _ := k.GetStringValue("CurrentBuildNumber")
	productName, _, _ := k.GetStringValue("ProductName")
	if productName == "" {
		productName = "Windows"
	}

	build := 0
	fmt.Sscanf(buildStr, "%d", &build)
	if build >= 22000 {
		return fmt.Sprintf("Windows 11 (Build %d) ✓", build)
	}
	if build >= 10240 {
		return fmt.Sprintf("Windows 10 (Build %d)", build)
	}
	return fmt.Sprintf("%s (Build %s)", productName, buildStr)
}

func getPortProcess(port int) string {
	script := fmt.Sprintf(
		`$c = Get-NetTCPConnection -LocalPort %d -ErrorAction SilentlyContinue | Select-Object -First 1; `+
			`if ($c) { (Get-Process -Id $c.OwningProcess -ErrorAction SilentlyContinue).ProcessName }`,
		port)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return "未知进程"
	}
	name := strings.TrimSpace(string(out))
	if name == "" {
		return "未知进程"
	}
	return name
}
