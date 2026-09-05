import { defineConfig } from '@playwright/test';
export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  use: {
    baseURL: process.env.AUDIT_BASE || 'https://localhost:9090',
    ignoreHTTPSErrors: true, // certificat local auto-signé (tests uniquement)
    trace: 'on',
    video: 'retain-on-failure',
  },
});
