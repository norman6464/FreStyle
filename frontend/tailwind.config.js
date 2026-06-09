import typography from '@tailwindcss/typography';

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#F0EFED',
          100: '#E3E2DF',
          200: '#D3D1CD',
          300: '#B5B2AC',
          400: '#918D85',
          500: '#5C5850',
          600: '#474440',
          700: '#353330',
          800: '#242220',
          900: '#181716',
        },
        // FreStyle ブランドカラー（ロゴの青 #2E7DF6）。アクション系ボタンの色をこれに統一する。
        // primary(taupe) はチャート / プログレス等の非ボタンアクセントに引き続き使う。
        brand: {
          50: '#EFF5FF',
          100: '#DBE8FE',
          200: '#BFD6FE',
          300: '#93BBFD',
          400: '#6098FA',
          500: '#2E7DF6',
          600: '#1B62E0',
          700: '#1A4FB8',
          800: '#1B4496',
          900: '#1C3D7A',
        },
        surface: 'var(--color-surface)',
        'surface-1': 'var(--color-surface-1)',
        'surface-2': 'var(--color-surface-2)',
        'surface-3': 'var(--color-surface-3)',
      },
      borderRadius: {
        xl: '1rem',
        '2xl': '1.5rem',
      },
      animation: {
        'fade-in': 'fadeIn 0.15s ease-in',
        'scale-in': 'scaleIn 0.2s ease-out',
        'skeleton': 'skeleton 1.5s ease-in-out infinite',
        // Toast 用: 上から落ちてバウンドする 0.6s のアニメーション。
        'toast-drop': 'toastDrop 0.6s cubic-bezier(0.34, 1.56, 0.64, 1)',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' },
        },
        skeleton: {
          '0%': { backgroundPosition: '200% 0' },
          '100%': { backgroundPosition: '-200% 0' },
        },
        toastDrop: {
          '0%': { transform: 'translateY(-120%)', opacity: '0' },
          '60%': { transform: 'translateY(12px)', opacity: '1' },
          '80%': { transform: 'translateY(-6px)' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
      },
    },
  },
  // @tailwindcss/typography: Markdown 描画 (ノート / 教材 / AI チャット) で `prose` クラスを使う。
  // これで `# 見出し` `表` `リスト` `引用` などが標準の Typography スタイルで描画される。
  plugins: [typography],
};
