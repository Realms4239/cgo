import { expect, test } from '@playwright/test';

// Contrat spatial §1.4 : sous 1024 px la réponse se replie en un bandeau
// épinglé sous la coque — un mot, un chiffre, un mot d'état, PREUVE — avec les
// quatre faits derrière un disclosure et l'Act juste dessous. La tête du champ
// reste visible ; rien ne déborde, rien ne défiler en page.

test('890×512 — bandeau compact, faits repliés, tête de champ visible, zéro débordement', async ({ page }) => {
  await page.setViewportSize({ width: 890, height: 512 });
  await page.goto('/');
  await expect(page.locator('#v-cockpit')).toBeVisible();

  const g = await page.evaluate(() => {
    const band = document.querySelector('.answer-band') as HTMLElement | null;
    const word = document.querySelector('#verdict-word') as HTMLElement | null;
    const proof = document.querySelector('.band-proof') as HTMLElement | null;
    const factsTrigger = document.querySelector('.facts-disclosure-trigger') as HTMLElement | null;
    const facts = document.querySelector('.answer-facts') as HTMLElement | null;
    const fieldHead = document.querySelector('.tf-header') as HTMLElement | null;
    const box = (e: HTMLElement | null) => e ? e.getBoundingClientRect() : null;
    const bandBox = box(band);
    const proofBox = box(proof);
    const headBox = box(fieldHead);
    return {
      bandHeight: bandBox?.height ?? -1,
      wordSize: word ? parseFloat(getComputedStyle(word).fontSize) : -1,
      proofW: proofBox?.width ?? -1,
      proofH: proofBox?.height ?? -1,
      expanded: factsTrigger?.getAttribute('aria-expanded') ?? null,
      factsVisible: facts
        ? getComputedStyle(facts).display !== 'none' && facts.getClientRects().length > 0
        : null,
      headVisible: headBox ? headBox.top >= 0 && headBox.bottom <= window.innerHeight + 1 : false,
      headTop: headBox?.top ?? -1,
      // « Aucun scroll de page » = le document/body ne défilent pas. Le
      // débordement interne de .layout-shell est clipé par body{overflow:hidden}
      // et vient du champ temporel haut à 890 (déjà admis par la suite).
      pageOverflow: Math.max(
        document.documentElement.scrollHeight - document.documentElement.clientHeight,
        document.body.scrollHeight - document.body.clientHeight,
      ),
      horizontalOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
    };
  });

  expect(g.bandHeight, 'bandeau présent').toBeGreaterThan(0);
  expect(g.bandHeight, 'bandeau compact ≤ 140 px').toBeLessThanOrEqual(140);
  expect(g.wordSize, 'mot du verdict à 32 px').toBe(32);
  expect(g.proofH, 'PREUVE cible ≥ 44 px de haut').toBeGreaterThanOrEqual(44);
  expect(g.proofW, 'PREUVE cible ≥ 44 px de large').toBeGreaterThanOrEqual(44);
  expect(g.expanded, 'quatre faits repliés par défaut').toBe('false');
  expect(g.factsVisible, 'faits masqués quand repliés').toBe(false);
  expect(g.headVisible, 'tête du champ reste dans le viewport').toBe(true);
  expect(g.pageOverflow, 'aucun scroll de page').toBeLessThanOrEqual(1);
  expect(g.horizontalOverflow, 'aucun débordement horizontal').toBeLessThanOrEqual(1);
});

test('890×512 — le disclosure révèle les faits dans le flux, sans page scroll', async ({ page }) => {
  await page.setViewportSize({ width: 890, height: 512 });
  await page.goto('/');
  await expect(page.locator('#v-cockpit')).toBeVisible();

  const trigger = page.locator('.facts-disclosure-trigger');
  await expect(trigger).toHaveAttribute('aria-expanded', 'false');
  await trigger.click();
  await expect(trigger).toHaveAttribute('aria-expanded', 'true');
  await expect(page.locator('.answer-facts')).toBeVisible();

  const g = await page.evaluate(() => {
    const main = document.getElementById('main')!;
    const facts = document.querySelector('.answer-facts') as HTMLElement;
    return {
      position: getComputedStyle(facts).position,
      slots: document.querySelectorAll('.answer-facts .state-slot').length,
      pageOverflow: main.scrollHeight - main.clientHeight,
      horizontalOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
    };
  });
  expect(g.position, 'les faits poussent le flux, ils ne le recouvrent pas').not.toBe('absolute');
  expect(g.position).not.toBe('fixed');
  expect(g.slots, 'quatre faits').toBe(4);
  expect(g.horizontalOverflow, 'aucun débordement horizontal une fois ouvert').toBeLessThanOrEqual(1);
});
