package model

import (
	"context"
	"fmt"

	"github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PeriodType constants
const (
	PoolPeriodMonthly   = "monthly"
	PoolPeriodQuarterly = "quarterly"
	PoolPeriodYearly    = "yearly"
	PoolPeriodOneoff    = "oneoff"
)

// PoolStatus constants
const (
	PoolStatusActive = "active"
	PoolStatusClosed = "closed"
)

// TransactionType constants
const (
	PoolTxnPurchase = "purchase"
	PoolTxnAllocate = "allocate"
	PoolTxnRecall   = "recall"
	PoolTxnAdjust   = "adjust"
)

// TransactionDirection constants
const (
	PoolTxnIn  = "in"
	PoolTxnOut = "out"
)

// GlobalPool represents a token budget pool.
type GlobalPool struct {
	Id         int    `json:"id" gorm:"primaryKey"`
	Name       string `json:"name" gorm:"type:varchar(100);not null"`
	TotalQuota int64  `json:"total_quota" gorm:"not null;default:0"`
	UsedQuota  int64  `json:"used_quota" gorm:"not null;default:0"`
	PeriodType string `json:"period_type" gorm:"type:varchar(20);not null"`
	PeriodKey  string `json:"period_key" gorm:"type:varchar(20);not null;index"`
	Status     string `json:"status" gorm:"type:varchar(20);not null;default:'active';index"`
	CreatedAt  int64  `json:"created_at" gorm:"not null"`
	ClosedAt   int64  `json:"closed_at"`
}

func (p *GlobalPool) AvailableQuota() int64 {
	return p.TotalQuota - p.UsedQuota
}

// PoolAllocation tracks quota allocated to a user from a pool.
type PoolAllocation struct {
	Id             int   `json:"id" gorm:"primaryKey"`
	PoolId         int   `json:"pool_id" gorm:"index;not null"`
	UserId         int   `json:"user_id" gorm:"index;not null"`
	AllocatedQuota int64 `json:"allocated_quota" gorm:"not null;default:0"`
	RecalledQuota  int64 `json:"recalled_quota" gorm:"not null;default:0"`
	CreatedAt      int64 `json:"created_at" gorm:"not null"`
	UpdatedAt      int64 `json:"updated_at" gorm:"not null"`
}

func (a *PoolAllocation) NetAllocated() int64 {
	return a.AllocatedQuota - a.RecalledQuota
}

// PoolTransaction records all pool-related money movements.
type PoolTransaction struct {
	Id           int    `json:"id" gorm:"primaryKey"`
	PoolId       int    `json:"pool_id" gorm:"index;not null"`
	UserId       int    `json:"user_id" gorm:"index"`
	Type         string `json:"type" gorm:"type:varchar(20);not null"`
	Amount       int64  `json:"amount" gorm:"not null"`
	Direction    string `json:"direction" gorm:"type:varchar(10);not null"`
	BalanceAfter int64  `json:"balance_after" gorm:"not null"`
	Remark       string `json:"remark" gorm:"type:text"`
	CreatedAt    int64  `json:"created_at" gorm:"not null"`
}

// ============================================================
// CRUD Functions
// ============================================================

// CreatePool creates a new budget pool.
func CreatePool(ctx context.Context, pool *GlobalPool) error {
	pool.CreatedAt = helper.GetTimestamp()
	result := DB.WithContext(ctx).Create(pool)
	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to create pool")
	}
	return nil
}

// GetPoolById retrieves a pool by ID.
func GetPoolById(ctx context.Context, id int) (*GlobalPool, error) {
	var pool GlobalPool
	result := DB.WithContext(ctx).First(&pool, id)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to get pool")
	}
	return &pool, nil
}

// GetAllPools returns paginated pools with optional filters.
func GetAllPools(ctx context.Context, periodType string, status string, page int, pageSize int) ([]*GlobalPool, int64, error) {
	var pools []*GlobalPool
	var total int64
	query := DB.WithContext(ctx).Model(&GlobalPool{})
	if periodType != "" {
		query = query.Where("period_type = ?", periodType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	result := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&pools)
	if result.Error != nil {
		return nil, 0, errors.Wrap(result.Error, "failed to get pools")
	}
	return pools, total, nil
}

// UpdatePool saves changes to an existing pool.
func UpdatePool(ctx context.Context, pool *GlobalPool) error {
	result := DB.WithContext(ctx).Save(pool)
	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to update pool")
	}
	return nil
}

