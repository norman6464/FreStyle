import { useState } from 'react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';
import InkwellTextField from './InkwellTextField';

/**
 * inkwell の入力欄。ラベルが枠の中から**枠線の上に浮き上がる**。
 *
 * 空のときはラベルが案内文の位置に居て、打ちはじめると小さくなって上へ移る。
 * ラベルが消えないので、打ち終わったあとでも「何を入れる欄だったか」が分かる。
 */
const meta = {
  title: 'shared/inkwell/InkwellTextField',
  component: InkwellTextField,
  parameters: { layout: 'centered' },
  decorators: [
    (Story) => (
      // ラベルは白地で枠線を切り欠く作りなので、白の上で見ないと切り欠きがずれて見える。
      <div className="w-80 bg-white p-6">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof InkwellTextField>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 空のとき。ラベルは枠の中に居る。 */
export const 空: Story = {
  args: { label: 'メールアドレス' },
};

/** 何か入っているとき。ラベルは上に浮いている。 */
export const 入力済み: Story = {
  args: { label: 'メールアドレス', defaultValue: 'takuma@example.com' },
};

/** 補足を添える。 */
export const 補足つき: Story = {
  args: { label: 'ユーザー名', helperText: '半角英数字とハイフンが使えます' },
};

/** 間違っているとき。枠も文字も赤くなる。 */
export const エラー: Story = {
  args: {
    label: 'メールアドレス',
    defaultValue: 'takuma',
    error: true,
    helperText: 'メールアドレスの形式が正しくありません',
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByLabelText('メールアドレス')).toHaveAttribute(
      'aria-invalid',
      'true',
    );
  },
};

/** 触れないとき。 */
export const 押せない: Story = {
  args: { label: 'メールアドレス', defaultValue: 'takuma@example.com', disabled: true },
};

/** 横幅いっぱい。 */
export const 幅いっぱい: Story = {
  args: { label: 'メールアドレス', fullWidth: true },
};

/** 打つとラベルが上へ移る。 */
export const 打つとラベルが浮く: Story = {
  args: { label: 'ユーザー名' },
  render: (args) => {
    function Interactive() {
      const [value, setValue] = useState('');
      return <InkwellTextField {...args} value={value} onChange={(e) => setValue(e.target.value)} />;
    }
    return <Interactive />;
  },
  play: async ({ canvasElement }) => {
    const input = within(canvasElement).getByLabelText('ユーザー名');
    await userEvent.type(input, 'takuma');
    await expect(input).toHaveValue('takuma');
  },
};
