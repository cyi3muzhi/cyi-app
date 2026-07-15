/**
 * Vite 配置
 */
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/admin/v1': {
        target: 'http://localhost:28000',
        changeOrigin: true
      }
    }
  }
})
