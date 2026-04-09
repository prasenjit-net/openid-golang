import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  base: '/',
  build: {
    outDir: '../ui/dist',
    assetsDir: 'assets',
    emptyOutDir: true,
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        // Function form: groups modules by path so tree-shaking still works.
        // All modules from the same library land in one chunk without forcing
        // full package entry points.
        manualChunks(id: string) {
          if (!id.includes('node_modules')) return undefined;
          // React core — include scheduler (react-dom peer dep)
          if (/node_modules\/(react|react-dom|scheduler)\//.test(id)) return 'vendor-react';
          // React Router + its Remix core deps
          if (/node_modules\/(react-router|react-router-dom|@remix-run)\//.test(id)) return 'vendor-router';
          // TanStack Query — both react-query wrapper and query-core
          if (id.includes('node_modules/@tanstack/')) return 'vendor-query';
          // MUI + Emotion styling engine
          if (/node_modules\/(@mui|@emotion)\//.test(id)) return 'vendor-mui';
          // antd + all its sub-packages: @ant-design/*, rc-* components, @rc-component/*
          if (/node_modules\/(antd|@ant-design|rc-[^/]+|@rc-component)\//.test(id)) return 'vendor-antd';
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
