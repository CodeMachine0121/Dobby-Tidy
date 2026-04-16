import { useEffect, useState } from 'react'
import { CheckCircle, XCircle, ArrowRight, RefreshCw } from 'lucide-react'
import { api } from '../lib/api'
import type { LogDTO, RuleDTO } from '../types'

function formatTime(iso: string) {
  return new Date(iso).toLocaleString('zh-TW', {
    month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  })
}

function shortPath(p: string) {
  return p.replace(/\\/g, '/').split('/').slice(-2).join('/')
}

export function Logs() {
  const [logs, setLogs] = useState<LogDTO[]>([])
  const [rules, setRules] = useState<RuleDTO[]>([])
  const [selectedRule, setSelectedRule] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  async function loadData(ruleId: string) {
    const logData = ruleId
      ? await api.logs.byRule(ruleId, 100)
      : await api.logs.recent(100)
    setLogs(logData)
  }

  useEffect(() => {
    Promise.all([api.rules.list(), api.logs.recent(100)]).then(([r, l]) => {
      setRules(r)
      setLogs(l)
    }).finally(() => setLoading(false))
  }, [])

  async function handleRuleFilter(ruleId: string) {
    setSelectedRule(ruleId)
    setRefreshing(true)
    await loadData(ruleId)
    setRefreshing(false)
  }

  async function handleRefresh() {
    setRefreshing(true)
    await loadData(selectedRule)
    setRefreshing(false)
  }

  const successCount = logs.filter((l) => l.status === 'success').length
  const errorCount = logs.filter((l) => l.status === 'error').length

  return (
    <div className="p-6 max-w-4xl">
      {/* Header */}
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-xl font-semibold text-slate-900">操作紀錄</h1>
          <p className="text-sm text-slate-500 mt-0.5">
            {logs.length} 筆紀錄・
            <span className="text-success">{successCount} 成功</span>
            {errorCount > 0 && <span className="text-destructive ml-1">{errorCount} 失敗</span>}
          </p>
        </div>
        <button
          type="button"
          onClick={handleRefresh}
          disabled={refreshing}
          className="btn-secondary"
        >
          <RefreshCw size={14} className={refreshing ? 'animate-spin' : ''} />
          重新整理
        </button>
      </div>

      {/* Filter bar */}
      <div className="flex items-center gap-3 mb-4">
        <label className="text-sm text-slate-600 font-medium flex-shrink-0">依規則篩選</label>
        <select
          className="input max-w-xs"
          value={selectedRule}
          onChange={(e) => handleRuleFilter(e.target.value)}
        >
          <option value="">全部規則</option>
          {rules.map((r) => (
            <option key={r.id} value={r.id}>{r.name}</option>
          ))}
        </select>
      </div>

      {/* Logs table */}
      <div className="card overflow-hidden">
        {loading ? (
          <div className="flex justify-center py-16">
            <div className="w-7 h-7 border-2 border-primary border-t-transparent rounded-full animate-spin" />
          </div>
        ) : logs.length === 0 ? (
          <div className="px-6 py-12 text-center">
            <p className="text-sm text-slate-400">
              {selectedRule ? '此規則尚無操作紀錄' : '尚無操作紀錄'}
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b border-border bg-muted/50">
                  <th className="text-left px-4 py-3 font-medium text-slate-500 w-6"></th>
                  <th className="text-left px-4 py-3 font-medium text-slate-500">原始檔案</th>
                  <th className="text-left px-3 py-3 font-medium text-slate-500 w-6"></th>
                  <th className="text-left px-4 py-3 font-medium text-slate-500">新路徑</th>
                  <th className="text-left px-4 py-3 font-medium text-slate-500">規則</th>
                  <th className="text-left px-4 py-3 font-medium text-slate-500">時間</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {logs.map((log) => (
                  <tr key={log.logId} className="hover:bg-muted/30 transition-colors group">
                    <td className="px-4 py-3">
                      {log.status === 'success'
                        ? <CheckCircle size={13} className="text-success" />
                        : <XCircle size={13} className="text-destructive" />
                      }
                    </td>
                    <td className="px-4 py-3 max-w-[180px]">
                      <p className="truncate text-slate-600 font-mono" title={log.originalPath}>
                        {shortPath(log.originalPath)}
                      </p>
                    </td>
                    <td className="px-3 py-3">
                      <ArrowRight size={10} className="text-slate-300" />
                    </td>
                    <td className="px-4 py-3 max-w-[200px]">
                      {log.status === 'success' ? (
                        <p className="truncate text-slate-800 font-mono font-medium" title={log.newPath}>
                          {shortPath(log.newPath)}
                        </p>
                      ) : (
                        <p className="text-destructive truncate" title={log.errorMessage}>
                          {log.errorMessage}
                        </p>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <span className="text-slate-500">{log.ruleName}</span>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-slate-400">
                      {formatTime(log.processedAt)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
