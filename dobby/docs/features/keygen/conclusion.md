# Keygen CLI — 驗證結論

**日期：** 2026-04-17  
**結果：** PASS

## 架構符合性

| 項目 | 結果 |
|------|------|
| 零 domain 修改 | ✓ |
| 新邏輯限於 `cmd/keygen/` | ✓ |
| 複用 `LicenseKeyValidator.GenerateKey` | ✓ |
| 錯誤寫入 stderr，輸出寫入 stdout/檔案 | ✓ |

## Gherkin 情境驗證

| Scenario | 結果 | 驗證方式 |
|----------|------|---------|
| S-1 預設 10000 行到 stdout | ✓ | 單元測試 + 手動執行 |
| S-2 指定 -n 50 | ✓ | 單元測試 `TestGenerateKeys_ReturnsExactCount` |
| S-3 輸出到檔案 | ✓ | `go run -n 3 -o /tmp/keys_test.txt` → 3 行 |
| S-4 同批次不重複 | ✓ | 單元測試 `TestGenerateKeys_AllUnique` |
| S-5 每個 key 通過 HMAC | ✓ | 單元測試 `TestGenerateKeys_EachKeyPassesHMAC` |
| S-6 count=0 空輸出 | ✓ | 單元測試 `TestGenerateKeys_ZeroCount_ReturnsEmptySlice` |
| S-7 count<0 非零退出 | ✓ | 單元測試 `TestGenerateKeys_NegativeCount_ReturnsError` + 手動 exit=1 |

## 單元測試結果

```
ok  github.com/dobby/filemanager/cmd/keygen  0.518s
5/5 PASS
```

## 產出檔案

- `cmd/keygen/main.go`
- `cmd/keygen/main_test.go`
- `features/keygen.feature`
- `docs/features/keygen/architecture.md`
- `docs/features/keygen/conclusion.md`
