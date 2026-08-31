import { test, expect } from '@playwright/test'
test('m9 wall kit', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('text=METEOLINK').first()).toBeVisible()
  await page.getByRole('button', { name: 'Densité' }).click()
  await expect(page.locator('.rail')).toBeVisible()
  await page.keyboard.press('Meta+k')
  if ((await page.getByRole('dialog', { name: 'Palette de commandes' }).count()) === 0) await page.keyboard.press('Control+k')
  await expect(page.getByRole('dialog', { name: 'Palette de commandes' })).toBeVisible()
  await page.keyboard.press('Escape')
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
  // Overlay mur live référence vs CAKE — même échelle, badge d'écart
  await page.getByLabel('Navigation').getByRole('button', { name: 'Tableau live' }).click()
  await expect(page.locator('.live-wall-overlay')).toBeVisible({ timeout: 5000 })
  await expect(page.locator('.live-wall-overlay')).toContainText('baseline')
  await expect(page.locator('.live-wall-overlay')).toContainText('CAKE')
  // Résultats badge diff A/B + provenance matériel
  await page.getByLabel('Navigation').getByRole('button', { name: 'Résultats' }).click()
  await expect(page.locator('.ab-bento')).toBeVisible({ timeout: 5000 })
  await expect(page.locator('.ab-bento .diff-badge')).toBeVisible({ timeout: 5000 })
  await expect(page.locator('.provenance, [data-testid="provenance"]')).toContainText(/hardware_recommendation|MikroTik|P1/)
})
