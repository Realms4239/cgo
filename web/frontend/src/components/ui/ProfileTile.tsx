// ProfileTile — carte de profil complète : le nom entier reste lisible,
// jamais de troncature. L'état sélectionné se dit par la bordure, la coche
// ET le mot, jamais par la couleur seule.
export interface ProfileTileData {
  title: string
  spec?: string
  meta?: string
}

export function ProfileTile({ tile, selected, onToggle }: {
  tile: ProfileTileData
  selected: boolean
  onToggle: () => void
}) {
  return (
    <button
      type="button"
      className="profile-tile ui-profile-tile"
      aria-pressed={selected}
      data-selected={selected ? 'true' : 'false'}
      onClick={onToggle}
    >
      <span className="profile-tile-title" style={{ textOverflow: 'clip', overflowWrap: 'anywhere' }}>{tile.title}</span>
      {tile.spec ? <span className="profile-tile-spec" style={{ overflowWrap: 'anywhere' }}>{tile.spec}</span> : null}
      {tile.meta ? <span className="profile-tile-meta" style={{ overflowWrap: 'anywhere' }}>{tile.meta}</span> : null}
      <span className="ui-profile-state" aria-hidden="true">{selected ? '✓' : '○'}</span>
      {selected ? <span className="ui-profile-word">sélectionné</span> : null}
    </button>
  )
}
