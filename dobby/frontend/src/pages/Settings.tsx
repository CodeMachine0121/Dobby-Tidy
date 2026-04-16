import { Info } from 'lucide-react'

export function Settings() {
  return (
    <div className="p-6 max-w-2xl">
      <div className="mb-6">
        <h1 className="text-xl font-semibold text-slate-900">設定</h1>
        <p className="text-sm text-slate-500 mt-0.5">系統設定與偏好</p>
      </div>

      {/* Notifications section */}
      <div className="card divide-y divide-border">
        <div className="px-5 py-4">
          <h2 className="text-sm font-semibold text-slate-900">通知</h2>
        </div>
        <div className="px-5 py-4 flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-slate-800">桌面通知</p>
            <p className="text-xs text-slate-400 mt-0.5">檔案被自動處理時發送系統通知</p>
          </div>
          <label className="relative inline-flex items-center cursor-pointer">
            <input type="checkbox" defaultChecked className="sr-only peer" />
            <div className="w-10 h-6 bg-slate-200 peer-focus:ring-2 peer-focus:ring-primary rounded-full peer
                          peer-checked:bg-primary transition-colors duration-150
                          after:content-[''] after:absolute after:top-0.5 after:left-0.5
                          after:bg-white after:rounded-full after:h-5 after:w-5
                          after:transition-all after:duration-150
                          peer-checked:after:translate-x-4" />
          </label>
        </div>
      </div>

      {/* About section */}
      <div className="card divide-y divide-border mt-4">
        <div className="px-5 py-4">
          <h2 className="text-sm font-semibold text-slate-900">關於</h2>
        </div>
        <div className="px-5 py-5">
          <div className="flex items-start gap-3">
            <div className="w-10 h-10 rounded-xl bg-primary flex items-center justify-center flex-shrink-0">
              <svg width="18" height="18" viewBox="0 0 16 16" fill="none">
                <path d="M2 4h12M2 8h8M2 12h10" stroke="white" strokeWidth="1.5" strokeLinecap="round"/>
              </svg>
            </div>
            <div>
              <p className="text-sm font-semibold text-slate-900">Dobby File Manager</p>
              <p className="text-xs text-slate-500 mt-0.5">版本 1.0.0</p>
              <p className="text-xs text-slate-400 mt-2 leading-relaxed">
                自動監控資料夾、依規則重新命名並歸檔檔案的桌面工具。
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Template vars reference */}
      <div className="card mt-4">
        <div className="px-5 py-4 border-b border-border flex items-center gap-2">
          <Info size={14} className="text-primary" />
          <h2 className="text-sm font-semibold text-slate-900">命名樣版變數參考</h2>
        </div>
        <div className="px-5 py-4">
          <table className="w-full text-xs">
            <thead>
              <tr className="text-left text-slate-400 border-b border-border">
                <th className="pb-2 font-medium">變數</th>
                <th className="pb-2 font-medium">說明</th>
                <th className="pb-2 font-medium">範例</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/50">
              {[
                ['{project}', '專案名稱', 'my-app'],
                ['{type}', '類型標籤', 'screenshot'],
                ['{YYYY}', '西元年', '2026'],
                ['{MM}', '月份', '04'],
                ['{DD}', '日期', '16'],
                ['{seq}', '當日序號', '001'],
                ['{original}', '原始檔名', 'Untitled'],
                ['{ext}', '副檔名', 'png'],
              ].map(([variable, desc, example]) => (
                <tr key={variable} className="text-slate-600">
                  <td className="py-2 font-mono text-primary">{variable}</td>
                  <td className="py-2 text-slate-500">{desc}</td>
                  <td className="py-2 font-mono text-slate-700">{example}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <div className="mt-3 p-3 bg-muted rounded-lg">
            <p className="text-xs text-slate-500 font-mono">
              {'{project}'}-{'{type}'}-{'{YYYY}'}{'{MM}'}{'{DD}'}-{'{seq}'}.{'{ext}'}
            </p>
            <p className="text-xs text-slate-400 mt-1">
              → my-app-screenshot-20260416-001.png
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
