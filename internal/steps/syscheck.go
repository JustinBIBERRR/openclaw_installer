package steps

import (
	"context"
	"fmt"
	"net"
	"time"

	"openclaw-manager/internal/mirror"
	"openclaw-manager/internal/state"
	"openclaw-manager/internal/ui"
)

const minDiskMB = 300

// RunSysCheck 执行系统预检，不写入 manifest（预检不是可续传步骤）
func RunSysCheck(m *state.Manifest) error {
	ui.PrintInfo("正在检查运行环境...")
	fmt.Println()

	var hasError bool

	// 1. 磁盘空间
	freeMB := getFreeDiskMB(m.InstallDir)
	if freeMB < minDiskMB {
		ui.PrintError(fmt.Sprintf("磁盘空间不足：需要 %d MB，剩余 %d MB", minDiskMB, freeMB))
		hasError = true
	} else {
		ui.PrintOK(fmt.Sprintf("磁盘空间: %d MB 可用", freeMB))
	}

	// 2. 端口检测
	port, err := resolvePort(m.GatewayPort)
	if err != nil {
		ui.PrintError(fmt.Sprintf("无法找到可用端口: %v", err))
		hasError = true
	} else {
		if port != m.GatewayPort {
			ui.PrintWarn(fmt.Sprintf("端口 %d 被占用，将使用备用端口 %d", m.GatewayPort, port))
			m.GatewayPort = port
		} else {
			ui.PrintOK(fmt.Sprintf("端口 %d: 可用", port))
		}
	}

	// 3. Windows 版本检测
	osInfo := getOSVersion()
	ui.PrintOK("操作系统: " + osInfo)

	// 4. 网络连通性（并发探测镜像）
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	m2, latency := mirror.SelectFastest(ctx)
	if latency == 0 {
		ui.PrintWarn("网络连接较慢，建议检查代理设置（安装将继续）")
	} else {
		ui.PrintOK(fmt.Sprintf("网络连通: %s (延迟 %dms)", m2.Name, latency.Milliseconds()))
	}

	fmt.Println()

	if hasError {
		return fmt.Errorf("系统预检未通过，请解决以上问题后重试")
	}
	return nil
}

// resolvePort 检测端口是否可用，自动尝试备用端口
func resolvePort(preferred int) (int, error) {
	candidates := []int{preferred, preferred + 1, preferred + 2}
	for _, port := range candidates {
		if isPortFree(port) {
			return port, nil
		}
		proc := getPortProcess(port)
		if proc != "" {
			ui.PrintWarn(fmt.Sprintf("端口 %d 已被 %s 占用", port, proc))
		}
	}
	return 0, fmt.Errorf("端口 %d/%d/%d 均被占用，请手动释放后重试",
		candidates[0], candidates[1], candidates[2])
}

func isPortFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
