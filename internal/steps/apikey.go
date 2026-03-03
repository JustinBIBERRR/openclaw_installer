package steps

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"openclaw-manager/internal/proxy"
	"openclaw-manager/internal/state"
	"openclaw-manager/internal/ui"
)

// providerConfig 定义每个 AI 服务商的校验方式
type providerConfig struct {
	name       string
	keyPattern *regexp.Regexp
	validateFn func(key string, client *http.Client) error
}

var providers = map[string]*providerConfig{
	"anthropic": {
		name:       "Anthropic (Claude)",
		keyPattern: regexp.MustCompile(`^sk-ant-`),
		validateFn: validateAnthropic,
	},
	"openai": {
		name:       "OpenAI (GPT)",
		keyPattern: regexp.MustCompile(`^sk-[A-Za-z0-9]`),
		validateFn: validateOpenAI,
	},
	"custom": {
		name:       "其他 OpenAI 兼容服务",
		keyPattern: regexp.MustCompile(`.+`),
		validateFn: nil, // 自定义服务不校验连通性
	},
}

// RunAPIKey 引导用户选择服务商、输入并验证 API Key
func RunAPIKey(m *state.Manifest) error {
	if m.IsDone(state.StepAPIKeySaved) {
		return nil
	}

	var providerID string
	var apiKey string
	var baseURL string // 仅 custom 服务商使用

	// 第一步：选择服务商
	providerForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("请选择 AI 服务商").
				Description("选择后将使用对应的 API Key 格式验证").
				Options(
					huh.NewOption("Anthropic (Claude)", "anthropic"),
					huh.NewOption("OpenAI (GPT)", "openai"),
					huh.NewOption("其他 OpenAI 兼容服务", "custom"),
				).
				Value(&providerID),
		),
	)
	if err := providerForm.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return fmt.Errorf("用户取消操作")
		}
		return err
	}

	cfg := providers[providerID]

	// 第二步：如果是自定义服务，询问 Base URL
	if providerID == "custom" {
		baseURLForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("请输入服务 Base URL").
					Description("例如: https://api.example.com/v1").
					Placeholder("https://").
					Value(&baseURL),
			),
		)
		if err := baseURLForm.Run(); err != nil {
			return err
		}
	}

	// 第三步：输入并验证 Key（最多重试 3 次）
	const maxRetries = 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		keyForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("请输入 %s API Key", cfg.name)).
					Description(getKeyHint(providerID)).
					EchoMode(huh.EchoModePassword).
					Validate(func(v string) error {
						v = strings.TrimSpace(v)
						if v == "" {
							return fmt.Errorf("API Key 不能为空")
						}
						if !cfg.keyPattern.MatchString(v) {
							return fmt.Errorf("Key 格式不符，请检查是否完整复制")
						}
						return nil
					}).
					Value(&apiKey),
			),
		)
		if err := keyForm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return fmt.Errorf("用户取消操作")
			}
			return err
		}
		apiKey = strings.TrimSpace(apiKey)

		// 提示识别结果
		ui.PrintOK(fmt.Sprintf("已识别为 %s Key", cfg.name))

		// 跳过自定义服务的网络校验
		if providerID == "custom" || cfg.validateFn == nil {
			break
		}

		// HTTP 连通性校验
		ui.PrintInfo("正在验证 Key 连通性...")
		client := proxy.NewHTTPClient(10 * time.Second)
		err := cfg.validateFn(apiKey, client)
		if err == nil {
			ui.PrintOK("验证通过")
			break
		}

		// 按错误类型分支处理
		switch classifyValidationError(err) {
		case errClassInvalidKey:
			if attempt < maxRetries {
				ui.PrintError(fmt.Sprintf("API Key 无效（第 %d/%d 次），请重新输入", attempt, maxRetries))
				apiKey = ""
				continue
			}
			// 第三次仍失败：询问是否跳过
			if ui.AskConfirm("连续验证失败，是否跳过验证直接继续？") {
				ui.PrintWarn("已跳过验证，API Key 将保存但未经确认")
				goto saveKey
			}
			return fmt.Errorf("API Key 验证失败，安装终止")

		case errClassNetwork:
			ui.PrintWarn(fmt.Sprintf("网络连接失败: %v", err))
			if ui.AskConfirm("是否跳过验证，直接保存 Key 继续安装？") {
				ui.PrintWarn("已跳过验证（建议检查代理设置）")
				goto saveKey
			}
			return fmt.Errorf("网络验证失败，安装终止")

		default:
			ui.PrintWarn(fmt.Sprintf("验证时发生未知错误: %v", err))
			goto saveKey
		}
	}

saveKey:
	meta := map[string]string{
		"provider": providerID,
	}
	if baseURL != "" {
		meta["base_url"] = baseURL
	}

	// 将 API Key 写入内存（config 步骤再写文件）
	m.Steps[state.StepAPIKeySaved] = &state.StepRecord{
		Status: "pending", // MarkDone 会设置为 done
	}
	// 临时存储到 meta，供 config 步骤使用
	m.Steps["_api_key_tmp"] = &state.StepRecord{
		Status: "tmp",
		Meta:   map[string]string{"key": apiKey, "provider": providerID, "base_url": baseURL},
	}

	return m.MarkDone(state.StepAPIKeySaved, meta)
}

// GetSavedAPIKey 从临时存储读取 API Key（供 config 步骤使用）
func GetSavedAPIKey(m *state.Manifest) (key, provider, baseURL string) {
	if r, ok := m.Steps["_api_key_tmp"]; ok && r.Meta != nil {
		return r.Meta["key"], r.Meta["provider"], r.Meta["base_url"]
	}
	return "", "", ""
}

func getKeyHint(provider string) string {
	switch provider {
	case "anthropic":
		return "格式: sk-ant-api03-... (从 console.anthropic.com 获取)"
	case "openai":
		return "格式: sk-... (从 platform.openai.com 获取)"
	default:
		return "请输入您的 API Key"
	}
}

type errClass int

const (
	errClassInvalidKey errClass = iota
	errClassNetwork
	errClassUnknown
)

func classifyValidationError(err error) errClass {
	if err == nil {
		return errClassUnknown
	}
	msg := err.Error()
	if strings.Contains(msg, "401") || strings.Contains(msg, "invalid") || strings.Contains(msg, "unauthorized") {
		return errClassInvalidKey
	}
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "connection") ||
		strings.Contains(msg, "no such host") || strings.Contains(msg, "dial") {
		return errClassNetwork
	}
	return errClassUnknown
}

func validateAnthropic(key string, client *http.Client) error {
	req, err := http.NewRequest(http.MethodGet, "https://api.anthropic.com/v1/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-api-key", key)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusForbidden, http.StatusTooManyRequests:
		return nil // Key 有效（403 = 无模型访问权限，429 = 限速，均视为有效）
	case http.StatusUnauthorized:
		return fmt.Errorf("401 unauthorized: Key 无效")
	default:
		return nil // 其他状态码不阻断
	}
}

func validateOpenAI(key string, client *http.Client) error {
	req, err := http.NewRequest(http.MethodGet, "https://api.openai.com/v1/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusForbidden, http.StatusTooManyRequests:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("401 unauthorized: Key 无效")
	default:
		return nil
	}
}
