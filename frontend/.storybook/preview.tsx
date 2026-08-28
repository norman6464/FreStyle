import type { Preview } from '@storybook/react-vite';
// アプリと同じ見た目で検証できるよう、本体のグローバル CSS と書体を通す。
// これが無いと --color-* トークンも prose も効かず、見本にならない。
import '@fontsource/noto-sans-jp/400.css';
import '@fontsource/noto-sans-jp/500.css';
import '@fontsource/noto-sans-jp/600.css';
import '@fontsource/noto-sans-jp/700.css';
import '../src/app/styles/index.css';

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
       color: /(background|color)$/i,
       date: /Date$/i,
      },
    },

    a11y: {
      // 'todo' - show a11y violations in the test UI only
      // 'error' - fail CI on a11y violations
      // 'off' - skip a11y checks entirely
      test: 'todo'
    }
  },
};

export default preview;