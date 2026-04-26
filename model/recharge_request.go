package model

import (
	"context"
	"fmt"

	"github.com/Laisky/errors/v2"
	"gorm.io/gorm"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
)

const (
	RechargeStatusPending  = 1 // 待审批
	RechargeStatusApproved = 2 // 已通过
	RechargeStatusRejected = 3 // 已拒绝
)

type RechargeRequest struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id" gorm:"index"`
	Amount       int64  `json:"amount"`       // 充值金额（内部单位）
	Quota        int64  `json:"quota"`        // 实际增加的 quota
	Status       int    `json:"status" gorm:"default:1"`
	Remark       string `json:"remark"`        // 用户备注
	AdminRemark  string `json:"admin_remark"`  // 管理员备注（拒绝原因等）
	CreatedTime  int64  `json:"created_time" gorm:"bigint"`
	ReviewedTime int64  `json:"reviewed_time" gorm:"bigint"`
	ReviewerId   int    `json:"reviewer_id"`   // 审批人ID
	CreatedAt    int64  `json:"created_at" gorm:"bigint;autoCreateTime:milli"`
	UpdatedAt    int64  `json:"updated_at" gorm:"bigint;autoUpdateTime:milli"`

	// 只读关联字段
	User *User `json:"user,omitempty" gorm:"foreignKey:UserId"`
}

var rechargeSortFields = map[string]string{
	"id":            "id",
	"user_id":       "user_id",
	"amount":        "amount",
	"status":        "status",
	"created_time":  "created_time",
	"reviewed_time": "reviewed_time",
	"created_at":    "created_at",
	"updated_at":    "updated_at",
}

// GetAllRechargeRequests 获取所有充值请求（管理员用，分页）
func GetAllRechargeRequests(startIdx int, num int) ([]*RechargeRequest, error) {
	var requests []*RechargeRequest
	err := DB.Preload("User").Order("id desc").Limit(num).Offset(startIdx).Find(&requests).Error
	return requests, err
}

// GetRechargeRequestCount 获取充值请求总数
func GetRechargeRequestCount() (count int64, err error) {
	err = DB.Model(&RechargeRequest{}).Count(&count).Error
	return count, err
}

// SearchRechargeRequests 搜索充值请求
func SearchRechargeRequests(keyword string, startIdx int, num int, sortBy string, sortOrder string) (requests []*RechargeRequest, total int64, err error) {
	db := DB.Model(&RechargeRequest{})
	if keyword != "" {
		db = db.Joins("LEFT JOIN users ON recharge_requests.user_id = users.id").
			Where("recharge_requests.id = ? OR users.username LIKE ? OR recharge_requests.remark LIKE ?", keyword, keyword+"%", keyword+"%")
	}
	db = db.Order(ValidateOrderClause(sortBy, sortOrder, rechargeSortFields, "id desc"))
	err = db.Count(&total).Limit(num).Offset(startIdx).Find(&requests).Error
	return requests, total, err
}

// GetRechargeRequestById 根据 ID 获取充值请求
func GetRechargeRequestById(id int) (*RechargeRequest, error) {
	if id == 0 {
		return nil, errors.New("id is empty!")
	}
	req := RechargeRequest{Id: id}
	var err error = nil
	err = DB.Preload("User").First(&req, "id = ?", id).Error
	return &req, err
}

// GetUserRechargeRequests 获取某用户的充值请求
func GetUserRechargeRequests(userId int, startIdx int, num int) ([]*RechargeRequest, error) {
	var requests []*RechargeRequest
	err := DB.Where("user_id = ?", userId).Order("id desc").Limit(num).Offset(startIdx).Find(&requests).Error
	return requests, err
}

// GetUserRechargeRequestCount 获取某用户的充值请求总数
func GetUserRechargeRequestCount(userId int) (count int64, err error) {
	err = DB.Model(&RechargeRequest{}).Where("user_id = ?", userId).Count(&count).Error
	return count, err
}

// CreateRechargeRequest 用户创建充值申请
func CreateRechargeRequest(ctx context.Context, userId int, amount int64, remark string) (*RechargeRequest, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	req := &RechargeRequest{
		UserId:      userId,
		Amount:      amount,
		Status:      RechargeStatusPending,
		Remark:      remark,
		CreatedTime: helper.GetTimestamp(),
	}
	if err := DB.Create(req).Error; err != nil {
		return nil, errors.Wrap(err, "create recharge request")
	}
	RecordLog(ctx, userId, LogTypeTopup, fmt.Sprintf("Submitted recharge request for %s", common.LogQuota(amount)))
	return req, nil
}

// ApproveRechargeRequest 管理员审批通过充值请求
func ApproveRechargeRequest(ctx context.Context, id int, reviewerId int, adminRemark string) error {
	req, err := GetRechargeRequestById(id)
	if err != nil {
		return errors.Wrap(err, "get recharge request")
	}
	if req.Status != RechargeStatusPending {
		return errors.New("request is not pending")
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		// 增加用户额度
		if err := tx.Model(&User{}).Where("id = ?", req.UserId).Update("quota", gorm.Expr("quota + ?", req.Amount)).Error; err != nil {
			return errors.Wrapf(err, "increase user %d quota", req.UserId)
		}
		// 更新请求状态
		now := helper.GetTimestamp()
		if err := tx.Model(req).Updates(map[string]interface{}{
			"status":        RechargeStatusApproved,
			"quota":         req.Amount,
			"admin_remark":  adminRemark,
			"reviewed_time": now,
			"reviewer_id":   reviewerId,
		}).Error; err != nil {
			return errors.Wrap(err, "update recharge request status")
		}
		return nil
	})
	if err != nil {
		return err
	}
	RecordLog(ctx, req.UserId, LogTypeTopup, fmt.Sprintf("Recharge request #%d approved: %s", id, common.LogQuota(req.Amount)))
	return nil
}

// RejectRechargeRequest 管理员拒绝充值请求
func RejectRechargeRequest(ctx context.Context, id int, reviewerId int, adminRemark string) error {
	req, err := GetRechargeRequestById(id)
	if err != nil {
		return errors.Wrap(err, "get recharge request")
	}
	if req.Status != RechargeStatusPending {
		return errors.New("request is not pending")
	}

	now := helper.GetTimestamp()
	err = DB.Model(req).Updates(map[string]interface{}{
		"status":        RechargeStatusRejected,
		"admin_remark":  adminRemark,
		"reviewed_time": now,
		"reviewer_id":   reviewerId,
	}).Error
	if err != nil {
		return errors.Wrap(err, "update recharge request status")
	}
	RecordLog(ctx, req.UserId, LogTypeTopup, fmt.Sprintf("Recharge request #%d rejected", id))
	return nil
}
