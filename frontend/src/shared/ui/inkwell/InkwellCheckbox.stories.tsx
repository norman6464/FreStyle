import { useState } from 'react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';
import InkwellCheckbox from './InkwellCheckbox';

/**
 * inkwell のチェックボックス。押すと丸い波紋が広がり、四角が塗られてレ点が出る。
 *
 * 本物の `<input type="checkbox">` は見えない場所に置いてあり、見た目だけを自前で描いている。
 * キーボードでの操作や読み上げは本物のほうが受け持つので、見た目を変えても壊れない。
 */
const meta = {
  title: 'shared/inkwell/InkwellCheckbox',
  component: InkwellCheckbox,
  parameters: { layout: 'centered' },
} satisfies Meta<typeof InkwellCheckbox>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 入っていないとき。 */
export const 未チェック: Story = {
  args: { label: '利用規約に同意する', checked: false, onChange: () => {} },
};

/** 入っているとき。 */
export const チェック済み: Story = {
  args: { label: '利用規約に同意する', checked: true, onChange: () => {} },
};

/** 触れないとき。 */
export const 押せない: Story = {
  args: { label: '利用規約に同意する', checked: true, disabled: true, onChange: () => {} },
};

/** ラベル無しで、表の中などに置く形。 */
export const ラベルなし: Story = {
  args: { checked: false, onChange: () => {}, 'aria-label': '行を選択' },
};

/** 実際に押して切り替える。 */
export const 押して切り替える: Story = {
  args: { label: 'メールで知らせを受け取る' },
  render: (args) => {
    function Interactive() {
      const [checked, setChecked] = useState(false);
      return (
        <InkwellCheckbox {...args} checked={checked} onChange={(e) => setChecked(e.target.checked)} />
      );
    }
    return <Interactive />;
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const box = canvas.getByRole('checkbox', { name: 'メールで知らせを受け取る' });
    await expect(box).not.toBeChecked();
    await userEvent.click(box);
    await expect(box).toBeChecked();
  },
};
