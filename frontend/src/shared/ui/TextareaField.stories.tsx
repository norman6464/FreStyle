import { useState } from 'react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import TextareaField from './TextareaField';

/**
 * 複数行の入力欄。
 *
 * `maxLength` を渡すと右下に「いま何文字 / 上限」が出る。上限の 9 割で黄色、
 * 上限に達すると赤 — **打ち終わってから怒られる**のを避けるための予告。
 */
const meta = {
  title: 'shared/TextareaField',
  component: TextareaField,
  parameters: { layout: 'centered' },
  args: { onChange: fn() },
  decorators: [
    (Story) => (
      <div className="w-96">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof TextareaField>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 空のとき。 */
export const 空: Story = {
  args: { label: '自己紹介', name: 'bio', value: '', placeholder: 'どんなことを学んでいますか？' },
};

/** 文字数の残りを見せる。まだ余裕があるので灰色。 */
export const 残り文字数: Story = {
  args: { label: '自己紹介', name: 'bio', value: 'Go と React を勉強しています。', maxLength: 100 },
};

/** 上限の 9 割。黄色で「そろそろ」を予告する。 */
export const 上限が近い: Story = {
  args: { label: '自己紹介', name: 'bio', value: 'あ'.repeat(92), maxLength: 100 },
};

/** 上限ちょうど。赤くなり、これ以上打てない。 */
export const 上限に達した: Story = {
  args: { label: '自己紹介', name: 'bio', value: 'あ'.repeat(100), maxLength: 100 },
};

/** 間違っているとき。 */
export const エラー: Story = {
  args: { label: '自己紹介', name: 'bio', value: '', error: '自己紹介を入力してください' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByLabelText('自己紹介')).toHaveAttribute(
      'aria-invalid',
      'true',
    );
  },
};

/** 実際に打てる形。打つと文字数の表示が増える。 */
export const 打ってみる: Story = {
  args: { label: 'ひとこと', name: 'memo', value: '', maxLength: 20 },
  render: (args) => {
    // この部品は自分では値を持たない（親が持つ）ので、story 側で持たせて動かす。
    function Interactive() {
      const [value, setValue] = useState('');
      return <TextareaField {...args} value={value} onChange={(e) => setValue(e.target.value)} />;
    }
    return <Interactive />;
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.type(canvas.getByLabelText('ひとこと'), 'こんにちは');
    await expect(canvas.getByText('5 / 20')).toBeVisible();
  },
};
