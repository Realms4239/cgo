import { useState } from 'react'
import { DEFAULTS, loadSettings, saveSettings, type Settings } from '../lib/settings'
import { useUIStore } from '../store/ui'

// Réglages — 10 paramètres, chaque borne est une décision de mesure (Q9).
// Persistés localement; les bornes serveur restent autoritaires.
const FIELDS: { key: keyof Settings; label: string; min: number; max: number; step?: number; unit?: string; hint: string }[] = [
  { key: 'warnMs', label: 'Seuil « dégradé »', min: 10, max: 200, unit: 'ms', hint: 'latence au-delà de laquelle une carte passe en jaune' },
  { key: 'critMs', label: 'Seuil « critique »', min: 50, max: 500, unit: 'ms', hint: 'latence au-delà de laquelle une carte passe en rouge' },
  { key: 'deadlineMs', label: 'Deadline small p95', min: 200, max: 5000, step: 100, unit: 'ms', hint: 'objectif de latence transmis à la campagne' },
  { key: 'ringMax', label: 'Fenêtre live', min: 600, max: 3600, step: 100, unit: 'pts', hint: 'profondeur d’historique du tableau live (10 pts/s)' },
  { key: 'watchMs', label: 'Cadence surveillance', min: 200, max: 2000, step: 100, unit: 'ms', hint: 'intervalle des sondes légères' },
  { key: 'shapeCap', label: 'Capacité du bord', min: 1, max: 1000, unit: 'Mbit/s', hint: 'capacité proposée par défaut au levier' },
  { key: 'linkDelayMs', label: 'Délai du lien', min: 0, max: 600, unit: 'ms', hint: 'délai proposé par défaut aux conditions du lien' },
  { key: 'linkJitterMs', label: 'Gigue', min: 0, max: 100, unit: 'ms', hint: 'gigue proposée par défaut' },
  { key: 'linkLossPct', label: 'Perte', min: 0, max: 10, step: 0.1, unit: '%', hint: 'perte proposée par défaut' },
  { key: 'target', label: 'Cible de mesure', min: 0, max: 64, hint: 'hôte sondé par la campagne et l’audit' },
]

export default function SettingsDrawer({ onClose }: { onClose: () => void }) {
  const [s, setS] = useState<Settings>(() => loadSettings())
  const setFlash = useUIStore((st: any) => st.setFlash)
  const apply = () => {
    saveSettings(s)
    setFlash({ type: 'success', msg: 'Réglages enregistrés' })
    onClose()
  }
  return (
    <div role="dialog" aria-label="Réglages" style={{ position: 'fixed', inset: 0, zIndex: 650, background: 'rgba(0,0,0,0.5)', display: 'flex', justifyContent: 'flex-end' }} onClick={onClose}>
      <div style={{ width: 380, height: '100%', background: 'var(--surface-soft, #0b0b0c)', borderLeft: '1px solid var(--hairline)', padding: 16, overflowY: 'auto' }} onClick={e => e.stopPropagation()}>
        <div className="card-head" style={{ marginBottom: 12 }}>Réglages</div>
        {FIELDS.map(f => (
          <div key={f.key} className="form-row" style={{ flexDirection: 'column', alignItems: 'stretch', gap: 2, marginBottom: 10 }}>
            <label className="mono" style={{ fontSize: 10, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--text-muted)' }}>
              {f.label}{f.unit ? ` (${f.unit})` : ''}
            </label>
            {f.key === 'target' ? (
              <input value={String(s[f.key])} maxLength={64} onChange={e => setS({ ...s, target: e.target.value })} style={{ background: 'var(--surface-card)', color: 'var(--text-body)', border: '1px solid var(--hairline)', padding: '6px 8px', fontFamily: 'var(--font-mono)', fontSize: 11 }} />
            ) : (
              <input type="number" min={f.min} max={f.max} step={f.step ?? 1} value={Number(s[f.key])} onChange={e => setS({ ...s, [f.key]: parseFloat(e.target.value) || 0 })} style={{ width: 120, background: 'var(--surface-card)', color: 'var(--text-body)', border: '1px solid var(--hairline)', padding: '6px 8px', fontFamily: 'var(--font-mono)', fontSize: 11 }} />
            )}
            <span className="mono" style={{ fontSize: 9, color: 'var(--text-faint)' }}>{f.hint}</span>
          </div>
        ))}
        <div className="form-row" style={{ gap: 8, marginTop: 12 }}>
          <button className="btn btn-primary" onClick={apply}>ENREGISTRER</button>
          <button className="btn" onClick={() => setS({ ...DEFAULTS })}>DÉFAUTS</button>
          <button className="btn" onClick={onClose}>ANNULER</button>
        </div>
      </div>
    </div>
  )
}
