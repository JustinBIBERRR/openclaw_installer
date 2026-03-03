package ui

import (
	"fmt"
	"strings"
	"time"
)

// PrintStepHeader 打印步骤标题，如 [2/5] 安装 OpenClaw CLI
func PrintStepHeader(current, total int, name string) {
	label := StyleStepHeader.Render(fmt.Sprintf(" %d/%d ", current, total))
	title := StyleBold.Render(" " + name)
	fmt.Printf("\n%s%s\n", label, title)
}

// PrintOK 打印成功行
func PrintOK(msg string) {
	fmt.Printf("  %s  %s\n", StyleSuccess.Render("✓"), msg)
}

// PrintSkip 打印跳过行
func PrintSkip(msg string) {
	fmt.Printf("  %s  %s\n", StyleSkip.Render("↷"), StyleSkip.Render(msg+" (已完成，跳过)"))
}

// PrintWarn 打印警告行
func PrintWarn(msg string) {
	fmt.Printf("  %s  %s\n", StyleWarn.Render("!"), StyleWarn.Render(msg))
}

// PrintInfo 打印信息行
func PrintInfo(msg string) {
	fmt.Printf("  %s  %s\n", StyleInfo.Render("·"), msg)
}

// PrintError 打印错误行
func PrintError(msg string) {
	fmt.Printf("\n  %s  %s\n", StyleError.Render("✗"), StyleError.Render(msg))
}

// PrintFatalError 打印安装失败终止信息
func PrintFatalError() {
	box := StyleBox.Render(
		StyleError.Render("安装中断") + "\n\n" +
			"进度已保存，下次运行将从断点继续。\n" +
			StyleDim.Render("如需重新安装，请使用 --uninstall 先卸载。"),
	)
	fmt.Println()
	fmt.Println(box)
	fmt.Println()
	fmt.Print("  按 Enter 退出...")
	fmt.Scanln()
}

// PrintSuccess 打印安装完成界面（带倒计时）
func PrintSuccess(url string) {
	content := StyleSuccess.Render("全部完成！") + "\n\n" +
		fmt.Sprintf("  Gateway 已启动 → %s\n", StyleInfo.Render(url)) +
		"  桌面快捷方式已创建\n\n" +
		"  下次双击此程序将直接进入管理界面。"

	fmt.Println()
	fmt.Println(StyleSuccessBox.Render(content))
	fmt.Println()

	for i := 5; i > 0; i-- {
		fmt.Printf("\r  正在打开浏览器，%d 秒后关闭窗口...  ", i)
		time.Sleep(time.Second)
	}
	fmt.Println()
}

// WithSpinner 在 fn 执行期间显示旋转动画
func WithSpinner(msg string, fn func() error) error {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan error, 1)

	go func() { done <- fn() }()

	i := 0
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case err := <-done:
			// 清除 Spinner 行
			fmt.Printf("\r%s\r", strings.Repeat(" ", 72))
			return err
		case <-ticker.C:
			frame := StyleAccent.Render(frames[i%len(frames)])
			fmt.Printf("\r  %s  %s", frame, msg)
			i++
		}
	}
}

// PrintDownloadProgress 显示下载进度（覆盖当前行）
func PrintDownloadProgress(downloaded, total int64, speed float64) {
	if total <= 0 {
		fmt.Printf("\r  正在下载... %s", formatBytes(downloaded))
		return
	}
	pct := float64(downloaded) / float64(total)
	width := 38
	filled := int(pct * float64(width))
	bar := StyleSuccess.Render(strings.Repeat("█", filled)) +
		StyleDim.Render(strings.Repeat("░", width-filled))
	speedStr := fmt.Sprintf("↓ %s/s", formatBytes(int64(speed)))
	fmt.Printf("\r  [%s]  %4.0f%%  %s / %s  %s  ",
		bar, pct*100, formatBytes(downloaded), formatBytes(total),
		StyleDim.Render(speedStr))
}

// PrintDownloadDone 完成下载后换行并打印完成信息
func PrintDownloadDone(total int64) {
	fmt.Printf("\r%s\r", strings.Repeat(" ", 80))
	PrintOK(fmt.Sprintf("下载完成 (%s)", formatBytes(total)))
}

// AskConfirm 询问 Y/N，返回 true 表示确认
func AskConfirm(prompt string) bool {
	fmt.Printf("\n  %s  %s [Y/n] ", StyleWarn.Render("?"), prompt)
	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "" || input == "y" || input == "yes"
}

// AskInput 简单文本输入
func AskInput(prompt string) string {
	fmt.Printf("\n  %s  %s: ", StyleInfo.Render("?"), prompt)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
