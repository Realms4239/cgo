import AxeBuilder from '@axe-core/playwright'
import { expect, test, type ConsoleMessage, type Page } from '@playwright/test'

// Parcours réel du binaire Go servi par playwright.config.ts : aucun page.route,
// aucun stub, aucun hôte distant, aucune écriture dans le corpus gelé, aucun
// banc live. Le harnais transmet le Bearer : l'authentification reste active.
const VIEWPORTS = [
  { width: 890, height: 512 },
  { width: 1280, height: 768 },
  { width: 1920, height: 1080 },
] as const
const AXE_TAGS = ['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa']
const TARGETS = '[data-mode], .chrome-trigger, .btn-primary, .btn-danger, .modeswitch-btn, .pr-step, .band-proof, .facts-disclosure-trigger, .inline-disclosure-trigger, .duel-proof, .rank-pin, .rank-sort, .proof-close, .proof-tab, .journal-disclosure-toggle'

// Raison vérifiée : ce 429 vient du limiteur déclaré par principal
// (60 lectures/min, rafale 20), pas d'une erreur applicative. Les tests
// précédents du même fichier épuisent le seau de lecture partagé ; le
// navigateur journalise alors uniquement le statut 429 du chargement de
// ressource. On n'exclut QUE ce message de statut : toute autre erreur
// console, et tout pageerror, continue de faire échouer le test.
const RATE_LIMIT_RESOURCE = /status of 429\b/

function watchErrors(page: Page): string[] {
  const errs: string[] = []
  page.on('console', (m: ConsoleMessage) => {
    if (m.type() !== 'error') return
    if (RATE_LIMIT_RESOURCE.test(m.text())) return
    errs.push(m.text())
  })
  page.on('pageerror', (e) => errs.push('PAGEERROR ' + e.message))
  return errs
}

async function overflow(page: Page) {
  return page.evaluate(() => ({
    h: document.documentElement.scrollWidth - document.documentElement.clientWidth,
    v: Math.max(
      document.documentElement.scrollHeight - document.documentElement.clientHeight,
      document.body.scrollHeight - document.body.clientHeight,
    ),
  }))
}

async function inertShell(page: Page): Promise<boolean> {
  return page.evaluate(() => (document.getElementById('shell') as HTMLElement).inert)
}

