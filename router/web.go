package router

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	"github.com/decardlabs/uniapi/common"
	"github.com/decardlabs/uniapi/common/config"
	"github.com/decardlabs/uniapi/controller"
	"github.com/decardlabs/uniapi/middleware"
)

func SetWebRouter(router *gin.Engine, buildFS fs.FS) {
	indexPageData, _ := fs.ReadFile(buildFS, fmt.Sprintf("web/build/%s/index.html", config.Theme))
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.GlobalWebRateLimit())
	router.Use(middleware.Cache())
	router.Use(static.Serve("/", common.EmbedFolder(buildFS, fmt.Sprintf("web/build/%s", config.Theme))))
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.RequestURI, "/v1") || strings.HasPrefix(c.Request.RequestURI, "/api") {
			controller.RelayNotFound(c)
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexPageData)
	})
}
