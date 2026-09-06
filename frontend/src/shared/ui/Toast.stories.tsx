import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import Toast from './Toast';

/**
 * 画面の上から落ちてくる短い知らせ。
 *
 * 4 秒で自分から消える。**消えても困らないこと**にだけ使う — 見逃すと進めなくなる情報は
 * その場（フォームの中）に出す。
 *
 * 置き場所（画面上部の中央）は ToastContainer 側の仕事で、この部品は見た目と自動で消える
 * ところだけを持つ。
 */
const meta = {
  title: 'shared/Toast',
  component: Toast,
  parameters: { layout: 'centered' },
  args: { onClose: fn() },
} satisfies Meta<typeof Toast>;

export default meta;
type Story = StoryObj<typeof meta>;

/** うまくいったとき。 */
export const 成功: Story = {
  args: { type: 'success', message: 'ページを保存しました' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('alert')).toHaveTextContent('ページを保存しました');
  },
};

/** 失敗したとき。 */
export const 失敗: Story = {
  args: { type: 'error', message: '保存できませんでした。通信を確認してください' },
};

/** ただのお知らせ。 */
export const お知らせ: Story = {
  args: { type: 'info', message: '共有リンクをコピーしました' },
};

/** 3 種類を並べて比べる。 */
export const 種類ぜんぶ: Story = {
  args: { type: 'success', message: '' },
  render: (args) => (
    <div className="flex flex-col gap-3">
      <Toast {...args} type="success" message="ページを保存しました" />
      <Toast {...args} type="error" message="保存できませんでした" />
      <Toast {...args} type="info" message="共有リンクをコピーしました" />
    </div>
  ),
};

/** 長い文でも折り返して収まる。 */
export const 長い文: Story = {
  args: {
    type: 'error',
    message:
      '保存できませんでした。ネットワークに接続していないか、ほかの人が同じページを編集しています。しばらく待ってからもう一度お試しください。',
  },
};

/** ✕ を押すと閉じる。 */
export const 閉じたとき: Story = {
  args: { type: 'info', message: '共有リンクをコピーしました' },
  play: async ({ args, canvasElement }) => {
    await userEvent.click(within(canvasElement).getByRole('button', { name: '閉じる' }));
    await expect(args.onClose).toHaveBeenCalled();
  },
};
