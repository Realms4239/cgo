import { test, expect } from '@playwright/test'
const BASE = process.env.AUDIT_BASE || '' // vide : baseURL du harnais, jamais d'hôte codé en dur

// Boucle contrôle-et-façonnage pilotée par l'UI (pivot ARG.md) :
// figer une référence depuis les anneaux live → appliquer CAKE au bord → écart live.
// BLOQUÉ — `[data-testid="live-wall-overlay"]` : 0 occurrence dans le DOM réel
// du binaire courant, dans `#v-cockpit` comme dans `#chapter-observer`. Aucun
// overlay équivalent ne porte le figer/CAKE/écart : le corps est conservé mot
// pour mot mais non exécuté. Réactiver quand une surface de contrôle exposera
// `[data-testid="live-wall-overlay"]` avec ses boutons « FIGER L'AVANT »/CAKE.
test('edge control demo', async ({ page }) => {
  test.setTimeout(180000)
  await page.goto(BASE + '/')
  // l'ancien panneau « live » est devenu le chapitre Observer du Labo.
  await page.locator('[data-mode="labo"]').click()
  await page.locator('.pr-step').nth(1).click()
  const overlay = page.locator('[data-testid="live-wall-overlay"]')

  // BLOQUÉ — cible supprimée, voir l'en-tête du fichier pour le déclencheur.
  test.skip(true, 'BLOQUÉ — [data-testid="live-wall-overlay"] : 0 occurrence ; aucun overlay de contrôle dans les deux modes ni les quatre chapitres')

  await overlay.scrollIntoViewIfNeeded()
  await expect(overlay).toBeVisible()
  // la démo contrôle tourne en SURVEILLANCE (watch), pas en campagne : le façonnage est
  // refusé 409 quand une campagne tient le shaper, mais watch compose avec —
  // cette composition (watch → figer l'avant → CAKE → écart) est le produit.
  const watchResp = await page.evaluate(async () => {
    const r = await fetch('/api/watch', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ on: true }) })
    return r.status
  })
  // campagne active sur l'hôte : la démo exige un shaper libre, skip honnête
  test.skip(watchResp === 409, 'campagne active sur l\'hôte — façonnage verrouillé par l\'API (409), comportement correct')
  // hôte observe (Windows) : pas de moteur watch/façonnage, 501 attendu
  test.skip(watchResp === 501, 'hôte en mode observation — le contrôle du bord exige le banc Linux (501), comportement correct')
  await page.waitForTimeout(15000)
  // vocabulaire Q11 — le bouton est «FIGER L'AVANT», l'état figé est une Pill
  // dont le nom accessible est `${label} ${value}`.
  const lockBtn = page.getByRole('button', { name: /FIGER L'AVANT/ })
  await expect(lockBtn).toBeEnabled({ timeout: 30000 })
  await lockBtn.click()
  await page.waitForTimeout(1200)
  await expect(page.getByRole('button', { name: /avant figé/ })).toBeVisible({ timeout: 10000 })
  // le texte de la Pill porte la valeur figée (spans label/valeur sans espace) ;
  // son nom accessible est l'infobulle
  await expect(page.getByRole('button', { name: /avant figé/ })).toHaveText(/avant\s*\d+(\.\d+)? ms @/)
  // les pills de façonnage portent title="appliquer <q> au bord" — ce title EST le
  // nom accessible (Pill : aria-label = title ?? label+valeur)
  await page.getByRole('button', { name: /appliquer cake au bord/ }).click()
  await expect(page.getByText(/appliqué au bord/)).toBeVisible({ timeout: 15000 })
  await page.waitForTimeout(12000)
  await overlay.scrollIntoViewIfNeeded()
  await page.screenshot({ path: 'shots/control-live-diff.png' })
  const txt = await overlay.textContent()
  console.log('OVERLAY-TEXT:', (txt ?? '').slice(0, 300).replace(/\s+/g, ' '))
  // leave a clean state — watch off
  await page.evaluate(async () => { await fetch('/api/watch', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ on: false }) }) })
})
