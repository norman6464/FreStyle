import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// 本番ビルドでは console.log/info/debug を除去し、必要最小限のログにする
export default defineConfig(({ mode }) => ({
  plugins: [react()],
  define: {
    global: 'window', // ここで global をブラウザの window に置き換える
  },
  esbuild: mode === 'production'
    ? {
        // console.error / console.warn は残し、info/log/debug は削除
        pure: ['console.log', 'console.debug', 'console.info'],
      }
    : undefined,
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // React本体
          'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          // 状態管理
          'vendor-state': ['@reduxjs/toolkit', 'react-redux', '@tanstack/react-query'],
          // TipTapエディタ（重いので分離）
          'vendor-tiptap': [
            '@tiptap/react',
            '@tiptap/starter-kit',
            '@tiptap/core',
          ],
        },
      },
    },
  },
}));
