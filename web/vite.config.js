import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': { target: 'http://localhost:8789', changeOrigin: true },
    },
  },
  build: {
    assetsDir: 'assets',
    chunkSizeWarningLimit: 900,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return

          const parts = id.split('node_modules/')[1]?.split('/') || []
          if (!parts.length) return 'vendor-misc'

          const packageName = parts[0].startsWith('@') ? `${parts[0]}-${parts[1] || 'pkg'}` : parts[0]

          if (packageName === 'element-plus') {
            return 'vendor-element-plus'
          }

          if (packageName === 'vue-router') {
            return 'vendor-router'
          }
          if (packageName === 'pinia') {
            return 'vendor-store'
          }
          if (packageName === 'axios') {
            return 'vendor-http'
          }
          if (packageName === 'vue' || packageName === '@vue-shared' || packageName === '@vue-runtime-core' || packageName === '@vue-runtime-dom' || packageName === '@vue-reactivity') {
            return 'vendor-vue'
          }

          return 'vendor-misc'
        },
      },
    },
  },
})
