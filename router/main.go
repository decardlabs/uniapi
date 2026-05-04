package router

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/decardlabs/uniapi/common/config"
	"github.com/decardlabs/uniapi/common/logger"
)

func SetRouter(router *gin.Engine, buildFS fs.FS) {
	SetApiRouter(router)
	SetDashboardRouter(router)
	SetRelayRouter(router)
	frontendBaseUrl := config.FrontendBaseURL
	if config.IsMasterNode && frontendBaseUrl != "" {
		frontendBaseUrl = ""
		logger.Logger.Info("FRONTEND_BASE_URL is ignored on master node")
	}
	if frontendBaseUrl == "" {
		SetWebRouter(router, buildFS)
	} else {
		router.NoRoute(func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s%s", frontendBaseUrl, c.Request.RequestURI))
		})
	}
}
