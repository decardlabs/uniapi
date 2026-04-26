package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	gmw "github.com/Laisky/gin-middlewares/v7"
	"github.com/songquanpeng/one-api/model"
)

// CreateRechargeRequest 用户提交充值申请
func CreateRechargeRequest(c *gin.Context) {
	ctx := gmw.Ctx(c)
	var req struct {
		Amount int64  `json:"amount" binding:"required,gt=0"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid request: amount must be greater than 0",
		})
		return
	}
	userId := c.GetInt(ctxkey.Id)
	recharge, err := model.CreateRechargeRequest(ctx, userId, req.Amount, req.Remark)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Recharge request submitted successfully",
		"data":    recharge,
	})
}

// GetAllRechargeRequests 管理员获取所有充值请求（分页）
func GetAllRechargeRequests(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 1 {
		p = 1
	}
	size, _ := strconv.Atoi(c.Query("size"))
	if size <= 0 {
		size = config.DefaultItemsPerPage
	}
	if size > config.MaxItemsPerPage {
		size = config.MaxItemsPerPage
	}

	requests, err := model.GetAllRechargeRequests((p-1)*size, size)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	total, err := model.GetRechargeRequestCount()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    requests,
		"total":   total,
	})
}

// GetUserRechargeRequests 获取当前用户的充值记录
func GetUserRechargeRequests(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 1 {
		p = 1
	}
	size, _ := strconv.Atoi(c.Query("size"))
	if size <= 0 {
		size = config.DefaultItemsPerPage
	}
	if size > config.MaxItemsPerPage {
		size = config.MaxItemsPerPage
	}

	requests, err := model.GetUserRechargeRequests(userId, (p-1)*size, size)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	total, err := model.GetUserRechargeRequestCount(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    requests,
		"total":   total,
	})
}

// ApproveRechargeRequest 管理员审批通过
func ApproveRechargeRequest(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	var req struct {
		AdminRemark string `json:"admin_remark"`
	}
	c.ShouldBindJSON(&req)

	reviewerId := c.GetInt(ctxkey.Id)
	err = model.ApproveRechargeRequest(ctx, id, reviewerId, req.AdminRemark)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Recharge request approved"})
}

// RejectRechargeRequest 管理员拒绝
func RejectRechargeRequest(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Invalid ID"})
		return
	}

	var req struct {
		AdminRemark string `json:"admin_remark" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "Reject reason is required"})
		return
	}

	reviewerId := c.GetInt(ctxkey.Id)
	err = model.RejectRechargeRequest(ctx, id, reviewerId, req.AdminRemark)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Recharge request rejected"})
}
