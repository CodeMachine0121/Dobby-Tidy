import { useEffect, useRef, useState } from 'react'
import {
  Plus, ToggleLeft, ToggleRight, Trash2, FolderOpen,
  ChevronDown, X, AlertCircle
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { SelectFolder } from '../../wailsjs/go/main/App'
import { api } from '../lib/api'
import type { CreateRuleRequest, RuleDTO } from '../types'

const TEMPLATE_VARS = ['{project}', '{type}', '{YYYY}', '{MM}', '{DD}', '{seq}', '{original}', '{ext}']
const DEFAULT_FORM: CreateRuleRequest = {
  name: '',
  watchFolder: '',
  recursive: false,
  filterExts: [],
  filterKeyword: '',
  nameTemplate: '{project}-{type}-{YYYY}{MM}{DD}-{seq}.{ext}',
  targetTemplate: '',
  project: '',
  typeLabel: '',
}

function ExtInput({ value, onChange }: { value: string[]; onChange: (v: string[]) => void }) {
  const { t } = useTranslation()
  const [inputVal, setInputVal] = useState('')

  function add() {
    const ext = inputVal.trim().replace(/^\./, '')
    if (ext && !value.includes(ext)) onChange([...value, ext])
    setInputVal('')
  }

  function remove(ext: string) {
    onChange(value.filter((e) => e !== ext))
  }

  return (
    <div>
      <div className="flex flex-wrap gap-1.5 mb-2">
        {value.map((ext) => (
          <span key={ext} className="inline-flex items-center gap-1 px-2 py-0.5 bg-primary-light text-primary text-xs rounded-full font-medium">
            .{ext}
            <button type="button" onClick={() => remove(ext)} className="hover:text-primary-hover cursor-pointer">
              <X size={10} />
            </button>
          </span>
        ))}
      </div>
      <div className="flex gap-2">
        <input
          type="text"
          className="input flex-1"
          placeholder="png, jpg, pdf..."
          value={inputVal}
          onChange={(e) => setInputVal(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); add() } }}
        />
        <button type="button" onClick={add} className="btn-secondary px-3">
          {t('rules.add')}
        </button>
      </div>
    </div>
  )
}