// Le parcours intégrateur : une seule marche, du bandeau compact au tiroir de
// preuve, en passant par la persistance du mode, le rail clavier et les quatre
// chapitres. Les specs statiques couvrent le DOM figé ; ici on wiring les gestes.
test('parcours réel unique — deux modes, bandeau, rail clavier, chapitres et tiroir', async ({ page }) => {
  test.setTimeout(90000)
  const errors = watchErrors(page)
  await page.setViewportSize({ width: 890, height: 512 })
  await page.goto('/')

  // Synthèse compacte : mot, chiffre, état, PREUVE, faits repliés.
  await expect(page.locator('#v-cockpit')).toBeVisible()
  await expect(page.locator('#verdict-word')).toHaveText('MESURE ABSENTE')
  await expect(page.locator('.band-number-absent')).toHaveText('— non mesuré')
  await expect(page.locator('.band-status')).toHaveText('Non mesurée')
  const compact = await page.evaluate(() => {
    const band = document.querySelector('.answer-band') as HTMLElement
    const word = document.querySelector('#verdict-word') as HTMLElement
    const proof = document.querySelector('.band-proof') as HTMLElement
    const box = (e: HTMLElement) => e.getBoundingClientRect()
    return {
      bandHeight: box(band).height,
      wordSize: parseFloat(getComputedStyle(word).fontSize),
      proofW: box(proof).width,
      proofH: box(proof).height,
    }
  })
  expect(compact.bandHeight, 'bandeau compact').toBeGreaterThan(0)
  expect(compact.bandHeight, 'bandeau ≤ 140 px').toBeLessThanOrEqual(140)
  expect(compact.wordSize, 'mot du verdict à 32 px').toBe(32)
  expect(compact.proofW, 'PREUVE ≥ 44 px de large').toBeGreaterThanOrEqual(44)
  expect(compact.proofH, 'PREUVE ≥ 44 px de haut').toBeGreaterThanOrEqual(44)
  const o0 = await overflow(page)
  expect(o0.h, 'aucun débordement horizontal').toBeLessThanOrEqual(1)
  expect(o0.v, 'aucun scroll de page').toBeLessThanOrEqual(1)

  // Disclosure des quatre faits : aria-expanded pilote le flux, sans scroll.
  const facts = page.locator('.facts-disclosure-trigger')
  await expect(facts).toHaveAttribute('aria-expanded', 'false')
  await facts.click()
  await expect(facts).toHaveAttribute('aria-expanded', 'true')
  await expect(page.locator('#answer-facts')).toBeVisible()
  await expect(page.locator('#answer-facts .state-slot')).toHaveCount(4)
  await facts.click()
  await expect(facts).toHaveAttribute('aria-expanded', 'false')
  await expect(page.locator('#answer-facts')).toBeHidden()

  // Bascule des deux modes, puis persistance au rechargement via localStorage.
  await page.locator('[data-mode="labo"]').click()
  await expect(page.locator('.labo-view')).toBeVisible()
  await expect(page.locator('[data-mode="labo"]')).toHaveAttribute('aria-current', 'page')
  expect(await page.evaluate(() => localStorage.getItem('meteolink-mode'))).toBe('labo')
  await page.reload()
  await expect(page.locator('.labo-view')).toBeVisible()
  expect(await page.evaluate(() => localStorage.getItem('meteolink-mode'))).toBe('labo')

  // Rail : flèches/Début/Fin déplacent le focus, Entrée seulement active.
  const steps = page.locator('.pr-step')
  const progress = page.locator('.pr-progress')
  await expect(progress).toHaveText('Étape 1 sur 4')
  await steps.first().focus()
  await expect(steps.first()).toBeFocused()
  await expect(steps.first()).toHaveAttribute('aria-current', 'step')
  await page.keyboard.press('ArrowRight')
  await expect(steps.nth(1), 'ArrowRight déplace le focus').toBeFocused()
  await expect(steps.first(), 'activation manuelle : aria-current attend Entrée').toHaveAttribute('aria-current', 'step')
  await expect(progress).toHaveText('Étape 1 sur 4')
  await page.keyboard.press('End')
  await expect(steps.nth(3), 'Fin = dernière étape').toBeFocused()
  await page.keyboard.press('Home')
  await expect(steps.first(), 'Début = première étape').toBeFocused()
  await page.keyboard.press('ArrowLeft')
  await expect(steps.nth(3), 'ArrowLeft boucle en fin de rail').toBeFocused()
  await page.keyboard.press('Enter')
  await expect(steps.nth(3), 'Entrée active l’étape').toHaveAttribute('aria-current', 'step')
  await expect(progress).toHaveText('Étape 4 sur 4')
  await expect(page.locator('#chapter-preuves-title'), 'focus rejoint le panneau').toBeFocused()
  await expect(page.locator('#chapter-preuves')).not.toHaveAttribute('inert')

  // CH4 Preuves : contenu réel du service, pas un panneau vide.
  await expect(page.locator('#chapter-preuves .preuves-lede')).toBeVisible()
  await expect(page.locator('#chapter-preuves .archive')).toBeVisible()
  await expect(page.locator('#chapter-preuves .corpus-line')).toBeVisible()
  const laboInline = page.locator('#chapter-preuves .inline-disclosure-trigger').first()
  await laboInline.click()
  await expect(laboInline).toHaveAttribute('aria-expanded', 'true')
  await laboInline.click()
  await expect(laboInline).toHaveAttribute('aria-expanded', 'false')

  // CH2 Observer : matrice, portes et chronologie réellement montées.
  await steps.nth(1).click()
  await expect(steps.nth(1)).toHaveAttribute('aria-current', 'step')
  await expect(page.locator('#chapter-observer .labo-matrix')).toBeVisible()
  await expect(page.locator('#chapter-observer .labo-gates')).toBeVisible()
  await expect(page.locator('#chapter-observer .labo-timeline')).toBeVisible()
  await expect(page.locator('#chapter-observer .event-timeline')).toBeVisible()
  await expect(page.locator('#chapter-observer .labo-gate-row')).toHaveCount(8)

  // CH3 Comparer : duel, classement, compromis ; le bouton PREUVE ouvre le tiroir.
  await steps.nth(2).click()
  await expect(steps.nth(2)).toHaveAttribute('aria-current', 'step')
  await expect(page.locator('#chapter-comparer .comparer')).toBeVisible()
  await expect(page.locator('#chapter-comparer .duel')).toBeVisible()
  await expect(page.locator('#chapter-comparer .rank-table')).toBeVisible()
  await expect(page.locator('#chapter-comparer .inline-disclosure-trigger')).toBeVisible()

  const trigger = page.locator('.duel-proof')
  await trigger.click()
  const dialog = page.locator('.proof-drawer[role="dialog"]')
  await expect(dialog).toBeVisible()
  await expect(dialog).toHaveAttribute('aria-modal', 'true')
  expect(await inertShell(page), 'arrière-plan inerte').toBe(true)
  await expect(page.locator('.proof-close'), 'focus entre dans le panneau').toBeFocused()
  for (let i = 0; i < 6; i++) {
    await page.keyboard.press('Tab')
    expect(await page.evaluate(() => !!document.activeElement?.closest('.proof-drawer-panel')), 'focus piégé').toBe(true)
  }
  await page.locator('.proof-close').focus()
  await page.keyboard.press('Shift+Tab')
  expect(await page.evaluate(() => document.activeElement?.id), 'Shift+Tab boucle').toBe('proof-tab-proof')
  await page.keyboard.press('Escape')
  await expect(dialog).toHaveCount(0)
  expect(await inertShell(page), 'arrière-plan rendu').toBe(false)
  await expect(trigger, 'focus rendu au déclencheur').toBeFocused()

  // CH1 Configurer : le quatrième panneau ferme la marche.
  await steps.first().click()
  await expect(steps.first()).toHaveAttribute('aria-current', 'step')
  await expect(page.locator('#chapter-configurer .configurer')).toBeVisible()
  await expect(page.locator('#chapter-configurer .inline-disclosure-trigger')).toBeVisible()

  expect(errors, 'zéro erreur console sur tout le parcours').toEqual([])
})

