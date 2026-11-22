// Configurer — la déclaration et les profils du référentiel produit. Aucune
// valeur n'est fabriquée : une déclaration absente garde son mot. Les contrôles
// ci-dessous ne sélectionnent qu'une configuration ; ils ne mesurent rien.
import { useEffect, useState } from 'react'
import { PROFILE_TILES } from '../../lib/profiles'
import { ProfileTile } from '../ui/ProfileTile'
import { InlineDisclosure } from '../ui/InlineDisclosure'
import { MetricRow } from '../ui/MetricRow'
import { StateSlot } from '../ui/StateSlot'
import { BenchCapabilities } from '../BenchCapabilities'
import { useCapabilities } from '../../lib/hooks'
import { loadSchema, param, type Param } from '../../lib/schema'
import type { Capabilities } from '../../lib/capabilities'

// Projection honnête de GET /api/calibrations : le serveur ne renvoie déjà que
// des sidecars vérifiés, le filtre client garde la porte fermée au repos.
export interface Calibration {
  name: string
  sha256: string
  samples: number
  p50_ms: number
  verified: boolean
}

function useSchemaParams(): Param[] {
  const [rows, setRows] = useState<Param[]>([])
  useEffect(() => {
    let alive = true
    loadSchema().then(p => { if (alive) setRows(p) }).catch(() => {})
    return () => { alive = false }
  }, [])
  return rows
}

function useCalibrations(): Calibration[] {
  const [rows, setRows] = useState<Calibration[]>([])
  useEffect(() => {
    let alive = true
    fetch('/api/calibrations').then(r => r.json())
      .then(j => { if (alive) setRows(Array.isArray(j?.calibrations) ? j.calibrations : []) })
      .catch(() => {})
    return () => { alive = false }
  }, [])
  return rows
}

export interface ConfigurerProps {
  site?: string
  link?: string
  duration?: number
  // Points d'entrée de test : la vue vit des endpoints réels en production.
  caps?: Capabilities | null
  params?: Param[] | null
  calibrations?: Calibration[] | null
}

const CHANNEL_LABELS: Record<string, string> = { statique: 'Statique', empirique: 'Empirique .dist' }
const DIRECTION_LABELS: Record<string, string> = { up: 'Egress WAN', down: 'Ingress IFB', both: 'Les deux' }

// Les libellés viennent du corpus, les valeurs du schéma : une valeur inconnue
// n'a pas d'option, donc pas de contrôle fabriqué.
const optionsOf = (p: Param | undefined, labels: Record<string, string>) =>
  (p?.enum ?? []).filter(v => typeof v === 'string' && labels[v])

const defaultOf = (p: Param | undefined, options: string[]) =>
  p?.default != null && options.includes(String(p.default)) ? String(p.default) : (options[0] ?? '')

export default function Configurer({ site, link, duration, caps, params, calibrations }: ConfigurerProps) {
  const [selected, setSelected] = useState<string[]>([])
  const [choices, setChoices] = useState<Record<string, string>>({})
  const liveCaps = useCapabilities()
  const liveParams = useSchemaParams()
  const liveCalibrations = useCalibrations()
  const host = caps === undefined ? liveCaps : caps
  const schema = params === undefined ? liveParams : (params ?? [])
  const cals = calibrations === undefined ? liveCalibrations : (calibrations ?? [])

  const toggle = (id: string) => setSelected(s =>
    s.includes(id) ? s.filter(x => x !== id) : [...s, id])

  const optionControl = (key: string, label: string, labels: Record<string, string>, word: string) => {
    const p = param(key, schema)
    const options = optionsOf(p, labels)
    const value = choices[key] ?? defaultOf(p, options)
    return (
      <div className="configurer-control" data-control={key}>
        <span className="configurer-control-label">{label}</span>
        {options.length ? (
          <div className="configurer-options" role="group" aria-label={label}>
            {options.map(v => (
              <button
                key={v} type="button" className="configurer-option"
                aria-pressed={value === v}
                onClick={() => setChoices(c => ({ ...c, [key]: v }))}
              >{labels[v]}</button>
            ))}
          </div>
        ) : (
          <span className="configurer-absent">— {word}</span>
        )}
      </div>
    )
  }

  const distParam = param('dist_name', schema)
  const verified = cals.filter(c => c && c.verified === true && typeof c.name === 'string' && c.name.length > 0)
  const distNames = verified.map(c => c.name)
  const distOptions = distParam?.enum?.length
    ? distNames.filter(n => distParam.enum!.includes(n))
    : distNames
  const distValue = choices.dist_name ?? ''
  const picked = verified.find(c => c.name === distValue)

  return (
    <div className="configurer">
      <p className="configurer-context">Ces valeurs viennent du rapport réel ; une déclaration absente garde son mot.</p>
      <MetricRow label="Déclaration">
        <StateSlot label="Site" value={site || null} word="site non déclaré" />
        <StateSlot label="Lien" value={link || null} word="lien non déclaré" />
        <StateSlot
          label="Durée"
          value={typeof duration === 'number' && duration > 0 ? `${duration} s` : null}
          word="durée non fixée"
        />
      </MetricRow>
      <div className="configurer-caps">
        <BenchCapabilities caps={host} />
      </div>
      <div className="configurer-controls">
        {optionControl('link_channel', 'Canal', CHANNEL_LABELS, 'canal indisponible')}
        {optionControl('link_direction', 'Direction', DIRECTION_LABELS, 'direction indisponible')}
        <div className="configurer-control" data-control="distribution">
          <span className="configurer-control-label">Distribution</span>
          {distOptions.length ? (
            <>
              <select
                className="configurer-select" aria-label="Distribution"
                value={distValue}
                onChange={e => setChoices(c => ({ ...c, dist_name: e.target.value }))}
              >
                {distOptions.map(n => <option key={n} value={n}>{n}</option>)}
              </select>
              {picked ? (
                <span className="configurer-calib-detail">
                  {picked.name} — {picked.samples} échantillons · p50 {picked.p50_ms} ms
                </span>
              ) : null}
            </>
          ) : (
            <>
              <span className="configurer-absent">— aucune calibration vérifiée</span>
              <select className="configurer-select" aria-label="Distribution" disabled value="" onChange={() => {}} />
            </>
          )}
        </div>
      </div>
      <InlineDisclosure label="Profils de campagne">
        <div className="profile-tiles" role="group" aria-label="Profils de campagne">
          {PROFILE_TILES.map(t => (
            <ProfileTile
              key={t.id}
              tile={{ title: t.title, spec: t.spec, meta: `${t.loss} · ${t.use}` }}
              selected={selected.includes(t.id)}
              onToggle={() => toggle(t.id)}
            />
          ))}
        </div>
      </InlineDisclosure>
    </div>
  )
}