function RuleModal({
  open,
  onClose,
  onSave,
}: {
  open: boolean
  onClose: () => void
  onSave: (req: CreateRuleRequest) => Promise<void>
}) {
  const { t } = useTranslation()
  const [form, setForm] = useState<CreateRuleRequest>(DEFAULT_FORM)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const overlayRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (open) { setForm(DEFAULT_FORM); setError('') }
  }, [open])

  function set<K extends keyof CreateRuleRequest>(key: K, val: CreateRuleRequest[K]) {
    setForm((f) => ({ ...f, [key]: val }))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.name.trim()) return setError(t('rules.error.nameRequired'))
    if (!form.watchFolder.trim()) return setError(t('rules.error.folderRequired'))
    if (!form.nameTemplate.trim()) return setError(t('rules.error.templateRequired'))
    setSaving(true)
    setError('')
    try {
      await onSave(form)
      onClose()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setSaving(false)
    }
  }

  if (!open) return null

  return (
    <div
      ref={overlayRef}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={(e) => { if (e.target === overlayRef.current) onClose() }}
    >
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-lg max-h-[90vh] flex flex-col animate-in fade-in zoom-in-95 duration-150">
        {/* Modal header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border flex-shrink-0">
          <h2 className="text-base font-semibold text-slate-900">{t('rules.modal.title')}</h2>
          <button type="button" onClick={onClose} className="p-1.5 rounded-lg hover:bg-slate-100 transition-colors cursor-pointer">
            <X size={16} className="text-slate-500" />
          </button>
        </div>

        {/* Modal body */}
        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto px-6 py-5 space-y-4">
          {error && (
            <div className="flex items-center gap-2 p-3 bg-destructive-light text-destructive rounded-lg text-sm">
              <AlertCircle size={14} />
              {error}
            </div>
          )}

          <div>
            <label className="label">{t('rules.modal.name')} <span className="text-destructive">*</span></label>
            <input className="input" placeholder={t('rules.modal.namePlaceholder')} value={form.name} onChange={(e) => set('name', e.target.value)} />
          </div>

          <div>
            <label className="label">{t('rules.modal.watchFolder')} <span className="text-destructive">*</span></label>
            <div className="relative">
              <input className="input pr-10" placeholder="C:/Users/me/Downloads" value={form.watchFolder} onChange={(e) => set('watchFolder', e.target.value)} />
              <button
                type="button"
                className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-primary transition-colors cursor-pointer"
                onClick={async () => { const p = await SelectFolder(); if (p) set('watchFolder', p) }}
                title={t('rules.modal.watchFolder')}
              >
                <FolderOpen size={15} />
              </button>
            </div>
            <label className="flex items-center gap-2 mt-2 cursor-pointer select-none">
              <input type="checkbox" className="w-4 h-4 rounded text-primary accent-primary" checked={form.recursive} onChange={(e) => set('recursive', e.target.checked)} />
              <span className="text-sm text-slate-600">{t('rules.modal.recursive')}</span>
            </label>
          </div>

          <div>
            <label className="label">{t('rules.modal.filterExts')}</label>
            <ExtInput value={form.filterExts} onChange={(v) => set('filterExts', v)} />
            <p className="text-xs text-slate-400 mt-1">{t('rules.modal.filterExtsHint')}</p>
          </div>

          <div>
            <label className="label">{t('rules.modal.filterKeyword')}</label>
            <input className="input" placeholder={t('rules.modal.filterKeywordPlaceholder')} value={form.filterKeyword} onChange={(e) => set('filterKeyword', e.target.value)} />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="label">{t('rules.modal.project')}</label>
              <input className="input" placeholder="my-project" value={form.project} onChange={(e) => set('project', e.target.value)} />
            </div>
            <div>
              <label className="label">{t('rules.modal.typeLabel')}</label>
              <input className="input" placeholder="screenshot" value={form.typeLabel} onChange={(e) => set('typeLabel', e.target.value)} />
            </div>
          </div>

          <div>
            <label className="label">{t('rules.modal.nameTemplate')} <span className="text-destructive">*</span></label>
            <input className="input font-mono text-xs" value={form.nameTemplate} onChange={(e) => set('nameTemplate', e.target.value)} />
            <div className="flex flex-wrap gap-1.5 mt-2">
              {TEMPLATE_VARS.map((v) => (
                <button
                  key={v}
                  type="button"
                  onClick={() => set('nameTemplate', form.nameTemplate + v)}
                  className="px-2 py-0.5 text-xs font-mono bg-muted text-slate-600 rounded border border-border hover:bg-primary-light hover:text-primary hover:border-primary transition-colors cursor-pointer"
                >
                  {v}
                </button>
              ))}
            </div>
            <p className="text-xs text-slate-400 mt-1.5">
              {t('rules.modal.nameTemplatePreview')}<span className="font-mono text-slate-600">{form.nameTemplate || '—'}</span>
            </p>
          </div>

          <div>
            <label className="label">{t('rules.modal.targetFolder')}</label>
            <div className="relative">
              <input className="input pr-10" placeholder="C:/Users/me/Projects/{project}/assets" value={form.targetTemplate} onChange={(e) => set('targetTemplate', e.target.value)} />
              <button
                type="button"
                className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-primary transition-colors cursor-pointer"
                onClick={async () => { const p = await SelectFolder(); if (p) set('targetTemplate', p) }}
                title={t('rules.modal.targetFolder')}
              >
                <FolderOpen size={15} />
              </button>
            </div>
            <p className="text-xs text-slate-400 mt-1">{t('rules.modal.targetFolderHint')}</p>
          </div>
        </form>

        {/* Modal footer */}
        <div className="flex items-center justify-end gap-2 px-6 py-4 border-t border-border flex-shrink-0">
          <button type="button" onClick={onClose} className="btn-secondary">{t('rules.modal.cancel')}</button>
          <button
            type="button"
            onClick={(e) => handleSubmit(e as unknown as React.FormEvent)}
            disabled={saving}
            className="btn-primary"
          >
            {saving ? (
              <>
                <div className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                {t('rules.modal.saving')}
              </>
            ) : t('rules.modal.create')}
          </button>
        </div>
      </div>
    </div>
  )
}

