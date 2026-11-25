import { test, expect, type APIRequestContext, type Page } from '@playwright/test';
import { retryOn429 } from './fixtures';

// Binaire Go réel servi par playwright.config.ts — aucun page.route, aucun stub.
// Les preuves viennent du serveur : auth activée, mode observation sur cet hôte.
const TOKEN = process.env.CGO_API_TOKEN ?? 'cgo-local-dev-7f3a91d2';

const MUTATIONS = [
  '/api/audit/start',
  '/api/audit/stop',
  '/api/run/start',
  '/api/run/stop',
  '/api/shape',
  '/api/watch',
  '/api/audit/toprofile',
];

// Contexte sans credential : ni Bearer ni cookie. La fixture `request` porte
// l'en-tête Bearer du harnais, donc on fabrique un contexte nu à part.
async function anonymousApi(playwright: any, baseURL: string | undefined): Promise<APIRequestContext> {
  if (!baseURL) throw new Error('baseURL requis pour le contexte anonyme');
  // newContext hérite des extraHTTPHeaders de la config : on écrase Bearer à vide.
  return playwright.request.newContext({ baseURL, extraHTTPHeaders: { Authorization: '' } });
}

// Lecture via le navigateur, sans l'en-tête Bearer du harnais : c'est le cookie
// de session seul qui doit prouver l'authentification.
async function browserReadStatus(page: Page): Promise<number> {
  return page.evaluate(async () => (await fetch('/api/state')).status);
}

test('GET /api/health — contrat ok / version / mode réels', async ({ request }) => {
  const res = await retryOn429(() => request.get('/api/health'));
  expect(res.status()).toBe(200);
  expect(res.headers()['content-type']).toContain('application/json');

  const body = await res.json();
  expect(body.ok).toBe(true);
  expect(typeof body.version).toBe('string');
  expect(body.version.length).toBeGreaterThan(0);
  expect(['observe', 'full', '']).toContain(body.mode);
  // La sonde annonce sa source, jamais une valeur de mesure fabriquée.
  expect(body.telemetry).toBeTruthy();
  expect(typeof body.telemetry.source).toBe('string');
});

test('auth — lecture sans credential 401, avec Bearer 200', async ({ playwright, request, baseURL }) => {
  const anon = await anonymousApi(playwright, baseURL);
  try {
    const denied = await anon.get('/api/state');
    expect(denied.status()).toBe(401);
    expect(denied.headers()['www-authenticate']).toContain('Bearer');
  } finally {
    await anon.dispose();
  }

  const allowed = await retryOn429(() => request.get('/api/state'));
  expect(allowed.status()).toBe(200);
});

test('auth — /api/login pose un cookie HttpOnly et le navigateur connecté lit', async ({ playwright, browser, baseURL }) => {
  const anon = await anonymousApi(playwright, baseURL);
  const login = await retryOn429(() => anon.post('/api/login', { data: { token: TOKEN } }));
  expect(login.status()).toBe(200);
  expect((await login.json()).ok).toBe(true);

  const setCookie = login.headers()['set-cookie'] ?? '';
  expect(setCookie).toContain('cgo_session=');
  expect(setCookie).toMatch(/HttpOnly/i);
  expect(setCookie).toContain('Path=/');
  expect(setCookie).toContain('SameSite=Strict');

  // Le cookie seul (aucun Bearer) suffit à lire.
  const cookieRead = await retryOn429(() => anon.get('/api/state'));
  expect(cookieRead.status()).toBe(200);
  await anon.dispose();

  // Navigateur sans l'en-tête du harnais : 401 avant cookie, 200 après.
  const context = await browser.newContext({ extraHTTPHeaders: {} });
  try {
    const page = await context.newPage();
    await page.goto('/');
    expect(await browserReadStatus(page)).toBe(401);

    const [name, value] = setCookie.split(';')[0].split('=');
    await context.addCookies([{ name, value, url: baseURL, httpOnly: true, sameSite: 'Strict' }]);
    expect(await browserReadStatus(page)).toBe(200);
  } finally {
    await context.close();
  }
});

