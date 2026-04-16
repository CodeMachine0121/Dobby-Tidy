# 檔案管家 — Feature Architecture

**日期：** 2026-04-16
**版本：** v1.0
**關聯規格：**
- `features/file-manager.feature`
- `docs/superpowers/specs/2026-04-16-file-manager-design.md`
- `docs/superpowers/specs/2026-04-16-file-manager-ddd-domain-design.md`
- `docs/superpowers/specs/2026-04-16-file-manager-go-project-structure.md`

---

## 1. 專案上下文

| 項目 | 說明 |
|------|------|
| 語言 | Go 1.21+ |
| 框架 | Wails v2（桌面 App：Go 後端 + WebView 前端） |
| 前端 | React + Vite（TypeScript） |
| 資料庫 | SQLite via `modernc.org/sqlite`（pure Go，無需 CGO） |
| 架構模式 | Tactical DDD（Aggregate Root、Value Object、Domain Service、Repository Interface） |
| 測試框架 | Go 標準 `testing` package + `github.com/stretchr/testify` |

### 依賴方向

```
frontend (React)
    ↓  Wails binding
app.go (薄殼)
    ↓
internal/application/
    ↓
internal/domain/        ← internal/infrastructure/ 實作 domain interfaces
internal/query/         ← internal/infrastructure/persistence/ 共用 db
```

外層依賴內層，內層不知外層存在。

---

## 2. 資料模型

### 2.1 Rule（Aggregate Root）
**Gherkin 來源：** S-1、S-2、S-3、S-4、S-5（行 15–65）

```go
// internal/domain/rule/rule.go
type Rule struct {
    id          RuleId
    name        string
    enabled     bool
    watchConfig WatchConfig
    filterSpec  FilterSpec
    nameTemplate  NamingTemplate
    targetTemplate TargetPathTemplate
    project     string
    typeLabel   string
    createdAt   time.Time
    updatedAt   time.Time
}
```

### 2.2 RuleId（Value Object）
**Gherkin 來源：** S-1（"規則擁有唯一的 RuleId"）

```go
// internal/domain/rule/value_objects.go
type RuleId struct { value string }
```

### 2.3 WatchConfig（Value Object）
**Gherkin 來源：** S-1、S-2（監控資料夾路徑唯一性約束）

```go
type WatchConfig struct {
    FolderPath string
    Recursive  bool
}
```

### 2.4 FilterSpec（Value Object）
**Gherkin 來源：** S-6、S-7、S-8、S-9、S-10（行 75–120）

```go
type FilterSpec struct {
    Extensions []string  // e.g. [".png", ".jpg"]；空清單 = 不過濾
    Keyword    string    // 原始檔名包含此字；空字串 = 不過濾
}
```

### 2.5 NamingTemplate（Value Object）
**Gherkin 來源：** S-11、S-12（行 122–143）

```go
type NamingTemplate struct {
    TemplateString string
}
```

### 2.6 TargetPathTemplate（Value Object）
**Gherkin 來源：** S-13（行 145–154）

```go
type TargetPathTemplate struct {
    PathTemplate string
}
```

### 2.7 ProcessingJob（Aggregate Root）
**Gherkin 來源：** S-16、S-17、S-18、S-19、S-20、S-21（行 173–221）

```go
// internal/domain/job/job.go
type ProcessingJob struct {
    id        JobId
    ruleId    rule.RuleId
    fileEvent FileEvent
    state     JobState
    context   *ProcessingContext  // nil until matched
    result    *ProcessingResult   // nil until succeeded/failed
}
```

### 2.8 JobId（Value Object）
**Gherkin 來源：** S-16（"Job 擁有唯一的 JobId"）

```go
// internal/domain/job/value_objects.go
type JobId struct { value string }
```

### 2.9 FileEvent（Value Object）
**Gherkin 來源：** S-16（行 173–177）

```go
type FileEvent struct {
    DetectedPath string
    OriginalName string
    Extension    string
    DetectedAt   time.Time
}
```

### 2.10 JobState（Enum）
**Gherkin 來源：** S-16 ~ S-21

```go
type JobState string
const (
    JobStatePending    JobState = "pending"
    JobStateMatched    JobState = "matched"
    JobStateProcessing JobState = "processing"
    JobStateSucceeded  JobState = "succeeded"
    JobStateFailed     JobState = "failed"
)
```

### 2.11 ProcessingContext（Value Object）
**Gherkin 來源：** S-11、S-12、S-13、S-17（行 122–154、185–190）

```go
type ProcessingContext struct {
    Project      string
    TypeLabel    string
    Date         time.Time
    Seq          string
    OriginalName string
    Extension    string
}
```

