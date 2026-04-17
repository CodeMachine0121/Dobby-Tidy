import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { FileCheck, ListChecks, ArrowRight, CheckCircle, XCircle } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { api } from '../lib/api'
import type { LogDTO, RuleDTO } from '../types'

function formatPath(p: string) {
  const parts = p.replace(/\\/g, '/').split('/')
  return parts[parts.length - 1]
}

function formatTime(iso: string) {
  const d = new Date(iso)
  return d.toLocaleString('zh-TW', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

export function Dashboard() {
  const { t } = useTranslation()
  const [todayCount, setTodayCount] = useState<number>(0)
  const [ruleCount, setRuleCount] = useState<number>(0)
  const [logs, setLogs] = useState<LogDTO[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      api.logs.todayCount(),
      api.rules.list(),
      api.logs.recent(8),
    ]).then(([count, rules, recentLogs]) => {
      setTodayCount(count)
      setRuleCount(rules.filter((r: RuleDTO) => r.enabled).length)
      setLogs(recentLogs)
    }).finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 border-2 border-primary border-t-transparent rounded-full animate-spin" />
          <p className="text-sm text-slate-500">{t('common.loading')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-4xl">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-xl font-semibold text-slate-900">{t('dashboard.title')}</h1>
        <p className="text-sm text-slate-500 mt-0.5">{t('dashboard.subtitle')}</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 gap-4 mb-6">
        <div className="card p-5">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-xs font-medium text-slate-500 uppercase tracking-wide">{t('dashboard.todayProcessed')}</p>
              <p className="text-3xl font-bold text-slate-900 mt-1">{todayCount}</p>
              <p className="text-xs text-slate-400 mt-1">{t('dashboard.filesUnit')}</p>
            </div>
            <div className="w-10 h-10 rounded-xl bg-accent-light flex items-center justify-center">
              <FileCheck size={20} className="text-accent" />
            </div>
          </div>
        </div>

        <div className="card p-5">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-xs font-medium text-slate-500 uppercase tracking-wide">{t('dashboard.activeRules')}</p>
              <p className="text-3xl font-bold text-slate-900 mt-1">{ruleCount}</p>
              <p className="text-xs text-slate-400 mt-1">{t('dashboard.rulesUnit')}</p>
            </div>
            <div className="w-10 h-10 rounded-xl bg-primary-light flex items-center justify-center">
              <ListChecks size={20} className="text-primary" />
            </div>
          </div>
        </div>
      </div>

      {/* Recent logs */}
      <div className="card">
        <div className="flex items-center justify-between px-5 py-4 border-b border-border">
          <h2 className="text-sm font-semibold text-slate-900">{t('dashboard.recentLogs')}</h2>
          <Link
            to="/logs"
            className="flex items-center gap-1 text-xs text-primary hover:text-primary-hover transition-colors"
          >
            {t('dashboard.viewAll')} <ArrowRight size={12} />
          </Link>
        </div>

        {logs.length === 0 ? (
          <div className="px-5 py-10 text-center">
            <p className="text-sm text-slate-400">{t('dashboard.noLogs')}</p>
            <p className="text-xs text-slate-300 mt-1">{t('dashboard.noLogsHint')}</p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {logs.map((log) => (
              <div key={log.logId} className="flex items-center gap-4 px-5 py-3 hover:bg-muted transition-colors">
                <div className="flex-shrink-0">
                  {log.status === 'success' ? (
                    <CheckCircle size={16} className="text-success" />
                  ) : (
                    <XCircle size={16} className="text-destructive" />
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-slate-400 truncate">{formatPath(log.originalPath)}</span>
                    <ArrowRight size={10} className="text-slate-300 flex-shrink-0" />
                    <span className="text-xs font-medium text-slate-700 truncate">{formatPath(log.newPath)}</span>
                  </div>
                  <p className="text-xs text-slate-400 mt-0.5">{log.ruleName}</p>
                </div>
                <span className="text-xs text-slate-400 flex-shrink-0">{formatTime(log.processedAt)}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
