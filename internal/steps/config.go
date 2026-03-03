package steps

import (
	"encoding/json"
	"fmt"
	"os"

	"openclaw-manager/internal/state"
	"openclaw-manager/internal/ui"
)

// openclawConfig 对应 openclaw.json 的结构
type openclawConfig struct {
	Provider string `json:"provider"`
	APIKey   string `json:"apiKey"`
	BaseURL  string `json:"baseUrl,omitempty"`
	Port     int    `json:"port"`
}

// RunConfig 将 API Key 和配置写入 openclaw.json
func RunConfig(m *state.Manifest) error {
	if m.IsDone(state.StepConfigWritten) {
		return nil
	}

	key, provider, baseURL := GetSavedAPIKey(m)
	if key == "" {
		return fmt.Errorf("未找到已保存的 API Key，请重新运行安装程序")
	}

	cfg := openclawConfig{
		Provider: provider,
		APIKey:   key,
		BaseURL:  baseURL,
		Port:     m.GatewayPort,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	configPath := m.ConfigFile()
	if err := os.MkdirAll(m.InstallDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 原子写：先写临时文件再重命名
	tmpPath := configPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}
	if err := os.Rename(tmpPath, configPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("保存配置失败: %w", err)
	}

	ui.PrintOK(fmt.Sprintf("配置已保存 → %s", configPath))

	// 清理临时 API Key（已写入磁盘，内存中不再需要）
	delete(m.Steps, "_api_key_tmp")

	return m.MarkDone(state.StepConfigWritten)
}
