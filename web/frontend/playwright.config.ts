import { defineConfig } from '@playwright/test';
export default defineConfig({
  testDir: './e2e',
  timeout: 15000,
  use: { baseURL: 'http://192.168.174.128:9090', trace: 'off' },
});
