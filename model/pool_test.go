package model

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/songquanpeng/one-api/common"
)

// testUserCounter provides unique IDs for test users across tests.
var testUserCounter int

// newTestUser creates a User in the given DB with guaranteed-unique fields.
func newTestUser(t *testing.T, db *gorm.DB) *User {
	t.Helper()
	testUserCounter++
	now := time.Now().UnixNano()
	user := &User{
		Username:    fmt.Sprintf("testuser-%d", testUserCounter),
		Role:        1,
		Quota:       0,
		Status:      UserStatusEnabled,
		AccessToken: fmt.Sprintf("tok-%d-%d", testUserCounter, now),
		AffCode:     fmt.Sprintf("aff-%d-%d", testUserCounter, now),
	}
	result := db.Create(user)
	require.NoError(t, result.Error, "failed to create test user")
	return user
}

// ============================================================
// Pure Unit Tests (no DB)
// ============================================================

func TestGlobalPool_AvailableQuota(t *testing.T) {
	tests := []struct {
		name  string
		total int64
		used  int64
		want  int64
	}{
		{"zero values", 0, 0, 0},
		{"partial usage", 1000, 300, 700},
		{"fully used", 5000, 5000, 0},
		{"large numbers", 500000000000000, 100000000000, 499900000000000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &GlobalPool{TotalQuota: tt.total, UsedQuota: tt.used}
			assert.Equal(t, tt.want, p.AvailableQuota())
		})
	}
}

func TestPoolAllocation_NetAllocated(t *testing.T) {
	tests := []struct {
		name      string
		allocated int64
		recalled  int64
		want      int64
	}{
		{"no recall", 1000, 0, 1000},
		{"partial recall", 1000, 400, 600},
		{"full recall", 1000, 1000, 0},
		{"over-recall edge case", 500, 600, -100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &PoolAllocation{AllocatedQuota: tt.allocated, RecalledQuota: tt.recalled}
			assert.Equal(t, tt.want, a.NetAllocated())
		})
	}
}

// ============================================================
// Integration Tests (SQLite in-memory)
// ============================================================

func setupPoolTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&GlobalPool{}, &PoolAllocation{}, &PoolTransaction{}, &User{}, &Log{})
	require.NoError(t, err)
	return db
}

// setupPoolTestDBWithLogDB sets up test DB and replaces both DB and LOG_DB.
// Callers must defer restoration of both globals.
func setupPoolTestDBWithLogDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupPoolTestDB(t)
	return db
}

// --- CreatePool ---

func TestCreatePool_Integration(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	DB = testDB
	LOG_DB = testDB
	defer func() { DB = originalDB; LOG_DB = originalLogDB }()

	ctx := context.Background()
	pool := &GlobalPool{
		Name:       "2026年4月预算池",
		TotalQuota: 500000000000000,
		PeriodType: PoolPeriodMonthly,
		PeriodKey:  "2026-04",
		Status:     PoolStatusActive,
	}
	err := CreatePool(ctx, pool)
	require.NoError(t, err)
	assert.NotZero(t, pool.Id)
	assert.Equal(t, int64(500000000000000), pool.TotalQuota)
	assert.Equal(t, "2026年4月预算池", pool.Name)
	assert.NotZero(t, pool.CreatedAt)
}

// --- AllocateFromPool ---

