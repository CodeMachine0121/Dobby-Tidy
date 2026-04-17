import { useEffect, useState } from 'react'
import { Info, Key, CheckCircle2, AlertTriangle, ShieldAlert } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import i18n from '../i18n'
import { api } from '../lib/api'
import type { LicenseInfo } from '../types'

// ── License Card ────────────────────────────────────────────────────────────────

function LicenseCard() {
  const { t } = useTranslation()
  const [info, setInfo] = useState<LicenseInfo | null>(null)
  const [key, setKey] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    api.license.info().then(setInfo).catch(() => null)
  }, [])

  async function handleActivate(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setSuccess(false)
    setSubmitting(true)
    try {
      await api.license.activate(key.trim())
      const updated = await api.license.info()
      setInfo(updated)
      setSuccess(true)
      setKey('')
    } catch (err: any) {
      setError(err?.message ?? t('license.defaultError'))
    } finally {
      setSubmitting(false)
    }
  }

  if (!info) return null

  return (
    <div className="card divide-y divide-border mt-4">
      <div className="px-5 py-4 flex items-center gap-2">
        <Key size={14} className="text-primary" />
        <h2 className="text-sm font-semibold text-slate-900">{t('license.title')}</h2>
      </div>

      {/* Status badge */}
      <div className="px-5 py-4">
        {info.status === 'activated' && (
          <div className="flex items-center gap-2.5">
            <CheckCircle2 size={18} className="text-success flex-shrink-0" />
            <div>
              <p className="text-sm font-medium text-slate-800">{t('license.activated')}</p>
              <p className="text-xs text-slate-400 mt-0.5">{t('license.activatedDesc')}</p>
            </div>
          </div>
        )}

        {info.status === 'active' && (
          <div className="flex items-center gap-2.5">
            <AlertTriangle size={18} className="text-amber-500 flex-shrink-0" />
            <div>
              <p className="text-sm font-medium text-slate-800">
                {t('license.trial', { days: info.daysRemaining })}
              </p>
              <p className="text-xs text-slate-400 mt-0.5">{t('license.trialDesc')}</p>
            </div>
          </div>
        )}

        {info.status === 'expired' && (
          <div className="flex items-center gap-2.5">
            <ShieldAlert size={18} className="text-destructive flex-shrink-0" />
            <div>
              <p className="text-sm font-medium text-slate-800">{t('license.expired')}</p>
              <p className="text-xs text-slate-400 mt-0.5">{t('license.expiredDesc')}</p>
            </div>
          </div>
        )}
      </div>

      {/* Key input — only shown when not yet activated */}
      {info.status !== 'activated' && (
        <div className="px-5 py-4 space-y-3">
          <p className="text-xs font-medium text-slate-500">{t('license.enterKey')}</p>

          {success && (
            <div className="flex items-center gap-2 p-2.5 rounded-lg bg-success/10 border border-success/20">
              <CheckCircle2 size={14} className="text-success" />
              <span className="text-xs text-success font-medium">{t('license.activateSuccess')}</span>
            </div>
          )}

          {error && (
            <div className="flex items-center gap-2 p-2.5 rounded-lg bg-destructive/10 border border-destructive/20">
              <ShieldAlert size={14} className="text-destructive" />
              <span className="text-xs text-destructive">{error}</span>
            </div>
          )}

          <form onSubmit={handleActivate} className="flex gap-2">
            <input
              type="text"
              value={key}
              onChange={e => setKey(e.target.value.toUpperCase())}
              placeholder="DOBBY-XXXX-XXXX-XXXX"
              className="flex-1 px-3 py-2 text-sm font-mono rounded-lg border border-border
                         bg-white text-slate-800 placeholder:text-slate-300
                         focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary
                         disabled:opacity-50"
              disabled={submitting}
              spellCheck={false}
              autoComplete="off"
            />
            <button
              type="submit"
              disabled={submitting || key.trim() === ''}
              className="px-4 py-2 text-sm font-medium text-white bg-primary rounded-lg
                         hover:bg-primary-hover disabled:opacity-40 disabled:cursor-not-allowed
                         transition-colors duration-150"
            >
              {submitting ? t('license.activating') : t('license.activate')}
            </button>
          </form>

          <p className="text-xs text-slate-400">
            {t('license.buyPrompt')}
            <a
              href="https://afternoonjames.gumroad.com/l/dobby-tidy"
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary hover:underline ml-1"
            >
              {t('license.buyLink')}
            </a>
          </p>
        </div>
      )}
    </div>
  )
}

