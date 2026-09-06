import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, waitFor, within } from 'storybook/test';
import { withApi, withRouter } from '../../../../.storybook/decorators';
import ExerciseLanguageSelectPage from './ExerciseLanguageSelectPage';

/**
 * 演習をはじめる前に、どの言語をやるか選ぶ画面。
 *
 * 各カードに「何問中いくつ解いたか」を出す。数が見えると、続きから戻ってきたときに
 * どこまでやったかを思い出さなくて済む。
 */
const meta = {
  title: 'pages/exercise-languages/ExerciseLanguageSelectPage',
  component: ExerciseLanguageSelectPage,
  parameters: { layout: 'fullscreen' },
  decorators: [
    withRouter,
    (Story) => (
      <div className="bg-surface p-6">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof ExerciseLanguageSelectPage>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 進み具合つき。 */
export const 既定: Story = {
  decorators: [
    withApi({
      '/exercises/summary': [
        { language: 'go', total: 24, solved: 7 },
        { language: 'typescript', total: 18, solved: 18 },
        { language: 'python', total: 12, solved: 0 },
        { language: 'docker', total: 9, solved: 3 },
      ],
    }),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByRole('heading', { level: 1 })).toBeVisible();
    });
  },
};

/** まだ 1 問も解いていないとき。 */
export const 未着手: Story = {
  decorators: [
    withApi({
      '/exercises/summary': [
        { language: 'go', total: 24, solved: 0 },
        { language: 'typescript', total: 18, solved: 0 },
      ],
    }),
  ],
};

/** 集計が取れなかったとき。 */
export const 取得に失敗: Story = {
  decorators: [withApi({})],
};
