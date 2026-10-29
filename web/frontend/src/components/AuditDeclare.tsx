// Déclarer — quatre décisions visibles, le reste replié. Les presets
// remplissent le formulaire ; ils ne lancent jamais l'audit.
import { useRef } from 'react'
import { ArmButton } from './ArmButton'
import { InlineField } from './InlineField'
import { animateShake } from '../lib/anime'
import {
  AUDIT_DURATIONS,
  durationChoice,
  directionLabel,
  validateAuditDeclaration,
} from '../lib/auditFlow'
import { useAuditStore } from '../store/audit'

const LINK_TYPES = ['fiber', '5g', '4g', 'vsat', 'other']

export function AuditDeclare() {
  const s = useAuditStore()
  const formRef = useRef<HTMLDivElement>(null)
  const validation = validateAuditDeclaration({
    site: s.site,
    link_type: s.link,
    duration: s.duration,
    direction: s.direction,
  })
  const custom = durationChoice(s.duration) === 'custom'

  const launch = () => {
    if (!validation.valid) {
      if (formRef.current) animateShake(formRef.current)
      return
    }
    void s.start()
  }

  const toggle = (
    label: string,
    active: boolean,
    onClick: () => void,
    ariaLabel: string,
  ) => (
    <button
      key={label}
      type="button"
      className="btn"
      aria-pressed={active}
      aria-label={ariaLabel}
      onClick={onClick}
      style={{
        padding: 'var(--sp-1) var(--sp-2)',
        borderColor: active ? 'var(--t-live)' : 'var(--hairline)',
        color: active ? 'var(--t-live)' : 'var(--text-muted)',
        background: active ? 'var(--live-dim)' : 'transparent',
      }}
    >
      {label}
    </button>
  )

  return (
    <div ref={formRef} className="audit-declare">
      {s.presets.length > 0 ? (
        <div className="audit-presets">
          {s.presets.map(p => (
            <button
              key={p.id}
              className="c-chip audit-preset"
              title={`${p.site} · ${p.link_type} · ${p.duration}s → ${p.target} — ${p.notes}`}
              onClick={() => s.applyPreset(p)}
            >
              ⚑ {p.label}
            </button>
          ))}
        </div>
      ) : null}
      {s.presetNote ? <div className="mono muted audit-preset-note">ⓘ {s.presetNote}</div> : null}

      <div className="audit-form">
        <InlineField label="Site" error={validation.errors.site} helper={!s.site ? 'requis' : undefined}>
          <input className="c-input" name="site" value={s.site} onChange={e => s.patch({ site: e.target.value })} />
        </InlineField>
        <InlineField label="Type de lien" error={validation.errors.link_type}>
          <select className="c-input" value={s.link} onChange={e => s.patch({ link: e.target.value })}>
            {LINK_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
          </select>
        </InlineField>
        <InlineField
          label="Durée"
          error={validation.errors.duration}
          helper={custom ? `durée libre : ${s.duration} s` : '20 ou 60 s — durée libre dans avancé'}
        >
          <div role="group" aria-label="Durée de l'audit" className="audit-pills">
            {AUDIT_DURATIONS.map(d =>
              toggle(`${d} s`, s.duration === d, () => s.patch({ duration: d }), `Durée ${d} secondes`),
            )}
          </div>
        </InlineField>
        <InlineField label="Sens" helper="fenêtre de charge exercée">
          <div role="group" aria-label="Sens de l'audit" className="audit-pills">
            {(['up', 'down', 'both'] as const).map(d =>
              toggle(directionLabel(d), s.direction === d, () => s.patch({ direction: d }), `Sens ${d}`),
            )}
          </div>
        </InlineField>
      </div>

      <button
        className="btn audit-advanced-toggle"
        aria-expanded={s.advanced}
        onClick={s.toggleAdvanced}
      >
        {s.advanced ? 'MASQUER LES CHAMPS AVANCÉS' : 'CHAMPS AVANCÉS'}
      </button>
      {s.advanced ? (
        <div className="audit-form">
          <InlineField label="Fournisseur" helper="ex. yas — gelé avec la ligne">
            <input className="c-input" name="audit-provider" value={s.provider} onChange={e => s.patch({ provider: e.target.value })} placeholder="yas" />
          </InlineField>
          <InlineField label="Cible ping" helper="vide = puits du banc détecté">
            <input className="c-input" name="audit-target" value={s.target} onChange={e => s.patch({ target: e.target.value })} placeholder="détecté" />
          </InlineField>
          <InlineField label="Interface Linux" helper="optionnel — ex. eth0, pour le script tc">
            <input className="c-input" name="audit-iface" value={s.iface} onChange={e => s.patch({ iface: e.target.value })} placeholder="eth0" />
          </InlineField>
          <InlineField label="Interface RouterOS" helper="optionnel — ex. ether1-wan">
            <input className="c-input" name="audit-ros-iface" value={s.rosIface} onChange={e => s.patch({ rosIface: e.target.value })} placeholder="ether1-wan" />
          </InlineField>
          <InlineField label="Durée libre (s)" helper="10–600 s">
            <input className="c-input" type="number" min={10} max={600} value={s.duration} onChange={e => s.patch({ duration: parseInt(e.target.value) || 0 })} />
          </InlineField>
        </div>
      ) : null}

      <div className="audit-launch">
        <ArmButton label="LANCER L'AUDIT" confirmLabel="CONFIRMER L'AUDIT" onConfirm={launch} disabled={!validation.valid} />
        <span className="mono muted">{s.msg}</span>
      </div>
    </div>
  )
}
