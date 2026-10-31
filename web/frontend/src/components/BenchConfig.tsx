// Configurer — tout ce qui change le résultat, une seule fois. Les imports
// de profil restent disponibles dans « Avancé ».
import { useEffect, useState } from 'react'
import { ArmButton } from './ArmButton'
import Explain from './Explain'
import { InlineField } from './InlineField'
import { PeekPopover } from './PeekPopover'
import { BenchCapabilities } from './BenchCapabilities'
import { useBenchStore } from '../store/bench'
import { useUIStore } from '../store/ui'
import { live as liveRing } from '../lib/live'
import { loadSettings } from '../lib/settings'
import { validate } from '../lib/validation'
import { formatBytesFR, groupDigits } from '../lib/format'
import { profileShort } from '../lib/profiles'
import { canalGate, effectiveChannelModel } from '../lib/canal'
import { useCapabilities } from '../lib/hooks'

export default function BenchConfig() {
  const caps = useCapabilities()
  const live = useUIStore(s => s.live)
  const pushToast = useUIStore(s => s.pushToast)
  const profiles = useBenchStore(s => s.profiles)
  const setProfiles = useBenchStore(s => s.setProfiles)
  const reps = useBenchStore(s => s.reps)
  const setReps = useBenchStore(s => s.setReps)
  const flows = useBenchStore(s => s.flows)
  const setFlows = useBenchStore(s => s.setFlows)
  const channelModel = useBenchStore(s => s.channelModel)
  const setChannelModel = useBenchStore(s => s.setChannelModel)
  const deadlineMs = useBenchStore(s => s.deadlineMs)
  const setDeadlineMs = useBenchStore(s => s.setDeadlineMs)

  const [allProfiles, setAllProfiles] = useState<{ id: string; imported: boolean }[]>([
    { id: 'P1', imported: false },
    { id: 'P2', imported: false },
  ])
  const [msg, setMsg] = useState('')
  const [importMsg, setImportMsg] = useState('')
  const [importForm, setImportForm] = useState(false)
  const [impId, setImpId] = useState('')
  const [impCap, setImpCap] = useState(20)
  const [impCapUp, setImpCapUp] = useState(0)
  const [impDelay, setImpDelay] = useState(100)
  const [impJitter, setImpJitter] = useState(2)
  const [impLoss, setImpLoss] = useState(0)
  const [peek, setPeek] = useState<{ rect: DOMRect; data: number[] } | null>(null)

  useEffect(() => {
    fetch('/api/profiles')
      .then(r => r.json())
      .then(j => {
        if (j?.profiles?.length) setAllProfiles(j.profiles.map((x: any) => ({ id: x.id, imported: !!x.imported })))
      })
      .catch(() => {})
  }, [])

  // la deadline suit les Réglages tant que l'opérateur n'a pas lancé.
  useEffect(() => {
    const on = () => setDeadlineMs(loadSettings().deadlineMs)
    window.addEventListener('meteolink-settings', on)
    return () => window.removeEventListener('meteolink-settings', on)
  }, [setDeadlineMs])

  const refProfiles = allProfiles.filter(p => /^P[1-4]$/.test(p.id))
  const canal = canalGate(profiles)
  const effectiveModel = effectiveChannelModel(channelModel, canal)
  useEffect(() => {
    if (!canal.allowed && channelModel === 'bursty') setChannelModel('clean')
  }, [canal.allowed, channelModel, setChannelModel])

  const importValidation = validate(
    { id: impId, capacity: impCap, delay: impDelay, jitter: impJitter, loss: impLoss },
    { id: { required: true }, capacity: { min: 0.1 }, delay: { min: 1 }, jitter: { min: 0 }, loss: { min: 0 } },
  )

  const importProfile = () => {
    if (!importValidation.valid) return
    const body: any = { id: impId, capacity_mbps: impCap, delay_ms: impDelay, jitter_ms: impJitter, loss_pct: impLoss }
    if (impCapUp > 0) body.capacity_up_mbps = impCapUp
    fetch('/api/profile/import', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
      .then(r => { if (!r.ok) throw new Error(String(r.status)); return r.json() })
      .then(() => {
        setImportMsg(`profil ${impId} importé`)
        pushToast(`Profil ${impId} importé`, 'ok')
        setImportForm(false)
      })
      .catch(e => {
        setImportMsg('échec : ' + e.message)
        pushToast('Échec import : ' + e.message, 'err')
      })
  }

  const start = async () => {
    setMsg('démarrage…')
    try {
      const r = await fetch('/api/run/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profiles,
          reps,
          flows,
          channel_model: effectiveModel,
          deadline_ms: deadlineMs,
          target: loadSettings().target,
        }),
      })
      if (!r.ok) {
        const txt = await r.text().catch(() => String(r.status))
        setMsg('échec : ' + txt.slice(0, 120))
        pushToast(txt.includes('déjà') || r.status === 409 ? "Campagne déjà en cours — arrêter d'abord" : 'Échec démarrage : ' + r.status, 'err')
        return
      }
      setMsg('campagne lancée')
      pushToast('Campagne lancée', 'ok')
      useUIStore.getState().setBenchStage('observe')
    } catch (e: any) {
      setMsg('échec : ' + (e?.message ?? 'réseau'))
      pushToast('Échec démarrage', 'err')
    }
  }

  const stop = async () => {
    await fetch('/api/run/stop', { method: 'POST' })
    setMsg('arrêt demandé')
    pushToast('Arrêt demandé', 'blue')
  }

  const toProfile = async () => {
    try {
      const r = await fetch('/api/audit/toprofile', { method: 'POST' })
      const j = await r.json().catch(() => ({}))
      if (!r.ok) { pushToast(j?.error ?? 'échec import audit', 'err'); return }
      pushToast(`Profil ${j.profile?.id ?? 'P-audit'} importé depuis l'audit — ${j.profile?.capacity_mbps} Mbit/s, ${j.profile?.delay_ms} ms, perte ${j.profile?.loss_pct ?? 0} %`, 'ok')
    } catch {
      pushToast('échec import audit', 'err')
    }
  }

  const peekCost = live?.wasted_bytes != null && live.cost_ar_per_h != null
    ? `${formatBytesFR(live.wasted_bytes).num} ${formatBytesFR(live.wasted_bytes).unit} gâchés · ${groupDigits(live.cost_ar_per_h, 2)} Ar/h`
    : '—'

  return (
    <div className="bench-config">
      {peek ? (
        <PeekPopover rect={peek.rect}>
          <div className="mono live-peek-label">live small — 20 pts</div>
          <div className="bench-peek-spark">
            {peek.data.map((v, i) => (
              <i key={i} style={{ flex: 1, height: `${Math.max(2, Math.min(28, (v / 200) * 28))}px`, background: 'var(--t-threshold)', borderRadius: 1, opacity: 0.7 + (i / peek.data.length) * 0.3 }} />
            ))}
          </div>
          <div className="mono bench-peek-val">{peek.data.at(-1)?.toFixed(1) ?? '—'} ms</div>
        </PeekPopover>
      ) : null}

      <div className="card bench-config-card">
        <div className="card-head">Campagne — ce qui change le résultat</div>
        <div className="form-row">
          <label>Profils</label>
          <div className="check-row bench-profiles">
            {refProfiles.map(p => (
              <label key={p.id} className="bench-profile">
                <input
                  type="checkbox"
                  checked={profiles.includes(p.id)}
                  onChange={e => setProfiles(e.target.checked ? [...profiles, p.id] : profiles.filter(x => x !== p.id))}
                />
                <span>
                  <Explain term={p.id}>{p.id}</Explain>
                  <small>{profileShort(p.id)}</small>
                </span>
              </label>
            ))}
          </div>
        </div>
        <BenchCapabilities caps={caps} />
        <div className="form-row">
          <label>Répétitions</label>
          <input className="c-input bench-num" type="number" min={1} max={5} value={reps} onChange={e => setReps(parseInt(e.target.value) || 1)} />
        </div>
        <div className="form-row">
          <label title=">1 sature un lien avec perte où un CUBIC mono-flux s'effondre">Flux de charge</label>
          <input className="c-input bench-num" type="number" min={1} max={8} value={flows} onChange={e => setFlows(Math.min(8, Math.max(1, parseInt(e.target.value) || 4)))} />
          <span className="mono muted bench-hint">1–8 · défaut 4</span>
        </div>
        <div className="form-row">
          <label><Explain term="canal">Canal</Explain></label>
          <div role="group" aria-label="Modèle de canal" className="bench-pills">
            {(['clean', 'bursty'] as const).map(m => {
              const blocked = m === 'bursty' && !canal.allowed
              const active = effectiveModel === m
              return (
                <button
                  key={m}
                  type="button"
                  className="btn"
                  aria-pressed={active}
                  disabled={blocked}
                  title={blocked ? canal.reason : m === 'bursty' ? 'rafales — canal propre P2/P3' : 'canal propre, sans rafales'}
                  onClick={() => setChannelModel(m)}
                  style={{
                    padding: 'var(--sp-1) var(--sp-2)',
                    borderColor: active ? 'var(--t-live)' : 'var(--hairline)',
                    color: active ? 'var(--t-live)' : 'var(--text-muted)',
                    background: active ? 'var(--live-dim)' : 'transparent',
                    opacity: blocked ? 'var(--disabled-opacity, 0.5)' : 1,
                    cursor: blocked ? 'not-allowed' : 'pointer',
                  }}
                >
                  {m}
                </button>
              )
            })}
          </div>
        </div>
        {!canal.allowed ? <div className="mono muted bench-hint" data-testid="canal-reason">{canal.reason}</div> : null}
        <div className="form-row">
          <label><Explain term="deadline">Deadline</Explain></label>
          <input className="c-input bench-num" type="number" min={200} max={5000} step={100} value={deadlineMs} onChange={e => setDeadlineMs(parseInt(e.target.value) || 1000)} title="objectif small p95 — voyage avec la campagne (bornes serveur 200–5000 ms)" />
          <span className="mono muted bench-hint">ms</span>
        </div>
        <div
          className="form-row bench-cost"
          onMouseEnter={e => setPeek({ rect: e.currentTarget.getBoundingClientRect(), data: liveRing.small.slice(-20).map(([, v]: [number, number]) => v) })}
          onMouseLeave={() => setPeek(null)}
        >
          <span className="mono bench-cost-label">coût en cours</span>
          <span className="mono bench-cost-val" data-measured={live?.wasted_bytes != null ? '1' : '0'}>{peekCost}</span>
          <span className="mono muted bench-hint bench-cost-note">octets gâchés × tarif unique</span>
        </div>
        <div className="form-row bench-launch">
          <span title={live?.running ? "campagne en cours — arrêter d'abord" : profiles.length === 0 ? 'choisir au moins un profil' : undefined}>
            <ArmButton label="DÉMARRER" onConfirm={start} disabled={live?.running || profiles.length === 0} />
          </span>
          <ArmButton label="ARRÊTER" confirmLabel="CONFIRMER L'ARRÊT" onConfirm={stop} disabled={!live?.running} />
          <span className="mono muted">{profiles.length === 0 ? 'aucun profil sélectionné' : msg}</span>
        </div>
      </div>

      <details className="bench-advanced">
        <summary className="mono">AVANCÉ — profil personnalisé, audit → profil</summary>
        <div className="bench-advanced-body">
          <div className="card">
            <div className="card-head bench-adv-head">
              <span>Profil personnalisé</span>
              {!importForm ? (
                <>
                  <button className="btn bench-mini" onClick={toProfile} title="transforme le dernier audit du lien réel en profil rejouable sur le banc">AUDIT → PROFIL</button>
                  <button className="btn bench-mini" onClick={() => setImportForm(true)}>IMPORTER</button>
                  <span className="mono muted bench-hint">{importMsg}</span>
                </>
              ) : null}
            </div>
            {importForm ? (
              <div className="bench-import">
                <InlineField label="Identifiant" error={importValidation.errors.id} helper="jamais P1–P4 (référentiel protégé) — ex. P-site-fibre">
                  <input className="c-input" name="import-id" value={impId} onChange={e => setImpId(e.target.value)} placeholder="P-site-fibre" />
                </InlineField>
                <InlineField label="Capacité (Mbit/s)" error={importValidation.errors.capacity} helper="> 0">
                  <input className="c-input bench-num" type="number" min={0.1} step={0.1} value={impCap} onChange={e => setImpCap(parseFloat(e.target.value) || 0)} />
                </InlineField>
                <InlineField label="Cap. montante" error={undefined} helper="0 = symétrique — ex. 5 pour 4G 20/5">
                  <input className="c-input bench-num" type="number" min={0} step={0.1} value={impCapUp} onChange={e => setImpCapUp(parseFloat(e.target.value) || 0)} />
                </InlineField>
                <InlineField label="RTT (ms)" error={importValidation.errors.delay} helper="> 0">
                  <input className="c-input bench-num" type="number" min={1} value={impDelay} onChange={e => setImpDelay(parseInt(e.target.value) || 0)} />
                </InlineField>
                <InlineField label="Gigue (ms)" error={importValidation.errors.jitter} helper="≥ 0 — ex. 30 pour VSAT">
                  <input className="c-input bench-num" type="number" min={0} step={0.5} value={impJitter} onChange={e => setImpJitter(parseFloat(e.target.value) || 0)} />
                </InlineField>
                <InlineField label="Perte (%)" error={importValidation.errors.loss} helper="≥ 0 — ex. 1 pour VSAT">
                  <input className="c-input bench-num" type="number" min={0} step={0.1} value={impLoss} onChange={e => setImpLoss(parseFloat(e.target.value) || 0)} />
                </InlineField>
                <div className="form-row bench-launch">
                  <ArmButton label="CONFIRMER L'IMPORT" onConfirm={importProfile} disabled={!importValidation.valid} />
                  <button className="btn" onClick={() => setImportForm(false)}>ANNULER</button>
                </div>
              </div>
            ) : null}
          </div>
        </div>
      </details>
    </div>
  )
}
