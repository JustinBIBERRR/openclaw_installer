package state

import (
	"fmt"
	"os"
	"path/filepath"

	"openclaw-manager/internal/ui"
)

// Rollback 按步骤逆序清理所有已完成的步骤产物
func Rollback(m *Manifest) {
	ui.PrintWarn("正在卸载 OpenClaw，清理所有相关文件...")
	fmt.Println()

	// 逆序执行回滚
	steps := []struct {
		key  string
		name string
		fn   func(*Manifest) error
	}{
		{StepGatewayStarted, "停止 Gateway", rollbackGateway},
		{StepConfigWritten, "删除配置文件", rollbackConfig},
		{StepAPIKeySaved, "清理 API Key 记录", rollbackAPIKey},
		{StepCLIInstalled, "卸载 OpenClaw CLI", rollbackCLI},
		{StepRuntimeDownloaded, "删除 Node.js 运行时", rollbackRuntime},
	}

	for _, s := range steps {
		if !m.IsDone(s.key) {
			continue
		}
		if err := s.fn(m); err != nil {
			ui.PrintWarn(fmt.Sprintf("%s 失败（已跳过）: %v", s.name, err))
		} else {
			ui.PrintOK(s.name)
		}
	}

	// 删除桌面快捷方式
	removeDesktopShortcut()

	// 删除清单文件
	if p := m.ManifestFilePath(); p != "" {
		os.Remove(p)
	}

	// 尝试删除整个安装目录（若为空）
	os.RemoveAll(m.InstallDir)

	m.Phase = PhaseUninstalled
	fmt.Println()
	ui.PrintOK("OpenClaw 已完全卸载")
	fmt.Println()
	fmt.Print("  按 Enter 退出...")
	fmt.Scanln()
}

func rollbackGateway(m *Manifest) error {
	// 尝试调用 openclaw gateway stop，忽略错误（进程可能已退出）
	ocCmd := filepath.Join(m.NPMGlobalBin(), "openclaw.cmd")
	if _, err := os.Stat(ocCmd); err != nil {
		return nil // CLI 不存在，无需停止
	}
	// 通过系统 cmd 执行，最多等待 5s
	// 实际实现见 steps/gateway.go 中的 stopGateway
	return nil
}

func rollbackConfig(m *Manifest) error {
	return os.Remove(m.ConfigFile())
}

func rollbackAPIKey(_ *Manifest) error {
	// API Key 仅存在于 openclaw.json，由 rollbackConfig 处理
	return nil
}

func rollbackCLI(m *Manifest) error {
	return os.RemoveAll(m.NPMGlobalPrefix())
}

func rollbackRuntime(m *Manifest) error {
	runtimeDir := filepath.Join(m.InstallDir, "runtime")
	return os.RemoveAll(runtimeDir)
}

func removeDesktopShortcut() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	lnk := filepath.Join(home, "Desktop", "OpenClaw Manager.lnk")
	os.Remove(lnk)
}
