// Configurer — la déclaration et les profils du référentiel produit. Aucune
// valeur n'est fabriquée : une déclaration absente garde son mot.
import { useState } from 'react'
import { PROFILE_TILES } from '../../lib/profiles'
import { ProfileTile } from '../ui/ProfileTile'
import { InlineDisclosure } from '../ui/InlineDisclosure'
import { MetricRow } from '../ui/MetricRow'
import { StateSlot } from '../ui/StateSlot'

export interface ConfigurerProps {
  site?: string
  link?: string
  duration?: number
}

export default function Configurer({ site, link, duration }: ConfigurerProps) {
  const [selected, setSelected] = useState<string[]>([])
  const toggle = (id: string) => setSelected(s =>
    s.includes(id) ? s.filter(x => x !== id) : [...s, id])

  return (
    <div className="configurer">
      <MetricRow label="Déclaration">
        <StateSlot label="Site" value={site || null} word="site non déclaré" />
        <StateSlot label="Lien" value={link || null} word="lien non déclaré" />
        <StateSlot
          label="Durée"
          value={typeof duration === 'number' && duration > 0 ? `${duration} s` : null}
          word="durée non fixée"
        />
      </MetricRow>
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
