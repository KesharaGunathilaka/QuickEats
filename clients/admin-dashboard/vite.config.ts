import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// Served by Kong under /admin, next to the API, so calls are same-origin.
// In `npm run dev`, /api is proxied to the local Kong gateway.
export default defineConfig({
  base: '/admin/',
  plugins: [react()],
  server: {
    port: 5173,
    proxy: { '/api': process.env.API_URL ?? 'http://localhost:8000' },
  },
});
