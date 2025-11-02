# 專案改進說明

本文檔說明最近對專案所做的架構改進。

## 📋 改進項目總覽

1. ✅ **Viper 配置管理** - 替換原有的 YAML 解析
2. ✅ **slog 結構化日誌** - 替換標準 log 包
3. ✅ **Trace ID 追蹤** - 實作請求追蹤機制

---

## 1. Viper 配置管理

### 特性

- ✅ 支援多種配置格式（TOML, YAML, JSON）
- ✅ 自動環境變數覆蓋（使用 `APP_` 前綴）
- ✅ 配置驗證
- ✅ 預設值設定

### 配置文件

主配置檔：`configs/config.toml`（TOML 格式範例）

```toml
[server]
port = 8080
host = "0.0.0.0"
mode = "debug"

[database]
driver = "postgres"
host = "localhost"
port = 5432
user = "postgres"
password = "postgres"
dbname = "go_clean_arch"

[logger]
level = "info"   # debug, info, warn, error
format = "json"  # json, text
```

### 環境變數覆蓋

使用 `APP_` 前綴覆蓋任何配置值：

```bash
# 覆蓋資料庫密碼
export APP_DATABASE_PASSWORD=secret_password

# 覆蓋 logger 級別
export APP_LOGGER_LEVEL=debug

# 覆蓋 server port
export APP_SERVER_PORT=9090
```

**命名規則：** 配置路徑中的 `.` 替換為 `_`

範例：
- `database.password` → `APP_DATABASE_PASSWORD`
- `logger.level` → `APP_LOGGER_LEVEL`
- `server.port` → `APP_SERVER_PORT`

### 配置文件範例

專案提供了三個配置文件範例：

1. `config.toml` - 預設配置（開發環境）
2. `config.example.toml` - 配置範本
3. `config.production.toml` - 生產環境範例

---

## 2. slog 結構化日誌系統

### 特性

- ✅ 結構化日誌（JSON/Text 格式）
- ✅ 多種日誌級別（debug, info, warn, error）
- ✅ 自動添加 trace ID
- ✅ 上下文感知（Context-aware logging）
- ✅ 向後兼容舊的 logger API

### 使用方式

#### 基本用法（無 context）

```go
import "go-clean-arch/pkg/logger"

// 簡單日誌
logger.Info("Server started", "port", 8080)
logger.Warn("High memory usage", "usage_percent", 85)
logger.Error("Database connection failed", "error", err)
logger.Debug("Request payload", "data", payload)
```

#### 使用 Context（推薦，支援 trace ID）

```go
// 在 HTTP handler 中
func (h *UserHandler) CreateUser(c *gin.Context) {
    logger.InfoContext(c.Request.Context(), "Creating user",
        "email", userEmail,
        "role", userRole,
    )
}

// 在 Worker/Cron 中
func (j *DailyReportJob) Run() error {
    ctx := context.Background()
    traceID := uuid.New().String()
    ctx = logger.WithTraceID(ctx, traceID)

    logger.InfoContext(ctx, "Starting daily report")
    // 所有日誌都會帶上 trace_id
}
```

#### 日誌格式範例

**JSON 格式**（推薦用於生產環境）：
```json
{
  "time": "2025-11-02T10:30:45.123Z",
  "level": "INFO",
  "source": {
    "file": "handler/user_handler.go",
    "line": 42
  },
  "msg": "HTTP request completed",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "path": "/api/v1/users",
  "status": 201,
  "latency_ms": 45
}
```

**Text 格式**（開發環境易讀）：
```
time=2025-11-02T10:30:45.123+08:00 level=INFO source=handler/user_handler.go:42 msg="HTTP request completed" trace_id=550e8400-e29b-41d4-a716-446655440000 method=POST path=/api/v1/users status=201 latency_ms=45
```

### 配置日誌級別和格式

在 `config.toml` 中配置：

```toml
[logger]
level = "info"   # 控制顯示哪些級別的日誌
format = "json"  # json 或 text
```

或使用環境變數：
```bash
export APP_LOGGER_LEVEL=debug
export APP_LOGGER_FORMAT=text
```

---

## 3. Trace ID 追蹤機制

### 特性

- ✅ 自動為每個 HTTP 請求生成 trace ID
- ✅ 支援外部傳入的 trace ID（X-Trace-ID header）
- ✅ 在所有日誌中自動記錄 trace ID
- ✅ 在 Worker 和 Cron Job 中支援 trace ID

### HTTP 請求追蹤

#### 自動生成 Trace ID

每個 HTTP 請求都會自動獲得一個 UUID 格式的 trace ID：

```bash
# 發送請求
curl http://localhost:8080/api/v1/users

# 響應 header 會包含
X-Trace-ID: 550e8400-e29b-41d4-a716-446655440000
```

#### 傳遞 Trace ID

客戶端可以傳遞自己的 trace ID（用於分佈式追蹤）：

