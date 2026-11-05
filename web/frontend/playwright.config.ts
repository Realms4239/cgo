import { defineConfig } from '@playwright/test';
export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
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
        command: 'npx vite preview --host 127.0.0.1 --port 4173 --strictPort',
        url: 'http://127.0.0.1:4173/',
        timeout: 120000,
        reuseExistingServer: true,
      },
});