// Régression ciblée du défaut trouvé par le parcours : sous 1024 px le
// discloser du journal était un bouton mort de 24 px (rail toujours visible).
// Correctif source : retiré sous 1024, cible ≥44 px et fonctionnelle au-delà.
test('journal — pas de cible morte sous 1024 px, discloser ≥44 px au-delà', async ({ page }) => {
  const errors = watchErrors(page)
  await page.setViewportSize({ width: 890, height: 512 })
  await page.goto('/')
  await expect(page.locator('.journal-rail')).toBeVisible()
  await expect(page.locator('.journal-disclosure-toggle'), 'bouton mort retiré').toBeHidden()

  await page.setViewportSize({ width: 1280, height: 768 })
  const toggle = page.locator('.journal-disclosure-toggle')
  await expect(toggle).toBeVisible()
  const h = await toggle.evaluate(e => e.getBoundingClientRect().height)
  expect(h, 'discloser ≥44 px').toBeGreaterThanOrEqual(44)
  await expect(page.locator('.journal-rail'), 'rail replié').toBeHidden()
  await toggle.click()
  await expect(toggle).toHaveAttribute('aria-expanded', 'true')
  await expect(page.locator('.journal-rail'), 'discloser ouvre le rail').toBeVisible()
  expect(errors, 'zéro erreur console').toEqual([])
})

// La géométrie de la décision épinglée : même contrat que la spec de layout,
// vérifié ici après le parcours pour prouver l'absence de scroll en navigation.
for (const [w, h, spineMin] of [[1280, 768, 300], [1920, 1080, 340]] as const) {
  test(`synthèse ${w}px — spine à gauche, champ à droite, zéro défilement`, async ({ page }) => {
    const errors = watchErrors(page)
    await page.setViewportSize({ width: w, height: h })
    await page.goto('/')
    await expect(page.locator('#v-cockpit')).toBeVisible()
    const g = await page.evaluate(() => {
      const spine = document.querySelector('.decision-spine')!.getBoundingClientRect()
      const field = document.querySelector('.cockpit-field')!.getBoundingClientRect()
      return {
        spineBottom: spine.bottom,
        spineRight: spine.right,
        fieldLeft: field.left,
        fieldBottom: field.bottom,
        viewportBottom: document.documentElement.clientHeight,
      }
    })
    expect(g.spineRight, 'décision à gauche du champ').toBeLessThan(g.fieldLeft)
    expect(g.spineRight, 'colonne épinglée').toBeGreaterThanOrEqual(spineMin)
    expect(g.fieldBottom, 'champ dans la zone de travail').toBeLessThanOrEqual(g.viewportBottom - 40)
    const o = await overflow(page)
    expect(o.h, 'aucun débordement horizontal').toBeLessThanOrEqual(1)
    expect(o.v, 'aucun scroll de page').toBeLessThanOrEqual(1)
    expect(errors, 'zéro erreur console').toEqual([])
  })
}

