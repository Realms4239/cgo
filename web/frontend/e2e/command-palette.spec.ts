import { test, expect } from '@playwright/test'

// Cmd-K — une entrée par destination, Escape referme. La palette rend l'écran inerté.

test('command palette — Ctrl+K ouvre, filtrage, Escape referme', async ({ page }) => {
  await page.route('**/api/**', route => {
    const path = new URL(route.request().url()).pathname
    if (path.endsWith('/api/audit/status') || path.endsWith('/api/audit/live')) {
      return route.fulfill({ json: { running: false, report: null } })
    }
    if (path.endsWith('/api/stream')) return route.fulfill({ status: 200, contentType: 'text/event-stream', body: '' })
    return route.fulfill({ json: {} })
  })
  await page.goto('/')
  await page.keyboard.press('Control+k')
  await expect(page.getByRole('dialog', { name: 'Palette de commandes' })).toBeVisible()
  await page.keyboard.type('auditer')
  const items = page.locator('.palette-item')
  await expect(items).toHaveCount(1)
  await expect(items.first()).toContainText('Auditer')
  await page.keyboard.press('Escape')
  await expect(page.getByRole('dialog', { name: 'Palette de commandes' })).toHaveCount(0)
})
