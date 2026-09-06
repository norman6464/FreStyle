import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
import GuidedHint from './GuidedHint';

/**
 * 画面の使い方を、その場で短く添える帯。
 *
 * 一度閉じたら二度と出さない（`storageKey` を渡したとき）。毎回出ると、慣れた人には
 * ただの邪魔になるため。
 *
 * この見本では `storageKey` を**わざと渡していない**。渡すと一度閉じたきり
 * 見本が空っぽになり、次に開いた人が何も確かめられなくなる。
 */
const meta = {
  title: 'shared/GuidedHint',
  component: GuidedHint,
  parameters: { layout: 'padded' },
  decorators: [
    (Story) => (
      <div className="max-w-2xl">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof GuidedHint>;

export default meta;
type Story = StoryObj<typeof meta>;

/** ふつうの案内。 */
export const 案内: Story = {
  args: {
    title: 'まずは言語を選びましょう',
    children: '選んだ言語の演習だけが一覧に出ます。あとから何度でも変えられます。',
  },
};

/** うまくいったことを伝える。 */
export const 成功: Story = {
  args: {
    tone: 'success',
    title: '保存されました',
    children: '書いた内容は自動で保存されます。手が止まってから少し待つだけで残ります。',
  },
};

/** 気をつけてほしいこと。 */
export const 注意: Story = {
  args: {
    tone: 'warning',
    title: '削除すると元に戻せません',
    children: 'ページを削除すると、中の子ページもまとめて消えます。',
  },
};

/** 3 つのトーンを並べて比べる。 */
export const トーンぜんぶ: Story = {
  args: { title: '', children: null },
  render: () => (
    <div className="flex flex-col gap-3">
      <GuidedHint title="案内">これから何をするかを伝えます。</GuidedHint>
      <GuidedHint tone="success" title="成功">うまくいったことを伝えます。</GuidedHint>
      <GuidedHint tone="warning" title="注意">気をつけてほしいことを伝えます。</GuidedHint>
    </div>
  ),
};

/** 閉じられない形（どうしても読んでほしいとき）。 */
export const 閉じられない: Story = {
  args: {
    dismissible: false,
    title: 'ここはお試し用の環境です',
    children: '書いた内容は毎晩消えます。本番のデータは入れないでください。',
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).queryByRole('button', { name: 'ヒントを閉じる' })).toBeNull();
  },
};

/** ✕ を押すと消える。 */
export const 閉じたとき: Story = {
  args: {
    title: 'まずは言語を選びましょう',
    children: '選んだ言語の演習だけが一覧に出ます。',
    onDismiss: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'ヒントを閉じる' }));
    await waitFor(async () => {
      await expect(canvas.queryByText('まずは言語を選びましょう')).toBeNull();
    });
    await expect(args.onDismiss).toHaveBeenCalled();
  },
};
