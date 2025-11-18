import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';

export default defineConfig({
  plugins: [preact()],
  resolve: {
    alias: {
      'react': 'preact/compat',
      'react-dom': 'preact/compat',
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/v1': {
        target: 'http://localhost:8082',
        changeOrigin: true,
        ws: true, // Enable WebSocket proxying for /v1/ws
      },
    },
  },
});
