import { defineConfig } from '@playwright/test';
export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  // PW_QUIET=1 — vérité sans réessai : un seul worker, zéro retry.
  workers: process.env.PW_QUIET === '1' ? 1 : '50%',
  retries: process.env.PW_QUIET === '1' ? 0 : 1,
  use: {
    baseURL: process.env.AUDIT_BASE || 'http://127.0.0.1:4173',
    ignoreHTTPSErrors: true, // certificat local auto-signé (tests uniquement)
    trace: 'on',
    video: 'retain-on-failure',
    launchOptions: process.env.PWCHROMIUM ? { executablePath: process.env.PWCHROMIUM } : undefined,
  },
  webServer: process.env.AUDIT_BASE
    ? undefined
    : {
        // Le binaire réel sert la SPA embarquée + /api : un seul serveur pour
        // les specs mockées (page.route écrase /api) et l' spéc unmocked.
        // vite preview s'est révélé instable sur ce poste (connexions coupées
        // en pleine suite) — cgo tient, preuve : e2e/unmocked.spec.ts 5/5.
        command: process.env.CGO_E2E_BIN
          ? `${process.env.CGO_E2E_BIN} --serve --mode observe --addr 127.0.0.1:4173 --tls=false`
          : 'go run ./cmd/cgo --serve --mode observe --addr 127.0.0.1:4173 --tls=false',
        cwd: '../..',
        url: 'http://127.0.0.1:4173/api/health',
        timeout: 120000,
        // Auth fail-closed : sans jeton, cgo serve refuse de démarrer (exit 2).
        // On fournit un jeton de test et une racine de données sous tmp/ pour ne
        // jamais écrire dans le corpus suivi (book/data, data).
        env: {
          CGO_API_TOKEN: 'cgo-e2e-axe-test-token-32-chars-minimum',
          CGO_DATA_DIR: 'tmp/axe-e2e-data',
        },
        reuseExistingServer: false,
      },
});