test('lectures — JSON valide et absence honnête, aucune donnée fabriquée', async ({ request }) => {
  const checks: Array<[string, (body: any) => void]> = [
    ['/api/state', (b) => {
      // Rien de mesuré n'est présenté : pas de campagne, pas d'événement.
      expect(b.running).toBe(false);
      expect(b.phase).toBe('');
      expect(b.total_events).toBe(0);
    }],
    ['/api/events', (b) => {
      expect(Array.isArray(b.events)).toBe(true);
      expect(b.events).toHaveLength(0);
    }],
    ['/api/integrity', (b) => {
      // Le sceau affiché vient du corpus gelé réel, pas d'un PASS inventé.
      expect(typeof b.available).toBe('boolean');
      expect(b.sha256).toMatch(/^[0-9a-f]{64}$/);
      expect(typeof b.integrity).toBe('string');
      expect(Array.isArray(b.breakdown)).toBe(true);
    }],
    ['/api/results', (b) => {
      expect(typeof b.available).toBe('boolean');
      expect(Array.isArray(b.groups)).toBe(true);
    }],
    ['/api/quarantine', (b) => {
      expect(Array.isArray(b.quarantines)).toBe(true);
    }],
    ['/api/capabilities', (b) => {
      expect(Array.isArray(b.qdiscs)).toBe(true);
      expect(Array.isArray(b.ccs)).toBe(true);
      expect(Array.isArray(b.known_ccs)).toBe(true);
      expect(typeof b.bbr3).toBe('boolean');
      expect(b.bbr3).toBe(b.ccs.includes('bbr3'));
    }],
    ['/api/profiles', (b) => {
      expect(Array.isArray(b.profiles)).toBe(true);
      expect(b.profiles.map((p: any) => p.id)).toEqual(expect.arrayContaining(['P1', 'P2', 'P3', 'P4']));
    }],
    ['/api/cost/tiers', (b) => {
      expect(Array.isArray(b.tiers)).toBe(true);
      expect(typeof b.default).toBe('string');
    }],
  ];

  for (const [path, check] of checks) {
    const res = await retryOn429(() => request.get(path));
    expect(res.status(), path).toBe(200);
    expect(res.headers()['content-type'], path).toContain('application/json');
    check(await res.json());
  }
});

test('absence honnête — figures et replay ne fabriquent ni image ni flux', async ({ request }) => {
  // Aucune figure générée localement : 404, pas une image ou un JSON inventé.
  const figures = await retryOn429(() => request.get('/api/figures/'));
  expect(figures.status()).toBe(404);
  expect(figures.headers()['content-type'] ?? '').not.toContain('image/');

  // Sans identifiant de run autorisé, le replay refuse au lieu d'émettre des trames.
  const replay = await retryOn429(() => request.get('/api/replay/stream'));
  expect(replay.status()).toBe(400);
  expect(await replay.text()).toContain('run invalide');
  expect(replay.headers()['content-type'] ?? '').not.toContain('text/event-stream');
});

test('observe — mutations 501 sans effet de bord', async ({ request }) => {
  const health = await (await retryOn429(() => request.get('/api/health'))).json();
  // Hôte Windows local : le mode auto résout observation, d'où les 501.
  expect(health.mode).toBe('observe');

  const stateBefore = await (await retryOn429(() => request.get('/api/state'))).json();
  const shapeBefore = await (await retryOn429(() => request.get('/api/shape'))).json();
  const eventsBefore = await (await retryOn429(() => request.get('/api/events'))).json();
  expect(stateBefore.running).toBe(false);
  expect(shapeBefore.applied).toBe(false);
  expect(eventsBefore.events).toHaveLength(0);

  for (const path of MUTATIONS) {
    const res = await retryOn429(() => request.post(path, {
      data: { site: 'x', link_type: 'fiber', profiles: ['P1'], reps: 1, qdisc: 'cake', capacity_mbps: 20, on: true },
      headers: { 'X-Confirm-Action': 'true' },
    }));
    expect(res.status(), path).toBe(501);
    expect((await res.json()).error, path).toContain('mode observation');
  }

  const stateAfter = await (await retryOn429(() => request.get('/api/state'))).json();
  const shapeAfter = await (await retryOn429(() => request.get('/api/shape'))).json();
  const eventsAfter = await (await retryOn429(() => request.get('/api/events'))).json();
  expect(stateAfter).toEqual(stateBefore);
  expect(shapeAfter).toEqual(shapeBefore);
  expect(eventsAfter).toEqual(eventsBefore);
});
