# 檔案管家 — Go 專案結構設計

**日期：** 2026-04-16
**版本：** v1.0
**關聯規格：**
- `2026-04-16-file-manager-design.md`
- `2026-04-16-file-manager-ddd-domain-design.md`

---

## 1. 設計決策

| 決策 | 選擇 | 理由 |
|------|------|------|
| 對外介面 | Wails binding only | 純桌面工具，不需 HTTP layer |
| 結構方式 | Wails-default + internal DDD 分層 | 符合 Go 慣例、DDD 邊界清晰、易測試 |
| Wails binding 位置 | `app.go`（薄殼） | 業務邏輯不滲入 binding，換框架時只改此層 |
| Repository interface | 定義於 domain 層 | 依賴反轉，infrastructure 實作 domain 宣告的合約 |

---

## 2. 完整目錄結構

```
dobby/
├── main.go                              # Wails 進入點、依賴組裝（DI root）
├── app.go                               # Wails binding 薄殼
├── wails.json
│
├── build/                               # Wails 打包資產
│   ├── appicon.png
│   └── windows/
│       └── wails.exe.manifest
│
├── frontend/                            # React + Vite（Wails 自動管理）
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Rules.tsx
│   │   │   ├── Logs.tsx
│   │   │   └── Settings.tsx
│   │   └── wailsjs/                     # 自動產生的 TypeScript binding（勿手改）
│   ├── package.json
│   └── vite.config.ts
│
└── internal/
    ├── domain/
    │   ├── rule/
    │   │   ├── rule.go                  # Rule aggregate root + constructor + methods
    │   │   ├── value_objects.go         # RuleId, WatchConfig, FilterSpec,
    │   │   │                            #   NamingTemplate, TargetPathTemplate
    │   │   ├── events.go                # RuleCreated, RuleUpdated, RuleDeleted,
    │   │   │                            #   RuleEnabled, RuleDisabled
    │   │   └── repository.go            # IRuleRepository interface（依賴反轉）
    │   │
    │   ├── job/
    │   │   ├── job.go                   # ProcessingJob aggregate root + state machine
    │   │   ├── value_objects.go         # JobId, FileEvent, ProcessingContext,
    │   │   │                            #   ProcessingResult, JobState（enum）
    │   │   ├── events.go                # FileDetected, RuleMatched,
    │   │   │                            #   ProcessingSucceeded, ProcessingFailed
    │   │   └── repository.go            # IProcessingJobRepository interface
    │   │
    │   └── service/                     # Domain Services（跨 Aggregate 的業務邏輯）
    │       ├── rule_matcher.go          # Match(FileEvent, Rule) bool
    │       ├── template_renderer.go     # Render(template, ProcessingContext) string
    │       └── sequence_generator.go   # Generate(ruleId, date) string（查 OperationLog）
    │
    ├── application/
    │   ├── rule_service.go              # CreateRule, UpdateRule, DeleteRule,
    │   │                               #   EnableRule, DisableRule, ListRules, GetRule
    │   └── job_service.go              # HandleFileEvent（orchestrate 整條處理流程）
    │
    ├── infrastructure/
    │   ├── persistence/
    │   │   ├── db.go                    # SQLite 連線初始化 + migrations
    │   │   ├── rule_repository.go       # IRuleRepository 實作（SQLite）
    │   │   ├── job_repository.go        # IProcessingJobRepository 實作（SQLite）
    │   │   └── operation_log_repository.go  # 寫入 OperationLog（Read Model projection）
    │   │
    │   ├── watcher/
    │   │   └── file_watcher.go          # fsnotify 封裝，偵測到新檔案後呼叫 JobService
    │   │
    │   ├── filesystem/
    │   │   └── file_system_service.go   # 實際執行 rename + move + mkdir
    │   │
    │   └── notification/
    │       └── notification_service.go  # Wails runtime 桌面通知
    │
    └── query/                           # Read Model queries（OperationLog 投影）
        └── operation_log_query.go       # ListLogs(filter), CountTodayByRule(ruleId)
```

---

## 3. 各層職責

### 3.1 `main.go`（DI Root）

唯一允許知道所有層的地方。負責：

1. 初始化 SQLite 連線（`persistence/db.go`）
2. 組裝所有 repository、service、infrastructure 實例
3. 將組裝好的 `*App` 傳給 Wails 啟動

