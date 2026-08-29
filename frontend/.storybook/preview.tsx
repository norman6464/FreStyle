import type { Preview } from '@storybook/react-vite';
// アプリと同じ見た目で検証できるよう、本体のグローバル CSS を通す。
// これが無いと --color-* トークンも prose も効かず、見本にならない。
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
      // 見つけたら落とす。story は CI でテストとして走るので、
      // 'todo'（表示だけ）にしておくと違反が積もっても誰も気づけない。
      test: 'error'
    }
  },
};

export default preview;