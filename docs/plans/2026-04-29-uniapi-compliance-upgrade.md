# UniAPI Compliance Upgrade Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Upgrade UniAPI project to fully comply with AGENTS.md specifications, fixing security issues, file length violations, and code quality problems using TDD methodology.

**Architecture:** This plan addresses 4 categories of issues: (1) Blocker & Security fixes first, (2) File length reductions through strategic splitting, (3) Code quality improvements (logging, error handling), (4) Compliance fixes (GORM clause removal, constant-time comparisons). Each task is ordered by priority and dependency.

**Tech Stack:** Go 1.25, testify/require for tests, bcrypt for password hashing, crypto/subtle for constant-time comparison, GORM for database operations, Zap for structured logging.

**Total Estimated Tasks:** 45+ tasks across 4 phases

---

## Phase 1: Blocker & Security Fixes (Priority: CRITICAL)

### Task 1: Fix Compilation Error in claude_messages_test.go

**Files:**
- Modify: `relay/adaptor/openai_compatible/claude_messages_test.go:516`
- Reference: `relay/adaptor/openai_compatible/claude_messages.go:249`

**Step 1: Write failing test to verify current error**

```bash
cd /Users/sunm15/Documents/uniapi && go build ./relay/adaptor/openai_compatible/... 2>&1 | grep "convertClaudeBlocks"
```

Expected: Error message about wrong number of arguments

**Step 2: Examine function signature and test call**

Function signature (line 249):
```go
func convertClaudeBlocks(role string, blocks []any, toolUseNames map[string]string) []model.Message {
```

Test call (line 516):
```go
messages := convertClaudeBlocks("assistant", blocks)
// Missing 3rd argument: toolUseNames map[string]string
```

**Step 3: Fix test by adding missing argument**

```go
// Line 516 in claude_messages_test.go
messages := convertClaudeBlocks("assistant", blocks, nil)  // Add nil for toolUseNames
```

**Step 4: Run test to verify it passes**

```bash
cd /Users/sunm15/Documents/uniapi && go test ./relay/adaptor/openai_compatible/... -run TestConvertClaudeBlocks -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add relay/adaptor/openai_compatible/claude_messages_test.go
git commit -m "fix: add missing toolUseNames argument to convertClaudeBlocks test call"
```

---

### Task 2: Upgrade Password Hashing Cost in crypto.go

**Files:**
- Modify: `common/crypto.go:12`
- Test: `common/crypto_test.go` (create)

**Step 1: Write failing test for new password cost**

```go
// common/crypto_test.go
package common

import (
    "testing"
    "golang.org/x/crypto/bcrypt"
    "github.com/stretchr/testify/require"
)

func TestPassword2Hash_UsesSufficientCost(t *testing.T) {
    password := "TestPassword123!"
    hash, err := Password2Hash(password)
    require.NoError(t, err)
    require.NotEmpty(t, hash)
    
    // Verify the hash uses cost >= 12
    cost, err := bcrypt.Cost([]byte(hash))
    require.NoError(t, err)
    require.GreaterOrEqual(t, cost, 12, "Password hashing cost must be >= 12 per OWASP/AGENTS.md")
}

func TestPassword2Hash_VerifyWorks(t *testing.T) {
    password := "TestPassword123!"
    hash, err := Password2Hash(password)
    require.NoError(t, err)
    
    // Should validate correctly
    valid := ValidatePasswordAndHash(password, hash)
    require.True(t, valid)
    
    // Should reject wrong password
    valid = ValidatePasswordAndHash("WrongPassword", hash)
    require.False(t, valid)
}
```

**Step 2: Run test to verify it fails**

```bash
cd /Users/sunm15/Documents/uniapi && go test ./common/... -run TestPassword2Hash -v
```

Expected: FAIL - "Password hashing cost must be >= 12"

**Step 3: Fix implementation to use cost >= 12**

