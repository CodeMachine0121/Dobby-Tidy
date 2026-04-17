# i18n 實作驗證結論

**日期：** 2026-04-16  
**狀態：** 通過

---

## 架構符合性

| 檢查項目 | 結果 |
|---------|------|
| `i18n/index.ts` — 初始化模組 | ✅ |
| `i18n/types.ts` — `TranslationKeys` 型別 | ✅ |
| `i18n/locales/zh-TW.ts` — 繁中語言檔 | ✅ |
| `i18n/locales/en.ts` — 英文語言檔 | ✅ |
| `main.tsx` 最頂端 import `./i18n` | ✅ |
| Settings 語言切換 card 位於 Notifications 上方 | ✅ |
| `resolveInitialLanguage` 匯出為純函式（可單元測試） | ✅ |

## Gherkin 情境符合性

| Scenario | 結果 |
|----------|------|
| S-1 無 localStorage 時預設 zh-TW | ✅ |
| S-2 從 localStorage 還原語言偏好 | ✅ |
| S-3 不支援的語言碼 fallback 到 zh-TW | ✅ |
| S-4 語言切換 card 在通知區塊上方 | ✅ |
| S-5 語言選單顯示目前語言 | ✅ |
| S-6 英文模式選單顯示 English | ✅ |
| S-7 切換至英文即時更新並持久化 | ✅ |
| S-8 切換至繁中即時更新並持久化 | ✅ |
| S-9 側邊欄導覽文字翻譯 | ✅ |
| S-10 Dashboard 文字翻譯 | ✅ |
| S-11 Rules subtitle 動態插值 | ✅ |
| S-12 Logs subtitle 動態插值 | ✅ |
| S-13 License 狀態動態插值 | ✅ |
| S-14 試用到期 banner 翻譯 | ✅ |
| S-15 zh-TW 與 en key 完全對齊 | ✅ TypeScript 型別 + 單元測試雙重保證 |

## 單元測試

```
Test Files  1 passed (1)
     Tests  7 passed (7)
```

## TypeScript

```
0 errors
```

## 未納入範圍（依規格）

- `formatTime()` 仍 hardcode `'zh-TW'` locale（規格明確排除）
- 後端動態字串不在前端 i18n 範圍內
- 第三語言架構支援但本次未實作