// Axe après interaction (divulgation ouverte) sur les trois largeurs et les
// deux modes : la valeur ajoutée est l'état modifié, pas le DOM figé des specs
// static. Les cibles ≥44 px et le débordement sont mesurés dans le même état.
for (const viewport of VIEWPORTS) {
  test(`axe WCAG 2.2 AA après interaction — ${viewport.width}×${viewport.height}`, async ({ page }) => {
    test.setTimeout(90000)
    const errors = watchErrors(page)
    await page.setViewportSize(viewport)
    await page.goto('/')
    for (const mode of ['cockpit', 'labo'] as const) {
      await page.locator(`[data-mode="${mode}"]`).click()
      if (mode === 'labo') await expect(page.locator('.labo-view')).toBeVisible()
      else await expect(page.locator('#v-cockpit')).toBeVisible()

      // Sous 1024 px, .proof-slot est masqué et le bandeau n'expose que le
      // disclosure des faits ; au-delà, c'est la divulgation inline qui vit.
      const disclosure = page.locator(mode === 'labo'
        ? '#chapter-configurer .inline-disclosure-trigger'
        : viewport.width < 1024 ? '.facts-disclosure-trigger'
          : '.decision-spine .inline-disclosure-trigger').first()
      await disclosure.click()
      await expect(disclosure).toHaveAttribute('aria-expanded', 'true')

      const results = await new AxeBuilder({ page }).withTags(AXE_TAGS).analyze()
      const bad = results.violations
        .filter(v => v.impact === 'critical' || v.impact === 'serious')
        .map(v => ({ id: v.id, nodes: v.nodes.map(n => n.target) }))
      expect(bad, `axe ${viewport.width} ${mode}`).toEqual([])

      const small = await page.evaluate((sel: string) => [...document.querySelectorAll(sel)]
        .filter(e => e.getClientRects().length && !e.closest('[inert], [aria-hidden="true"]'))
        .map(e => ({
          word: (e.textContent || '').trim().slice(0, 20),
          w: e.getBoundingClientRect().width,
          h: e.getBoundingClientRect().height,
        }))
        .filter(x => x.w < 44 || x.h < 44), TARGETS)
      expect(small, `cibles <44 px ${viewport.width} ${mode}`).toEqual([])

      const o = await overflow(page)
      expect(o.h, `débordement horizontal ${viewport.width} ${mode}`).toBeLessThanOrEqual(1)
      expect(o.v, `scroll de page ${viewport.width} ${mode}`).toBeLessThanOrEqual(1)
      await disclosure.click()
      await expect(disclosure).toHaveAttribute('aria-expanded', 'false')
    }
    expect(errors, 'zéro erreur console').toEqual([])
  })
}

test('contraste forcé et mouvement réduit — rendu sans erreur dans les deux modes', async ({ page }) => {
  test.setTimeout(90000)
  const errors = watchErrors(page)
  await page.setViewportSize({ width: 1280, height: 768 })
  await page.emulateMedia({ reducedMotion: 'reduce', forcedColors: 'active' })
  await page.goto('/')

  await expect(page.locator('#v-cockpit')).toBeVisible()
  await expect(page.locator('#verdict-word')).toBeVisible()
  // À 1280, .band-proof est masqué (contrôle du seul bandeau compact) : on
  // vérifie la colonne de décision, pas un bouton volontairement retiré.
  await expect(page.locator('.decision-spine')).toBeVisible()
  await page.locator('[data-mode="labo"]').click()
  await expect(page.locator('.labo-view')).toBeVisible()
  await expect(page.locator('.pipeline-rail')).toBeVisible()
  await expect(page.locator('#chapter-configurer')).toBeVisible()

  const slideMs = await page.evaluate(() => {
    const track = document.querySelector('.panel-track')!
    const v = getComputedStyle(track).transitionDuration
    return parseFloat(v) * (v.includes('ms') ? 1 : 1000)
  })
  expect(slideMs, 'glissement instantané en mouvement réduit').toBeLessThanOrEqual(1)
  const o = await overflow(page)
  expect(o.h, 'aucun débordement horizontal').toBeLessThanOrEqual(1)
  expect(o.v, 'aucun scroll de page').toBeLessThanOrEqual(1)
  expect(errors, 'zéro erreur console').toEqual([])
})
