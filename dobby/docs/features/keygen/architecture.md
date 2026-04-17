# Keygen CLI — 架構設計

## 專案上下文

- **語言：** Go 1.26
- **模組：** `github.com/dobby/filemanager`
- **架構模式：** Domain-driven，domain layer 不依賴 infra
- **測試框架：** `testing` + `testify/assert` + `testify/mock`
- **測試目錄規範：** 與被測套件同目錄，`_test.go` 後綴，`package xxx_test`

## 資料模型

### KeygenConfig（來自 S-1, S-2, S-3, S-6, S-7）
```
KeygenConfig {
  Count  int     // -n 旗標；預設 10000
  Output string  // -o 旗標；空字串代表 stdout
}
```

> 純輸入參數結構，無需持久化。

## 服務介面

### KeyGenerator（來自 S-1 ~ S-5）

**職責：** 接收 count，產生指定數量的唯一、有效 DOBBY license keys

```go
// internal/domain/license（已存在，無需新增）
// LicenseKeyValidator.GenerateKey(p1, p2 string) string
// LicenseKeyValidator.Validate(key string) error
```

KeyGenerator 為 `cmd/keygen` 內的私有邏輯，直接呼叫現有 `LicenseKeyValidator`，不新增 domain 層介面。

### RandSegment（來自 S-4, S-5）

**職責：** 用 `crypto/rand` 產生 4 字元 A-Z0-9 隨機字串，作為 p1/p2 輸入

## 架構決策

1. **零 domain 修改：** 所有新邏輯限於 `cmd/keygen/`，domain layer 保持不變。
2. **crypto/rand：** 使用 `crypto/rand` 而非 `math/rand`，確保統計分佈均勻，降低碰撞機率。
3. **in-memory 去重：** `map[string]struct{}` 追蹤已產生的 keys；10000 筆佔用記憶體可忽略。
4. **stderr 錯誤：** 所有錯誤訊息寫入 stderr，key 輸出寫入 stdout 或檔案，兩者不混。

## 情境對應

| Scenario | 對應邏輯 |
|----------|---------|
| S-1 | Count 預設 10000，Output 空 → stdout |
| S-2 | Count = 50 |
| S-3 | Output = "keys.txt" → os.Create |
| S-4 | map 去重邏輯 |
| S-5 | validator.Validate(key) == nil |
| S-6 | Count = 0 → 空迴圈，正常結束 |
| S-7 | Count < 0 → 印 stderr，os.Exit(1) |

## 檔案結構規劃

```
cmd/
└── keygen/
    ├── main.go               # CLI 入口、flag 解析、輸出邏輯
    └── main_test.go          # 單元測試
```

新增單一目錄，不動任何現有檔案。
