import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { fileURLToPath, URL } from 'node:url';

// 本番ビルドでは console.log/info/debug を除去し、必要最小限のログにする
export default defineConfig(({ mode }) => ({
  plugins: [react()],
  // '@' → src の絶対パス。FSD は層をまたぐ参照を絶対パスで書く前提なので、
  // tsconfig.json の paths と同じ内容をビルド側にも定義する（FRESTYLE-155）。
  // 型チェック・ビルド・テストの 3 か所すべてに無いと、
  // 「型は通るがビルドで落ちる」状態になるため必ず揃える。
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
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
        // vite 8 のデフォルトバンドラ(Rolldown)はオブジェクト形式の manualChunks を
        // 受け付けず関数形式のみ対応するため、id ベースの関数で同じ分割を表現する。
        // react-router-dom / react-dom を react より先に判定する(部分一致の取りこぼし防止)。
        manualChunks(id) {
          if (!id.includes('node_modules')) return undefined;
          if (id.includes('node_modules/monaco-editor')) return 'vendor-monaco';
          if (
            id.includes('node_modules/@reduxjs/toolkit') ||
            id.includes('node_modules/react-redux') ||
            id.includes('node_modules/@tanstack/react-query')
          ) {
            return 'vendor-state';
          }
          if (
            id.includes('node_modules/react-router-dom') ||
            id.includes('node_modules/react-dom') ||
            id.includes('node_modules/react/')
          ) {
            return 'vendor-react';
          }
          return undefined;
        },
      },
    },
  },
}));