function RuleCard({ rule, onToggle, onDelete }: {
  rule: RuleDTO
  onToggle: () => void
  onDelete: () => void
}) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)

  return (
    <div className={`card overflow-hidden transition-all duration-200 ${!rule.enabled ? 'opacity-60' : ''}`}>
      {/* Rule header */}
      <div className="flex items-center gap-4 px-5 py-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className={`w-2 h-2 rounded-full flex-shrink-0 ${rule.enabled ? 'bg-success' : 'bg-slate-300'}`} />
            <h3 className="text-sm font-semibold text-slate-900 truncate">{rule.name}</h3>
          </div>
          <p className="text-xs text-slate-400 mt-0.5 truncate ml-4">{rule.watchFolder}</p>
        </div>

        <div className="flex items-center gap-1 flex-shrink-0">
          <button
            type="button"
            onClick={onToggle}
            className="p-2 rounded-lg hover:bg-muted transition-colors cursor-pointer"
            title={rule.enabled ? t('rules.card.disable') : t('rules.card.enable')}
          >
            {rule.enabled
              ? <ToggleRight size={18} className="text-primary" />
              : <ToggleLeft size={18} className="text-slate-400" />
            }
          </button>

          {confirmDelete ? (
            <div className="flex items-center gap-1">
              <button type="button" onClick={() => setConfirmDelete(false)} className="btn-secondary text-xs px-2 py-1">{t('rules.card.cancelDelete')}</button>
              <button type="button" onClick={onDelete} className="btn-danger text-xs px-2 py-1">{t('rules.card.confirmDelete')}</button>
            </div>
          ) : (
            <button
              type="button"
              onClick={() => setConfirmDelete(true)}
              className="p-2 rounded-lg hover:bg-destructive-light transition-colors cursor-pointer"
              title={t('rules.card.delete')}
            >
              <Trash2 size={15} className="text-slate-400 hover:text-destructive" />
            </button>
          )}

          <button
            type="button"
            onClick={() => setExpanded((v) => !v)}
            className="p-2 rounded-lg hover:bg-muted transition-colors cursor-pointer"
          >
            <ChevronDown size={15} className={`text-slate-400 transition-transform duration-200 ${expanded ? 'rotate-180' : ''}`} />
          </button>
        </div>
      </div>

      {/* Expanded details */}
      {expanded && (
        <div className="px-5 pb-4 border-t border-border pt-4 space-y-3">
          <div className="grid grid-cols-2 gap-x-6 gap-y-2 text-xs">
            <div>
              <span className="text-slate-400">{t('rules.card.nameTemplate')}</span>
              <p className="font-mono text-slate-700 mt-0.5">{rule.nameTemplate || '—'}</p>
            </div>
            <div>
              <span className="text-slate-400">{t('rules.card.targetFolder')}</span>
              <p className="font-mono text-slate-700 mt-0.5">{rule.targetTemplate || '—'}</p>
            </div>
            <div>
              <span className="text-slate-400">{t('rules.card.filterExts')}</span>
              <p className="text-slate-700 mt-0.5">
                {rule.filterExts.length > 0 ? rule.filterExts.map((e) => `.${e}`).join(', ') : t('rules.card.filterExtsAll')}
              </p>
            </div>
            <div>
              <span className="text-slate-400">{t('rules.card.filterKeyword')}</span>
              <p className="text-slate-700 mt-0.5">{rule.filterKeyword || '—'}</p>
            </div>
            {rule.project && (
              <div>
                <span className="text-slate-400">{t('rules.card.project')}</span>
                <p className="text-slate-700 mt-0.5">{rule.project}</p>
              </div>
            )}
            {rule.typeLabel && (
              <div>
                <span className="text-slate-400">{t('rules.card.typeLabel')}</span>
                <p className="text-slate-700 mt-0.5">{rule.typeLabel}</p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export function Rules() {
  const { t } = useTranslation()
  const [rules, setRules] = useState<RuleDTO[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)

  async function loadRules() {
    const data = await api.rules.list()
    setRules(data)
  }

  useEffect(() => {
    loadRules().finally(() => setLoading(false))
  }, [])

  async function handleCreate(req: CreateRuleRequest) {
    await api.rules.create(req)
    await loadRules()
  }

  async function handleToggle(rule: RuleDTO) {
    if (rule.enabled) await api.rules.disable(rule.id)
    else await api.rules.enable(rule.id)
    await loadRules()
  }

  async function handleDelete(id: string) {
    await api.rules.delete(id)
    await loadRules()
  }

  return (
    <div className="p-6 max-w-3xl">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-xl font-semibold text-slate-900">{t('rules.title')}</h1>
          <p className="text-sm text-slate-500 mt-0.5">
            {t('rules.subtitle', { total: rules.length, active: rules.filter((r) => r.enabled).length })}
          </p>
        </div>
        <button type="button" onClick={() => setModalOpen(true)} className="btn-primary">
          <Plus size={15} />
          {t('rules.addRule')}
        </button>
      </div>

      {/* Rules list */}
      {loading ? (
        <div className="flex justify-center py-16">
          <div className="w-7 h-7 border-2 border-primary border-t-transparent rounded-full animate-spin" />
        </div>
      ) : rules.length === 0 ? (
        <div className="card px-6 py-12 text-center">
          <div className="w-12 h-12 rounded-2xl bg-primary-light flex items-center justify-center mx-auto mb-3">
            <Plus size={22} className="text-primary" />
          </div>
          <p className="text-sm font-medium text-slate-700">{t('rules.noRules')}</p>
          <p className="text-xs text-slate-400 mt-1">{t('rules.noRulesHint')}</p>
          <button type="button" onClick={() => setModalOpen(true)} className="btn-primary mt-4 mx-auto">
            <Plus size={14} />
            {t('rules.addRule')}
          </button>
        </div>
      ) : (
        <div className="space-y-3">
          {rules.map((rule) => (
            <RuleCard
              key={rule.id}
              rule={rule}
              onToggle={() => handleToggle(rule)}
              onDelete={() => handleDelete(rule.id)}
            />
          ))}
        </div>
      )}

      <RuleModal open={modalOpen} onClose={() => setModalOpen(false)} onSave={handleCreate} />
    </div>
  )
}