```bash
curl -H "X-Trace-ID: my-custom-trace-id" \
     http://localhost:8080/api/v1/users
```

### 在代碼中使用 Trace ID

#### HTTP Handler 中

```go
import "go-clean-arch/pkg/middleware"

func (h *UserHandler) CreateUser(c *gin.Context) {
    // 獲取 trace ID
    traceID := middleware.GetTraceID(c)

    // 使用 context 記錄日誌（自動包含 trace ID）
    logger.InfoContext(c.Request.Context(), "Processing request")
}
```

#### Worker 中

```go
func (c *OrderConsumer) ConsumeMessage(messageBody []byte) error {
    // 為消息處理創建 trace ID
    ctx := context.Background()
    traceID := uuid.New().String()
    ctx = logger.WithTraceID(ctx, traceID)

    logger.InfoContext(ctx, "Processing message")
    // 所有後續日誌都會包含這個 trace ID
}
```

### Trace ID 的用途

1. **請求追蹤** - 在日誌中追蹤單個請求的完整生命週期
2. **問題排查** - 快速定位特定請求的所有相關日誌
3. **分佈式追蹤** - 跨服務傳遞 trace ID 實現端到端追蹤
4. **效能分析** - 分析特定請求的處理時間

### 日誌查詢範例

```bash
# 查詢特定 trace ID 的所有日誌
grep "550e8400-e29b-41d4-a716-446655440000" app.log

# 使用 jq 過濾 JSON 日誌
cat app.log | jq 'select(.trace_id == "550e8400-e29b-41d4-a716-446655440000")'
```

---

## 🚀 快速開始

### 1. 準備配置文件

```bash
# 複製範例配置
cp configs/config.example.toml configs/config.toml

# 編輯配置（或使用環境變數）
vim configs/config.toml
```

### 2. 啟動應用

```bash
# API 模式
./bin/app -mode api -config configs/config.toml

# Worker 模式
./bin/app -mode worker -config configs/config.toml

# Cron 模式
./bin/app -mode cron -config configs/config.toml
```

### 3. 使用環境變數

```bash
# 覆蓋配置
export APP_DATABASE_PASSWORD=secret
export APP_LOGGER_LEVEL=debug
export APP_LOGGER_FORMAT=text

# 啟動
./bin/app -mode api
```

---

## 📊 對比：改進前後

### 配置管理

| 功能 | 改進前 | 改進後 |
|------|--------|--------|
| 格式支援 | YAML only | TOML, YAML, JSON |
| 環境變數 | 4 個硬編碼 | 全自動，任意配置 |
| 驗證 | 無 | 啟動時驗證 |
| 預設值 | 無 | 完整預設值 |

### 日誌系統

| 功能 | 改進前 | 改進後 |
|------|--------|--------|
| 格式 | 純文字 | JSON/Text 結構化 |
| 級別控制 | 無 | debug/info/warn/error |
| Trace ID | 無 | 自動記錄 |
| 上下文 | 無 | Context-aware |
| 生產就緒 | ❌ | ✅ |

### 可觀測性

| 功能 | 改進前 | 改進後 |
|------|--------|--------|
| 請求追蹤 | ❌ | ✅ UUID trace ID |
| 日誌關聯 | ❌ | ✅ 自動關聯 |
| 分佈式追蹤 | ❌ | ✅ 支援傳遞 |
| Worker 追蹤 | ❌ | ✅ 支援 |

---

## 🔧 故障排查

### 配置加載失敗

```bash
# 錯誤：failed to read config file
# 解決：確認配置文件路徑正確
ls -la configs/config.toml

# 錯誤：invalid logger level
# 解決：使用正確的日誌級別：debug, info, warn, error
```

### 日誌未顯示

```bash
# 檢查日誌級別設定
# debug < info < warn < error
# 如果設定為 error，則 info 和 debug 不會顯示

# 查看當前配置
cat configs/config.toml | grep -A2 "logger"
```

### Trace ID 未出現

```bash
# 確認使用了 Context 版本的日誌函數
logger.InfoContext(ctx, "message")  # ✅ 會包含 trace ID
logger.Info("message")              # ❌ 不會包含 trace ID
```

---

## 📚 延伸閱讀

- [Viper 官方文檔](https://github.com/spf13/viper)
- [Go slog 包文檔](https://pkg.go.dev/log/slog)
- [TOML 語法指南](https://toml.io/)
- [結構化日誌最佳實踐](https://www.innoq.com/en/blog/structured-logging/)

---

## 🎯 下一步建議

如果專案持續成長，可以考慮：

1. **Wire 依賴注入** - 減少手動 DI 代碼重複
2. **OpenTelemetry** - 完整的分佈式追蹤和指標
3. **配置熱更新** - Viper 支援配置文件變更監聽
4. **日誌輪轉** - 添加檔案日誌和自動輪轉
5. **Cobra CLI** - 更好的命令列介面體驗
