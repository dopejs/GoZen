import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// API port: use VITE_API_PORT env var or default to 19840 (production)
const apiPort = process.env.VITE_API_PORT || '19840'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: `http://127.0.0.1:${apiPort}`,
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
