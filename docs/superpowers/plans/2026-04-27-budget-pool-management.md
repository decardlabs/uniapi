# Budget Pool Management Implementation Plan (TDD)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a complete budget pool management system — root purchases tokens into a global pool, allocates budgets to sub-users, users consume via API calls, and root performs periodic reconciliation (monthly/quarterly/yearly).

**Architecture:** 3 new GORM models (`GlobalPool`, `PoolAllocation`, `PoolTransaction`) in `model/pool.go`, new controller `controller/pool.go`, new route group `/api/pool/*`, new frontend page `BudgetPoolsPage`.

**Tech Stack:** Go (Gin + GORM) backend, React 18 + TypeScript + Tailwind + shadcn/ui frontend

**Design spec:** Agreed in conversation — "Prepaid Pool + Real-time Credit Line" model

**TDD Cycle:** Each step follows **RED → GREEN → REFACTOR**:
- 🔴 **RED**: Write failing test first
- 🟢 **GREEN**: Write minimal code to pass
- 🔵 **REFACTOR**: Clean up while keeping tests green

---

## File Structure

- **Create:** `model/pool.go` — Pool models and business logic
- **Create:** `model/pool_test.go` — Pool model tests (TDD)
- **Create:** `controller/pool.go` — Pool management API handlers
- **Create:** `controller/pool_test.go` — Pool controller tests (TDD)
- **Modify:** `model/main.go` — Add AutoMigrate for pool models
- **Modify:** `router/api.go` — Add `/api/pool/` route group
- **Create:** `web/modern/src/pages/pools/BudgetPoolsPage.tsx` — Pool management UI
- **Create:** `web/modern/src/pages/pools/__tests__/BudgetPoolsPage.test.tsx` — Page tests (TDD)
- **Modify:** `web/modern/src/App.tsx` — Add `/pools` route
- **Modify:** `web/modern/src/components/layout/navigation.ts` — Add pools nav entry
- **Modify:** `web/modern/src/i18n/locales/*/common.json` — Add i18n keys

---

## Key Conventions

### Quota Units
The project uses an internal quota unit. Display conversion: `quota / 500000 = USD`. The frontend `renderQuota()` utility handles this. All backend logic works in raw quota units.

### Error Handling Pattern
All handlers return `gin.H{"success": bool, "message": string, "data": any}`. The frontend `api` lib checks `success` field.

### Auth Middleware
- `middleware.AdminAuth()` — Admin-only endpoints (root + admin)
- `middleware.UserAuth()` — Any authenticated user

### Timestamp Format
Project uses `int64` Unix timestamps via `helper.GetTimestamp()`.

### Frontend Component Patterns
- Pages use `useAuthStore()` for auth state
- API calls via `import { api } from '@/lib/api'`
- Tables use `@tanstack/react-table` + shadcn `<Table>` components
- Forms use shadcn `<Dialog>` for create/edit modals
- Notifications via `useNotification()` from `@/components/ui/notifications`

### Go Testing Patterns (from codebase survey)
- **Framework:** `testify/assert` + `testify/require`
- **DB setup:** Replace global `DB` with SQLite in-memory (`gorm.Open(sqlite.Open(":memory:"))`)
- **SQLite flag:** `common.UsingSQLite.Store(true)` before tests, restore after
- **Parallel:** `t.Parallel()` where safe (not with global DB replacement)
- **Table-driven:** Use `t.Run()` subtests for case variations
- **Dependencies confirmed:** `github.com/stretchr/testify v1.11.1`, `gorm.io/driver/sqlite v1.6.0` already in `go.mod`

### Frontend Testing Patterns (from codebase survey)
- **Framework:** Vitest + `@testing-library/react` + jsdom
- **Setup:** `web/modern/src/test/setup.ts` — i18n mock, Radix polyfills, `matchMedia`, `ResizeObserver`
- **Mock pattern:** Radix Select → native `<select>`, API → `vi.fn()`
- **Reference:** `ChannelsPage.test.tsx` — best page-level test example
- **Run:** `cd web/modern && npx vitest run`

---

## Task 1: Backend — Pool Models & Database Migration

**Files:**
- Create: `model/pool.go`
- Create: `model/pool_test.go`
- Modify: `model/main.go` (add AutoMigrate calls)

