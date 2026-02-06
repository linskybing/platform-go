package filebrowser

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/pkg/response"
)

// ProxyConfig holds configuration for reverse proxy to FileBrowser
type ProxyConfig struct {
	ServiceName string
	Namespace   string
	PathPrefix  string
}

// ProxyHandler creates a reverse proxy handler for FileBrowser access
func ProxyHandler(cfg ProxyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", cfg.ServiceName, cfg.Namespace, DefaultPort)

		remote, err := url.Parse(targetURL)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Invalid target URL")
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(remote)

		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)

			// Remove the Gin route prefix so FileBrowser receives the correct path
			if cfg.PathPrefix != "" {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, cfg.PathPrefix)
			}

			req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
			req.Header.Set("X-Forwarded-Proto", "http")
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