// ── Settings Page ────────────────────────────────────────────────────────────────

export function Settings() {
  const { t } = useTranslation()

  function handleLanguageChange(lang: string) {
    i18n.changeLanguage(lang)
    localStorage.setItem('language', lang)
  }

  return (
    <div className="p-6 max-w-2xl">
      <div className="mb-6">
        <h1 className="text-xl font-semibold text-slate-900">{t('settings.title')}</h1>
        <p className="text-sm text-slate-500 mt-0.5">{t('settings.subtitle')}</p>
      </div>

      {/* Language section */}
      <div className="card divide-y divide-border">
        <div className="px-5 py-4">
          <h2 className="text-sm font-semibold text-slate-900">{t('settings.language')}</h2>
        </div>
        <div className="px-5 py-4 flex items-center justify-between">
          <p className="text-sm font-medium text-slate-800">{t('settings.interfaceLanguage')}</p>
          <select
            className="input max-w-[160px]"
            value={i18n.language}
            onChange={(e) => handleLanguageChange(e.target.value)}
          >
            <option value="zh-TW">繁體中文</option>
            <option value="en">English</option>
          </select>
        </div>
      </div>

      {/* Notifications section */}
      <div className="card divide-y divide-border mt-4">
        <div className="px-5 py-4">
          <h2 className="text-sm font-semibold text-slate-900">{t('settings.notifications')}</h2>
        </div>
        <div className="px-5 py-4 flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-slate-800">{t('settings.desktopNotifications')}</p>
            <p className="text-xs text-slate-400 mt-0.5">{t('settings.desktopNotificationsDesc')}</p>
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

      <LicenseCard />

      {/* About section */}
      <div className="card divide-y divide-border mt-4">
        <div className="px-5 py-4">
          <h2 className="text-sm font-semibold text-slate-900">{t('settings.about')}</h2>
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
              <p className="text-xs text-slate-500 mt-0.5">{t('settings.version')} 1.0.0</p>
              <p className="text-xs text-slate-400 mt-2 leading-relaxed">{t('settings.appDesc')}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Template vars reference */}
      <div className="card mt-4">
        <div className="px-5 py-4 border-b border-border flex items-center gap-2">
          <Info size={14} className="text-primary" />
          <h2 className="text-sm font-semibold text-slate-900">{t('settings.templateRef')}</h2>
        </div>
        <div className="px-5 py-4">
          <table className="w-full text-xs">
            <thead>
              <tr className="text-left text-slate-400 border-b border-border">
                <th className="pb-2 font-medium">{t('settings.templateCol.var')}</th>
                <th className="pb-2 font-medium">{t('settings.templateCol.desc')}</th>
                <th className="pb-2 font-medium">{t('settings.templateCol.example')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/50">
              {([
                ['{project}', t('settings.templateVars.project'), 'my-app'],
                ['{type}', t('settings.templateVars.type'), 'screenshot'],
                ['{YYYY}', t('settings.templateVars.YYYY'), '2026'],
                ['{MM}', t('settings.templateVars.MM'), '04'],
                ['{DD}', t('settings.templateVars.DD'), '16'],
                ['{seq}', t('settings.templateVars.seq'), '001'],
                ['{original}', t('settings.templateVars.original'), 'Untitled'],
                ['{ext}', t('settings.templateVars.ext'), 'png'],
              ] as [string, string, string][]).map(([variable, desc, example]) => (
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