### 🔴 RED Phase — Write Tests First

- [ ] **Step 1.1: Create `model/pool_test.go` — Pure unit tests (no DB)**

```go
package model

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGlobalPool_AvailableQuota(t *testing.T) {
	tests := []struct {
		name       string
		total      int64
		used       int64
		want       int64
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
		{"over-recall edge case", 500, 600, -100}, // valid: negative net
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &PoolAllocation{AllocatedQuota: tt.allocated, RecalledQuota: tt.recalled}
			assert.Equal(t, tt.want, a.NetAllocated())
		})
	}
}
```

**Expected:** `go test ./model/ -run TestGlobalPool_AvailableQuota` fails (compilation error — `GlobalPool` not defined yet).

- [ ] **Step 1.2: Add `model/pool.go` — Models only (minimum to pass unit tests)**

Create `model/pool.go` with just the 3 struct definitions and 2 method functions (`AvailableQuota`, `NetAllocated`). No CRUD functions yet.

```go
package model

type GlobalPool struct {
	Id          int    `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"type:varchar(100);not null"`
	TotalQuota  int64  `json:"total_quota" gorm:"not null;default:0"`
	UsedQuota   int64  `json:"used_quota" gorm:"not null;default:0"`
	PeriodType  string `json:"period_type" gorm:"type:varchar(20);not null"`
	PeriodKey   string `json:"period_key" gorm:"type:varchar(20);not null;index"`
	Status      string `json:"status" gorm:"type:varchar(20);not null;default:'active';index"`
	CreatedAt   int64  `json:"created_at" gorm:"not null"`
	ClosedAt    int64  `json:"closed_at"`
}

func (p *GlobalPool) AvailableQuota() int64 {
	return p.TotalQuota - p.UsedQuota
}

type PoolAllocation struct {
	Id              int   `json:"id" gorm:"primaryKey"`
	PoolId          int   `json:"pool_id" gorm:"index;not null"`
	UserId          int   `json:"user_id" gorm:"index;not null"`
	AllocatedQuota  int64 `json:"allocated_quota" gorm:"not null;default:0"`
	RecalledQuota   int64 `json:"recalled_quota" gorm:"not null;default:0"`
	CreatedAt       int64 `json:"created_at" gorm:"not null"`
	UpdatedAt       int64 `json:"updated_at" gorm:"not null"`
}

func (a *PoolAllocation) NetAllocated() int64 {
	return a.AllocatedQuota - a.RecalledQuota
}

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
```

**Expected:** Unit tests pass. Run: `cd uniapi && go test ./model/ -run "TestGlobalPool_AvailableQuota|TestPoolAllocation_NetAllocated" -v`

### 🟢 GREEN Phase — Integration Tests for CRUD

- [ ] **Step 1.3: Add CRUD integration tests to `model/pool_test.go`**

Add these tests using SQLite in-memory DB (matching `ability_test.go` pattern):

