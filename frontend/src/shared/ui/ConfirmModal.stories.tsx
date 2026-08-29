import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, screen, within } from 'storybook/test';
import ConfirmModal from './ConfirmModal';

/**
 * 確認モーダル。ブラウザ標準の confirm/alert の代わりに使う
 * （見た目が周りと揃い、文言とボタンの並びをこちらで統べられる）。
 * サイドバーのページ削除がこの部品で確かめる。
 */
const meta = {
  title: 'shared/ConfirmModal',
  component: ConfirmModal,
  parameters: { layout: 'fullscreen' },
  args: { onConfirm: fn(), onCancel: fn() },
} satisfies Meta<typeof ConfirmModal>;

export default meta;
type Story = StoryObj<typeof meta>;

/** サイドバーの「削除」で開く形。 */
export const ページの削除: Story = {
  args: {
    isOpen: true,
    title: 'ページを削除',
    message:
      '「ネスト2」を中のページごと削除します（アーカイブ済みの子ページも含みます）。元に戻せません。',
    confirmText: '削除',
    isDanger: true,
  },
  play: async ({ args, canvasElement }) => {
    // モーダルは document.body へポータルされるので canvasElement の中には無い
    // （呼び出し元の DOM に閉じ込めないための設計。ConfirmModal のコメント参照）。
    await expect(within(canvasElement).queryByRole('dialog')).toBeNull();

    const dialog = screen.getByRole('dialog', { name: 'ページを削除' });
    await expect(dialog).toBeInTheDocument();
    // 危険側のボタンを押すと onConfirm が 1 回だけ呼ばれる。
    (await within(dialog).findByRole('button', { name: '削除' })).click();
    await expect(args.onConfirm).toHaveBeenCalledOnce();
  },
};

/** 危険ではない確認（青いボタン側）。 */
export const 通常の確認: Story = {
  args: {
    isOpen: true,
    title: '確認',
    message: 'この内容で送信しますか？',
    confirmText: '送信',
    isDanger: false,
  },
};