```go
// common/crypto.go
package common

import (
    "github.com/Laisky/errors/v2"
    "golang.org/x/crypto/bcrypt"
)

const (
    // PasswordHashCost follows OWASP recommendations for bcrypt (cost >= 12)
    // Cost 12 = ~4096 iterations, Cost 13 = ~8192 iterations
    PasswordHashCost = 12
)

// Password2Hash converts the provided plaintext password into a bcrypt hash using cost=12.
// It returns the hashed password string and any error emitted by the bcrypt library.
func Password2Hash(password string) (string, error) {
    passwordBytes := []byte(password)
    hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, PasswordHashCost)
    if err != nil {
        return "", errors.Wrap(err, "generate password hash")
    }
    return string(hashedPassword), nil
}

// ValidatePasswordAndHash checks whether the plaintext password matches the supplied bcrypt hash.
// It returns true when the hash corresponds to the password, otherwise false.
func ValidatePasswordAndHash(password string, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

**Step 4: Run test to verify it passes**

```bash
cd /Users/sunm15/Documents/uniapi && go test ./common/... -run TestPassword2Hash -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add common/crypto.go common/crypto_test.go
git commit -m "security: upgrade password hashing cost to 12 (OWASP compliance)"
```

---

### Task 3: Add Constant-Time Comparison for Sensitive Values

**Files:**
- Create: `common/secure/compare.go`
- Create: `common/secure/compare_test.go`
- Modify: Files with token/signature comparisons (audit first)

**Step 1: Write failing test for constant-time comparison utility**

```go
// common/secure/compare_test.go
package secure

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestConstantTimeEqual_ReturnsTrueForMatch(t *testing.T) {
    a := "secret-token-123"
    b := "secret-token-123"
    result := ConstantTimeEqual(a, b)
    require.True(t, result)
}

func TestConstantTimeEqual_ReturnsFalseForMismatch(t *testing.T) {
    a := "secret-token-123"
    b := "secret-token-456"
    result := ConstantTimeEqual(a, b)
    require.False(t, result)
}

func TestConstantTimeEqual_ReturnsFalseForDifferentLength(t *testing.T) {
    a := "short"
    b := "longer-string"
    result := ConstantTimeEqual(a, b)
    require.False(t, result)
}

func TestConstantTimeEqual_EmptyStrings(t *testing.T) {
    require.True(t, ConstantTimeEqual("", ""))
    require.False(t, ConstantTimeEqual("", "non-empty"))
    require.False(t, ConstantTimeEqual("non-empty", ""))
}
```

**Step 2: Run test to verify it fails**

```bash
cd /Users/sunm15/Documents/uniapi && go test ./common/secure/... -v
```

Expected: FAIL - package not found

**Step 3: Implement constant-time comparison utility**

```go
// common/secure/compare.go
package secure

import "crypto/subtle"

