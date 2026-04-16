# Dobby — 檔案管家

自動將新進檔案依照使用者定義的規則重新命名並移動至指定資料夾的桌面應用程式。

---

## 技術棧

| 層 | 技術 |
|----|------|
| 桌面框架 | [Wails v2](https://wails.io)（Go 後端 + 原生 WebView） |
| 後端語言 | Go 1.21+ |
| 架構模式 | Tactical DDD（Aggregate、Repository、Domain Service） |
| 資料庫 | SQLite via `modernc.org/sqlite`（pure Go，不需 CGO） |
| 前端 | React + Vite（TypeScript）— 待建立 |
| 檔案監控 | Go `fsnotify`（待實作） |

---

## 前置需求

### 必要

| 工具 | 版本 | 安裝方式 |
|------|------|---------|
| Go | 1.21+ | https://go.dev/dl/ |
| Wails CLI | v2.x | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| Node.js | 18+ | https://nodejs.org |
| npm | 9+ | 隨 Node.js 附帶 |

### Windows 額外需求

- **WebView2 Runtime**：Windows 11 已內建；Windows 10 請至 [Microsoft 官網](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) 下載安裝。
- **Build Tools**：執行 `wails doctor` 檢查環境是否齊全。

### macOS 額外需求

- Xcode Command Line Tools：`xcode-select --install`

---

## 本地開發

### 1. Clone 專案

```bash
git clone <repo-url>
cd Dobby-Files
```

### 2. 安裝 Go 依賴

```bash
cd dobby
go mod download
```

### 3. 確認環境

```bash
wails doctor
```

所有項目應顯示 ✅。若有缺少，依照輸出提示安裝對應工具。

### 4. 建立前端

> 目前 `frontend/dist/` 僅有 placeholder。  
> 前端尚未實作時，可先跳過此步驟，後端邏輯仍可正常開發與測試。

前端就緒後，在 `dobby/` 下執行：

```bash
cd frontend
npm install
npm run build   # 產出 frontend/dist/
```

### 5. 啟動開發模式

```bash
cd dobby
wails dev
```

Wails 會：
- 編譯 Go 後端
- 啟動前端 Vite dev server（hot reload）
- 開啟原生視窗，透過 WebView 載入前端
- 自動產生 `frontend/src/wailsjs/` TypeScript bindings

若前端尚未建立，想單純跑後端邏輯（CLI 模式），可直接：

```bash
go run .
```

> 注意：`go run .` 會嘗試開啟 Wails 視窗，需要 WebView2 Runtime。純後端測試請使用單元測試（見下方）。

---

## 執行測試

### 單元測試（無外部依賴）

```bash
cd dobby
go test ./...
```

預期輸出：

```
ok  github.com/dobby/filemanager/internal/domain/job      (29 tests)
ok  github.com/dobby/filemanager/internal/domain/rule
ok  github.com/dobby/filemanager/internal/domain/service
```

### 指定套件

```bash
go test ./internal/domain/...         # 所有 domain 測試
go test ./internal/domain/rule/... -v # 詳細輸出
```

---

## 建置正式版本

```bash
cd dobby
wails build
```

產出位置：`dobby/build/bin/dobby`（Windows 為 `dobby.exe`）

---

## 資料庫位置

SQLite 資料庫在應用程式**首次啟動時自動建立**，位於：

| 平台 | 路徑 |
|------|------|
| Windows | `%USERPROFILE%\.dobby\dobby.db` |
| macOS / Linux | `~/.dobby/dobby.db` |

Schema migration 為 **code-first**，由 `internal/infrastructure/persistence/db.go` 在啟動時自動套用，無需手動執行 SQL。

---

## 專案結構

```
Dobby-Files/
├── dobby/                              # Go 專案根目錄
│   ├── main.go                         # Wails 進入點、DI 組裝
│   ├── app.go                          # Wails binding 薄殼（委派給 application layer）
│   ├── wails.json                      # Wails 專案設定
│   ├── frontend/
│   │   └── dist/                       # 編譯後的前端（由 npm run build 產生）
│   └── internal/
│       ├── domain/
│       │   ├── rule/                   # Rule aggregate、value objects、repository interface
│       │   ├── job/                    # ProcessingJob 狀態機
│       │   └── service/               # RuleMatcher、TemplateRenderer、SequenceGenerator
│       ├── application/
│       │   ├── rule_service.go         # Rule CRUD use cases
│       │   └── log_service.go          # Operation log read use cases
│       ├── infrastructure/
│       │   └── persistence/           # SQLite repositories + code-first migrations
│       └── query/
│           └── operation_log_query.go  # Operation log read model
├── features/
│   └── file-manager.feature           # Gherkin 行為規格（21 scenarios）
└── docs/
    └── features/file-manager/
        ├── architecture.md            # 架構設計文件
        └── conclusion.md             # SDD 驗證報告
```

### 依賴方向

```
frontend  (React)
    ↓  Wails JS binding（自動產生）
app.go    (薄殼)
    ↓
application/
    ↓
domain/  ←──  infrastructure/
              （實作 domain interfaces）
query/   ←──  infrastructure/persistence/
              （共用 db 連線）
```

**規則：外層依賴內層，內層不知道外層存在。**

---

## 開發流程

### 新增規則 use case

1. 在 `internal/domain/rule/` 定義或修改 domain model
2. 在 `internal/application/rule_service.go` 加入 use case 方法
3. 在 `app.go` 加入對應的 Wails binding method
4. 前端呼叫 `window.go.main.App.<MethodName>()`（Wails 自動產生 TypeScript binding）

### 新增資料庫欄位

在 `internal/infrastructure/persistence/db.go` 的 `migrations` slice 尾端**新增**一條 DDL（不修改既有條目），重新啟動後自動套用。

---

## 常見問題

**Q: `wails doctor` 顯示 WebView2 未安裝**  
A: Windows 10 使用者請至 [此連結](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) 下載 Evergreen Bootstrapper 並安裝。

**Q: `go build` 報錯找不到 `modernc.org/sqlite`**  
A: 執行 `go mod download` 重新下載依賴。

**Q: 想直接查看資料庫內容**  
A: 使用任意 SQLite GUI 工具（如 [DB Browser for SQLite](https://sqlitebrowser.org/)）開啟 `~/.dobby/dobby.db`。
