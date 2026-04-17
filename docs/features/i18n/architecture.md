# i18n 架構設計文件

**功能：** 雙語語言切換（zh-TW / en）  
**日期：** 2026-04-16  
**Gherkin：** `features/i18n.feature`

---

## 專案上下文

- **技術棧：** React 18 + TypeScript + Vite + Tailwind CSS + Wails（Go desktop）
- **架構模式：** 以頁面為單位的元件架構（Layout + pages/）
- **命名慣例：** PascalCase 元件，camelCase 函式，kebab-case 路徑
- **狀態管理：** React local state + Context（無 Redux）

---

## 函式庫選擇

採用 **react-i18next**（`i18next` + `react-i18next`）：
- 支援動態插值語法 `{{days}}`、`{{total}}` 等（對應 S-11、S-12、S-13）
- 初始化一次即可全 app 使用，無需額外 Provider
- 未來擴充第三語言只需新增語言檔

---

## 資料模型

### `TranslationKeys` 型別（`i18n/types.ts`）

對應 Gherkin S-15（TypeScript 型別保證 key 一致性）。

```ts
export interface TranslationKeys {
  nav: { dashboard: string; rules: string; logs: string; settings: string;
         trialExpired: string; trialExpiredDesc: string; enterLicenseKey: string }
  dashboard: { title: string; subtitle: string; todayProcessed: string;
               filesUnit: string; activeRules: string; rulesUnit: string;
               recentLogs: string; viewAll: string; noLogs: string; noLogsHint: string }
  rules: { title: string; subtitle: string; addRule: string; noRules: string;
           noRulesHint: string; add: string;
           modal: { title: string; name: string; namePlaceholder: string;
                    watchFolder: string; recursive: string; filterExts: string;
                    filterExtsHint: string; filterKeyword: string;
                    filterKeywordPlaceholder: string; project: string;
                    typeLabel: string; nameTemplate: string;
                    nameTemplatePreview: string; targetFolder: string;
                    targetFolderHint: string; cancel: string; create: string; saving: string }
           card: { enable: string; disable: string; delete: string;
                   confirmDelete: string; cancelDelete: string; nameTemplate: string;
                   targetFolder: string; filterExts: string; filterExtsAll: string;
                   filterKeyword: string; project: string; typeLabel: string }
           error: { nameRequired: string; folderRequired: string; templateRequired: string } }
  logs: { title: string; subtitle: string; subtitleWithError: string;
          refresh: string; filterByRule: string; allRules: string;
          noLogs: string; noLogsForRule: string;
          col: { original: string; newPath: string; rule: string; time: string } }
  settings: { title: string; subtitle: string; language: string;
              interfaceLanguage: string; notifications: string;
              desktopNotifications: string; desktopNotificationsDesc: string;
              about: string; version: string; appDesc: string;
              templateRef: string;
              templateCol: { var: string; desc: string; example: string }
              templateVars: { project: string; type: string; YYYY: string;
                              MM: string; DD: string; seq: string;
                              original: string; ext: string } }
  license: { title: string; activated: string; activatedDesc: string;
             trial: string; trialDesc: string; expired: string; expiredDesc: string;
             enterKey: string; activate: string; activating: string;
             activateSuccess: string; buyPrompt: string; buyLink: string;
             defaultError: string }
  common: { loading: string }
}
```

---

## 服務介面

### `i18n/index.ts` — 初始化模組（對應 S-1、S-2、S-3）

```ts
// 在 main.tsx 最頂端 import，執行一次
// 從 localStorage.getItem('language') 讀取初始語言，fallback 到 'zh-TW'
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import zhTW from './locales/zh-TW'
import en from './locales/en'

const savedLang = localStorage.getItem('language')
const supportedLangs = ['zh-TW', 'en']
const initialLang = savedLang && supportedLangs.includes(savedLang) ? savedLang : 'zh-TW'

i18n.use(initReactI18next).init({
  resources: { 'zh-TW': { translation: zhTW }, en: { translation: en } },
  lng: initialLang,
  fallbackLng: 'zh-TW',
  interpolation: { escapeValue: false },
})

export default i18n
```

### `handleLanguageChange(lang: string)` — 語言切換函式（對應 S-7、S-8）

```ts
function handleLanguageChange(lang: string) {
  i18n.changeLanguage(lang)
  localStorage.setItem('language', lang)
}
```

### 各 component 取用翻譯

```ts
const { t, i18n } = useTranslation()
// t('nav.dashboard')
// t('rules.subtitle', { total, active })
```

---

## 架構決策

| 決策 | 選項 | 理由 |
|------|------|------|
| 語言儲存 | localStorage | Desktop app，無需 server-side session（對應 S-2） |
| 翻譯 key 結構 | 扁平化 + 頁面前綴 | 避免 namespace 過多，易於維護 |
| Fallback 語言 | zh-TW | 主要市場，對應 S-3 |
| TypeScript 型別 | `TranslationKeys` interface | 確保兩份語言檔 key 一致，對應 S-15 |
| `formatTime()` | 不改動 | 規格明確排除，hardcode zh-TW locale |

---

## 情境對應表

| Gherkin Tag | 對應模組/元件 |
|-------------|--------------|
| S-1, S-2, S-3 | `i18n/index.ts` 初始化邏輯 |
| S-4, S-5, S-6 | `pages/Settings.tsx` 語言切換 card |
| S-7, S-8 | `handleLanguageChange()` 函式 |
| S-9 | `components/Layout.tsx` navItems |
| S-10 | `pages/Dashboard.tsx` |
| S-11 | `pages/Rules.tsx` subtitle 插值 |
| S-12 | `pages/Logs.tsx` subtitle 插值 |
| S-13 | `pages/Settings.tsx` LicenseCard |
| S-14 | `components/Layout.tsx` trial expired banner |
| S-15 | `i18n/types.ts` + `i18n/locales/zh-TW.ts` + `i18n/locales/en.ts` |

---

## 檔案結構規劃

```
frontend/src/
├── i18n/
│   ├── index.ts          # i18next 初始化（新建）
│   ├── types.ts          # TranslationKeys 型別（新建）
│   └── locales/
│       ├── zh-TW.ts      # 繁中翻譯（新建）
│       └── en.ts         # 英文翻譯（新建）
├── main.tsx              # 加入 i18n import（修改）
├── components/
│   └── Layout.tsx        # navItems 及 banner 使用 t()（修改）
└── pages/
    ├── Dashboard.tsx     # 使用 t()（修改）
    ├── Rules.tsx         # 使用 t()（修改）
    ├── Logs.tsx          # 使用 t()（修改）
    └── Settings.tsx      # 加入語言切換 card，使用 t()（修改）
```