func TestAllocateFromPool_Success(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	user := newTestUser(t, testDB)

	err := AllocateFromPool(ctx, pool.Id, user.Id, 3000, "测试分配")
	require.NoError(t, err)

	// Verify pool state
	updated, err := GetPoolById(ctx, pool.Id)
	require.NoError(t, err)
	assert.Equal(t, int64(3000), updated.UsedQuota)
	assert.Equal(t, int64(7000), updated.AvailableQuota())

	// Verify user quota
	var dbUser User
	testDB.First(&dbUser, user.Id)
	assert.Equal(t, int64(3000), dbUser.Quota)

	// Verify transaction record
	var txns []PoolTransaction
	testDB.Where("pool_id = ?", pool.Id).Find(&txns)
	assert.Len(t, txns, 1)
	assert.Equal(t, PoolTxnAllocate, txns[0].Type)
	assert.Equal(t, PoolTxnOut, txns[0].Direction)
	assert.Equal(t, int64(3000), txns[0].Amount)
	assert.Equal(t, int64(7000), txns[0].BalanceAfter)

	// Verify allocation record
	var allocs []PoolAllocation
	testDB.Where("pool_id = ? AND user_id = ?", pool.Id, user.Id).Find(&allocs)
	assert.Len(t, allocs, 1)
	assert.Equal(t, int64(3000), allocs[0].AllocatedQuota)
}

func TestAllocateFromPool_InsufficientBalance(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	user := newTestUser(t, testDB)

	err := AllocateFromPool(ctx, pool.Id, user.Id, 20000, "超额分配")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient")
}

func TestAllocateFromPool_InactivePool(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusClosed}
	require.NoError(t, CreatePool(ctx, pool))

	user := newTestUser(t, testDB)

	err := AllocateFromPool(ctx, pool.Id, user.Id, 1000, "应该失败")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

func TestAllocateFromPool_NonexistentPool(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	err := AllocateFromPool(ctx, 9999, 1, 1000, "不存在的池")
	require.Error(t, err)
}

// --- RecallFromPool ---

func TestRecallFromPool_Success(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	user := newTestUser(t, testDB)

	// Allocate first
	require.NoError(t, AllocateFromPool(ctx, pool.Id, user.Id, 3000, "初始分配"))

	// Recall 1000
	err := RecallFromPool(ctx, pool.Id, user.Id, 1000, "部分回收")
	require.NoError(t, err)

	// Verify pool state
	updated, err := GetPoolById(ctx, pool.Id)
	require.NoError(t, err)
	assert.Equal(t, int64(2000), updated.UsedQuota)
	assert.Equal(t, int64(8000), updated.AvailableQuota())

	// Verify user quota
	var dbUser User
	testDB.First(&dbUser, user.Id)
	assert.Equal(t, int64(2000), dbUser.Quota)

	// Verify allocation updated
	var alloc PoolAllocation
	testDB.Where("pool_id = ? AND user_id = ?", pool.Id, user.Id).First(&alloc)
	assert.Equal(t, int64(3000), alloc.AllocatedQuota)
	assert.Equal(t, int64(1000), alloc.RecalledQuota)
	assert.Equal(t, int64(2000), alloc.NetAllocated())

	// Verify transaction record
	var txns []PoolTransaction
	testDB.Where("pool_id = ?", pool.Id).Order("id desc").Find(&txns)
	assert.Len(t, txns, 2) // allocate + recall
	assert.Equal(t, PoolTxnRecall, txns[0].Type)
	assert.Equal(t, PoolTxnIn, txns[0].Direction)
	assert.Equal(t, int64(1000), txns[0].Amount)
}

func TestRecallFromPool_ExceedsNetAllocated(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	user := newTestUser(t, testDB)

	require.NoError(t, AllocateFromPool(ctx, pool.Id, user.Id, 3000, "初始分配"))

	err := RecallFromPool(ctx, pool.Id, user.Id, 5000, "超额回收")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot recall more than net allocated")
}

func TestRecallFromPool_UserQuotaInsufficient(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	user := newTestUser(t, testDB)

	require.NoError(t, AllocateFromPool(ctx, pool.Id, user.Id, 3000, "初始分配"))

	// Simulate user consuming some quota (reduce directly)
	testDB.Model(&User{}).Where("id = ?", user.Id).Update("quota", 500)

	// Try to recall 1000 when user only has 500
	err := RecallFromPool(ctx, pool.Id, user.Id, 1000, "超额回收")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user quota insufficient")
}

