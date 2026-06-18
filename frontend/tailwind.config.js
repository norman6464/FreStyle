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
        // FreStyle ブランドカラー（Tailwind blue スケール準拠）。
        // #2E7DF6 は彩度が高すぎて長時間凝視すると目が疲れるため、
        // 科学的に「認知負荷が低く集中力が上がる」中明度青（#3B82F6 = Tailwind blue-500）に調整。
        // 使い分けルール:
        //   brand-* → CTA ボタン / フォーカスリング / インタラクティブ要素
        //   primary-* (taupe) → ステップインジケーター / プログレスバー / アバター背景などの非ボタンアクセント
        brand: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        },
        surface: 'var(--color-surface)',
        'surface-1': 'var(--color-surface-1)',
        'surface-2': 'var(--color-surface-2)',
        'surface-3': 'var(--color-surface-3)',
      },
      // 角丸ルール:
      //   rounded-xl  → カード / パネル / モーダル / フローティングメニュー
      //   rounded-lg  → ボタン / インプット / セレクト / タグ / バッジ
      //   rounded-full → アバター / 円形アイコンボタン / ドット / ピル
      //   rounded-md  → 小さいインライン要素（コードブロック枠 / ミニアイコンボタン等）
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