// ClosePool marks a pool as closed.
func ClosePool(ctx context.Context, id int) error {
	result := DB.WithContext(ctx).Model(&GlobalPool{}).Where("id = ?", id).
		Updates(map[string]any{"status": PoolStatusClosed, "closed_at": helper.GetTimestamp()})
	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to close pool")
	}
	return nil
}

// PurchaseToPool adds tokens to a pool (procurement).
func PurchaseToPool(ctx context.Context, poolId int, amount int64, remark string) error {
	tx := DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var pool GlobalPool
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pool, poolId).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to lock pool")
	}

	if pool.Status != PoolStatusActive {
		tx.Rollback()
		return errors.Errorf("pool %d is not active", poolId)
	}

	oldTotalQuota := pool.TotalQuota
	if err := tx.Model(&pool).Update("total_quota", gorm.Expr("total_quota + ?", amount)).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to update pool total_quota")
	}

	balanceAfter := (oldTotalQuota + amount) - pool.UsedQuota
	txn := PoolTransaction{
		PoolId:       poolId,
		Type:         PoolTxnPurchase,
		Amount:       amount,
		Direction:    PoolTxnIn,
		BalanceAfter: balanceAfter,
		Remark:       remark,
		CreatedAt:    helper.GetTimestamp(),
	}
	if err := tx.Create(&txn).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to create transaction record")
	}

	return tx.Commit().Error
}

// AllocateFromPool allocates tokens from a pool to a user.
func AllocateFromPool(ctx context.Context, poolId int, userId int, amount int64, remark string) error {
	tx := DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Lock pool row
	var pool GlobalPool
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pool, poolId).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to lock pool")
	}

	if pool.Status != PoolStatusActive {
		tx.Rollback()
		return errors.Errorf("pool %d is not active (status: %s)", poolId, pool.Status)
	}

	available := pool.AvailableQuota()
	if available < amount {
		tx.Rollback()
		return errors.Errorf("insufficient pool balance: available %s, requested %s",
			common.LogQuota(available), common.LogQuota(amount))
	}

	// 2. Update pool used_quota
	oldUsedQuota := pool.UsedQuota
	if err := tx.Model(&pool).Update("used_quota", pool.UsedQuota+amount).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to update pool used_quota")
	}

	// 3. Update user quota
	if err := tx.Model(&User{}).Where("id = ?", userId).
		Update("quota", gorm.Expr("quota + ?", amount)).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to increase user quota")
	}

	// 4. Create allocation record
	allocation := PoolAllocation{
		PoolId:         poolId,
		UserId:         userId,
		AllocatedQuota: amount,
		CreatedAt:      helper.GetTimestamp(),
		UpdatedAt:      helper.GetTimestamp(),
	}
	if err := tx.Create(&allocation).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to create allocation record")
	}

	// 5. Create transaction record
	balanceAfter := pool.TotalQuota - (oldUsedQuota + amount)
	txn := PoolTransaction{
		PoolId:       poolId,
		UserId:       userId,
		Type:         PoolTxnAllocate,
		Amount:       amount,
		Direction:    PoolTxnOut,
		BalanceAfter: balanceAfter,
		Remark:       remark,
		CreatedAt:    helper.GetTimestamp(),
	}
	if err := tx.Create(&txn).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to create transaction record")
	}

	// 6. Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 7. Record topup log (after commit, using global DB)
	RecordTopupLog(ctx, userId, fmt.Sprintf("从预算池[%s]分配", pool.Name), int(amount))
	return nil
}

// RecallFromPool recalls allocated tokens from a user back to the pool.
func RecallFromPool(ctx context.Context, poolId int, userId int, amount int64, remark string) error {
	tx := DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var pool GlobalPool
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pool, poolId).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to lock pool")
	}

	if pool.Status != PoolStatusActive {
		tx.Rollback()
		return errors.Errorf("pool %d is not active", poolId)
	}

	var allocation PoolAllocation
	if err := tx.Where("pool_id = ? AND user_id = ?", poolId, userId).First(&allocation).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "no allocation found for user in this pool")
	}

	netAllocated := allocation.NetAllocated()
	if netAllocated < amount {
		tx.Rollback()
		return errors.Errorf("cannot recall more than net allocated: net=%s, requested=%s",
			common.LogQuota(netAllocated), common.LogQuota(amount))
	}

	var user User
	if err := tx.First(&user, userId).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to get user")
	}
	if user.Quota < amount {
		tx.Rollback()
		return errors.Errorf("user quota insufficient for recall: user quota=%s, recall amount=%s",
			common.LogQuota(user.Quota), common.LogQuota(amount))
	}

	if err := tx.Model(&user).Update("quota", gorm.Expr("quota - ?", amount)).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to decrease user quota")
	}

	if err := tx.Model(&allocation).Updates(map[string]any{
		"recalled_quota": gorm.Expr("recalled_quota + ?", amount),
		"updated_at":     helper.GetTimestamp(),
	}).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to update allocation")
	}

	oldUsedQuota := pool.UsedQuota
	if err := tx.Model(&pool).Update("used_quota", gorm.Expr("used_quota - ?", amount)).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to release pool used_quota")
	}

	balanceAfter := pool.TotalQuota - (oldUsedQuota - amount)
	txn := PoolTransaction{
		PoolId:       poolId,
		UserId:       userId,
		Type:         PoolTxnRecall,
		Amount:       amount,
		Direction:    PoolTxnIn,
		BalanceAfter: balanceAfter,
		Remark:       remark,
		CreatedAt:    helper.GetTimestamp(),
	}
	if err := tx.Create(&txn).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to create transaction record")
	}

	return tx.Commit().Error
}

