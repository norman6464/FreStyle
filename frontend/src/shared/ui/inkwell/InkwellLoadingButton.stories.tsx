import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
import InkwellLoadingButton from './InkwellLoadingButton';

/**
 * 押すと中身がぐるぐるに変わり、終わるとレ点（成功）か × （失敗）になるボタン。
 *
 * **二重送信を防ぐ**のが主な役目。ただし押せなくする（`disabled`）のではなく、
 * 「いま処理中」という状態にして受け付けないだけにしてある。押せなくするとフォーカスが
 * 外れてしまい、読み上げソフトを使っている人が結果を聞き逃すため。
 *
 * 幅は固定してあるので、中身が入れ替わってもボタンの大きさは変わらない。
 */
const meta = {
  title: 'shared/inkwell/InkwellLoadingButton',
  component: InkwellLoadingButton,
  parameters: { layout: 'centered' },
} satisfies Meta<typeof InkwellLoadingButton>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 押す前。ふつうのボタンに見える。 */
export const 押す前: Story = {
  args: { children: '送信する', onAction: fn(async () => {}) },
};

/**
 * うまくいったとき。ぐるぐるのあとレ点が描かれる。
 *
 * 失敗・成功の作り方に注意: `fn().mockResolvedValue(...)` は使えない。Storybook が story
 * ごとに fn() を reset するため、あとから足した振る舞いは消える。`fn(実装)` の形で渡す。
 */
export const 成功: Story = {
  args: {
    children: '送信する',
    onAction: fn(async () => {
      await new Promise((resolve) => setTimeout(resolve, 300));
    }),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button'));
    await waitFor(async () => {
      await expect(canvas.getByRole('status')).toHaveTextContent('完了しました');
    });
  },
};

/** 失敗したとき。× になる。 */
export const 失敗: Story = {
  args: {
    children: '送信する',
    onAction: fn(async () => {
      await new Promise((resolve) => setTimeout(resolve, 200));
      throw new Error('送信に失敗しました');
    }),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button'));
    await waitFor(async () => {
      await expect(canvas.getByRole('status')).toHaveTextContent('失敗しました');
    });
  },
};

/** 処理中に押しても二重に走らない。 */
export const 二重送信しない: Story = {
  args: {
    children: '送信する',
    onAction: fn(async () => {
      await new Promise((resolve) => setTimeout(resolve, 400));
    }),
  },
  play: async ({ args, canvasElement }) => {
    const button = within(canvasElement).getByRole('button');
    await userEvent.click(button);
    await userEvent.click(button);
    await userEvent.click(button);
    await expect(args.onAction).toHaveBeenCalledTimes(1);
  },
};

/** 読み上げる文言を変える。 */
export const 文言を変える: Story = {
  args: {
    children: '削除する',
    color: 'error',
    loadingLabel: '削除しています',
    successLabel: '削除しました',
    errorLabel: '削除できませんでした',
    onAction: fn(async () => {}),
  },
};