### 2.12 ProcessingResult（Value Object）
**Gherkin 來源：** S-19、S-20（行 196–210）

```go
type ProcessingResult struct {
    NewPath      string
    ErrorMessage string
    ProcessedAt  time.Time
}
```

### 2.13 OperationLog（Read Model，非 Aggregate）
**Gherkin 來源：** S-14、S-15（行 157–170）— SequenceGenerator 查詢此表

```go
// internal/query/operation_log_query.go
type OperationLog struct {
    LogId        string
    RuleId       string
    RuleName     string
    OriginalPath string
    NewPath      string
    Status       string  // "success" | "error"
    ErrorMessage string
    ProcessedAt  time.Time
}
```

---

## 3. 服務介面

### 3.1 IRuleRepository（Domain Interface）
**Gherkin 來源：** S-1、S-2、S-3、S-4、S-5

```go
// internal/domain/rule/repository.go
type IRuleRepository interface {
    Save(ctx context.Context, rule *Rule) error
    FindById(ctx context.Context, id RuleId) (*Rule, error)
    FindByFolderPath(ctx context.Context, folderPath string) (*Rule, error)
    ExistsByFolderPath(ctx context.Context, folderPath string) (bool, error)
    Delete(ctx context.Context, id RuleId) error
    ListAll(ctx context.Context) ([]*Rule, error)
}
```

### 3.2 IProcessingJobRepository（Domain Interface）
**Gherkin 來源：** S-16 ~ S-21

```go
// internal/domain/job/repository.go
type IProcessingJobRepository interface {
    Save(ctx context.Context, job *ProcessingJob) error
    FindById(ctx context.Context, id JobId) (*ProcessingJob, error)
}
```

### 3.3 IOperationLogRepository（Infrastructure → Query）
**Gherkin 來源：** S-14、S-15

```go
// internal/infrastructure/persistence/operation_log_repository.go（實作）
// internal/query/operation_log_query.go（讀取）
type IOperationLogRepository interface {
    Save(ctx context.Context, log *OperationLog) error
    CountSuccessByRuleAndDate(ctx context.Context, ruleId string, date time.Time) (int, error)
}
```

### 3.4 RuleMatcher（Domain Service）
**Gherkin 來源：** S-6、S-7、S-8、S-9、S-10

```go
// internal/domain/service/rule_matcher.go
type RuleMatcher struct{}
func (m *RuleMatcher) Match(event FileEvent, rule *rule.Rule) bool
```

比對邏輯（AND 條件）：
1. `len(rule.FilterSpec.Extensions) == 0` **或** `event.Extension ∈ rule.FilterSpec.Extensions`
2. `rule.FilterSpec.Keyword == ""` **或** `strings.Contains(event.OriginalName, rule.FilterSpec.Keyword)`

### 3.5 TemplateRenderer（Domain Service）
**Gherkin 來源：** S-11、S-12、S-13

```go
// internal/domain/service/template_renderer.go
type TemplateRenderer struct{}
func (r *TemplateRenderer) RenderName(template rule.NamingTemplate, ctx job.ProcessingContext) string
func (r *TemplateRenderer) RenderPath(template rule.TargetPathTemplate, ctx job.ProcessingContext) string
```

支援變數：`{project}`, `{type}`, `{YYYY}`, `{MM}`, `{DD}`, `{seq}`, `{original}`, `{ext}`

### 3.6 SequenceGenerator（Domain Service）
**Gherkin 來源：** S-14、S-15

```go
// internal/domain/service/sequence_generator.go
type SequenceGenerator struct {
    logRepo IOperationLogRepository
}
func (g *SequenceGenerator) Generate(ctx context.Context, ruleId string, date time.Time) (string, error)
```

邏輯：查詢 `CountSuccessByRuleAndDate`，結果 +1 後格式化為三位數 `%03d`。

### 3.7 Rule Aggregate Methods（行為）
**Gherkin 來源：** S-3、S-4

```go
// internal/domain/rule/rule.go
func (r *Rule) Enable() error
func (r *Rule) Disable() error
```

### 3.8 ProcessingJob Aggregate Methods（狀態機）
**Gherkin 來源：** S-17、S-18、S-19、S-20、S-21

```go
// internal/domain/job/job.go
func NewProcessingJob(event FileEvent) *ProcessingJob
func (j *ProcessingJob) MarkMatched(ruleId rule.RuleId, ctx ProcessingContext) error
func (j *ProcessingJob) MarkProcessing() error
func (j *ProcessingJob) MarkSucceeded(newPath string) error
func (j *ProcessingJob) MarkFailed(errorMessage string) error
```

非法狀態轉移回傳 `ErrInvalidStateTransition`。

---

## 4. 架構決策

