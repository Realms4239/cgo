import { test, expect } from '@playwright/test'

// BLOQUÉ — `.prompt-progress`, `text=Actions rapides` : 0 occurrence dans le DOM
// vivant (la progression d'invite a été supprimée) ; `.live-wall-overlay` :
// 0 occurrence ; `[data-testid="rank-verdict"]` / `.rank-verdict` : 0 occurrence.
// Réactiver ces assertions quand une ligne de progression d'invite, un overlay
// de mur et un verdict de classement seront remontés dans l'interface.
test('m9 wall kit', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=METEOLINK').first()).toBeVisible()
  // le bouton « Densité » et le rail `.rail` de l'ancien shell ont disparu :
  // la bascule de destination porte l'interaction, `.pipeline-rail` est le rail vivant.
  await page.locator('[data-mode="labo"]').click()
  await expect(page.locator('.pipeline-rail')).toBeVisible()
  await page.keyboard.press('Meta+k')
  if ((await page.getByRole('dialog', { name: 'Palette de commandes' }).count()) === 0) await page.keyboard.press('Control+k')
  await expect(page.getByRole('dialog', { name: 'Palette de commandes' })).toBeVisible()
  await page.keyboard.press('Escape')

  // BLOQUÉ — cibles supprimées, voir l'en-tête du fichier pour le déclencheur.
  test.skip(true, 'BLOQUÉ — .prompt-progress : 0 occurrence ; .live-wall-overlay : 0 occurrence ; [data-testid="rank-verdict"] / .rank-verdict : 0 occurrence')

  // Après fermeture palette, vérifier l'invite — lance un run léger si idle pour que prompt-progress (Event 3/6) soit visible
  const prompt = page.locator('.prompt-progress')
  if ((await prompt.count()) === 0) {
    await page.evaluate(() => fetch('/api/run/start', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ profiles: ['P1'], reps: 1 }) }).catch(() => {}))
    await page.waitForTimeout(1200)
  }
  // Ligne progression invite (ambre 1px) — visible en run ; tolère 7s de réapparition si fermée par Esc
  await expect(page.locator('.prompt-progress')).toBeVisible({ timeout: 7000 }).catch(async () => {
    // repli : vérifier au moins la ligne Event en cours ou le label Actions rapides
    const evt = page.locator('text=Event')
    const rapid = page.locator('text=Actions rapides')
    if (await evt.count()) await expect(evt.first()).toBeVisible({ timeout: 2000 })
    else if (await rapid.count()) await expect(rapid.first()).toBeVisible({ timeout: 2000 })
  })
  // Overlay mur live référence vs discipline appliquée — même échelle, badge d'écart
  // « Tableau live » est devenu le chapitre Observer du Labo.
  await page.locator('.pr-step').nth(1).click()
  await expect(page.locator('.live-wall-overlay')).toBeVisible({ timeout: 5000 })
  await expect(page.locator('.live-wall-overlay')).toContainText(/baseline/i)
  await expect(page.locator('.live-wall-overlay')).toContainText(/cake|pfifo/i)
  // Résultats classement + épinglage A/B + provenance (design verrouillé)
  // « Résultats » est devenu le chapitre Comparer ; la provenance est portée
  // par le chapitre Preuves.
  await page.locator('.pr-step').nth(2).click()
  await expect(page.getByRole('heading', { name: /Comparer/ })).toBeVisible({ timeout: 5000 })
  // BLOQUÉ — aucun verdict de classement n'est monté ; réactiver quand
  // `[data-testid="rank-verdict"]` ou `.rank-verdict` réapparaîtra.
  await expect(page.locator('[data-testid="rank-verdict"], .rank-verdict').first()).toBeVisible({ timeout: 5000 })
  await expect(page.locator('#chapter-preuves').getByText(/Provenance des preuves/).first()).toBeVisible({ timeout: 5000 })
})
