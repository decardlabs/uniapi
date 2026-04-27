package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	gmw "github.com/Laisky/gin-middlewares/v7"
	"github.com/songquanpeng/one-api/model"
)

// --- Request/Response structs ---

type createPoolRequest struct {
	Name       string `json:"name" binding:"required"`
	TotalQuota int64  `json:"total_quota" binding:"required,min=1"`
	PeriodType string `json:"period_type" binding:"required,oneof=monthly quarterly yearly oneoff"`
	PeriodKey  string `json:"period_key" binding:"required"`
}

type purchaseToPoolRequest struct {
	Amount int64  `json:"amount" binding:"required,gt=0"`
	Remark string `json:"remark"`
}

type allocateFromPoolRequest struct {
	UserId int    `json:"user_id" binding:"required"`
	Amount int64  `json:"amount" binding:"required,gt=0"`
	Remark string `json:"remark"`
}

type recallFromPoolRequest struct {
	UserId int    `json:"user_id" binding:"required"`
	Amount int64  `json:"amount" binding:"required,gt=0"`
	Remark string `json:"remark"`
}

type rollOverPoolRequest struct {
	NewPeriodKey string `json:"new_period_key" binding:"required"`
	NewName      string `json:"new_name" binding:"required"`
}

// --- Handlers ---

func CreatePool(c *gin.Context) {
	var req createPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	ctx := gmw.Ctx(c)
	pool := &model.GlobalPool{
		Name:       req.Name,
		TotalQuota: req.TotalQuota,
		PeriodType: req.PeriodType,
		PeriodKey:  req.PeriodKey,
		Status:     model.PoolStatusActive,
	}
	if err := model.CreatePool(ctx, pool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": pool})
}

func GetAllPools(c *gin.Context) {
	ctx := gmw.Ctx(c)
	periodType := c.Query("period_type")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	pools, total, err := model.GetAllPools(ctx, periodType, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": gin.H{
		"items": pools, "total": total, "page": page, "page_size": pageSize,
	}})
}

func GetPool(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	pool, err := model.GetPoolById(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "pool not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": pool})
}

func PurchaseToPool(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	var req purchaseToPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := model.PurchaseToPool(ctx, id, req.Amount, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

func AllocateFromPool(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	var req allocateFromPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := model.AllocateFromPool(ctx, id, req.UserId, req.Amount, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

func RecallFromPool(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	var req recallFromPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := model.RecallFromPool(ctx, id, req.UserId, req.Amount, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

func GetPoolAllocations(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	allocs, err := model.GetPoolAllocations(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": allocs})
}

func GetPoolTransactions(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	txns, total, err := model.GetPoolTransactions(ctx, id, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": gin.H{
		"items": txns, "total": total, "page": page, "page_size": pageSize,
	}})
}

func GetPoolReconciliation(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	pool, allocs, err := model.GetPoolReconciliation(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": gin.H{
		"pool": pool, "allocations": allocs,
	}})
}

func ClosePool(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	if err := model.ClosePool(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
}

func RollOverPool(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}
	var req rollOverPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	newPool, err := model.RollOverPool(ctx, id, req.NewPeriodKey, req.NewName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": newPool})
}
