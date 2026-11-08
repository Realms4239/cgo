import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.ts', 'src/**/*.test.tsx'],
    exclude: ['e2e/**', 'node_modules/**', 'dist/**'],
    // BenchConfig.dom.test.tsx does vi.resetModules() + dynamic import; under full-suite
    // load the cold transform exceeded the 5s default (flaky timeout 2026-11-08).
    testTimeout: 15000,
  },
})
