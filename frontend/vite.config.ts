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
