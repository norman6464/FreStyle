import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import KbSuggestDraftBanner from './KbSuggestDraftBanner';

/**
 * ドラフトモード中に本文の上へ出す帯。KbVersionPreviewBanner と同じ見た目
 * （role="status"・amber 系の左ボーダー）で、「送信」「キャンセル」の 2 ボタンを持つ。
 */
const meta = {
  title: 'pages/kb/KbSuggestDraftBanner',
  component: KbSuggestDraftBanner,
  parameters: { layout: 'padded' },
  args: { submitting: false, error: null, onSubmit: fn(), onCancel: fn() },
} satisfies Meta<typeof KbSuggestDraftBanner>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 通常状態。文言と両ボタンが見える。 */
export const 通常: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('status')).toHaveTextContent('提案として保存されます');
    await expect(canvas.getByRole('button', { name: '送信' })).toBeEnabled();
    await expect(canvas.getByRole('button', { name: 'キャンセル' })).toBeEnabled();
  },
};

/** 送信を押すと onSubmit が呼ばれる。 */
export const 送信を押す: Story = {
  play: async ({ canvasElement, args }) => {
    await userEvent.click(within(canvasElement).getByRole('button', { name: '送信' }));
    await expect(args.onSubmit).toHaveBeenCalledTimes(1);
  },
};

/** キャンセルを押すと onCancel が呼ばれる。 */
export const キャンセルを押す: Story = {
  play: async ({ canvasElement, args }) => {
    await userEvent.click(within(canvasElement).getByRole('button', { name: 'キャンセル' }));
    await expect(args.onCancel).toHaveBeenCalledTimes(1);
  },
};

/** 送信中は両ボタンとも押せない。 */
export const 送信中: Story = {
  args: { submitting: true },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('button', { name: '送信' })).toBeDisabled();
    await expect(canvas.getByRole('button', { name: 'キャンセル' })).toBeDisabled();
  },
};

/** 失敗すると帯の中にエラーが出る。ドラフトモードは終わらない（呼び出し側の約束）。 */
export const 送信に失敗: Story = {
  args: { error: '提案を送信できませんでした。' },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('alert')).toHaveTextContent('提案を送信できませんでした');
    // 失敗しても両ボタンは押せる（消えていない）。
    await expect(canvas.getByRole('button', { name: '送信' })).toBeEnabled();
  },
};
