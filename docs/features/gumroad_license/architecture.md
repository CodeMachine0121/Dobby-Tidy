# Gumroad License Verification — 架構設計

## 專案上下文

- **語言：** Go 1.26
- **模組：** `github.com/dobby/filemanager`
- **架構模式：** Domain-driven；application layer 定義 port 介面，infrastructure 實作
- **測試框架：** `testing` + `testify/assert` + `testify/mock`
- **測試目錄：** 與被測套件同目錄，`package xxx_test`

## 資料模型

### License（已存在，無修改）
`internal/domain/license/license.go`

欄位不變；`Activate(key, machineID string)` 接受任意字串，與 key 格式無關。

### GumroadVerifyResponse（新增，infra 私有）
`internal/infrastructure/license/gumroad_verifier.go`（私有 struct）

```
gumroadVerifyResponse {
  Success bool        // "success" JSON 欄位（來自 S-1, S-2, S-3）
  Uses    int         // "uses" JSON 欄位（來自 S-1, S-3）
}
```

> 來源：S-1 line 4-5、S-2 line 9、S-3 line 14

## 服務介面

### IGumroadVerifier（新增，application layer port）
`internal/application/license_service.go`（來自 S-1 ~ S-4）

```go
type IGumroadVerifier interface {
    Verify(ctx context.Context, licenseKey string) error
}
```

**語意：**
- `nil` → 驗證通過（success=true, uses=1）
- `ErrInvalidLicenseKey` → Gumroad 回傳 success=false（S-2）
- `ErrLicenseAlreadyUsed` → uses > 1（S-3）
- 網路錯誤 → wrapped error with 中文訊息（S-4）

> 來源：S-1 line 4、S-2 line 9、S-3 line 14、S-4 line 19

### LicenseService.ActivateLicense（修改）
`internal/application/license_service.go`（來自 S-1 ~ S-5）

移除 `validator *license.LicenseKeyValidator`，改注入 `verifier IGumroadVerifier`。
移除 `strings.ToUpper`（Gumroad key 大小寫由 API 決定）。

新流程：
1. 若 license 已啟用 → 回傳 `ErrAlreadyActivated`，不呼叫 API（S-5）
2. `verifier.Verify(ctx, key)` → 失敗則回傳錯誤（S-2, S-3, S-4）
3. `machineId.MachineID()` → 取得機器 ID
4. `l.Activate(key, machineID)` → 啟用
5. `repo.Save(l)` → 持久化（S-1）

> 來源：S-1 line 3-6、S-5 line 24-26

## 架構決策

1. **IGumroadVerifier 放在 application layer**：application 定義需求，infra 滿足，符合 DIP。
2. **key 不做本地格式驗證**：格式驗證完全交給 Gumroad API，減少重複邏輯。
3. **key 保持原始大小寫**：不強制 ToUpper，Gumroad API 已做 case-insensitive 處理。
4. **ErrAlreadyActivated 在呼叫 verifier 前檢查**：避免對已啟用的機器重複消耗 API 的 uses 計數。

## 情境對應

| Scenario | 對應邏輯 |
|----------|---------|
| S-1 | verifier 成功 → Activate → Save |
| S-2 | verifier 回傳 ErrInvalidLicenseKey |
| S-3 | verifier 回傳 ErrLicenseAlreadyUsed |
| S-4 | verifier 回傳 network error |
| S-5 | Load 發現已啟用 → ErrAlreadyActivated，不呼叫 verifier |
| S-6 | GetLicenseInfo 讀本機，不涉及 verifier |

## 移除項目

| 路徑 | 原因 |
|------|------|
| `internal/domain/license/license_key_validator.go` | HMAC 邏輯不再需要 |
| `cmd/keygen/` | 預產 key 方案已廢棄 |

## Value Objects 變更（`value_objects.go`）

移除：
- `ErrInvalidLicenseKeyFormat`

保留/更新：
- `ErrAlreadyActivated`（不變）
- `ErrInvalidLicenseKey`（保留名稱，語意改為「Gumroad 拒絕此 key」）

新增：
- `ErrLicenseAlreadyUsed`（key 已在其他機器使用）

## 檔案結構規劃

```
internal/
├── application/
│   └── license_service.go          # 修改：移除 validator，注入 verifier
├── domain/license/
│   ├── license.go                  # 不變
│   ├── license_key_validator.go    # 刪除
│   ├── repository.go               # 不變
│   └── value_objects.go            # 修改：移除 ErrInvalidLicenseKeyFormat，新增 ErrLicenseAlreadyUsed
└── infrastructure/license/
    ├── gumroad_verifier.go         # 新增
    ├── local_license_repository.go # 不變
    ├── machine_id_windows.go       # 不變
    └── machine_id_other.go         # 不變

cmd/keygen/                         # 刪除整個目錄
```
