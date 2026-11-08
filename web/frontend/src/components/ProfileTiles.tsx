// Cartes sélectionnables — nom complet, spec complète, jamais de troncature.
import { PROFILE_TILES } from '../lib/profiles'

export type ProfileSelection = {
  profiles: Array<{ id: string; imported: boolean }>
  selected: string[]
  onToggle: (id: string) => void
}

export default function ProfileTiles({ profiles, selected, onToggle }: ProfileSelection) {
  // Les quatre profils de référence restent visibles même si l’API renvoie
  // une liste partielle : l’opérateur ne doit pas perdre un choix à cause d’un fetch raté.
  const imported = profiles.filter(profile => profile.imported)

  return (
    <div className="profile-stack">
      <div className="profile-tiles" role="group" aria-label="Profils de campagne">
        {PROFILE_TILES.map(tile => {
          const isSelected = selected.includes(tile.id)
          return (
            <button
              key={tile.id}
              type="button"
              className="profile-tile"
              aria-pressed={isSelected}
              data-selected={isSelected ? 'true' : 'false'}
              onClick={() => onToggle(tile.id)}
            >
              <span className="profile-tile-title">{tile.title}</span>
              <span className="profile-tile-spec">{tile.spec}</span>
              <span className="profile-tile-meta">{tile.loss} · {tile.use}</span>
            </button>
          )
        })}
      </div>
      {imported.length > 0 ? (
        <p className="profile-imported">Profils importés disponibles dans le mode recherche : {imported.map(profile => profile.id).join(', ')}</p>
      ) : null}
    </div>
  )
}
