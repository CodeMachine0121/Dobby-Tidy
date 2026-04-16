# 檔案管家 — Tactical DDD Domain Design

**日期：** 2026-04-16
**版本：** v1.0
**關聯規格：** `2026-04-16-file-manager-design.md`

---

## 1. 設計原則

- **兩個核心業務關切點同等重要**：規則設定（Rule）與檔案處理（ProcessingJob）
- **同一份 Rule 模型**：UI 與背景服務共用，不分離 Read/Write 模型
- **有狀態的處理流程**：ProcessingJob 保留執行過程的狀態流轉
- **OperationLog 為 Read Model**：從 ProcessingJob 完成事件投影而來，不是獨立 Aggregate

---

## 2. Domain Invariants

| 約束 | 說明 |
|------|------|
| `Rule.WatchConfig.folderPath` 全域唯一 | 一個資料夾只能對應一條規則 |
| Repository 在新增/更新 Rule 前必須檢查 `folderPath` 唯一性 | 由 Application Service 層強制 |

---

## 3. Aggregates

### 3.1 Rule（設定層）

```
Rule (Aggregate Root)
│
├── RuleId              (Value Object)
├── name: string
├── enabled: bool
│
├── WatchConfig         (Value Object)
│     folderPath: string   ← Domain Invariant: 全域唯一
│     recursive: bool
│
├── FilterSpec          (Value Object)
│     extensions: string[]  (e.g. [".png", ".jpg"])
│     keyword?: string       (原始檔名包含此字)
│
├── NamingTemplate      (Value Object)
│     templateString: string
│     render(ctx: ProcessingContext) → string
│
├── TargetPathTemplate  (Value Object)
│     pathTemplate: string
│     render(ctx: ProcessingContext) → string
│
├── project: string     (供 {project} 變數使用)
└── typeLabel: string   (供 {type} 變數使用)
```

**Domain Events：**
- `RuleCreated`
- `RuleUpdated`
- `RuleDeleted`
- `RuleEnabled`
- `RuleDisabled`

---

### 3.2 ProcessingJob（執行層）

```
ProcessingJob (Aggregate Root)
│
├── JobId               (Value Object)
├── ruleId: RuleId      (reference，不 embed Rule)
│
├── FileEvent           (Value Object)
│     detectedPath: string
│     originalName: string
│     extension: string
│     detectedAt: time
│
├── state: JobState     (enum)
│     pending → matched → processing → succeeded
│                                    ↘ failed
│
├── ProcessingContext   (Value Object，state = matched 後建立)
│     project: string
│     typeLabel: string
│     date: time
│     seq: string       (e.g. "001")
│     originalName: string
│     extension: string
│
└── ProcessingResult    (Value Object，state = succeeded/failed 後填入)
      newPath?: string
      errorMessage?: string
      processedAt?: time
```

**State Transitions：**
```
pending
  → matched      (RuleMatcher 找到對應 Rule，建立 ProcessingContext)
  → processing   (FileSystemService 開始執行)
  → succeeded    (rename + move 成功)
  → failed       (任何步驟失敗)
```

**Domain Events：**
- `FileDetected`
- `RuleMatched`
- `ProcessingSucceeded`
- `ProcessingFailed`

---

## 4. Read Model

### OperationLog

ProcessingJob 完成後的投影（Projection），不是獨立 Aggregate。

```
OperationLog
  logId         唯一識別碼
  ruleId        套用的規則 ID
  ruleName      規則名稱（快照，避免 Rule 刪除後資料遺失）
  originalPath  原始檔案完整路徑
  newPath       處理後檔案完整路徑
  status        success / error
  errorMessage  錯誤訊息（若失敗）
  processedAt   處理時間
```

由 `ProcessingSucceeded` / `ProcessingFailed` 事件寫入，供 UI 查詢歷史紀錄。

---

## 5. Domain Services

### 5.1 RuleMatcher
```
input:  FileEvent + Rule
output: bool

邏輯（AND 條件）：
  1. FileEvent.extension ∈ Rule.FilterSpec.extensions
  2. FileEvent.originalName contains Rule.FilterSpec.keyword（若有設定）
```

> **注意：** 原規格書的 `source_folder` filter 條件在此設計中已由 `WatchConfig.folderPath` 取代。
> 背景服務以 `folderPath` 為索引鍵查詢 Rule（1 folder = 1 rule），
> 因此進入 RuleMatcher 時來源資料夾已確定符合，無需再次驗證。

### 5.2 TemplateRenderer
```
input:  NamingTemplate (或 TargetPathTemplate) + ProcessingContext
output: string（渲染後的檔名或路徑）

負責將 {project}, {YYYY}, {MM}, {DD}, {seq}, {original}, {ext}, {type}
等變數替換為 ProcessingContext 中的實際值
```

### 5.3 SequenceGenerator
```
input:  ruleId + date
output: string (e.g. "001")

從 OperationLog 計算當天該 rule 已成功處理的筆數，加一後格式化為三位數字串
```

---

## 6. Infrastructure（Domain 外部）

| 元件 | 職責 |
|------|------|
| `FileWatcher` | 封裝 fsnotify，偵測到新檔案後觸發 ProcessingJob 建立 |
| `FileSystemService` | 實際執行 rename / move 操作 |
| `NotificationService` | 監聽 `ProcessingSucceeded` event，發送桌面通知 |
| `RuleRepository` | SQLite CRUD for Rule（含 folderPath 唯一性查詢） |
| `ProcessingJobRepository` | SQLite CRUD for ProcessingJob（短期狀態保存） |
| `OperationLogRepository` | SQLite Read Model，供 UI 查詢歷史紀錄（保留最近 1000 筆） |

---

## 7. 背景服務執行流程

```
FileWatcher 偵測到新檔案
  → 建立 ProcessingJob (state: pending)
  → RuleRepository 依 folderPath 取得對應 Rule
  → RuleMatcher 確認 FilterSpec 條件符合
  → JobState → matched，SequenceGenerator 產生 seq
  → TemplateRenderer 渲染 newName + targetPath
  → ProcessingContext 建立完成
  → JobState → processing
  → FileSystemService 執行 rename + move
  → JobState → succeeded / failed
  → OperationLog 寫入 (Read Model)
  → NotificationService 發送桌面通知
```

---

## 8. SQLite 資料表對應

| Aggregate / Read Model | 資料表 |
|------------------------|--------|
| Rule | `rules` |
| ProcessingJob | `processing_jobs`（短期，可定期清除） |
| OperationLog | `operation_logs`（長期，最近 1000 筆） |
