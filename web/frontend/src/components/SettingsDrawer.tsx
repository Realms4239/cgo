import { useEffect, useState } from 'react'
import { DEFAULTS, loadSettings, saveSettings, type Settings } from '../lib/settings'
import { useUIStore } from '../store/ui'
import { loadSchema, param, type Param } from '../lib/schema'

// Réglages — rendu depuis /api/schema (source unique): bornes, pas,
// défauts en placeholder. Les bornes serveur restent autoritaires.
const FIELDS: { key: keyof Settings; skey: string; label: string; unit?: string }[] = [
  { key: 'warnMs', skey: 'warn_ms', label: 'Seuil « dégradé »' },
  { key: 'critMs', skey: 'crit_ms', label: 'Seuil « critique »' },
  { key: 'deadlineMs', skey: 'deadline_ms', label: 'Deadline small p95' },
  { key: 'ringMax', skey: 'ring_max', label: 'Fenêtre live' },
  { key: 'watchMs', skey: 'watch_ms', label: 'Cadence surveillance' },
  { key: 'shapeCap', skey: 'shape_capacity_mbps', label: 'Capacité du bord' },
  { key: 'linkDelayMs', skey: 'link_delay_ms', label: 'Délai du lien' },
  { key: 'linkJitterMs', skey: 'link_jitter_ms', label: 'Gigue' },
  { key: 'linkLossPct', skey: 'link_loss_pct', label: 'Perte' },
  { key: 'target', skey: 'target', label: 'Cible de mesure' },
]

export default function SettingsDrawer({ onClose }: { onClose: () => void }) {
  const [s, setS] = useState<Settings>(() => loadSettings())
  const [schema, setSchema] = useState<Param[]>([])
  const [tiers, setTiers] = useState<{ name: string; ar_per_gb: number }[]>([])
  const [tier, setTier] = useState('')
  const setFlash = useUIStore((st: any) => st.setFlash)
  useEffect(() => { loadSchema().then(setSchema).catch(() => {}) }, [])
  useEffect(() => {
    fetch('/api/cost/tiers').then(r => r.json()).then(j => {
      setTiers(j.tiers || [])
      setTier(j.default || '')
    }).catch(() => {})
  }, [])
  const apply = async () => {
    saveSettings(s)
    // le palier est un état SERVEUR (coûts affichés partout) — l'échec se dit
    if (tier) {
      try {
        const r = await fetch('/api/cost/tier', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ tier }) })
        if (!r.ok) throw new Error(String(r.status))
      } catch {
        setFlash({ type: 'danger', msg: 'Réglages locaux OK — palier serveur refusé' })
        return
      }
    }
    setFlash({ type: 'success', msg: 'Réglages enregistrés' })
    onClose()
  }
  return (
    <div role="dialog" aria-label="Réglages" style={{ position: 'fixed', inset: 0, zIndex: 650, background: 'rgba(0,0,0,0.5)', display: 'flex', justifyContent: 'flex-end' }} onClick={onClose}>
      <div style={{ width: 380, height: '100%', background: 'var(--surface-soft, #0b0b0c)', borderLeft: '1px solid var(--hairline)', padding: 16, overflowY: 'auto' }} onClick={e => e.stopPropagation()}>
        <div className="card-head" style={{ marginBottom: 12 }}>Réglages</div>
        {FIELDS.map(f => {
          const p = param(f.skey, schema)
          const def = p?.default ?? (DEFAULTS as any)[f.key]
          return (
            <div key={f.key} className="form-row" style={{ flexDirection: 'column', alignItems: 'stretch', gap: 2, marginBottom: 12 }}>
              <label className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#a8aeb7' }}>
                {f.label}{p?.unit ? ` (${p.unit})` : ''}
              </label>
              {f.key === 'target' ? (
                <input value={String(s[f.key])} maxLength={Number(p?.max ?? 64)} placeholder={String(def)} onChange={e => setS({ ...s, target: e.target.value })} style={{ background: 'var(--surface-card)', color: 'var(--text-body)', border: '1px solid var(--hairline)', padding: '6px 8px', fontFamily: 'var(--font-mono)', fontSize: 11 }} />
              ) : (
                <input type="number" min={p?.min} max={p?.max} step={p?.step ?? 1} placeholder={String(def)} value={Number(s[f.key])} onChange={e => setS({ ...s, [f.key]: parseFloat(e.target.value) || 0 })} style={{ width: 120, background: 'var(--surface-card)', color: 'var(--text-body)', border: '1px solid var(--hairline)', padding: '6px 8px', fontFamily: 'var(--font-mono)', fontSize: 11 }} />
              )}
              <span className="mono" style={{ fontSize: 9, color: '#9aa0a8' }}>{p?.desc ?? f.label}</span>
            </div>
          )
        })}
        <div className="form-row" style={{ flexDirection: 'column', alignItems: 'stretch', gap: 2, marginBottom: 12 }}>
          <label className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#a8aeb7' }}>
            Palier tarifaire (coûts affichés)
          </label>
          <select data-testid="tier-select" value={tier} onChange={e => setTier(e.target.value)} style={{ background: 'var(--surface-card)', color: 'var(--text-body)', border: '1px solid var(--hairline)', padding: '6px 8px', fontFamily: 'var(--font-mono)', fontSize: 11 }}>
            {tiers.length === 0 && <option value="">—</option>}
            {tiers.map(t => <option key={t.name} value={t.name}>{t.name} — {t.ar_per_gb} Ar/Go</option>)}
          </select>
          <span className="mono" style={{ fontSize: 9, color: '#9aa0a8' }}>le forfait où vit l'institution — tous les coûts affichés suivent</span>
        </div>
        <div className="form-row" style={{ gap: 8, marginTop: 12 }}>
          <button className="btn btn-primary" onClick={apply}>ENREGISTRER</button>
          <button className="btn" onClick={() => setS({ ...DEFAULTS })}>DÉFAUTS</button>
          <button className="btn" onClick={onClose}>ANNULER</button>
        </div>
      </div>
    </div>
  )
}