```go
func setupPoolTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&GlobalPool{}, &PoolAllocation{}, &PoolTransaction{}, &User{})
	require.NoError(t, err)
	return db
}

func TestCreatePool_Integration(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB
	DB = testDB
	defer func() { DB = originalDB }()

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
}

func TestAllocateFromPool_Success(t *testing.T) {
	testDB := setupPoolTestDB(t)
	originalDB := DB; DB = testDB
	originalSQLite := common.UsingSQLite.Value()
	common.UsingSQLite.Store(true)
	defer func() { DB = originalDB; common.UsingSQLite.Store(originalSQLite) }()

	ctx := context.Background()
	// Create pool
	pool := &GlobalPool{Name: "test", TotalQuota: 10000, PeriodType: PoolPeriodMonthly, PeriodKey: "2026-04", Status: PoolStatusActive}
	require.NoError(t, CreatePool(ctx, pool))

	// Create user
	user := &User{Username: "testuser", Role: 1, Quota: 0}
	testDB.Create(user)

	// Allocate
	err := AllocateFromPool(ctx, pool.Id, user.Id, 3000, "测试分配")
	require.NoError(t, err)

	// Verify pool state
	updated, _ := GetPoolById(ctx, pool.Id)
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
}

func TestAllocateFromPool_InsufficientBalance(t *testing.T) {
	// Same setup as above, but request 20000 from pool with 10000
	// Expect error containing "insufficient"
}

func TestAllocateFromPool_InactivePool(t *testing.T) {
	// Create pool with status "closed", expect error containing "not active"
}

func TestRecallFromPool_Success(t *testing.T) {
	// Setup: create pool + user + allocate 3000
	// Recall 1000
	// Verify: pool.UsedQuota=2000, user.Quota=2000, allocation.recalled=1000
}

func TestRecallFromPool_ExceedsNetAllocated(t *testing.T) {
	// Allocate 3000, try to recall 5000 → error "cannot recall more than net allocated"
}

func TestRecallFromPool_UserQuotaInsufficient(t *testing.T) {
	// Allocate 3000, user consumes 2500 (quota=500), try to recall 1000 → error
}

func TestPurchaseToPool_Success(t *testing.T) {
	// Create pool with total=10000, purchase 5000
	// Verify: pool.TotalQuota=15000, transaction record exists
}

func TestPurchaseToPool_InactivePool(t *testing.T) {
	// Closed pool → error "not active"
}

func TestGetAllPools_WithFilters(t *testing.T) {
	// Create 3 pools (monthly, quarterly, yearly), some active, some closed
	// Query with period_type="monthly" → expect 1
	// Query with status="closed" → expect correct count
	// Query with pagination → verify page size
}

func TestClosePool_Integration(t *testing.T) {
	// Create active pool → ClosePool → verify status="closed", closed_at set
}

func TestGetPoolReconciliation_Integration(t *testing.T) {
	// Create pool, allocate to 2 users
	// GetPoolReconciliation → verify pool + allocations returned
}

func TestGetPoolTransactions_WithPagination(t *testing.T) {
	// Create pool, do 5 allocations
	// GetPoolTransactions(page=1, pageSize=3) → expect 3 items, total=5
}

func TestRollOverPool_Integration(t *testing.T) {
	// Create pool (monthly, 2026-04), purchase 10000, allocate 3000 to user
	// RollOver to 2026-05
	// Verify: old pool closed, new pool created with total=7000, user recalled
}

func TestRecallAllFromPool_Integration(t *testing.T) {
	// Create pool, allocate to 3 users, recall all
	// Verify: all users quota decreased, pool.UsedQuota=0
}
```

**Expected:** `go test ./model/ -run TestCreatePool_Integration` fails (CreatePool not implemented).

- [ ] **Step 1.4: Add CRUD functions to `model/pool.go`**

Implement: `CreatePool`, `GetPoolById`, `GetAllPools`, `UpdatePool`, `ClosePool`, `AllocateFromPool`, `RecallFromPool`, `PurchaseToPool`, `GetPoolAllocations`, `GetPoolTransactions`, `GetPoolReconciliation`, `RecallAllFromPool`, `RollOverPool`.

Full implementations in original plan (lines 149-500). Key additions needed for imports:
```go
import (
	"context"
	"fmt"
	"github.com/Laisky/errors/v2"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/helper"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)
```

**Expected:** All integration tests pass. Run: `cd uniapi && go test ./model/ -run "Test.*Pool" -v`

### 🔵 REFACTOR Phase

- [ ] **Step 1.5: Refactor pool.go** — Extract common patterns (lock-and-check) into helper if duplicated >2 times. Ensure all error messages are consistent.

- [ ] **Step 1.6: Add AutoMigrate to `model/main.go`**

In `migrateDB()`, add after the `RechargeRequest` migration block:

```go
if err = DB.AutoMigrate(&GlobalPool{}); err != nil {
    return errors.Wrapf(err, "failed to migrate GlobalPool")
}
if err = DB.AutoMigrate(&PoolAllocation{}); err != nil {
    return errors.Wrapf(err, "failed to migrate PoolAllocation")
}
if err = DB.AutoMigrate(&PoolTransaction{}); err != nil {
    return errors.Wrapf(err, "failed to migrate PoolTransaction")
}
```

- [ ] **Step 1.7: Verify compilation** — `cd uniapi && go build ./...` and `cd uniapi && go test ./model/ -run "Test.*Pool" -v`

**Commit:** `feat(pool): add pool models, CRUD operations, and integration tests`

---

## Task 2: Backend — Pool Controller & Routes

**Files:**
- Create: `controller/pool.go`
- Create: `controller/pool_test.go`
- Modify: `router/api.go`

