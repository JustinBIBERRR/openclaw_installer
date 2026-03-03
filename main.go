package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"

	"openclaw-manager/internal/state"
	"openclaw-manager/internal/steps"
	"openclaw-manager/internal/ui"
)

// Version 由构建时 ldflags 注入
var Version = "1.0.0"

func main() {
	uninstallFlag := flag.Bool("uninstall", false, "卸载 OpenClaw 及所有相关文件")
	versionFlag := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	ui.PrintBanner(Version)

	if *versionFlag {
		fmt.Printf("  OpenClaw Manager v%s\n\n", Version)
		return
	}

	m, loadErr := state.LoadManifest()

	// 卸载模式
	if *uninstallFlag {
		if loadErr != nil || m == nil || m.Phase == state.PhaseUninstalled {
			ui.PrintInfo("未检测到 OpenClaw 安装，无需卸载。")
			fmt.Println()
			return
		}
		if ui.AskConfirm("确定要卸载 OpenClaw 及所有相关文件吗？") {
			state.Rollback(m)
		}
		return
	}

	// 按 manifest 状态分发
	switch {
	case loadErr != nil:
		if !errors.Is(loadErr, os.ErrNotExist) {
			// 文件存在但损坏
			ui.PrintWarn(fmt.Sprintf("清单文件异常: %v", loadErr))
			if !ui.AskConfirm("是否重新开始安装？（现有安装数据将被清理）") {
				return
			}
		}
		// 全新安装
		runInstaller(state.NewManifest(Version))

	case m.Phase == state.PhaseComplete:
		// 管理器模式
		runManager(m)

	case m.Phase == state.PhaseFailed:
		// 失败恢复引导
		runRecovery(m)

	default:
		// installing 态（程序上次异常退出）
		ui.PrintInfo("检测到上次安装未完成，正在从断点继续...")
		fmt.Println()
		runInstaller(m)
	}
}

// ─────────────────────────────────────────────
// 安装器模式
// ─────────────────────────────────────────────

type installStep struct {
	stepKey string // 对应 manifest 中的 step key（空 = 不可跳过）
	name    string
	fn      func(*state.Manifest) error
}

var installSteps = []installStep{
	{"", "系统预检", steps.RunSysCheck},
	{state.StepRuntimeDownloaded, "下载 Node.js 运行时", steps.RunRuntimeDownload},
	{state.StepCLIInstalled, "安装 OpenClaw CLI", steps.RunOpenClawInstall},
	{state.StepAPIKeySaved, "配置 AI 服务", steps.RunAPIKey},
	{state.StepConfigWritten, "写入配置文件", steps.RunConfig},
	{state.StepGatewayStarted, "启动本地 Gateway", steps.RunGateway},
}

func runInstaller(m *state.Manifest) {
	total := len(installSteps)

	for i, s := range installSteps {
		// 可跳过的步骤：已完成则跳过
		if s.stepKey != "" && m.IsDone(s.stepKey) {
			ui.PrintSkip(s.name)
			continue
		}

		ui.PrintStepHeader(i+1, total, s.name)

		if err := s.fn(m); err != nil {
			ui.PrintError(err.Error())
			m.Phase = state.PhaseFailed
			_ = m.Save()
			ui.PrintFatalError()
			os.Exit(1)
		}
	}

	m.Phase = state.PhaseComplete
	_ = m.Save()
}

// ─────────────────────────────────────────────
// 管理器模式（已安装后双击进入）
// ─────────────────────────────────────────────

