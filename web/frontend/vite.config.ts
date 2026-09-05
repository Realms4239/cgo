import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: { proxy: { '/api': { target: 'https://localhost:9090', secure: false } } }, // backend local auto-signé
  build: { outDir: 'dist', sourcemap: false },
});
