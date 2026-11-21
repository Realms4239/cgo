import { defineConfig } from '@playwright/test';

// Le serveur réel, jamais un stub : les preuves viennent du binaire Go qui
// sert le dist embedé. Un serveur périmé sur le port est refusé, pas réutilisé.
const PORT = 19091;
const baseURL = process.env.AUDIT_BASE || `http://127.0.0.1:${PORT}`;

export default defineConfig({
  testDir: './e2e',
  timeout: 90000,
  workers: 1,
  reporter: 'list',
  use: {
    baseURL,
    ignoreHTTPSErrors: true, // certificat local auto-signé (tests uniquement)
    // main impose un Bearer sur /api : le harnais transmet le jeton de test,
    // l'authentification du serveur reste intégralement active.
    extraHTTPHeaders: process.env.CGO_API_TOKEN
      ? { Authorization: `Bearer ${process.env.CGO_API_TOKEN}` }
      : undefined,
    trace: 'retain-on-failure',
    video: 'retain-on-failure',
    launchOptions: process.env.PWCHROMIUM ? { executablePath: process.env.PWCHROMIUM } : undefined,
  },
  webServer: process.env.AUDIT_BASE
    ? undefined
    : {
        command: process.platform === 'win32'
          ? `.\\cgo-test-server.exe --serve --addr 127.0.0.1:${PORT} --tls=false --http-addr=`
          : `./cgo-test-server --serve --addr 127.0.0.1:${PORT} --tls=false --http-addr=`,
        url: `${baseURL}/api/health`,
        timeout: 120000,
        reuseExistingServer: false, // piège du serveur périmé : jamais de dist obsolète
      },
});
