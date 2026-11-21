// Portes de structure du document Labo retargetées T6 : quatre chapters dans
// un seul flux, un en-tête de commande épinglé, deux modes seulement, et les
// anciennes vues disparues — pas seulement démontées.
import { describe, expect, it } from 'vitest'
// @ts-ignore — types node:fs inutiles sous jsdom.
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const SRC = join(process.cwd(), 'src')
const read = (p: string): string => readFileSync(join(SRC, p), 'utf8')
const absent = (...parts: string[]) => expect(existsSync(join(SRC, ...parts)), parts.join('/')).toBe(false)

describe('document Labo linéaire', () => {
  it('quatre chapters nommés, aucune taxonomie imbriquée', () => {
    const b = read('views/LaboView.tsx')
    for (const id of ['configurer', 'observer', 'comparer', 'preuves']) {
      expect(b, `chapter ${id}`).toContain(`id="${id}"`)
    }
    expect(read('components/labo/Chapter.tsx')).toContain('id={`chapter-${id}`}')
    expect(b).toContain('<CommandHeader')
    expect(b).toContain('<PipelineRail')
    expect(b).toContain('<Configurer')
    expect(b).toContain('<Observer')
    expect(b).toContain('<Comparer')
    expect(b).toContain('<Preuves')
    expect(b, 'aucun sous-onglet').not.toContain('role="tablist"')
  })

  it('Configurer garde chaque décision de campagne', () => {
    const c = read('components/labo/Configurer.tsx')
    for (const token of ['PROFILE_TILES', 'ProfileTile', 'InlineDisclosure', 'StateSlot', 'Déclaration']) {
      expect(c, token).toContain(token)
    }
  })

  it('Observer garde l’instrument et le journal', () => {
    const o = read('components/labo/Observer.tsx')
    expect(o).toContain('<TimeField rings={instrument}')
    expect(o).toContain('<JournalRail events={instrument.events}')
  })

  it('les anciennes vues et le pont historique ont disparu', () => {
    absent('views', 'CampagneView.tsx')
    absent('views', 'DashboardView.tsx')
    absent('views', 'AuditView.tsx')
    absent('components', 'VerdictHero.tsx')
    absent('components', 'PhaseStepper.tsx')
    for (const f of ['LiveView', 'BenchView', 'EvidenceView', 'ResultatsView', 'IntegriteView']) {
      absent('views', `${f}.tsx`)
    }
    for (const f of ['DonutJFI', 'EventBadge', 'Beam',
      'CompareView', 'InterpretationView', 'PromptProgressLine', 'LiveGuide']) {
      absent('components', `${f}.tsx`)
    }
    const css = read('styles/index.css')
    for (const dead of ['promptDrain', '.prompt-progress', '.interp-card', '.live-guide']) {
      expect(css, `fixture CSS morte : ${dead}`).not.toContain(dead)
    }
    expect(css).not.toContain('.cockpit')
  })

  it('deux modes, laboratoire différé, sans identifiant historique', () => {
    const app = read('App.tsx')
    expect(app).toContain("import SynthèseView from './views/SynthèseView'")
    expect(app, 'le laboratoire reste différé').toContain("lazy(() => import('./views/LaboView'))")
    expect(app).toContain('<SynthèseView />')
    expect(app).toContain('<LaboView />')
    expect(app).toContain('<Suspense')
    expect(app).toContain('id="v-cockpit"')
    expect(app).toContain('id="v-labo"')
    expect(app.startsWith('import { lazy, Suspense'), 'premier rendu léger').toBe(true)
    // Un identifiant de mode est une chaîne littérale OU un adaptateur nommé.
    // Les chemins d'import (`./store/audit`) et les noms de données
    // (`benchStage`, `audit_id`) ne sont pas des modes et restent légitimes.
    for (const lit of ["'dashboard'", "'audit'", "'bench'", "'evidence'",
      '"dashboard"', '"audit"', '"bench"', '"evidence"', 'v-bench', 'v-dashboard', 'v-evidence',
      'legacyMode', 'shellMode', 'AnyModeId']) {
      expect(app, lit).not.toContain(lit)
    }
    const store = read('store/ui.ts')
    for (const legacy of ['LEGACY_MODES', 'LegacyModeId', 'AnyModeId', 'LEGACY_TO_MODE', 'shellMode', 'legacyMode']) {
      expect(store, legacy).not.toContain(legacy)
    }
    // Le préchauffage revient, mais deux modes seulement : le chunk nommé est
    // Labo, et aucun identifiant historique n'entre dans la table d'imports.
    expect(store).toContain("import { preloadMode } from '../lib/preload'")
    expect(store).toContain('preloadMode(mode)')
    const preload = read('lib/preload.ts')
    expect(preload).toContain("labo: () => import('../views/LaboView')")
    for (const lit of ["'dashboard'", "'audit'", "'bench'", "'evidence'"]) {
      expect(preload, lit).not.toContain(lit)
    }
    const palette = read('components/CommandPalette.tsx')
    expect(palette).toContain('MODES.map')
    expect(palette).not.toContain('LEGACY_MODES')
    const nav = read('components/shell/ModeSwitch.tsx')
    expect(nav).toContain('MODES.map')
    expect(nav).toContain('preloadMode(m.id)')
    expect(nav).not.toContain('shellMode')
  })

  it('hors destination, l’instrument ne se re-rend pas au rythme du SSE', () => {
    expect(read('components/cockpit/TimeField.tsx')).toContain('useRafLoop')
    expect(read('components/cockpit/TimeField.tsx')).toContain('rings.small.dirty')
    expect(read('components/cockpit/JournalRail.tsx')).toContain('events.dirty')
    expect(read('components/labo/Comparer.tsx')).not.toContain('useRafLoop')
  })

  it('une commande sans matière à agir se nomme avant le clic', () => {
    const p = read('components/labo/Preuves.tsx')
    expect(p).toContain('Preuve manquante')
    const s = read('components/labo/SealInspector.tsx')
    expect(s).toContain('Vérifier le sceau')
    expect(s).toContain('integrityVerified(j)')
    expect(s).toContain("fetch('/api/integrity')")
    expect(s, 'jamais vérifié par présence seule').not.toContain('setVerified(true)')
    const r = read('components/ArchiveTable.tsx')
    expect(r).toContain('Preuve manquante')
  })

  it('la commande d’arrêt reste épinglée dans l’en-tête', () => {
    const h = read('components/labo/CommandHeader.tsx')
    expect(h).toContain('<header className="command-header" data-pinned="true"')
    expect(h).toContain('ARRÊTER')
    expect(h).toContain('ARMER')
    const css = read('styles/components.css')
    expect(css).toContain('.command-header { position: sticky; top: 0;')
  })
})