| 決策 | 選擇 | 理由 |
|------|------|------|
| 監控資料夾唯一性 | Repository 層強制（ExistsByFolderPath） | 不污染 domain aggregate，由 Application Service 呼叫 |
| 空副檔名清單語意 | 空 = 不過濾（接受所有副檔名） | 見 S-10；最大彈性，明確業務語意 |
| SequenceGenerator 依賴 | 依賴 IOperationLogRepository interface | 依賴反轉，domain service 不知道 SQLite 存在 |
| ProcessingResult | pointer（`*ProcessingResult`） | nil 表示尚未有結果，比 zero-value 語意更明確 |
| 非法狀態轉移 | 回傳 error（不 panic） | 業務錯誤，讓 Application Service 決定如何處理 |

---

## 5. Scenario 對應表

| Scenario | 涉及元件 |
|----------|---------|
| S-1 | Rule 建立、RuleId 生成、IRuleRepository.Save |
| S-2 | IRuleRepository.ExistsByFolderPath、Application 層 duplicate 檢查 |
| S-3 | Rule.Disable() |
| S-4 | Rule.Enable() |
| S-5 | IRuleRepository.Delete |
| S-6 | RuleMatcher.Match（extension match，無 keyword） |
| S-7 | RuleMatcher.Match（extension mismatch） |
| S-8 | RuleMatcher.Match（extension + keyword match） |
| S-9 | RuleMatcher.Match（extension match，keyword mismatch） |
| S-10 | RuleMatcher.Match（空 extension list） |
| S-11 | TemplateRenderer.RenderName（完整變數） |
| S-12 | TemplateRenderer.RenderName（{original} 變數） |
| S-13 | TemplateRenderer.RenderPath（動態路徑） |
| S-14 | SequenceGenerator.Generate（無歷史紀錄） |
| S-15 | SequenceGenerator.Generate（已有 3 筆） |
| S-16 | NewProcessingJob、JobId 生成 |
| S-17 | ProcessingJob.MarkMatched |
| S-18 | ProcessingJob.MarkProcessing |
| S-19 | ProcessingJob.MarkSucceeded |
| S-20 | ProcessingJob.MarkFailed |
| S-21 | 非法狀態轉移（ErrInvalidStateTransition） |

---

## 6. 檔案結構規劃

```
dobby/
├── main.go
├── app.go
├── wails.json
│
├── internal/
│   ├── domain/
│   │   ├── rule/
│   │   │   ├── rule.go                    ← Rule aggregate（S-1~S-5）
│   │   │   ├── value_objects.go           ← RuleId, WatchConfig, FilterSpec,
│   │   │   │                                NamingTemplate, TargetPathTemplate
│   │   │   ├── events.go                  ← RuleCreated, RuleUpdated...
│   │   │   └── repository.go              ← IRuleRepository interface
│   │   │
│   │   ├── job/
│   │   │   ├── job.go                     ← ProcessingJob aggregate（S-16~S-21）
│   │   │   ├── value_objects.go           ← JobId, FileEvent, ProcessingContext,
│   │   │   │                                ProcessingResult, JobState
│   │   │   ├── events.go                  ← FileDetected, RuleMatched...
│   │   │   └── repository.go              ← IProcessingJobRepository interface
│   │   │
│   │   └── service/
│   │       ├── rule_matcher.go            ← RuleMatcher（S-6~S-10）
│   │       ├── template_renderer.go       ← TemplateRenderer（S-11~S-13）
│   │       └── sequence_generator.go      ← SequenceGenerator（S-14~S-15）
│   │
│   ├── application/
│   │   ├── rule_service.go
│   │   └── job_service.go
│   │
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   ├── db.go
│   │   │   ├── rule_repository.go
│   │   │   ├── job_repository.go
│   │   │   └── operation_log_repository.go
│   │   ├── watcher/
│   │   │   └── file_watcher.go
│   │   ├── filesystem/
│   │   │   └── file_system_service.go
│   │   └── notification/
│   │       └── notification_service.go
│   │
│   └── query/
│       └── operation_log_query.go
│
└── internal/domain/service/  ← 測試目錄（同套件）
    ├── rule_matcher_test.go
    ├── template_renderer_test.go
    └── sequence_generator_test.go
```

---

## 7. 測試策略

| 層 | 方式 |
|----|------|
| `domain/rule/` | Pure unit test（無外部依賴） |
| `domain/job/` | Pure unit test（狀態機） |
| `domain/service/` | Unit test（RuleMatcher、TemplateRenderer 純函數；SequenceGenerator mock IOperationLogRepository） |
| `application/` | Mock repository interface，unit test use case |
| `infrastructure/persistence/` | Integration test（in-memory SQLite） |
