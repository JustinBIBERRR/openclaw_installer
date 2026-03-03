package env

import (
	"os"
	"path/filepath"
	"strings"

	"openclaw-manager/internal/state"
)

// ShadowEnv 返回注入了内置 Node.js Runtime 路径的环境变量切片。
// 不修改任何注册表或全局系统变量，完全隔离。
func ShadowEnv(m *state.Manifest) []string {
	nodeDir := m.NodeDir()
	npmBinDir := m.NPMGlobalBin()
	npmPrefix := m.NPMGlobalPrefix()

	sep := string(os.PathListSeparator)
	newPath := nodeDir + sep + npmBinDir + sep + os.Getenv("PATH")

	// 过滤掉已有的 PATH / OPENCLAW_HOME / NPM_CONFIG_PREFIX，再追加我们的版本
	base := os.Environ()
	env := make([]string, 0, len(base)+4)
	for _, kv := range base {
		key := strings.SplitN(kv, "=", 2)[0]
		upper := strings.ToUpper(key)
		if upper == "PATH" || upper == "OPENCLAW_HOME" || upper == "NPM_CONFIG_PREFIX" {
			continue
		}
		env = append(env, kv)
	}

	env = append(env,
		"PATH="+newPath,
		"OPENCLAW_HOME="+m.InstallDir,
		"NPM_CONFIG_PREFIX="+npmPrefix,
		// 禁用 npm 的 update-notifier，避免安装时额外输出
		"NO_UPDATE_NOTIFIER=1",
		// npm 缓存目录也隔离在我们的安装目录内
		"NPM_CONFIG_CACHE="+filepath.Join(m.InstallDir, "npm-cache"),
	)

	return env
}
