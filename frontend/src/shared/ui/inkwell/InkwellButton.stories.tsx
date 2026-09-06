import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import InkwellButton from './InkwellButton';

/**
 * inkwell のボタン。押した場所から波紋が広がり、面が少し浮く。
 *
 * inkwell は、アプリ本体とは別に用意してある「触った感じ」を確かめるための部品の一群。
 * 本体の Button とは見た目の作法が違う（大文字・波紋・影の段）ので、混ぜて使わない。
 *
 * - `contained` … 面を色で塗る。いちばん強い
 * - `outlined` … 縁だけ
 * - `text` … 文字だけ
 */
const meta = {
  title: 'shared/inkwell/InkwellButton',
  component: InkwellButton,
  parameters: { layout: 'centered' },
  args: { children: 'ボタン', onClick: fn() },
} satisfies Meta<typeof InkwellButton>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 面を塗る形。 */
export const 塗り: Story = {
  args: { variant: 'contained', children: '送信する' },
};

/** 縁だけの形。 */
export const 枠線: Story = {
  args: { variant: 'outlined', children: '送信する' },
};

/** 文字だけの形。 */
export const 文字のみ: Story = {
  args: { variant: 'text', children: '送信する' },
};

/** 3 つの形 × 3 つの色を並べて見る。 */
export const 形と色: Story = {
  render: (args) => (
    <div className="flex flex-col gap-3">
      {(['contained', 'outlined', 'text'] as const).map((variant) => (
        <div key={variant} className="flex items-center gap-3">
          {(['primary', 'secondary', 'error'] as const).map((color) => (
            <InkwellButton key={color} {...args} variant={variant} color={color}>
              {variant}
            </InkwellButton>
          ))}
        </div>
      ))}
    </div>
  ),
};

/** 大きさ 3 段。 */
export const 大きさ: Story = {
  render: (args) => (
    <div className="flex items-center gap-3">
      <InkwellButton {...args} size="small">小</InkwellButton>
      <InkwellButton {...args} size="medium">中</InkwellButton>
      <InkwellButton {...args} size="large">大</InkwellButton>
    </div>
  ),
};

/** 押せないとき。色が抜け、波紋も出ない。 */
export const 押せない: Story = {
  args: { disabled: true, children: '送信する' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('button')).toBeDisabled();
  },
};

/** 横幅いっぱい。 */
export const 幅いっぱい: Story = {
  args: { fullWidth: true, children: '送信する' },
  decorators: [
    (Story) => (
      <div className="w-72">
        <Story />
      </div>
    ),
  ],
};

/** 押すと onClick が呼ばれる。 */
export const 押したとき: Story = {
  args: { children: '押してみる' },
  play: async ({ args, canvasElement }) => {
    await userEvent.click(within(canvasElement).getByRole('button'));
    await expect(args.onClick).toHaveBeenCalledTimes(1);
  },
};
