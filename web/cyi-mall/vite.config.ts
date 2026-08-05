/**
 * cyi-mall Vite 配置
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
    port: 3001,
    proxy: {
      // 用户端接口 -> interface 服务（端口 18000）
      '/user/v1': {
        target: 'http://localhost:18000',
        changeOrigin: true
      },
      '/product/v1': {
        target: 'http://localhost:18000',
        changeOrigin: true
      },
      '/category/v1': {
        target: 'http://localhost:18000',
        changeOrigin: true
      },
      '/cart/v1': {
        target: 'http://localhost:18000',
        changeOrigin: true
      },
      '/order/v1': {
        target: 'http://localhost:18000',
        changeOrigin: true
      },
      '/settings/v1': {
        target: 'http://localhost:18000',
        changeOrigin: true
      },
      '/message/v1': {
        target: 'http://localhost:18000',
        changeOrigin: true
      },
      '/api/v1': {
        target: 'http://localhost:18000',
        changeOrigin: true
      }
    }
  }
})