// GetPoolAllocations returns all allocations for a pool.
func GetPoolAllocations(ctx context.Context, poolId int) ([]*PoolAllocation, error) {
	var allocations []*PoolAllocation
	result := DB.WithContext(ctx).Where("pool_id = ?", poolId).Find(&allocations)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to get allocations")
	}
	return allocations, nil
}

// GetPoolTransactions returns paginated transactions for a pool.
func GetPoolTransactions(ctx context.Context, poolId int, page int, pageSize int) ([]*PoolTransaction, int64, error) {
	var transactions []*PoolTransaction
	var total int64
	query := DB.WithContext(ctx).Model(&PoolTransaction{}).Where("pool_id = ?", poolId)
	query.Count(&total)
	result := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&transactions)
	if result.Error != nil {
		return nil, 0, errors.Wrap(result.Error, "failed to get transactions")
	}
	return transactions, total, nil
}

// GetPoolReconciliation returns pool state and allocations for reconciliation.
func GetPoolReconciliation(ctx context.Context, poolId int) (*GlobalPool, []*PoolAllocation, error) {
	pool, err := GetPoolById(ctx, poolId)
	if err != nil {
		return nil, nil, err
	}
	allocations, err := GetPoolAllocations(ctx, poolId)
	if err != nil {
		return nil, nil, err
	}
	return pool, allocations, nil
}

// RecallAllFromPool recalls all net allocations from all users in a pool.
func RecallAllFromPool(ctx context.Context, poolId int) (int64, error) {
	allocations, err := GetPoolAllocations(ctx, poolId)
	if err != nil {
		return 0, err
	}
	var totalRecalled int64
	for _, alloc := range allocations {
		net := alloc.NetAllocated()
		if net > 0 {
			if err := RecallFromPool(ctx, poolId, alloc.UserId, net, "批量回收（关闭池）"); err != nil {
				logger.Logger.Error("failed to recall from user",
					zap.Int("pool_id", poolId), zap.Int("user_id", alloc.UserId), zap.Error(err))
				continue
			}
			totalRecalled += net
		}
	}
	return totalRecalled, nil
}

// RollOverPool closes the current pool and creates a new one with remaining (unallocated) balance.
// Note: this does NOT recall allocated quotas from users. Use RecallAllFromPool first if needed.
func RollOverPool(ctx context.Context, poolId int, newPeriodKey string, newName string) (*GlobalPool, error) {
	pool, err := GetPoolById(ctx, poolId)
	if err != nil {
		return nil, err
	}

	remaining := pool.AvailableQuota()

	if err := ClosePool(ctx, poolId); err != nil {
		return nil, err
	}

	newPool := &GlobalPool{
		Name:        newName,
		TotalQuota:  remaining,
		UsedQuota:   0,
		PeriodType:  pool.PeriodType,
		PeriodKey:   newPeriodKey,
		Status:      PoolStatusActive,
		CreatedAt:   helper.GetTimestamp(),
	}
	if err := CreatePool(ctx, newPool); err != nil {
		return nil, err
	}

	if remaining > 0 {
		txn := PoolTransaction{
			PoolId:       newPool.Id,
			Type:         PoolTxnPurchase,
			Amount:       remaining,
			Direction:    PoolTxnIn,
			BalanceAfter: remaining,
			Remark:       fmt.Sprintf("从池[%s]结转", pool.Name),
			CreatedAt:    helper.GetTimestamp(),
		}
		DB.Create(&txn)
	}

	return newPool, nil
}
