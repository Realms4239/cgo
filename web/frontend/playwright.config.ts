import { defineConfig } from '@playwright/test';
export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  use: {
    baseURL: process.env.AUDIT_BASE || 'http://localhost:9090',
    trace: 'on',
    video: 'retain-on-failure',
  },
});
