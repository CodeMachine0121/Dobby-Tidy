# 架構設計：Monetization — 試用期與授權啟用

## 專案上下文

- **技術棧：** Go 1.26 + Wails v2 + React（TypeScript）
- **架構模式：** DDD — domain / application / infrastructure 三層
- **模組路徑：** `github.com/dobby/filemanager`
- **測試框架：** `testify`（assert + mock）
- **既有慣例：**
  - Domain 聚合根暴露 accessor，不公開欄位
  - Port 介面定義在使用者那層（application 定義 infra port）
  - 命名：`I<Name>` 為 interface，`New<Name>` 為建構子，`Reconstitute` 為 repository 重建

---

## 資料模型

### `License`（Domain Aggregate Root） — `domain/license/license.go`
> 來源：S-1（記錄試用日期）、S-2/S-3（狀態查詢）、S-5/S-8（啟用）

| 欄位 | 型別 | 說明 |
|------|------|------|
| `trialStartedAt` | `time.Time` | 首次啟動時間（混淆儲存） |
| `activatedKey` | `string` | 啟用的 license key，空字串代表未啟用 |
| `machineID` | `string` | 綁定的機器 ID（啟用時記錄） |

**行為方法：**
- `IsActivated() bool`
- `IsTrialActive(now time.Time) bool` — S-2/S-3
- `CanRun(now time.Time) bool` — S-2/S-3/S-4/S-8
- `Status(now time.Time) LicenseStatus` — S-2/S-3/S-5/S-8
- `DaysRemaining(now time.Time) int` — S-1
- `Activate(key, machineID string) error` — S-5

### `LicenseStatus`（Value Object） — `domain/license/value_objects.go`
> 來源：S-1/S-2/S-3/S-5/S-8

```
"active"    // 試用期內
"expired"   // 試用到期，未啟用
"activated" // 已付費啟用
```

---

## 服務介面

### Domain Port：`ILicenseRepository` — `domain/license/repository.go`
> 來源：S-1（儲存首次啟動）、S-2/S-3/S-5（讀取）

```go
Load(ctx) (*License, bool, error)  // bool = found
Save(ctx, *License) error
```

### Domain Service：`LicenseKeyValidator` — `domain/license/license_key_validator.go`
> 來源：S-5（有效 key）、S-6（checksum 錯誤）、S-7（格式錯誤）

- `Validate(key string) error`
- `GenerateKey(p1, p2 string) string`（工具用途）

**Key 格式：** `DOBBY-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}`
**校驗邏輯：** `HMAC-SHA256(appSecret, P1+P2)` 前 4 bytes hex uppercase = 第 4 段

### Application Port：`IMachineIdProvider` — `application/license_service.go`
> 來源：S-5（key 綁定 machine）

```go
MachineID() (string, error)
```

### Application Service：`LicenseService` — `application/license_service.go`
> 來源：S-1/S-2/S-3/S-4/S-5/S-6/S-7/S-8

```go
InitializeTrial(ctx) error                        // S-1
GetLicenseInfo(ctx) (*LicenseInfo, error)         // S-1/S-2/S-3/S-8
ActivateLicense(ctx, key string) error            // S-5/S-6/S-7
CanRunBackgroundProcessor(ctx) (bool, error)      // S-2/S-3/S-4/S-8
```

### Application Port：`ILicenseGuard` — `application/background_processor.go`
> 來源：S-4（攔截 processor）

```go
CanRunBackgroundProcessor(ctx) (bool, error)
```

注入 `BackgroundProcessorService`，`ScanAndProcess` 開頭檢查授權。

---

## 架構決策

| 決策 | 說明 |
|------|------|
| Domain port 在 domain 層 | `ILicenseRepository` 放 `domain/license`，符合既有 `IRuleRepository` 慣例 |
| Machine ID port 在 application 層 | 屬於 infrastructure 細節，domain 不需知道 |
| LicenseGuard 用 interface 注入 | 避免 BackgroundProcessorService 直接依賴 LicenseService，保持可測試性 |
| 混淆儲存 = XOR + Base64 | MVP 只需防止「直接刪檔重試」，不需強加密 |
| 無後端驗證 | MVP 階段，key 格式 + HMAC 即可驗證正常購買流程 |
| build tag `windows` for MachineId | 使用 Windows Registry `MachineGuid`，加上 `!windows` fallback 確保測試可執行 |

---

## 情境對應

| Scenario | 觸發路徑 |
|----------|----------|
| S-1 | `LicenseService.InitializeTrial` → `ILicenseRepository.Load`（not found）→ `NewLicense` → `Save` |
| S-2 | `LicenseService.CanRunBackgroundProcessor` → `License.CanRun` = true |
| S-3 | `LicenseService.CanRunBackgroundProcessor` → `License.CanRun` = false |
| S-4 | `BackgroundProcessorService.ScanAndProcess` → `ILicenseGuard.CanRunBackgroundProcessor` = false → 直接 return |
| S-5 | `LicenseService.ActivateLicense` → `Validate` → `Activate` → `Save` |
| S-6 | `LicenseService.ActivateLicense` → `Validate` → `ErrInvalidLicenseKey` |
| S-7 | `LicenseService.ActivateLicense` → `Validate` → `ErrInvalidLicenseKeyFormat` |
| S-8 | `LicenseService.CanRunBackgroundProcessor` → `License.IsActivated()` = true → `CanRun` = true |

---

## 檔案結構規劃

```
features/
└── monetization.feature

docs/features/monetization/
├── architecture.md
└── conclusion.md

dobby/internal/domain/license/
├── license.go                     # License aggregate root
├── value_objects.go               # LicenseStatus + errors
├── repository.go                  # ILicenseRepository port
└── license_key_validator.go       # domain service

dobby/internal/application/
├── license_service.go             # LicenseService + LicenseInfo + ports
└── license_service_test.go        # unit tests (S-1~S-8)

dobby/internal/infrastructure/license/
├── local_license_repository.go    # XOR+base64 obfuscated file store
├── machine_id_windows.go          # Windows Registry MachineGuid (build: windows)
└── machine_id_other.go            # stub fallback (build: !windows)

修改：
- dobby/internal/application/background_processor.go  （加 ILicenseGuard 欄位與檢查）
- dobby/main.go                                        （wire LicenseService）
- dobby/app.go                                         （加 GetLicenseInfo / ActivateLicense binding）
```
