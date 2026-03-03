//go:build windows

package proxy

import (
	"net/url"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func fromEnv() *url.URL {
	for _, key := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"} {
		if v := os.Getenv(key); v != "" {
			if u, err := url.Parse(v); err == nil && u.Host != "" {
				return u
			}
		}
	}
	return nil
}

func fromWinINet() *url.URL {
	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return nil
	}
	defer k.Close()

	enabled, _, err := k.GetIntegerValue("ProxyEnable")
	if err != nil || enabled == 0 {
		return nil
	}

	server, _, err := k.GetStringValue("ProxyServer")
	if err != nil || server == "" {
		return nil
	}

	// ProxyServer 格式可能是 "host:port" 或 "http=host:port;https=host:port;..."
	if strings.Contains(server, "=") {
		// 解析分号分隔的协议=地址格式
		for _, part := range strings.Split(server, ";") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "https=") {
				server = strings.TrimPrefix(part, "https=")
				break
			}
			if strings.HasPrefix(part, "http=") {
				server = strings.TrimPrefix(part, "http=")
				// 继续查找 https=
			}
		}
	}

	if !strings.Contains(server, "://") {
		server = "http://" + server
	}

	u, err := url.Parse(server)
	if err != nil || u.Host == "" {
		return nil
	}
	return u
}
