import { expect, test, type Page } from '@playwright/test';

// Binaire Go réel servi par playwright.config.ts — aucun page.route, aucun stub,
// aucun hôte distant. Le serveur local tourne en mode observation : /api/stream
// publie des trames idle honnêtes (running:false, aucune mesure). Le test
// n'invente ni campagne, ni valeur mesurée, ni sceau.

// Champs qui, s'ils apparaissaient dans une trame idle, simuleraient une mesure.
const MEASURED_KEYS = [
  'rtt_p50_ms', 'rtt_p95_ms', 'small_p95_ms', 'bulk_goodput_mbps',
  'drops', 'deadline_ok_pct', 'wasted_bytes', 'cost_ar_per_h',
] as const;

interface RawSample {
  contentType: string;
  raw: string;
  frames: Record<string, unknown>[];
  times: number[];
  events: string[];
}

// Lecture du vrai flux SSE depuis le navigateur : fetch + ReadableStream reader.
// Fenêtre bornée (~1 s) et garde-fou AbortController : aucune horloge murale
// précise n'est exigée, seulement une enveloppe de cadence.
async function readRealStream(page: Page, windowMs: number): Promise<RawSample> {
  return page.evaluate(async (ms) => {
    const ac = new AbortController();
    const timer = setTimeout(() => ac.abort(), ms + 250);
    const decoder = new TextDecoder();
    let all = '';
    let buf = '';
    const frames: Record<string, unknown>[] = [];
    const times: number[] = [];
    const events: string[] = [];

    let res: Response;
    try {
      res = await fetch('/api/stream', {
        headers: { Accept: 'text/event-stream' },
        signal: ac.signal,
      });
    } catch (e) {
      clearTimeout(timer);
      throw e;
    }
    const contentType = res.headers.get('content-type') ?? '';
    const reader = res.body!.getReader();
    const t0 = performance.now();

    try {
      while (performance.now() - t0 < ms) {
        let chunk: ReadableStreamReadResult<Uint8Array>;
        try {
          chunk = await reader.read();
        } catch {
          break; // abort de sécurité : le garde-fou a coupé la lecture
        }
        if (chunk.done) break;
        const text = decoder.decode(chunk.value, { stream: true });
        all += text;
        buf += text;
        let i: number;
        while ((i = buf.indexOf('\n\n')) !== -1) {
          const block = buf.slice(0, i);
          buf = buf.slice(i + 2);
          const lines = block.split('\n');
          for (const l of lines) {
            if (l.startsWith('event:')) events.push(l.slice(6).trim());
          }
          for (const l of lines) {
            if (!l.startsWith('data:')) continue;
            try {
              const obj = JSON.parse(l.slice(5).trim());
              frames.push(obj);
              times.push(performance.now() - t0);
            } catch {
              // ligne non-JSON : ignorée, jamais acceptée comme trame
            }
          }
        }
      }
    } finally {
      clearTimeout(timer);
      try { reader.cancel(); } catch { /* déjà fermé */ }
      ac.abort();
    }
    return { contentType, raw: all, frames, times, events };
  }, windowMs);
}

test('flux SSE réel — trames idle honnêtes, cadence ~20 Hz et reprise annoncée', async ({ page }) => {
  await page.goto('/');
  const s = await readRealStream(page, 1000);

  expect(s.contentType, 'content-type du flux').toContain('text/event-stream');
  // directive de reconnexion EventSource émise par le serveur, pas simulée.
  expect(s.raw, 'directive de reprise').toContain('retry: 2000');
  expect(s.frames.length, 'au moins une vraie trame JSON').toBeGreaterThan(0);

  // Contrat idle : running:false partout, aucune mesure réelle, aucun sceau/verdict.
  // Le binaire sérialise tous les champs du Snapshot, donc l'idle porte des
  // zéros (« trous de pompe »). La frontière de vérité du front (lib/sse.ts)
  // les traite comme non-mesures ; les tests UI ci-dessous prouvent qu'aucun
  // zéro n'est présenté comme valeur mesurée.
  for (const f of s.frames) {
    expect(f.running, 'trame idle').toBe(false);
    for (const k of MEASURED_KEYS) {
      if (k in f) expect(f[k], `mesure réelle ${k} absente`).toBe(0);
    }
    expect(f).not.toHaveProperty('sha256');
    expect(f).not.toHaveProperty('seal');
    expect(f).not.toHaveProperty('operator_verdict');
    expect(f).not.toHaveProperty('verdict');
    expect(typeof f.ts, 'horodatage réel').toBe('number');
  }
  // Aucun jalon de campagne ne est publié en observation (backpressure seul toléré).
  expect(s.events.filter(e => e !== 'backpressure')).toEqual([]);

  // Enveloppe de cadence (smoke, pas un oracle de précision) :
  // ≥2 trames en 500 ms, bande large <60 sur 1 s — jamais de timing strict.
  const firstHalf = s.times.filter(t => t <= 500).length;
  expect(firstHalf, 'trames reçues en 500 ms').toBeGreaterThanOrEqual(2);
  expect(s.frames.length, 'trames reçues en 1 s (plancher)').toBeGreaterThanOrEqual(6);
  expect(s.frames.length, 'trames reçues en 1 s (plafond)').toBeLessThan(60);
});

