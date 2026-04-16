# Background Processor — 架構設計

## 專案上下文

- **技術棧**：Go 1.26、Wails v2、SQLite（modernc）
- **架構模式**：DDD Layered Architecture（Domain / Application / Infrastructure）
- **模組路徑**：`github.com/dobby/filemanager`
- **命名慣例**：
  - 介面以 `I` 前綴（`IFileSystem`、`IRuleRepository`）
  - Application service 放 `internal/application/`
  - Infrastructure 實作放 `internal/infrastructure/`
  - 測試使用 `testify`（assert + mock）

---

## 資料模型

既有模型（不新增，直接使用）：

| 模型 | 套件 | 說明 | Gherkin 來源 |
|------|------|------|--------------|
| `ProcessingJob` | `domain/job` | 追蹤單次檔案處理生命週期 | S-1, S-4, S-7, S-9 |
| `FileEvent` | `domain/job` | 被偵測到的檔案資訊 | S-1, S-7, S-9 |
| `ProcessingContext` | `domain/job` | 渲染模板所需的變數集合 | S-9 |
| `Rule` | `domain/rule` | 規則聚合根 | S-1 ~ S-8 |
| `OperationLog`（query） | `query` | 操作日誌讀模型 | S-1, S-4 |

---

## 服務介面

### 新增 Port：`IFileSystem`

```go
// internal/application/background_processor.go（或新建 port 檔案）
type IFileSystem interface {
    // ListFiles 列出指定目錄下的所有檔案（非目錄）
    // recursive=true 時遞迴掃描子目錄
    ListFiles(ctx context.Context, dir string, recursive bool) ([]FileInfo, error)

    // MoveFile 將 src 搬移（含重新命名）至 dst
    MoveFile(ctx context.Context, src, dst string) error

    // EnsureDir 確保目標目錄存在（等同 os.MkdirAll）
    EnsureDir(ctx context.Context, dir string) error
}

type FileInfo struct {
    Path      string    // 完整路徑
    Name      string    // 不含副檔名的檔名
    Extension string    // 含點，如 ".pdf"
    DetectedAt time.Time
}
```

> Gherkin 來源：S-1（file move）、S-4（move failure）、S-7（recursive）、S-8（non-recursive）

### 新增 Application Service：`BackgroundProcessorService`

```go
// internal/application/background_processor.go
type BackgroundProcessorService struct {
    ruleRepo      rule.IRuleRepository
    jobRepo       job.IProcessingJobRepository
    logRepo       IOperationLogWriter          // 寫端（新介面，見下方）
    fileSystem    IFileSystem
    ruleMatcher   *domainservice.RuleMatcher
    renderer      *domainservice.TemplateRenderer
    seqGen        *domainservice.SequenceGenerator
    running       atomic.Bool                  // 防止並發掃描（S-6）
}

// ScanAndProcess 掃描所有啟用規則的監控目錄，處理符合條件的檔案
// 若已有掃描在執行中則立即返回 nil（S-6）
func (s *BackgroundProcessorService) ScanAndProcess(ctx context.Context) error
```

> Gherkin 來源：S-1 ~ S-9（核心動詞）

### 新增 Port：`IOperationLogWriter`

```go
type IOperationLogWriter interface {
    Save(ctx context.Context, log *query.OperationLog) error
}
```

> 說明：`SQLiteOperationLogRepository` 已實作此方法，只需定義介面讓 application layer 依賴
> Gherkin 來源：S-1（log written）、S-4（error log）

---

## 架構決策

1. **並發保護用 `sync/atomic.Bool`**（S-6）：輕量、無鎖、無需額外依賴
2. **`IFileSystem` 抽象**（S-4）：讓單元測試能 mock 檔案搬移失敗情境
3. **ScanAndProcess 設計為無狀態單次呼叫**：呼叫端（Wails `startup` goroutine 或 ticker）決定排程頻率，service 本身不持有 ticker
4. **`IOperationLogWriter` 從 query 套件分離讀寫職責**：application layer 只需注入寫端介面
5. **ProcessingContext 在 application layer 組裝**（S-9）：由 `BackgroundProcessorService` 呼叫 `SequenceGenerator` 後建立

---

## 情境對應

| Scenario | 觸發路徑 | 關鍵邏輯 |
|----------|----------|----------|
| S-1 | `ScanAndProcess` → match → move → log success | 完整快樂路徑 |
| S-2 | `ScanAndProcess` → RuleMatcher 不符副檔名 → 跳過 | 過濾邏輯 |
| S-3 | `ScanAndProcess` → RuleMatcher 不符 keyword → 跳過 | 過濾邏輯 |
| S-4 | `ScanAndProcess` → move 失敗 → MarkFailed → log error | 錯誤處理 |
| S-5 | `ScanAndProcess` → ListAll 回傳空/全停用 → 直接返回 | 邊界條件 |
| S-6 | `ScanAndProcess` → `running.CompareAndSwap(false,true)` 失敗 → 返回 nil | 並發保護 |
| S-7 | `IFileSystem.ListFiles(recursive=true)` | 遞迴掃描 |
| S-8 | `IFileSystem.ListFiles(recursive=false)` | 非遞迴掃描 |
| S-9 | `SequenceGenerator.Generate` → `ProcessingContext{...}` | Context 組裝 |

---

## 檔案結構規劃

```
internal/
├── application/
│   └── background_processor.go          # BackgroundProcessorService + ports
├── infrastructure/
│   └── filesystem/
│       └── os_filesystem.go             # IFileSystem 的 OS 實作
```

測試檔案：
```
internal/application/background_processor_test.go
```
