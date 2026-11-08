// Configurer — tout ce qui change le résultat, une seule fois. Les imports
// de profil restent disponibles dans « Avancé ».
import { useEffect, useState } from 'react'
import { ArmButton } from './ArmButton'
import Explain from './Explain'
import { InlineField } from './InlineField'
import { PeekPopover } from './PeekPopover'
import { BenchCapabilities } from './BenchCapabilities'
import { DuelCard } from './DuelCard'
import ProfileTiles from './ProfileTiles'
import { useBenchStore } from '../store/bench'
import { useUIStore } from '../store/ui'
import { live as liveRing } from '../lib/live'
import { CampaignStartError, loadDuel, startCampaign, type DuelPair } from '../lib/bench'
import { loadSettings } from '../lib/settings'
import { validate } from '../lib/validation'
import { formatBytesFR, groupDigits } from '../lib/format'
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
  const research = useBenchStore(s => s.research)
  const setResearch = useBenchStore(s => s.setResearch)

  const [duel, setDuel] = useState<DuelPair | null>(null)

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

  useEffect(() => {
    let alive = true
    loadDuel('P2')
      .then(pair => { if (alive) setDuel(pair) })
      .catch(() => { if (alive) setDuel(null) })
    return () => { alive = false }
  }, [])

  // la deadline suit les Réglages tant que l'opérateur n'a pas lancé.
  useEffect(() => {
    const on = () => setDeadlineMs(loadSettings().deadlineMs)
    window.addEventListener('meteolink-settings', on)
    return () => window.removeEventListener('meteolink-settings', on)
  }, [setDeadlineMs])

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

  const launch = async (opts: {
    profiles: string[]
    reps: number
    flows: number
    channel_model: typeof effectiveModel
    deadline_ms: number
    target: string
  }) => {
    setMsg('démarrage…')
    try {
      await startCampaign(opts)
      setMsg('campagne lancée')
      pushToast('Campagne lancée', 'ok')
      useUIStore.getState().setBenchStage('observe')
    } catch (e) {
      const err = e as Partial<CampaignStartError>
      const txt = typeof err?.message === 'string' ? err.message : 'réseau'
      const status = typeof err?.status === 'number' ? err.status : undefined
      setMsg('échec : ' + txt.slice(0, 120))
      pushToast(txt.includes('déjà') || status === 409 ? "Campagne déjà en cours — arrêter d'abord" : `Échec démarrage : ${status ?? txt.slice(0, 60)}`, 'err')
    }
  }

  const start = () => launch({
    profiles,
    reps,
    flows,
    channel_model: effectiveModel,
    deadline_ms: deadlineMs,
    target: loadSettings().target,
  })

  const launchDuel = () => launch({
    profiles: ['P2'],
    reps: 3,
    flows: 4,
    channel_model: 'clean',
    deadline_ms: duel?.deadline_ms ?? 220,
    target: loadSettings().target,
  })

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

  const wasted = live?.running ? live?.wasted_bytes : undefined
  const liveCostAr = live?.running ? live?.cost_ar_per_h : undefined
  const liveCost = wasted != null && liveCostAr != null
  const peekCost = liveCost
    ? `${formatBytesFR(wasted).num} ${formatBytesFR(wasted).unit} gâchés · ${groupDigits(liveCostAr, 2)} Ar/h`
    : 'sans campagne en cours'

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

      <DuelCard pair={duel} onLaunch={launchDuel} disabled={!!live?.running} />

      <div className="card bench-config-card">
        <div className="card-head bench-config-head">
          <span>Campagne — ce qui change le résultat</span>
          <button
            type="button"
            className="btn bench-research"
            data-research
            aria-pressed={research}
            onClick={() => setResearch(!research)}
          >
            Mode recherche
          </button>
        </div>
        <ProfileTiles
          profiles={allProfiles}
          selected={profiles}
          onToggle={id => setProfiles(profiles.includes(id) ? profiles.filter(x => x !== id) : [...profiles, id])}
        />
        {research ? (
          <>
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
              <input className="c-input bench-num" type="number" min={200} max={5000} step={100} value={deadlineMs} onChange={e => setDeadlineMs(parseInt(e.target.value) || 220)} title="objectif small p95 — voyage avec la campagne (bornes serveur 200–5000 ms)" />
              <span className="mono muted bench-hint">ms</span>
            </div>
          </>
        ) : null}
        <div
          className="form-row bench-cost"
          onMouseEnter={e => setPeek({ rect: e.currentTarget.getBoundingClientRect(), data: liveRing.small.slice(-20).map(([, v]: [number, number]) => v) })}
          onMouseLeave={() => setPeek(null)}
        >
          <span className="mono bench-cost-label">coût en cours</span>
          <span className="mono bench-cost-val" data-measured={liveCost ? '1' : '0'}>{peekCost}</span>
          <span className="mono muted bench-hint bench-cost-note">octets gâchés × tarif unique</span>
        </div>
        <div className="form-row bench-launch">
          <span title={live?.running ? "campagne en cours — arrêter d'abord" : profiles.length === 0 ? 'choisir au moins un profil' : undefined}>
            <ArmButton label="DÉMARRER" onConfirm={start} disabled={live?.running || profiles.length === 0} />
          </span>
          <ArmButton label="ARRÊTER" confirmLabel="CONFIRMER L'ARRÊT" onConfirm={stop} disabled={!live?.running} />
          <span className="mono muted">{profiles.length === 0 ? 'Sélectionnez au moins un profil' : msg}</span>
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
