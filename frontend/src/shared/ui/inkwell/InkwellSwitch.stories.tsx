import { useState } from 'react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';
import InkwellSwitch from './InkwellSwitch';

/**
 * inkwell の切り替えスイッチ。つまみがレールの上を滑る。
 *
 * チェックボックスとの使い分け: スイッチは**その場ですぐ効く**設定（通知を受け取る / 受け取らない）、
 * チェックボックスは**あとで送信する**フォームの項目に使う。
 */
const meta = {
  title: 'shared/inkwell/InkwellSwitch',
  component: InkwellSwitch,
  parameters: { layout: 'centered' },
} satisfies Meta<typeof InkwellSwitch>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 切のとき。 */
export const 切: Story = {
  args: { label: '通知を受け取る', checked: false, onChange: () => {} },
};

/** 入のとき。 */
export const 入: Story = {
  args: { label: '通知を受け取る', checked: true, onChange: () => {} },
};

/** 触れないとき。 */
export const 押せない: Story = {
  args: { label: '通知を受け取る', checked: true, disabled: true, onChange: () => {} },
};

/** 実際に押して切り替える。 */
export const 押して切り替える: Story = {
  args: { label: '通知を受け取る' },
  render: (args) => {
    function Interactive() {
      const [checked, setChecked] = useState(false);
      return (
        <InkwellSwitch {...args} checked={checked} onChange={(e) => setChecked(e.target.checked)} />
      );
    }
    return <Interactive />;
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const toggle = canvas.getByRole('checkbox', { name: '通知を受け取る' });
    await userEvent.click(toggle);
    await expect(toggle).toBeChecked();
  },
};
