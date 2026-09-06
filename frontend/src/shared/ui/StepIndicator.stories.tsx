import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import StepIndicator from './StepIndicator';

/**
 * 何段階かある手順の「いまどこか」を示す帯。
 *
 * 済んだところは緑のチェック、いまのところは色つきの丸、これからは灰色。
 * 色だけに頼らず、読み上げソフト向けに「ステップ 2 / 3・実行中」という文も隠して持たせてある。
 */
const meta = {
  title: 'shared/StepIndicator',
  component: StepIndicator,
  parameters: { layout: 'padded' },
  args: {
    steps: [
      { label: 'シナリオを選ぶ' },
      { label: '会話する' },
      { label: 'スコアを見る' },
    ],
  },
  decorators: [
    (Story) => (
      <div className="max-w-2xl">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof StepIndicator>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 始めたところ。 */
export const 最初: Story = {
  args: { currentStep: 0 },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('list', { name: '進行状況' })).toBeVisible();
  },
};

/** 途中。左は済み、真ん中がいま。 */
export const 途中: Story = {
  args: { currentStep: 1 },
};

/** 最後まで来たところ。 */
export const 最後: Story = {
  args: { currentStep: 2 },
};

/** 全部済んだところ。 */
export const 全部済み: Story = {
  args: { currentStep: 3 },
};

/** 各段に補足を添える。 */
export const 補足つき: Story = {
  args: {
    currentStep: 1,
    steps: [
      { label: 'シナリオを選ぶ', description: '場面を決めます' },
      { label: '会話する', description: 'AI と話します' },
      { label: 'スコアを見る', description: '5 つの観点で振り返ります' },
    ],
  },
};