func TestRecallFromPool_NoAllocation(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	user := newTestUser(t, testDB)

	err := RecallFromPool(ctx, pool.Id, user.Id, 1000, "没有分配记录")
	require.Error(t, err)
}

// --- PurchaseToPool ---

func TestPurchaseToPool_Success(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	err := PurchaseToPool(ctx, pool.Id, 5000, "追加采购")
	require.NoError(t, err)

	updated, err := GetPoolById(ctx, pool.Id)
	require.NoError(t, err)
	assert.Equal(t, int64(15000), updated.TotalQuota)
	assert.Equal(t, int64(0), updated.UsedQuota)
	assert.Equal(t, int64(15000), updated.AvailableQuota())

	// Verify transaction
	var txns []PoolTransaction
	testDB.Where("pool_id = ?", pool.Id).Find(&txns)
	assert.Len(t, txns, 1)
	assert.Equal(t, PoolTxnPurchase, txns[0].Type)
	assert.Equal(t, PoolTxnIn, txns[0].Direction)
	assert.Equal(t, int64(5000), txns[0].Amount)
	assert.Equal(t, int64(15000), txns[0].BalanceAfter)
}

func TestPurchaseToPool_InactivePool(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusClosed}
	require.NoError(t, CreatePool(ctx, pool))

	err := PurchaseToPool(ctx, pool.Id, 5000, "应该失败")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

// --- GetAllPools ---