test('UI — état live honnête pour des trames idle (connexion, absence de mesure, backpressure)', async ({ page }) => {
  await page.goto('/');

  await expect(page.locator('[data-slot="connection"] .trust-word')).toHaveText('EN LIGNE', { timeout: 15000 });
  await page.waitForFunction(() => ((window as any).__CGO_SSE?.frameCount ?? 0) >= 2, null, { timeout: 15000 });

  // Le navigateur a consommé de vraies trames, sans jamais écrire de mesure dans les anneaux.
  const rings = await page.evaluate(() => {
    const live = (window as any).__CGO_LIVE;
    const sse = (window as any).__CGO_SSE;
    return {
      frames: sse.frameCount,
      ringLen: sse.ringLen,
      dropped: sse.dropped,
      small: live.small.length,
      rtt50: live.rtt50.length,
      rtt95: live.rtt95.length,
      goodput: live.goodput.length,
      deadline: live.deadline.length,
    };
  });
  expect(rings.frames).toBeGreaterThanOrEqual(2);
  expect(rings.ringLen).toBe(0);
  expect(rings.small).toBe(0);
  expect(rings.rtt50).toBe(0);
  expect(rings.rtt95).toBe(0);
  expect(rings.goodput).toBe(0);
  expect(rings.deadline).toBe(0);
  expect(rings.dropped).toBe(0);

  // L'absence garde son mot : ni zéro, ni verdict, ni état « en cours » fabriqué.
  await expect(page.locator('#verdict-word')).toContainText('MESURE ABSENTE');
  await expect(page.locator('.health-word')).toHaveText('Non mesurée');
  await expect(page.locator('.band-number-absent')).toHaveText('— non mesuré');
  await expect(page.locator('.answer-sentence')).toHaveText('Aucune mesure pour ce lien.');
  await expect(page.locator('.act-slot')).toHaveAttribute('data-state', 'idle');
  await expect(page.getByText('Mesure en cours', { exact: false })).toHaveCount(0);

  // Backpressure : mot honnête (absence ou aucune perte), jamais une perte inventée.
  await expect(page.locator('[data-slot="backpressure"] .trust-word')).not.toContainText('trames perdues');
  const bp = await page.locator('[data-slot="backpressure"] .trust-word').textContent();
  expect(['ABSENT', 'AUCUNE PERTE']).toContain((bp ?? '').trim());
});

test('ARM — refus 501 honnête (« mode observation »), jamais un état en cours fabriqué', async ({ page, request }) => {
  const health = await (await request.get('/api/health')).json();
  expect(health.mode, 'hôte de vérification attendu en observation').toBe('observe');

  await page.goto('/');
  await expect(page.locator('[data-slot="connection"] .trust-word')).toHaveText('EN LIGNE', { timeout: 15000 });

  const arm = page.getByRole('button', { name: 'Démarrer la mesure' });
  await expect(arm).toBeVisible();

  const [resp] = await Promise.all([
    page.waitForResponse(r => r.url().includes('/api/audit/start') && r.request().method() === 'POST', { timeout: 15000 }),
    arm.click(),
  ]);

  // Le mot de refus vient du serveur réel, pas d'une copie dans le test.
  expect(resp.status(), 'le service doit refuser en observation').toBe(501);
  const body = await resp.json();
  expect(body.error, 'mot de refus réel du serveur').toContain('mode observation');

  // Aucune campagne n'a démarré : l'UI reste dans l'absence et affiche le refus.
  await expect(page.locator('.act-slot')).toHaveAttribute('data-state', 'quiet');
  await expect(page.locator('.act-slot [role="status"]')).toContainText('Démarrage impossible');
  await expect(page.locator('.act-slot [role="status"]')).toContainText('501');
  await expect(page.locator('#verdict-word')).toContainText('MESURE ABSENTE');
  await expect(page.locator('.band-status')).not.toHaveText('Mesure en cours');
  await expect(page.getByText('Mesure en cours', { exact: false })).toHaveCount(0);
  await expect(page.locator('.toast')).toContainText('Échec audit');

  // Aucun effet de bord : l'état reste idle après le refus.
  const state = await (await request.get('/api/state')).json();
  expect(state.running).toBe(false);
});