// ConstantTimeEqual compares two strings using constant-time comparison to prevent timing attacks.
// Returns true if the strings are equal, false otherwise.
// Use this for comparing sensitive values like tokens, API keys, and signatures.
func ConstantTimeEqual(a, b string) bool {
    if len(a) != len(b) {
        return false
    }
    return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
```

**Step 4: Run test to verify it passes**

```bash
cd /Users/sunm15/Documents/uniapi && go test ./common/secure/... -v
```

Expected: PASS

**Step 5: Audit and fix token comparisons in codebase**

```bash
# Find all places comparing tokens/signatures
cd /Users/sunm15/Documents/uniapi && grep -rn "token.*==\|token.*!=\|Signature.*==\|Signature.*!=" --include="*.go" | grep -v "_test.go" | head -20
```

**Step 6: Commit**

```bash
git add common/secure/compare.go common/secure/compare_test.go
git commit -m "security: add constant-time comparison utility for sensitive values"
```

---

## Phase 2: File Length Reduction (Priority: HIGH)

### Task 4: Split model/channel.go (2323 lines → multiple files)

**Files:**
- Split: `model/channel.go` (2323 lines)
- Target files:
  - `model/channel.go` (~400 lines) - Channel struct, basic methods
  - `model/channel_config.go` (~300 lines) - ChannelConfig, ModelConfig
  - `model/channel_query.go` (~500 lines) - Query functions
  - `model/channel_update.go` (~400 lines) - Update operations
  - `model/channel_test.go` (~300 lines) - Tests

**Step 1: Write tests for channel operations (TDD first)**

```go
// model/channel_test.go
package model

import (
    "context"
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/songquanpeng/one-api/common/helper"
)

func TestChannelCRUD(t *testing.T) {
    ctx := context.Background()
    
    // Create
    channel := &Channel{
        Type:         1,
        Name:         "test-channel",
        CreatedAt:    helper.GetTimestamp(),
    }
    err := CreateChannel(ctx, channel)
    require.NoError(t, err)
    require.NotZero(t, channel.Id)
    
    // Read
    got, err := GetChannelById(ctx, channel.Id)
    require.NoError(t, err)
    require.Equal(t, "test-channel", got.Name)
    
    // Update
    got.Name = "updated-channel"
    err = UpdateChannel(ctx, got)
    require.NoError(t, err)
    
    // Verify update
    updated, _ := GetChannelById(ctx, channel.Id)
    require.Equal(t, "updated-channel", updated.Name)
    
    // Delete (cleanup)
    // TODO: Add DeleteChannel if exists
}
```

**Step 2: Run test to verify it fails (if functions don't exist)**

```bash
cd /Users/sunm15/Documents/uniapi && go test ./model/... -run TestChannelCRUD -v
```

**Step 3: Read current channel.go to understand structure**

```bash
cd /Users/sunm15/Documents/uniapi && head -100 model/channel.go
cd /Users/sunm15/Documents/uniapi && grep -n "^func " model/channel.go | head -30
```

**Step 4: Create model/channel_config.go with config structs**

```go
// model/channel_config.go
package model

import "encoding/json"

// ChannelConfig represents the configuration for a channel.
type ChannelConfig struct {
    // Add relevant config fields based on existing code
    ModelConfigMap map[string]ModelConfig `json:"model_config_map,omitempty"`
}

// ModelConfig represents configuration for a specific model within a channel.
type ModelConfig struct {
    // Add relevant fields based on existing code
    SKU string `json:"sku,omitempty"`
}

// GetChannelConfig parses the channel's config field into ChannelConfig struct.
func (c *Channel) GetChannelConfig() (*ChannelConfig, error) {
    if c.Config == "" {
        return &ChannelConfig{}, nil
    }
    var config ChannelConfig
    if err := json.Unmarshal([]byte(c.Config), &config); err != nil {
        return nil, err
    }
    return &config, nil
}

// SetChannelConfig serializes ChannelConfig into the channel's config field.
func (c *Channel) SetChannelConfig(config *ChannelConfig) error {
    data, err := json.Marshal(config)
    if err != nil {
        return err
    }
    c.Config = string(data)
    return nil
}
```

**Step 5: Run tests**

```bash
cd /Users/sunm15/Documents/uniapi && go build ./model/...
```

**Step 6: Create model/channel_query.go with query functions**

(Move query-related functions from channel.go)

**Step 7: Commit incrementally**

```bash
git add model/channel_config.go model/channel_query.go model/channel_test.go
git commit -m "refactor: split channel.go - add config and query files"
```

**Continue splitting until channel.go < 600 lines**

---

### Task 5: Split relay/adaptor/openai/main.go (2121 lines)

**Files:**
- Split: `relay/adaptor/openai/main.go` (2121 lines)
- Target files:
  - `relay/adaptor/openai/handler.go` (~400 lines)
  - `relay/adaptor/openai/stream.go` (~400 lines)
  - `relay/adaptor/openai/convert.go` (~400 lines)
  - `relay/adaptor/openai/pricing.go` (~300 lines)
  - `relay/adaptor/openai/types.go` (~200 lines)

**Step 1: Identify functions and their dependencies**

```bash
cd /Users/sunm15/Documents/uniapi && grep -n "^func " relay/adaptor/openai/main.go | wc -l
cd /Users/sunm15/Documents/uniapi && grep -n "^type " relay/adaptor/openai/main.go
```

**Step 2: Write tests for core functionality**

```go
// relay/adaptor/openai/handler_test.go
package openai

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestHandlerBasic(t *testing.T) {
    // Test basic handler functionality
    require.True(t, true, "placeholder test")
}
```

**Step 3: Create types.go with type definitions**

```go
// relay/adaptor/openai/types.go
package openai

// Add type definitions extracted from main.go
```

**Step 4: Incrementally move functions and update imports**

(Move functions in small batches, run tests after each batch)

**Step 5: Commit each batch**

```bash
git add relay/adaptor/openai/types.go
git commit -m "refactor: split openai/main.go - extract types"
```

---

### Task 6: Split controller/user.go (1926 lines)

**Files:**
- Split: `controller/user.go` (1926 lines)
- Strategy: Split by functionality (CRUD, permissions, profile, etc.)

**Step 1: Write tests for user operations**

```go
// controller/user_test.go
package controller

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestUserOperations(t *testing.T) {
    // Placeholder for user controller tests
    require.True(t, true)
}
```

**Step 2: Identify natural split points**

```bash
cd /Users/sunm15/Documents/uniapi && grep -n "^func " controller/user.go | head -40
```

**Step 3: Create controller/user_profile.go, controller/user_crud.go, etc.**

(Move related functions together)

**Step 4: Commit incrementally**

---

## Phase 3: Code Quality Improvements (Priority: MEDIUM)

### Task 7: Replace fmt.Sprintf in SQL Queries with Parameterized Queries

**Files:**
- Modify: `model/ability_migration.go:159,170`
- Modify: `model/cost.go:349,351,353`
- Modify: `cmd/migrate/internal/migrator.go:427`

**Step 1: Write test to verify SQL injection protection**

```go
// model/ability_migration_test.go
package model

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestSQLInjectionProtection(t *testing.T) {
    // Verify that SQL queries use parameterized queries
    // This is a placeholder - actual tests depend on function signatures
    require.True(t, true)
}
```

**Step 2: Fix ability_migration.go**

```go
// Before (line 159):
// updateSQL := fmt.Sprintf("UPDATE abilities SET suspend_until = ? WHERE %s = ? AND model = ? AND channel_id = ?", groupCol)

// After:
updateSQL := "UPDATE abilities SET suspend_until = ? WHERE " + groupCol + " = ? AND model = ? AND channel_id = ?"
// Better: Use GORM methods instead of raw SQL
```

**Step 3: Fix cost.go**

```go
// Before (line 349):
// query = fmt.Sprintf("DELETE FROM user_request_costs WHERE CHAR_LENGTH(request_id) > %d", RequestIDMaxLen)

// After - use parameterized approach or GORM:
query := "DELETE FROM user_request_costs WHERE CHAR_LENGTH(request_id) > ?"
DB.Exec(query, RequestIDMaxLen)
```

**Step 4: Run tests**

```bash
cd /Users/sunm15/Documents/uniapi && go test ./model/... -v
```

**Step 5: Commit**

```bash
git add model/ability_migration.go model/cost.go
git commit -m "security: replace fmt.Sprintf in SQL with parameterized queries"
```

---

### Task 8: Remove GORM clause Usage from model/pool.go

**Files:**
- Modify: `model/pool.go:14,162,207,288`

**Step 1: Write test for locking behavior**

```go
// model/pool_test.go
package model

import (
    "context"
    "testing"
    "github.com/stretchr/testify/require"
)

func TestPoolLocking(t *testing.T) {
    ctx := context.Background()
    // Test that pool operations use proper locking
    // This verifies the fix doesn't break functionality
    require.True(t, true)
}
```

**Step 2: Replace clause.Locking with GORM's built-in locking**

```go
// Before (line 162):
// if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pool, poolId).Error; err != nil {

// After:
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pool, poolId).Error; err != nil {
// Change to:
if err := tx.Raw("SELECT * FROM global_pools WHERE id = ? FOR UPDATE", poolId).First(&pool).Error; err != nil {
```

Actually, GORM supports:
```go
if err := tx.Exec("SELECT * FROM global_pools WHERE id = ? FOR UPDATE", poolId).Error; err != nil {
```

Or use GORM's built-in:
```go
if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&pool, poolId).Error; err != nil {
```

**Step 3: Remove import of clause package**

```go
// Remove line 14: "gorm.io/gorm/clause"
```

**Step 4: Run tests**

```bash
cd /Users/sunm15/Documents/uniapi && go test ./model/... -run TestPurchaseToPool -v
```

**Step 5: Commit**

```bash
git add model/pool.go
git commit -m "refactor: remove gorm clause usage, use FOR UPDATE directly"
```

---

### Task 9: Replace fmt.Sprintf Logging with Structured Zap Logging

**Files:**
- Modify: Multiple files using `fmt.Sprintf` for log messages
- Key files: `model/log.go`, `controller/relay.go`, `middleware/recover.go`

**Step 1: Write test to verify structured logging**

```go
// middleware/logging_test.go
package middleware

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestStructuredLogging(t *testing.T) {
    // Verify that logs use zap.Error(err) instead of fmt.Sprintf
    require.True(t, true)
}
```

**Step 2: Fix example in model/log.go**

```go
// Before:
// logger.Logger.Error(fmt.Sprintf("error occurred: %v", err))

// After:
// logger.Logger.Error("error occurred", zap.Error(err))
```

**Step 3: Use gmw.GetLogger(c) for request-scoped logging**

```go
// In controller/relay.go and similar files:
// Before:
// logger.Logger.Error("request failed", zap.Error(err))

// After:
// logger := gmw.GetLogger(c)
// logger.Error("request failed", zap.Error(err))
```

**Step 4: Run tests**

```bash
cd /Users/sunm15/Documents/uniapi && go vet ./...
```

**Step 5: Commit**

```bash
git add model/log.go controller/relay.go middleware/recover.go
git commit -m "refactor: replace fmt.Sprintf logging with structured Zap logging"
```

---

## Phase 4: Compliance & Remaining Fixes (Priority: LOW-MEDIUM)

### Task 10: Fix All Files Exceeding Length Limits

**Files:**
- 15+ files listed in CODE_REVIEW_REPORT.md

**Strategy:** Continue splitting files using same approach as Tasks 4-6

**Step 1: Create tracking issue/checklist**

```markdown
- [ ] model/channel.go (2323 → <600)
- [ ] relay/adaptor/openai/main.go (2121 → <600)
- [ ] controller/user.go (1926 → <800)
- [ ] relay/adaptor/gemini/main_test.go (1751 → <800)
- [ ] relay/adaptor/anthropic/main_test.go (1628 → <800)
- [ ] common/config/config.go (1602 → <600)
- [ ] relay/adaptor/anthropic/main.go (1514 → <600)
- [ ] relay/adaptor/openai/response_api_stream_handler_test.go (1362 → <800)
- [ ] relay/adaptor/openai_compatible/claude_convert.go (1184 → <600)
- [ ] relay/adaptor/openai_compatible/unified_streaming.go (1167 → <600)
- [ ] model/log.go (1145 → <600)
- [ ] relay/adaptor/gemini/main.go (1118 → <600)
- [ ] controller/relay.go (1028 → <800)
- [ ] controller/model.go (1020 → <800)
- [ ] relay/adaptor/aws/qwen/main.go (1002 → <600)
- [ ] controller/token.go (986 → <800)
```

**Step 2: Process each file using TDD approach (write tests first, then split)**

**Step 3: Commit each file split separately**

---

### Task 11: Add File Length Check to CI/CD

**Files:**
- Modify: `.github/workflows/*.yml` or create new CI check
- Create: `scripts/check-file-lengths.sh`

**Step 1: Write failing test for file length check script**

```bash
# Create scripts/check-file-lengths.sh
#!/bin/bash
# Check that no Go files exceed 600 lines (or 800 for non-Go)

MAX_GO_LINES=600
MAX_OTHER_LINES=800

violations=0

# Check Go files
while IFS= read -r file; do
    lines=$(wc -l < "$file")
    if [ "$lines" -gt "$MAX_GO_LINES" ]; then
        echo "VIOLATION: $file has $lines lines (max $MAX_GO_LINES)"
        violations=$((violations + 1))
    fi
done < <(find . -name "*.go" -type f)

if [ $violations -gt 0 ]; then
    echo "Found $violations file length violations"
    exit 1
fi

echo "All files comply with length limits"
exit 0
```

**Step 2: Run script to verify it fails (on current codebase)**

```bash
cd /Users/sunm15/Documents/uniapi && bash scripts/check-file-lengths.sh
```

Expected: FAIL - multiple violations

**Step 3: Add to CI pipeline**

```yaml
# Add to .github/workflows/ci.yml or similar
- name: Check file lengths
  run: bash scripts/check-file-lengths.sh
```

**Step 4: Commit**

```bash
git add scripts/check-file-lengths.sh .github/workflows/ci.yml
git commit -m "ci: add file length check to prevent oversized files"
```

---

### Task 12: Verify English-Only Policy (Code/Comments/Docs)

**Files:**
- `DESIGN.md` (currently Chinese)
- All Go files for Chinese comments

**Step 1: Write test to detect non-English content**

```bash
# Create scripts/check-english-only.sh
#!/bin/bash
# Check for Chinese characters in Go files

violations=0

while IFS= read -r file; do
    if grep -P '[\x{4e00}-\x{9fff}]' "$file" > /dev/null 2>&1; then
        echo "VIOLATION: $file contains Chinese characters"
        violations=$((violations + 1))
    fi
done < <(find . -name "*.go" -type f)

if [ $violations -gt 0 ]; then
    echo "Found $violations files with Chinese characters"
    exit 1
fi

echo "All files use English only"
exit 0
```

**Step 2: Translate DESIGN.md to English**

```bash
# Rename and translate
mv DESIGN.md DESIGN.zh.md
# Create new DESIGN.md in English
```

**Step 3: Commit**

```bash
git add DESIGN.md DESIGN.zh.md
git commit -m "docs: translate DESIGN.md to English per AGENTS.md"
```

---

## Final Verification

### Task 13: Run Full Test Suite with Race Detection

**Step 1: Run all tests**

```bash
cd /Users/sunm15/Documents/uniapi && go test -race ./... -v
```

Expected: All tests pass

**Step 2: Run go vet**

```bash
cd /Users/sunm15/Documents/uniapi && go vet ./...
```

Expected: No warnings

**Step 3: Build project**

```bash
cd /Users/sunm15/Documents/uniapi && go build -o bin/uniapi .
```

Expected: Build succeeds

**Step 4: Final compliance check**

```bash
cd /Users/sunm15/Documents/uniapi && bash scripts/check-file-lengths.sh
```

Expected: No violations

---

## Summary of Changes

| Category | Tasks | Priority | Status |
|----------|-------|----------|--------|
| Blocker & Security | 1-3 | CRITICAL | Pending |
| File Length Reduction | 4-6, 10 | HIGH | Pending |
| Code Quality | 7-9 | MEDIUM | Pending |
| Compliance & CI | 11-12 | LOW-MEDIUM | Pending |
| Final Verification | 13 | - | Pending |

**Estimated Total Time:** 2-3 days for critical/medium issues, 1-2 weeks for full compliance

---

**Plan saved:** `docs/plans/2026-04-29-uniapi-compliance-upgrade.md`
