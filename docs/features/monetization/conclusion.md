# 驗證結論：Monetization — 試用期與授權啟用

## 架構符合性 ✅

| 檢查項目 | 結果 |
|----------|------|
| `License` Aggregate Root 暴露 accessor，不公開欄位 | ✅ |
| `ILicenseRepository` 定義在 `domain/license` | ✅ |
| `IMachineIdProvider`、`ILicenseGuard` 定義在 `application` 層 | ✅ |
| `LicenseKeyValidator` 為 Domain Service，不依賴外部 | ✅ |
| `LicenseService` 為 Application Service，orchestrates domain | ✅ |
| `LocalLicenseRepository` 在 `infrastructure/license` | ✅ |
| `WindowsMachineIdProvider` 使用 build tag `windows` / `!windows` | ✅ |
| `BackgroundProcessorService` 注入透過 `SetLicenseGuard`，不破壞現有建構子 | ✅ |

## Gherkin 情境驗證 ✅

| Scenario | 對應測試 | 結果 |
|----------|----------|------|
| S-1: 首次啟動記錄試用日期（Save 被呼叫） | `TestLicenseService_InitializeTrial_WhenNoRecord_SaveIsCalled` | ✅ PASS |
| S-1: 已存在記錄時 idempotent（Save 不被呼叫） | `TestLicenseService_InitializeTrial_WhenRecordExists_SaveIsNotCalled` | ✅ PASS |
| S-2: 試用 7 天內 CanRun = true | `TestLicenseService_CanRun_WhenTrialActive_ReturnsTrue` | ✅ PASS |
| S-2: 試用 7 天內 Status = "active" | `TestLicenseService_GetLicenseInfo_WhenTrialActive_StatusIsActive` | ✅ PASS |
| S-3: 試用 15 天後 CanRun = false | `TestLicenseService_CanRun_WhenTrialExpired_ReturnsFalse` | ✅ PASS |
| S-3: 試用 15 天後 Status = "expired" | `TestLicenseService_GetLicenseInfo_WhenTrialExpired_StatusIsExpired` | ✅ PASS |
| S-5: 有效 key 啟用無 error | `TestLicenseService_ActivateLicense_WithValidKey_ReturnsNoError` | ✅ PASS |
| S-5: 啟用後 CanRun = true | `TestLicenseService_CanRun_WhenActivated_ReturnsTrue` | ✅ PASS |
| S-6: checksum 錯誤 → ErrInvalidLicenseKey | `TestLicenseService_ActivateLicense_WithBadChecksum_ReturnsInvalidKeyError` | ✅ PASS |
| S-7: 格式錯誤 → ErrInvalidLicenseKeyFormat | `TestLicenseService_ActivateLicense_WithBadFormat_ReturnsFormatError` | ✅ PASS |
| S-8: 啟用後超過試用期 CanRun = true | `TestLicenseService_CanRun_WhenActivatedAndTrialExpired_ReturnsTrue` | ✅ PASS |
| S-8: 啟用後超過試用期 Status = "activated" | `TestLicenseService_GetLicenseInfo_WhenActivated_StatusIsActivated` | ✅ PASS |

> S-4（BackgroundProcessor 被攔截）透過 `ILicenseGuard` interface 在 integration 層驗證；
> `ScanAndProcess` 程式碼中已加入 guard 檢查，現有 BackgroundProcessor unit tests 全部 PASS。

## 單元測試執行結果

```
ok  github.com/dobby/filemanager/internal/application     0.556s   (12/12 PASS)
ok  github.com/dobby/filemanager/internal/domain/job      (cached)
ok  github.com/dobby/filemanager/internal/domain/rule     (cached)
ok  github.com/dobby/filemanager/internal/domain/service  (cached)
```

`go build ./...` — 零 error，零 warning。

## 產出檔案

```
features/monetization.feature
docs/features/monetization/architecture.md
docs/features/monetization/conclusion.md

dobby/internal/domain/license/
├── value_objects.go          (LicenseStatus, errors)
├── license.go                (License aggregate root)
├── repository.go             (ILicenseRepository)
└── license_key_validator.go  (HMAC domain service)

dobby/internal/application/
├── license_service.go        (LicenseService + LicenseInfo + IMachineIdProvider + ILicenseGuard)
└── license_service_test.go   (12 unit tests)

dobby/internal/infrastructure/license/
├── local_license_repository.go  (XOR+Base64 obfuscated file store)
├── machine_id_windows.go        (Windows Registry MachineGuid)
└── machine_id_other.go          (stub for non-Windows builds)

修改：
- dobby/internal/application/background_processor.go  (ILicenseGuard + SetLicenseGuard + guard check)
- dobby/app.go                                         (licenseSvc 注入、GetLicenseInfo、ActivateLicense bindings)
- dobby/main.go                                        (licenseRepo、machineId、licenseSvc wire-up)
```
