# slog 結構化日誌使用範例

## 基本使用

### 簡單日誌

```go
import "go-clean-arch/pkg/logger"

// Info 級別
logger.Info("Server started successfully", "port", 8080, "mode", "release")

// Warning 級別
logger.Warn("Cache miss", "key", "user:123", "retry_count", 3)

// Error 級別
logger.Error("Failed to connect to database", "error", err, "retry_after", "30s")

// Debug 級別（只在 debug 模式顯示）
logger.Debug("Processing request", "user_id", 123, "action", "login")
```

### 帶 Context 的日誌（推薦）

```go
// 在 HTTP Handler 中
func (h *UserHandler) CreateUser(c *gin.Context) {
    // 自動包含 trace ID
    logger.InfoContext(c.Request.Context(),
        "Creating new user",
        "email", req.Email,
        "role", req.Role,
    )

    user, err := h.userInteractor.CreateUser(req)
    if err != nil {
        logger.ErrorContext(c.Request.Context(),
            "Failed to create user",
            "email", req.Email,
            "error", err,
        )
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    logger.InfoContext(c.Request.Context(),
        "User created successfully",
        "user_id", user.ID,
        "email", user.Email,
    )
}
```

## 在不同場景中使用

### HTTP Handler 範例

```go
package handler

import (
    "go-clean-arch/pkg/logger"
    "go-clean-arch/pkg/middleware"
    "github.com/gin-gonic/gin"
)

func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")

    logger.InfoContext(c.Request.Context(),
        "Fetching user",
        "user_id", userID,
    )

    user, err := h.userInteractor.GetUserByID(userID)
    if err != nil {
        logger.WarnContext(c.Request.Context(),
            "User not found",
            "user_id", userID,
        )
        c.JSON(404, gin.H{"error": "User not found"})
        return
    }

    logger.InfoContext(c.Request.Context(),
        "User fetched successfully",
        "user_id", user.ID,
        "email", user.Email,
    )

    c.JSON(200, user)
}
```

### Worker 範例

```go
package consumer

import (
    "context"
    "go-clean-arch/pkg/logger"
    "github.com/google/uuid"
)

func (c *OrderConsumer) ConsumeMessage(messageBody []byte) error {
    // 創建帶 trace ID 的 context
    ctx := context.Background()
    traceID := uuid.New().String()
    ctx = logger.WithTraceID(ctx, traceID)

    logger.InfoContext(ctx, "Processing message")

    var msg OrderMessage
    if err := json.Unmarshal(messageBody, &msg); err != nil {
        logger.ErrorContext(ctx,
            "Failed to unmarshal message",
            "error", err,
            "raw_message", string(messageBody),
        )
        return err
    }

    logger.InfoContext(ctx,
        "Message parsed",
        "type", msg.Type,
        "order_id", msg.OrderID,
    )

    // 處理消息...
    return nil
}
```

### Cron Job 範例

```go
package job

import (
    "context"
    "go-clean-arch/pkg/logger"
    "github.com/google/uuid"
)

func (j *DailyReportJob) Run() error {
    // 為每次執行創建 trace ID
    ctx := context.Background()
    traceID := uuid.New().String()
    ctx = logger.WithTraceID(ctx, traceID)

    logger.InfoContext(ctx, "Starting daily report generation")

    users, err := j.userInteractor.ListUsers(1, 100)
    if err != nil {
        logger.ErrorContext(ctx,
            "Failed to fetch users",
            "error", err,
        )
        return err
    }

    logger.InfoContext(ctx,
        "Daily report completed",
        "total_users", len(users),
        "execution_time", time.Since(start).Seconds(),
    )

    return nil
}
```

### Repository 層範例

```go
package repository

import (
    "go-clean-arch/pkg/logger"
)

func (r *UserRepository) Create(user *domain.User) error {
    logger.Debug("Inserting user into database",
        "email", user.Email,
        "table", "users",
    )

    result := r.db.Create(user)
    if result.Error != nil {
        logger.Error("Database insert failed",
            "error", result.Error,
            "email", user.Email,
        )
        return result.Error
    }

    logger.Info("User created in database",
        "user_id", user.ID,
        "email", user.Email,
    )

    return nil
}
```

