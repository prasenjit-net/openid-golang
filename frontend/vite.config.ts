import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  base: '/',
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    emptyOutDir: true,
    chunkSizeWarningLimit: 1500,
    rollupOptions: {
      output: {
        // Function form: groups modules by path so tree-shaking still works.
        // All modules from the same library land in one chunk without forcing
        // full package entry points.
        // Keep react, antd, and MUI in one chunk — antd and MUI both call
        // React APIs at module init time, so they must share a chunk with React
        // to avoid "Cannot read properties of undefined" TDZ crashes.
        // Only router and query are truly standalone and safe to split.
        manualChunks(id: string) {
          if (!id.includes('node_modules')) return undefined;
          if (/node_modules\/(react-router|react-router-dom|@remix-run)\//.test(id)) return 'vendor-router';
          if (id.includes('node_modules/@tanstack/')) return 'vendor-query';
          return 'vendor-libs';
        },
      },
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/authorize': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/login': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/consent': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/token': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/.well-known': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/jwks': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/userinfo': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/setup': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  }
})