```go
// 組裝順序示意
db := persistence.NewDB("dobby.db")
ruleRepo := persistence.NewRuleRepository(db)
logRepo  := persistence.NewOperationLogRepository(db)
logQuery := query.NewOperationLogQuery(db)

seqGen       := service.NewSequenceGenerator(logRepo)
renderer     := service.NewTemplateRenderer()
matcher      := service.NewRuleMatcher()
fsService    := filesystem.NewFileSystemService()
notifService := notification.NewNotificationService(wailsCtx)

ruleService := application.NewRuleService(ruleRepo)
jobService  := application.NewJobService(ruleRepo, matcher, seqGen, renderer, fsService, logRepo, notifService)

watcher := watcher.NewFileWatcher(ruleRepo, jobService)
app     := NewApp(ruleService, logQuery, watcher)
```

---

### 3.2 `app.go`（Wails Binding 薄殼）

只做兩件事：接收前端呼叫、委派給 application layer。**禁止**在此寫業務邏輯。

```go
type App struct {
    ruleService *application.RuleService
    logQuery    *query.OperationLogQuery
    watcher     *watcher.FileWatcher
}

// ── Rules ──────────────────────────────────────────
func (a *App) CreateRule(req CreateRuleRequest) (*RuleDTO, error)
func (a *App) UpdateRule(id string, req UpdateRuleRequest) (*RuleDTO, error)
func (a *App) DeleteRule(id string) error
func (a *App) EnableRule(id string) error
func (a *App) DisableRule(id string) error
func (a *App) ListRules() ([]*RuleDTO, error)

// ── Operation Logs ──────────────────────────────────
func (a *App) ListOperationLogs(filter LogFilter) ([]*OperationLogDTO, error)

// ── Service Control ─────────────────────────────────
func (a *App) StartWatcher() error
func (a *App) StopWatcher() error
func (a *App) GetServiceStatus() ServiceStatus
```

---

### 3.3 `internal/domain/`（Domain Layer）

- **不依賴任何外層**（無 import infra / app）
- Aggregate root 持有並保護內部 value objects
- Repository **interface** 定義於此，由 infrastructure **實作**（依賴反轉）
- Domain service 只依賴 domain 內部的型別

---

### 3.4 `internal/application/`（Application Layer）

- 只依賴 domain layer 和 domain service
- 透過 interface 使用 repository（不知道 SQLite 存在）
- `RuleService`：Rule 的 CRUD + 啟用/停用
- `JobService`：`HandleFileEvent()` 依 DDD spec 的背景服務執行流程 orchestrate 整條狀態流轉

---

### 3.5 `internal/infrastructure/`（Infrastructure Layer）

- 依賴 domain interfaces，不被 domain 依賴
- `persistence/`：SQLite 實作，schema migration 在 `db.go` 啟動時執行
- `watcher/`：封裝 fsnotify，僅負責偵測事件並呼叫 `JobService.HandleFileEvent()`
- `filesystem/`：rename + move，若目標資料夾不存在則自動 `MkdirAll`
- `notification/`：持有 Wails `context.Context`，呼叫 `runtime.EventsEmit` 發送通知

---

### 3.6 `internal/query/`（Read Model）

- `OperationLogQuery` 直接查詢 SQLite，不經過 domain aggregate
- 供 `app.go` 的 `ListOperationLogs` 使用
- 保留最近 1000 筆的清除邏輯也放在此處

---

## 4. 依賴方向

```
frontend
    ↓  (Wails binding)
app.go
    ↓
application/
    ↓
domain/ ←── infrastructure/
              (實作 domain interfaces)
query/  ←── infrastructure/persistence/
              (共用 db 連線)
```

規則：**外層依賴內層，內層不知道外層存在。**

---

## 5. SQLite Schema 對應

| Go 層 | 資料表 | 說明 |
|-------|--------|------|
| `domain/rule` + `infrastructure/persistence` | `rules` | Rule aggregate 持久化 |
| `domain/job` + `infrastructure/persistence` | `processing_jobs` | 短期狀態，可定期清除 |
| `query/` + `infrastructure/persistence` | `operation_logs` | Read Model，保留最近 1000 筆 |

---

## 6. 測試策略

| 層 | 測試方式 |
|----|---------|
| `domain/` | 純 unit test，無外部依賴 |
| `domain/service/` | 純 unit test（RuleMatcher、TemplateRenderer 輸入輸出明確） |
| `application/` | Mock repository interface，unit test use case 流程 |
| `infrastructure/persistence/` | Integration test（in-memory SQLite） |
| `app.go` | 可跳過或做 smoke test（薄殼邏輯少） |