func TestGetAllPools_WithFilters(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	DB = testDB
	LOG_DB = testDB
	defer func() { DB = originalDB; LOG_DB = originalLogDB }()

	ctx := context.Background()
	// Create pools
	testDB.Create(&GlobalPool{Name: "monthly1", TotalQuota: 1000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive, CreatedAt: 1})
	testDB.Create(&GlobalPool{Name: "quarterly1", TotalQuota: 2000, PeriodType: PoolPeriodQuarterly, PeriodKey: "2026-Q1", Status: PoolStatusActive, CreatedAt: 2})
	testDB.Create(&GlobalPool{Name: "monthly2", TotalQuota: 3000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-03", Status: PoolStatusClosed, CreatedAt: 3})

	// Filter by period_type
	pools, total, err := GetAllPools(ctx, PoolPeriodMonthly, "", 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)

	// Filter by status
	pools, total, err = GetAllPools(ctx, "", PoolStatusClosed, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, PoolStatusClosed, pools[0].Status)

	// Filter by both
	pools, total, err = GetAllPools(ctx, PoolPeriodMonthly, PoolStatusActive, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)

	// No filters
	pools, total, err = GetAllPools(ctx, "", "", 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)

	// Pagination
	pools, total, err = GetAllPools(ctx, "", "", 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, pools, 2)

	_ = pools // suppress unused warning in test context
}

// --- ClosePool ---

func TestClosePool_Integration(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	DB = testDB
	LOG_DB = testDB
	defer func() { DB = originalDB; LOG_DB = originalLogDB }()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive, CreatedAt: 1}
	testDB.Create(pool)

	err := ClosePool(ctx, pool.Id)
	require.NoError(t, err)

	updated, err := GetPoolById(ctx, pool.Id)
	require.NoError(t, err)
	assert.Equal(t, PoolStatusClosed, updated.Status)
	assert.NotZero(t, updated.ClosedAt)
}

// --- GetPoolById ---

func TestGetPoolById_NotFound(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	DB = testDB
	LOG_DB = testDB
	defer func() { DB = originalDB; LOG_DB = originalLogDB }()

	ctx := context.Background()
	_, err := GetPoolById(ctx, 9999)
	require.Error(t, err)
}

// --- GetPoolAllocations ---

func TestGetPoolAllocations_Integration(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	user1 := newTestUser(t, testDB)
	user2 := newTestUser(t, testDB)

	require.NoError(t, AllocateFromPool(ctx, pool.Id, user1.Id, 1000, ""))
	require.NoError(t, AllocateFromPool(ctx, pool.Id, user2.Id, 2000, ""))

	allocs, err := GetPoolAllocations(ctx, pool.Id)
	require.NoError(t, err)
	assert.Len(t, allocs, 2)
}

// --- GetPoolTransactions ---

func TestGetPoolTransactions_WithPagination(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 100000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	user := newTestUser(t, testDB)

	// Create 5 transactions via allocations
	for i := 0; i < 5; i++ {
		require.NoError(t, AllocateFromPool(ctx, pool.Id, user.Id, 100, ""))
	}

	txns, total, err := GetPoolTransactions(ctx, pool.Id, 1, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, txns, 3)

	// Page 2
	txns, total, err = GetPoolTransactions(ctx, pool.Id, 2, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, txns, 2)
}

// --- GetPoolReconciliation ---

func TestGetPoolReconciliation_Integration(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	user1 := newTestUser(t, testDB)
	user2 := newTestUser(t, testDB)

	require.NoError(t, AllocateFromPool(ctx, pool.Id, user1.Id, 3000, ""))
	require.NoError(t, AllocateFromPool(ctx, pool.Id, user2.Id, 2000, ""))

	reconPool, allocs, err := GetPoolReconciliation(ctx, pool.Id)
	require.NoError(t, err)
	assert.Equal(t, int64(5000), reconPool.UsedQuota)
	assert.Equal(t, int64(5000), reconPool.AvailableQuota())
	assert.Len(t, allocs, 2)
}

// --- RecallAllFromPool ---

func TestRecallAllFromPool_Integration(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	user1 := newTestUser(t, testDB)
	user2 := newTestUser(t, testDB)
	user3 := newTestUser(t, testDB)

	require.NoError(t, AllocateFromPool(ctx, pool.Id, user1.Id, 2000, ""))
	require.NoError(t, AllocateFromPool(ctx, pool.Id, user2.Id, 3000, ""))
	require.NoError(t, AllocateFromPool(ctx, pool.Id, user3.Id, 1000, ""))

	recalled, err := RecallAllFromPool(ctx, pool.Id)
	require.NoError(t, err)
	assert.Equal(t, int64(6000), recalled)

	// Verify pool state
	updated, err := GetPoolById(ctx, pool.Id)
	require.NoError(t, err)
	assert.Equal(t, int64(0), updated.UsedQuota)
	assert.Equal(t, int64(10000), updated.AvailableQuota())
}

// --- RollOverPool ---

func TestRollOverPool_Integration(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	originalLogDB := LOG_DB
	originalSQLite := common.UsingSQLite.Load()
	DB = testDB
	LOG_DB = testDB
	common.UsingSQLite.Store(true)
	defer func() {
		DB = originalDB; LOG_DB = originalLogDB
		common.UsingSQLite.Store(originalSQLite)
	}()

	ctx := context.Background()
	pool := &GlobalPool{Name: "2026-04", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	user := newTestUser(t, testDB)

	require.NoError(t, AllocateFromPool(ctx, pool.Id, user.Id, 3000, ""))

	newPool, err := RollOverPool(ctx, pool.Id, "2026-05", "2026-05")
	require.NoError(t, err)
	require.NotNil(t, newPool)

	// Verify new pool
	assert.Equal(t, "2026-05", newPool.Name)
	assert.Equal(t, int64(7000), newPool.TotalQuota) // 10000 - 3000 (still allocated)
	assert.Equal(t, int64(0), newPool.UsedQuota)
	assert.Equal(t, "2026-05", newPool.PeriodKey)
	assert.Equal(t, PoolStatusActive, newPool.Status)
	assert.NotEqual(t, pool.Id, newPool.Id)

	// Verify old pool closed
	oldPool, err := GetPoolById(ctx, pool.Id)
	require.NoError(t, err)
	assert.Equal(t, PoolStatusClosed, oldPool.Status)

	// User quota unchanged (no recall in rollover)
	var dbUser User
	testDB.First(&dbUser, user.Id)
	assert.Equal(t, int64(3000), dbUser.Quota)
}
