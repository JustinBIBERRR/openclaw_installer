package steps

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"openclaw-manager/internal/env"
	"openclaw-manager/internal/state"
	"openclaw-manager/internal/ui"
)

const npmPackage = "openclaw"

// RunOpenClawInstall 通过 npm 安装 OpenClaw CLI
func RunOpenClawInstall(m *state.Manifest) error {
	if m.IsDone(state.StepCLIInstalled) {
		return nil
	}

	if err := os.MkdirAll(m.NPMGlobalPrefix(), 0755); err != nil {
		return fmt.Errorf("创建 npm 目录失败: %w", err)
	}

	shadowEnv := env.ShadowEnv(m)

	// npm install -g openclaw，通过 cmd /c 调用以便找到 npm.cmd
	args := []string{"/c", "npm", "install", "-g", npmPackage, "--no-audit", "--no-fund"}
	cmd := exec.Command("cmd", args...)
	cmd.Env = shadowEnv
	setHideWindow(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 npm 失败: %w", err)
	}

	// 合并并流式打印 npm 输出
	done := make(chan struct{})
	go func() {
		defer close(done)
		merged := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(merged)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			// 只显示有意义的行，过滤 npm 冗余日志
			if shouldShowNPMLine(line) {
				fmt.Printf("  %s  %s\n", ui.StyleDim.Render("›"), ui.StyleDim.Render(line))
			}
		}
	}()

	<-done
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("npm install 失败: %w", err)
	}

	ui.PrintOK(fmt.Sprintf("OpenClaw CLI 安装成功"))
	return m.MarkDone(state.StepCLIInstalled)
}

// shouldShowNPMLine 过滤 npm 输出中无意义的噪音行
func shouldShowNPMLine(line string) bool {
	lower := strings.ToLower(line)
	noise := []string{
		"npm warn", "npm notice",
		"gyp verb", "gyp info",
		"added 0 packages",
	}
	for _, n := range noise {
		if strings.HasPrefix(lower, n) {
			return false
		}
	}
	return true
}
