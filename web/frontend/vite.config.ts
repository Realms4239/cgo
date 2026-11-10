import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  // Le proxy borne l'attente : sans backend joignable, la connexion traîne
  // au lieu d'échouer vite, et le serveur preview (monothread) bloque les
  // assets derrière elle — un chunk tardif faisait tomber le cockpit.
  server: { proxy: { '/api': { target: 'https://localhost:9090', secure: false, timeout: 3000, proxyTimeout: 3000 } } },
  preview: { proxy: { '/api': { target: 'https://localhost:9090', secure: false, timeout: 3000, proxyTimeout: 3000 } } },
  build: { outDir: 'dist', sourcemap: false },
});