### 🔴 RED Phase — Write Controller Tests First

- [ ] **Step 2.1: Create `controller/pool_test.go`**

Use `httptest` + `gin.TestMode`. Pattern reference: existing controller tests in project.

```go
package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupPoolRouter() *gin.Engine {
	r := gin.New()
	// Mount pool routes (without auth middleware for testing)
	api := r.Group("/api")
	{
		pool := api.Group("/pool")
		{
			pool.GET("/", GetAllPools)
			pool.POST("/", CreatePool)
			pool.GET("/:id", GetPool)
			pool.POST("/:id/purchase", PurchaseToPool)
			pool.POST("/:id/allocate", AllocateFromPool)
			pool.POST("/:id/recall", RecallFromPool)
			pool.GET("/:id/allocations", GetPoolAllocations)
			pool.GET("/:id/transactions", GetPoolTransactions)
			pool.GET("/:id/reconciliation", GetPoolReconciliation)
			pool.POST("/:id/close", ClosePool)
			pool.POST("/:id/rollover", RollOverPool)
		}
	}
	return r
}

func TestCreatePool_Handler(t *testing.T) {
	// NOTE: Requires model.DB to be set up with SQLite
	// For unit handler tests, we test request parsing and response format
	// Integration with model layer covered in model/pool_test.go

	router := setupPoolRouter()
	body := `{"name":"test pool","total_quota":10000,"period_type":"monthly","period_key":"2026-04"}`
	req := httptest.NewRequest("POST", "/api/pool/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Response format check (may fail if model not set up)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
}

func TestGetAllPools_Handler_QueryParams(t *testing.T) {
	router := setupPoolRouter()
	req := httptest.NewRequest("GET", "/api/pool/?period_type=monthly&status=active&page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
}

func TestAllocateFromPool_Handler_Validation(t *testing.T) {
	router := setupPoolRouter()

	// Missing required fields
	body := `{}`
	req := httptest.NewRequest("POST", "/api/pool/1/allocate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return error (missing user_id, amount)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPurchaseToPool_Handler_Validation(t *testing.T) {
	router := setupPoolRouter()
	body := `{"amount": 0}`  // Zero amount should be rejected
	req := httptest.NewRequest("POST", "/api/pool/1/purchase", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

**Expected:** `go test ./controller/ -run TestCreatePool_Handler` fails (CreatePool handler not defined).

### 🟢 GREEN Phase — Implement Controller

- [ ] **Step 2.2: Create `controller/pool.go`**

All handler functions follow the standard pattern: `gin.H{"success": bool, "message": string, "data": any}`.
Context helper: `gmw.Ctx(c)` (matching `controller/user.go` pattern).

```go
package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/controller/gmw"
	"github.com/songquanpeng/one-api/model"
)

type createPoolRequest struct {
	Name       string `json:"name" binding:"required"`
	TotalQuota int64  `json:"total_quota" binding:"required,min=1"`
	PeriodType string `json:"period_type" binding:"required,oneof=monthly quarterly yearly oneoff"`
	PeriodKey  string `json:"period_key" binding:"required"`
}

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
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 10 }

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

// PurchaseToPoolRequest, AllocateRequest, RecallRequest, RollOverRequest structs
// PurchaseToPool, AllocateFromPool, RecallFromPool, GetPoolAllocations,
// GetPoolTransactions, GetPoolReconciliation, ClosePool, RollOverPool handlers
// — all follow the same pattern as above
```

Full implementations for remaining handlers follow identical patterns.

**Expected:** `cd uniapi && go test ./controller/ -run "Test.*Pool" -v` passes.

### 🔵 REFACTOR Phase

- [ ] **Step 2.3: Extract common response helpers** if duplicated across handlers.

- [ ] **Step 2.4: Add routes to `router/api.go`**

Add after the `rechargeRoute` block:

```go
		// Budget pool management routes
		poolRoute := apiRouter.Group("/pool")
		poolRoute.Use(middleware.AdminAuth())
		{
			poolRoute.GET("/", controller.GetAllPools)
			poolRoute.POST("/", controller.CreatePool)
			poolRoute.GET("/:id", controller.GetPool)
			poolRoute.POST("/:id/purchase", controller.PurchaseToPool)
			poolRoute.POST("/:id/allocate", controller.AllocateFromPool)
			poolRoute.POST("/:id/recall", controller.RecallFromPool)
			poolRoute.GET("/:id/allocations", controller.GetPoolAllocations)
			poolRoute.GET("/:id/transactions", controller.GetPoolTransactions)
			poolRoute.GET("/:id/reconciliation", controller.GetPoolReconciliation)
			poolRoute.POST("/:id/close", controller.ClosePool)
			poolRoute.POST("/:id/rollover", controller.RollOverPool)
		}