## 最佳實踐

### ✅ 推薦做法

```go
// 1. 使用 Context 版本的函數
logger.InfoContext(ctx, "message", "key", value)

// 2. 使用結構化欄位而非字串插值
logger.Info("User logged in", "user_id", 123, "ip", clientIP)
// ❌ 不要: logger.Info(fmt.Sprintf("User %d logged in from %s", 123, clientIP))

// 3. 錯誤日誌包含足夠的上下文
logger.ErrorContext(ctx, "Failed to process order",
    "order_id", orderID,
    "user_id", userID,
    "error", err,
    "retry_count", retries,
)

// 4. 使用合適的日誌級別
logger.Debug("Detailed debugging info")  // 開發時使用
logger.Info("Normal flow events")        // 重要業務事件
logger.Warn("Unexpected but handled")    // 可恢復的問題
logger.Error("Serious problems")         // 需要關注的錯誤
```

### ❌ 避免的做法

```go
// 1. 不要在循環中記錄過多日誌
for _, item := range items {
    logger.Info("Processing item", "id", item.ID)  // ❌ 太多日誌
}
// 改為：
logger.Info("Processing items", "count", len(items))
logger.Debug("Item details", "items", items)  // Debug 級別

// 2. 不要記錄敏感信息
logger.Info("User logged in", "password", password)  // ❌ 危險
logger.Info("User logged in", "user_id", userID)     // ✅ 安全

// 3. 不要忽略 Context
logger.Info("message")  // ❌ 缺少 trace ID
logger.InfoContext(ctx, "message")  // ✅ 包含 trace ID
```

## 日誌輸出範例

### Text 格式（開發環境）

```
time=2025-11-02T10:30:45.123+08:00 level=INFO source=handler/user_handler.go:42 msg="User created successfully" trace_id=550e8400-e29b-41d4-a716-446655440000 user_id=123 email=user@example.com
```

### JSON 格式（生產環境）

```json
{
  "time": "2025-11-02T10:30:45.123Z",
  "level": "INFO",
  "source": {
    "file": "handler/user_handler.go",
    "line": 42
  },
  "msg": "User created successfully",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": 123,
  "email": "user@example.com"
}
```

## 日誌級別指南

| 級別 | 何時使用 | 範例 |
|------|---------|------|
| **Debug** | 詳細的調試信息，開發時有用 | 函數參數、中間變量、詳細流程 |
| **Info** | 正常業務流程的重要事件 | 用戶登入、訂單創建、作業完成 |
| **Warn** | 異常但可處理的情況 | Cache miss、重試、降級服務 |
| **Error** | 需要關注的錯誤 | 數據庫錯誤、外部 API 失敗 |

## 查詢日誌技巧

### 使用 grep 查詢

```bash
# 查詢特定 trace ID
grep "550e8400-e29b-41d4-a716-446655440000" app.log

# 查詢錯誤日誌
grep "level=ERROR" app.log

# 查詢特定用戶的操作
grep "user_id=123" app.log
```

### 使用 jq 查詢 JSON 日誌

```bash
# 過濾特定 trace ID
cat app.log | jq 'select(.trace_id == "550e8400...")'

# 只顯示錯誤
cat app.log | jq 'select(.level == "ERROR")'

# 提取特定欄位
cat app.log | jq '{time, trace_id, msg, error}'

# 計算錯誤數量
cat app.log | jq 'select(.level == "ERROR")' | wc -l
```

## 與舊 logger 的兼容性

專案保留了舊的 logger API 用於向後兼容：

```go
// 舊的 API（仍可使用，但不推薦）
logger.LogInfo("User created: %s", email)
logger.LogWarning("Cache miss for key: %s", key)
logger.LogError("Database error: %v", err)

// 新的 API（推薦）
logger.Info("User created", "email", email)
logger.Warn("Cache miss", "key", key)
logger.Error("Database error", "error", err)
```

建議逐步遷移到新的結構化日誌 API。
