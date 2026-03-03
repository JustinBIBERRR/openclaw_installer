//go:build !windows

package proxy

import (
	"net/url"
	"os"
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
	return nil
}
