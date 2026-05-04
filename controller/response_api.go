package controller

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/decardlabs/uniapi/common/helper"
	"github.com/decardlabs/uniapi/monitor"
	rcontroller "github.com/decardlabs/uniapi/relay/controller"
	metalib "github.com/decardlabs/uniapi/relay/meta"
)

func RelayResponseGet(c *gin.Context) {
	meta := metalib.GetByContext(c)
	startTime := time.Now()

	PrometheusMonitor.RecordChannelRequest(meta, startTime)

	if bizErr := rcontroller.RelayResponseAPIGetHelper(c); bizErr != nil {
		PrometheusMonitor.RecordRelayRequest(c, meta, startTime, false, 0, 0, 0)
		monitor.Emit(meta.ChannelId, false)

		requestId := c.GetString(helper.RequestIdKey)
		bizErr.Error.Message = helper.MessageWithRequestId(bizErr.Error.Message, requestId)
		c.JSON(bizErr.StatusCode, gin.H{"error": bizErr.Error})
		return
	}

	monitor.Emit(meta.ChannelId, true)
	PrometheusMonitor.RecordRelayRequest(c, meta, startTime, true, 0, 0, 0)
}

func RelayResponseDelete(c *gin.Context) {
	meta := metalib.GetByContext(c)
	startTime := time.Now()

	PrometheusMonitor.RecordChannelRequest(meta, startTime)

	if bizErr := rcontroller.RelayResponseAPIDeleteHelper(c); bizErr != nil {
		PrometheusMonitor.RecordRelayRequest(c, meta, startTime, false, 0, 0, 0)
		monitor.Emit(meta.ChannelId, false)

		requestId := c.GetString(helper.RequestIdKey)
		bizErr.Error.Message = helper.MessageWithRequestId(bizErr.Error.Message, requestId)
		c.JSON(bizErr.StatusCode, gin.H{"error": bizErr.Error})
		return
	}

	monitor.Emit(meta.ChannelId, true)
	PrometheusMonitor.RecordRelayRequest(c, meta, startTime, true, 0, 0, 0)
}

func RelayResponseCancel(c *gin.Context) {
	meta := metalib.GetByContext(c)
	startTime := time.Now()

	PrometheusMonitor.RecordChannelRequest(meta, startTime)

	if bizErr := rcontroller.RelayResponseAPICancelHelper(c); bizErr != nil {
		PrometheusMonitor.RecordRelayRequest(c, meta, startTime, false, 0, 0, 0)
		monitor.Emit(meta.ChannelId, false)

		requestId := c.GetString(helper.RequestIdKey)
		bizErr.Error.Message = helper.MessageWithRequestId(bizErr.Error.Message, requestId)
		c.JSON(bizErr.StatusCode, gin.H{"error": bizErr.Error})
		return
	}

	monitor.Emit(meta.ChannelId, true)
	PrometheusMonitor.RecordRelayRequest(c, meta, startTime, true, 0, 0, 0)
}