func runManager(m *state.Manifest) {
	running := steps.IsGatewayRunning(m)
	gatewayURL := fmt.Sprintf("http://localhost:%d", m.GatewayPort)

	statusLine := ui.StyleDim.Render("Gateway: ") + ui.StyleError.Render("● 已停止")
	if running {
		statusLine = ui.StyleDim.Render("Gateway: ") + ui.StyleSuccess.Render("● 运行中") +
			ui.StyleDim.Render(fmt.Sprintf(" (%s)", gatewayURL))
	}
	fmt.Println("  " + statusLine)
	fmt.Println()

	var choice string
	var menuOptions []huh.Option[string]

	if running {
		menuOptions = []huh.Option[string]{
			huh.NewOption("打开浏览器界面", "open"),
			huh.NewOption("重启 Gateway", "restart"),
			huh.NewOption("停止 Gateway", "stop"),
			huh.NewOption("更新 OpenClaw CLI", "update"),
			huh.NewOption("更换 API Key", "rekey"),
			huh.NewOption("───────────────", "sep"),
			huh.NewOption("卸载 OpenClaw", "uninstall"),
			huh.NewOption("退出", "exit"),
		}
	} else {
		menuOptions = []huh.Option[string]{
			huh.NewOption("启动 Gateway", "start"),
			huh.NewOption("更新 OpenClaw CLI", "update"),
			huh.NewOption("更换 API Key", "rekey"),
			huh.NewOption("───────────────", "sep"),
			huh.NewOption("卸载 OpenClaw", "uninstall"),
			huh.NewOption("退出", "exit"),
		}
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("请选择操作").
				Options(menuOptions...).
				Value(&choice),
		),
	)

	if err := form.Run(); err != nil {
		return
	}

	switch choice {
	case "sep":
		runManager(m) // 忽略分隔线，重新选择

	case "open":
		steps.OpenBrowserURL(gatewayURL)
		ui.PrintOK("已在浏览器中打开 " + gatewayURL)

	case "start":
		ui.PrintInfo("正在启动 Gateway...")
		if err := steps.StartGateway(m); err != nil {
			ui.PrintError(fmt.Sprintf("启动失败: %v", err))
		} else {
			ui.PrintOK(fmt.Sprintf("Gateway 已启动 → %s", gatewayURL))
			steps.OpenBrowserURL(gatewayURL)
		}

	case "restart":
		ui.PrintInfo("正在重启 Gateway...")
		_ = steps.StopGateway(m)
		if err := steps.StartGateway(m); err != nil {
			ui.PrintError(fmt.Sprintf("重启失败: %v", err))
		} else {
			ui.PrintOK("Gateway 已重启")
		}

	case "stop":
		ui.PrintInfo("正在停止 Gateway...")
		if err := steps.StopGateway(m); err != nil {
			ui.PrintWarn(fmt.Sprintf("停止命令返回错误: %v（可能已停止）", err))
		} else {
			ui.PrintOK("Gateway 已停止")
		}

	case "update":
		ui.PrintStepHeader(1, 1, "更新 OpenClaw CLI")
		if err := steps.RunOpenClawUpdate(m); err != nil {
			ui.PrintError(fmt.Sprintf("更新失败: %v", err))
		} else {
			ui.PrintOK("更新完成")
		}

	case "rekey":
		m.Steps[state.StepAPIKeySaved] = &state.StepRecord{Status: state.StatusPending}
		m.Steps[state.StepConfigWritten] = &state.StepRecord{Status: state.StatusPending}
		_ = m.Save()
		ui.PrintStepHeader(1, 2, "更换 API Key")
		if err := steps.RunAPIKey(m); err != nil {
			ui.PrintError(fmt.Sprintf("更换失败: %v", err))
			return
		}
		ui.PrintStepHeader(2, 2, "写入配置文件")
		if err := steps.RunConfig(m); err != nil {
			ui.PrintError(fmt.Sprintf("写入失败: %v", err))
			return
		}
		ui.PrintOK("API Key 已更换")
		if running {
			ui.PrintInfo("正在重启 Gateway 以应用新配置...")
			_ = steps.StopGateway(m)
			_ = steps.StartGateway(m)
		}

	case "uninstall":
		if ui.AskConfirm("确定要卸载 OpenClaw 及所有相关文件吗？") {
			state.Rollback(m)
		}

	case "exit":
		return
	}

	fmt.Println()
	fmt.Print("  按 Enter 返回菜单或关闭窗口退出...")
	fmt.Scanln()
}

// ─────────────────────────────────────────────
// 失败恢复模式
// ─────────────────────────────────────────────

func runRecovery(m *state.Manifest) {
	fmt.Println("  " + ui.StyleWarn.Render("⚠ 检测到上次安装未完成"))
	fmt.Println()

	// 显示已完成和未完成的步骤
	for _, s := range installSteps {
		if s.stepKey == "" {
			continue
		}
		if m.IsDone(s.stepKey) {
			fmt.Printf("  %s  %s\n", ui.StyleSuccess.Render("✓"), s.name)
		} else {
			fmt.Printf("  %s  %s\n", ui.StyleError.Render("✗"), s.name)
		}
	}
	fmt.Println()

	var choice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("请选择").
				Options(
					huh.NewOption("从中断处继续安装", "resume"),
					huh.NewOption("重新开始（清理已下载文件）", "restart"),
					huh.NewOption("退出", "exit"),
				).
				Value(&choice),
		),
	)

	if err := form.Run(); err != nil || choice == "exit" {
		return
	}

	if choice == "restart" {
		state.Rollback(m)
		newM := state.NewManifest(Version)
		runInstaller(newM)
		return
	}

	// 续传
	runInstaller(m)
}
