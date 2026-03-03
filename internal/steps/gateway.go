package steps

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"openclaw-manager/internal/env"
	"openclaw-manager/internal/state"
	"openclaw-manager/internal/ui"
)

// RunGateway 启动 Gateway，等待健康检查通过，打开浏览器，创建桌面快捷方式
func RunGateway(m *state.Manifest) error {
	if m.IsDone(state.StepGatewayStarted) {
		// 已安装过，检查 Gateway 是否在运行
		gatewayURL := fmt.Sprintf("http://localhost:%d", m.GatewayPort)
		if isHealthy(gatewayURL, 1*time.Second) {
			ui.PrintOK(fmt.Sprintf("Gateway 已在运行 → %s", gatewayURL))
			return nil
		}
		// 未在运行，重新启动
		ui.PrintInfo("Gateway 未运行，正在重新启动...")
	}

	shadowEnv := env.ShadowEnv(m)
	gatewayURL := fmt.Sprintf("http://localhost:%d", m.GatewayPort)

	// 启动 Gateway（后台运行）
	ocCmd := filepath.Join(m.NPMGlobalBin(), "openclaw.cmd")
	var cmd *exec.Cmd

	if _, err := os.Stat(ocCmd); err == nil {
		// Windows: 直接使用 openclaw.cmd
		cmd = exec.Command("cmd", "/c", ocCmd, "gateway", "start",
			"--port", fmt.Sprintf("%d", m.GatewayPort))
	} else {
		// 回退：通过 cmd 调用
		cmd = exec.Command("cmd", "/c", "openclaw", "gateway", "start",
			"--port", fmt.Sprintf("%d", m.GatewayPort))
	}

	cmd.Env = shadowEnv
	setDetachedProcess(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 Gateway 失败: %w", err)
	}

	// 轮询健康检查（最多 30 秒）
	ui.PrintInfo("等待 Gateway 就绪...")
	if err := waitForGateway(gatewayURL, 30*time.Second); err != nil {
		return fmt.Errorf("Gateway 启动超时（30s），请检查配置后手动运行 openclaw gateway start")
	}

	ui.PrintOK(fmt.Sprintf("Gateway 已启动 → %s", gatewayURL))

	// 创建桌面快捷方式
	exePath, _ := os.Executable()
	if err := createDesktopShortcut(exePath, "OpenClaw Manager"); err != nil {
		ui.PrintWarn(fmt.Sprintf("桌面快捷方式创建失败（已跳过）: %v", err))
	} else {
		ui.PrintOK("桌面快捷方式已创建")
	}

	if err := m.MarkDone(state.StepGatewayStarted); err != nil {
		return err
	}

	// 打开浏览器 + 完成界面
	openBrowser(gatewayURL)
	ui.PrintSuccess(gatewayURL)

	return nil
}

// StartGateway 在管理器模式下重新启动 Gateway
func StartGateway(m *state.Manifest) error {
	shadowEnv := env.ShadowEnv(m)
	gatewayURL := fmt.Sprintf("http://localhost:%d", m.GatewayPort)

	ocCmd := filepath.Join(m.NPMGlobalBin(), "openclaw.cmd")
	cmd := exec.Command("cmd", "/c", ocCmd, "gateway", "start",
		"--port", fmt.Sprintf("%d", m.GatewayPort))
	cmd.Env = shadowEnv
	setDetachedProcess(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动失败: %w", err)
	}

	return waitForGateway(gatewayURL, 20*time.Second)
}

// StopGateway 停止 Gateway
func StopGateway(m *state.Manifest) error {
	shadowEnv := env.ShadowEnv(m)
	ocCmd := filepath.Join(m.NPMGlobalBin(), "openclaw.cmd")
	cmd := exec.Command("cmd", "/c", ocCmd, "gateway", "stop")
	cmd.Env = shadowEnv
	setHideWindow(cmd)
	return cmd.Run()
}

// IsGatewayRunning 检查 Gateway 是否在线
func IsGatewayRunning(m *state.Manifest) bool {
	return isHealthy(fmt.Sprintf("http://localhost:%d", m.GatewayPort), 2*time.Second)
}

func waitForGateway(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if isHealthy(url, 2*time.Second) {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("超时 %v", timeout)
}

func isHealthy(baseURL string, timeout time.Duration) bool {
	client := &http.Client{Timeout: timeout}
	for _, path := range []string{"/health", "/api/health", "/"} {
		resp, err := client.Get(baseURL + path)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return true
			}
		}
	}
	return false
}

func openBrowser(url string) {
	cmd := exec.Command("cmd", "/c", "start", url)
	setHideWindow(cmd)
	_ = cmd.Start()
}

func createDesktopShortcut(targetExe, name string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	lnkPath := filepath.Join(home, "Desktop", name+".lnk")
	workDir := filepath.Dir(targetExe)

	script := fmt.Sprintf(
		`$ws = New-Object -ComObject WScript.Shell; `+
			`$s = $ws.CreateShortcut('%s'); `+
			`$s.TargetPath = '%s'; `+
			`$s.WorkingDirectory = '%s'; `+
			`$s.IconLocation = '%s,0'; `+
			`$s.Description = 'OpenClaw 本地 AI 网关管理器'; `+
			`$s.Save()`,
		lnkPath, targetExe, workDir, targetExe,
	)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	setHideWindow(cmd)
	return cmd.Run()
}