```

- [ ] **Step 2.5: Verify compilation** — `cd uniapi && go build ./...`

**Commit:** `feat(pool): add pool controller, routes, and handler tests`

---

## Task 3: Frontend — Budget Pool Management Page

**Files:**
- Create: `web/modern/src/pages/pools/BudgetPoolsPage.tsx`
- Create: `web/modern/src/pages/pools/__tests__/BudgetPoolsPage.test.tsx`
- Modify: `web/modern/src/App.tsx`
- Modify: `web/modern/src/components/layout/navigation.ts`
- Modify: `web/modern/src/i18n/locales/*/common.json`

### 🔴 RED Phase — Write Frontend Tests First

- [ ] **Step 3.1: Add i18n keys**

| File | Key | Value |
|------|-----|-------|
| `zh/common.json` | `"pools"` | `"预算池"` |
| `en/common.json` | `"pools"` | `"Budget Pools"` |
| `ja/common.json` | `"pools"` | `"予算プール"` |
| `es/common.json` | `"pools"` | `"Pools de presupuesto"` |
| `fr/common.json` | `"pools"` | `"Pools budgétaires"` |

Additional keys for pool page:
| Key (zh) | Value |
|----------|-------|
| `pool.create` | `"新建预算池"` |
| `pool.name` | `"名称"` |
| `pool.period_type` | `"周期类型"` |
| `pool.period_key` | `"周期Key"` |
| `pool.total_quota` | `"采购总额"` |
| `pool.available_quota` | `"可用余额"` |
| `pool.allocated` | `"已分配"` |
| `pool.status` | `"状态"` |
| `pool.purchase` | `"采购"` |
| `pool.allocate` | `"分配"` |
| `pool.recall` | `"回收"` |
| `pool.reconciliation` | `"盘点"` |
| `pool.close` | `"关闭"` |
| `pool.rollover` | `"结转"` |

- [ ] **Step 3.2: Create `BudgetPoolsPage.test.tsx`**

```tsx
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

// Mock Radix Select (same pattern as ChannelsPage.test.tsx)
vi.mock('@/components/ui/select', () => ({
  Select: ({ value, onValueChange, children }: any) => (
    <select value={value || ''} onChange={(e) => onValueChange?.(e.target.value)} data-testid="select">
      {children}
    </select>
  ),
  SelectTrigger: ({ children }: any) => <div data-testid="select-trigger">{children}</div>,
  SelectValue: ({ placeholder }: any) => <span>{placeholder}</span>,
  SelectContent: ({ children }: any) => <>{children}</>,
  SelectItem: ({ children, value }: any) => <option value={value}>{children}</option>,
  SelectGroup: ({ children }: any) => <>{children}</>,
  SelectLabel: ({ children }: any) => <>{children}</>,
}));

// Mock Dialog
vi.mock('@/components/ui/dialog', () => ({
  Dialog: ({ children, open }: any) => open ? <div data-testid="dialog">{children}</div> : null,
  DialogTrigger: ({ children }: any) => <>{children}</>,
  DialogContent: ({ children }: any) => <div data-testid="dialog-content">{children}</div>,
  DialogHeader: ({ children }: any) => <>{children}</>,
  DialogTitle: ({ children }: any) => <h2>{children}</h2>,
  DialogDescription: ({ children }: any) => <p>{children}</p>,
  DialogFooter: ({ children }: any) => <div>{children}</div>,
}));

// Mock API
const mockApiGet = vi.fn();
const mockApiPost = vi.fn();
vi.mock('@/lib/api', () => ({
  api: {
    get: (...args: any[]) => mockApiGet(...args),
    post: (...args: any[]) => mockApiPost(...args),
  },
}));

// Mock useAuthStore
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ user: { role: 100, username: 'root' }, isAdmin: true }),
}));

import BudgetPoolsPage from '../BudgetPoolsPage';

describe('BudgetPoolsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockApiGet.mockResolvedValue({
      data: { success: true, data: { items: [], total: 0, page: 1, page_size: 10 } },
    });
  });

  it('renders page title and create button', async () => {
    render(<BudgetPoolsPage />);
    expect(screen.getByRole('heading', { name: /预算池/i })).toBeInTheDocument();
  });

  it('shows empty state when no pools', async () => {
    render(<BudgetPoolsPage />);
    await waitFor(() => {
      expect(screen.getByText(/暂无数据/i)).toBeInTheDocument();
    });
  });

  it('renders pool table with data', async () => {
    mockApiGet.mockResolvedValue({
      data: {
        success: true,
        data: {
          items: [
            { id: 1, name: '2026年4月预算池', total_quota: 100000, used_quota: 30000, period_type: 'monthly', period_key: '2026-04', status: 'active', created_at: 1746000000 },
          ],
          total: 1, page: 1, page_size: 10,
        },
      },
    });
    render(<BudgetPoolsPage />);
    await waitFor(() => {
      expect(screen.getByText('2026年4月预算池')).toBeInTheDocument();
    });
  });

  it('opens create dialog on button click', async () => {
    const user = userEvent.setup();
    render(<BudgetPoolsPage />);
    await user.click(screen.getByRole('button', { name: /新建预算池/i }));
    await waitFor(() => {
      expect(screen.getByTestId('dialog-content')).toBeInTheDocument();
    });
  });
});
```

**Expected:** `cd web/modern && npx vitest run src/pages/pools/` fails (BudgetPoolsPage not found).

### 🟢 GREEN Phase — Implement Page

- [ ] **Step 3.3: Add navigation entry to `navigation.ts`**

1. Import: `import { PiggyBank } from 'lucide-react'`
2. iconMap: `'/pools': PiggyBank`
3. buildAuthenticatedNavItems: `{ name: t('common.pools'), to: '/pools', show: isAdmin, requiresAdmin: true }` (after recharges)
4. groupNavItems adminPaths: add `'/pools'`

- [ ] **Step 3.4: Add route to `App.tsx`**

Lazy import:
```tsx
const BudgetPoolsPage = lazy(() => import('@/pages/pools/BudgetPoolsPage'));
```
Route (protected Layout group, after recharges):
```tsx
<Route path="pools" element={<BudgetPoolsPage />} />
```

- [ ] **Step 3.5: Create `BudgetPoolsPage.tsx`** (minimum to pass tests)

Start with: page title, create button, empty state, API fetch, table rendering.

```tsx
import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Plus } from 'lucide-react';
import { api } from '@/lib/api';
import { useNotification } from '@/components/ui/notifications';
import { Button } from '@/components/ui/button';
import { renderQuota } from '@/utils/render';

export default function BudgetPoolsPage() {
  const { t } = useTranslation();
  const { addNotification } = useNotification();
  const [pools, setPools] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  const fetchPools = useCallback(async () => {
    try {
      const res = await api.get('/api/pool/', { params: { page: 1, page_size: 10 } });
      if (res.data.success) {
        setPools(res.data.data.items);
        setTotal(res.data.data.total);
      }
    } catch (e: any) {
      addNotification({ type: 'error', message: e.message });
    } finally {
      setLoading(false);
    }
  }, [addNotification]);

  useEffect(() => { fetchPools(); }, [fetchPools]);

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">{t('common.pools')}</h1>
        <Button><Plus className="mr-2 h-4 w-4" />{t('pool.create')}</Button>
      </div>
      {pools.length === 0 && !loading && (
        <div className="text-center py-12 text-muted-foreground">暂无数据</div>
      )}
    </div>
  );
}
```

**Expected:** `cd web/modern && npx vitest run src/pages/pools/` passes (basic tests).

### 🔵 REFACTOR Phase — Full Feature Implementation

- [ ] **Step 3.6: Build complete BudgetPoolsPage**

The page has these views:

**Pool List View (default):**
- Filter bar: period type dropdown, status dropdown
- Table: ID, 名称, 周期类型, 周期Key, 采购总额, 已分配, 可用余额, 状态, 创建时间
- Actions per row: 采购, 分配, 盘点, 关闭, 结转
- "新建预算池" button

**Create Pool Dialog:**
- Fields: 名称, 采购总额, 周期类型 (select), 周期Key
- Submit: POST `/api/pool/`

**Purchase Dialog:**
- Fields: 追加金额, 备注
- Submit: POST `/api/pool/:id/purchase`

**Allocate Dialog:**
- Fields: 用户 (select + search), 分配金额, 备注
- Submit: POST `/api/pool/:id/allocate`

**Recall Dialog:**
- Fields: 用户 (pre-selected), 回收金额 (max=net allocated), 备注
- Submit: POST `/api/pool/:id/recall`

**Reconciliation View (Dialog):**
- Pool summary: 采购总额, 已分配, 未分配
- User table: 用户名, 分配额, 已消费, 剩余, 可回收
- Actions: 回收全部剩余, 结转下期, 关闭预算池
- API: GET `/api/pool/:id/reconciliation`

**RollOver Dialog:**
- Fields: 新周期Key, 新名称
- Submit: POST `/api/pool/:id/rollover`

Use same UI patterns as RechargesPage (Dialog forms, TanStack Table, shadcn components).

- [ ] **Step 3.7: Add remaining frontend tests**

```tsx
it('renders pool table columns correctly', async () => { /* verify column headers */ });
it('calls purchase API when purchase dialog submitted', async () => { /* mock api.post, fill form, submit */ });
it('shows error notification when API fails', async () => { /* mock api.get rejection */ });
it('disables closed pool actions', async () => { /* pool with status=closed should not show allocate/purchase */ });
it('filters by period type', async () => { /* select filter, verify API called with params */ });
```

- [ ] **Step 3.8: Verify build** — `cd web/modern && npm run build` and `cd web/modern && npx vitest run src/pages/pools/`

**Commit:** `feat(pool): add BudgetPoolsPage with full CRUD UI and tests`

---

## Task 4: Backend — AdminTopUp Pool Integration (Optional)

**Files:**
- Modify: `controller/user.go`

- [ ] **Step 4.1: Add optional `pool_id` to AdminTopUp**

```go
type adminTopUpRequest struct {
	UserId int    `json:"user_id"`
	Quota  int    `json:"quota"`
	Remark string `json:"remark"`
	PoolId int    `json:"pool_id"` // optional: allocate from pool
}
```

Logic:
```go
if req.PoolId > 0 {
    err = model.AllocateFromPool(ctx, req.PoolId, req.UserId, int64(req.Quota), req.Remark)
} else {
    err = model.IncreaseUserQuota(ctx, req.UserId, int64(req.Quota))
}
```

Existing callers without `pool_id` work unchanged.

**Commit:** `feat(pool): add optional pool_id to AdminTopUp`

---

## Task 5: Frontend — Polish & Dashboard Integration (Optional)

- [ ] **Step 5.1: Add pool summary to Dashboard** (if desired)

- [ ] **Step 5.2: Ensure consistent quota formatting** using `renderQuota()` across all pool tables.

---

## Execution Notes

### Database Compatibility
- **SQLite**: `clause.Locking{Strength: "UPDATE"}` works in GORM for SQLite via `BEGIN EXCLUSIVE`.
- **MySQL/PostgreSQL**: Standard row locking.

### Context Helper
The project uses `gmw.Ctx(c)` (see `controller/user.go:1514`). Use this in the controller.

### Dependencies needed in `model/pool.go`
```go
import (
    "context"
    "fmt"
    "github.com/Laisky/errors/v2"
    "github.com/songquanpeng/one-api/common"
    "github.com/songquanpeng/one-api/common/logger"
    "github.com/songquanpeng/one-api/common/helper"
    "go.uber.org/zap"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)
```

### Dependencies needed in `controller/pool.go`
```go
import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/songquanpeng/one-api/controller/gmw"
    "github.com/songquanpeng/one-api/model"
)
```

### Test Commands Summary
```bash
# Backend model tests
cd uniapi && go test ./model/ -run "Test.*Pool" -v

# Backend controller tests
cd uniapi && go test ./controller/ -run "Test.*Pool" -v

# Frontend tests
cd uniapi/web/modern && npx vitest run src/pages/pools/

# Full build verification
cd uniapi && go build ./...
cd uniapi/web/modern && npm run build
```
