import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
import KbPageTitle from './KbPageTitle';

/**
 * ナレッジのページ見出し。書ける人には、見出しそのものが入力欄になる。
 *
 * 「編集」ボタンを別に置いていないのは、押してから直す一手間を挟まないため。読むだけの人には
 * ただの見出しに見え、書ける人はそのまま打ち替えられる。
 *
 * Enter か欄外を押すと確定。Escape で打ちかけを捨てる。空のまま確定したときは**改名しない**
 * （題名が空だと、木の中で押す場所そのものが無くなる）。
 *
 * 失敗しても入力は消さない。サイドバーの改名・作成フォームと同じ約束にしてある。
 */
const meta = {
  title: 'pages/kb/KbPageTitle',
  component: KbPageTitle,
  parameters: { layout: 'padded' },
  args: { title: '設計メモ', canEdit: true, onRename: fn(async () => {}) },
  decorators: [
    (Story) => (
      <div className="max-w-2xl p-4">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof KbPageTitle>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 書ける人が見たとき。 */
export const 書ける: Story = {
  args: {},
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByDisplayValue('設計メモ')).toBeInTheDocument();
  },
};

/** 読むだけの人が見たとき。ただの見出しになる。 */
export const 読むだけ: Story = {
  args: { canEdit: false },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('設計メモ')).toBeVisible();
    await expect(canvas.queryByRole('textbox')).toBeNull();
  },
};

/** 題名が長いとき。折り返して収まる。 */
export const 長い題名: Story = {
  args: {
    title: '新しく入った人が最初の 1 週間で読むべきものと、その順番についての決めごと',
  },
};

/** 打ち替えて Enter。確定すると親へ渡り、本文へ移る合図も出る。 */
export const 打ち替えて確定: Story = {
  args: { onEnter: fn() },
  play: async ({ args, canvasElement }) => {
    const input = within(canvasElement).getByDisplayValue('設計メモ');
    await userEvent.clear(input);
    await userEvent.type(input, '設計メモ（改訂）{Enter}');
    await expect(args.onRename).toHaveBeenCalledWith('設計メモ（改訂）');
    await expect(args.onEnter).toHaveBeenCalled();
  },
};

/** 空のまま確定したとき。改名しない。 */
export const 空にはできない: Story = {
  args: {},
  play: async ({ args, canvasElement }) => {
    const input = within(canvasElement).getByDisplayValue('設計メモ');
    await userEvent.clear(input);
    await userEvent.type(input, '{Enter}');
    await expect(args.onRename).not.toHaveBeenCalled();
    await waitFor(async () => {
      await expect(input).toHaveValue('設計メモ');
    });
  },
};

/** Escape で打ちかけを捨てる。 */
export const 取り消す: Story = {
  args: {},
  play: async ({ args, canvasElement }) => {
    const input = within(canvasElement).getByDisplayValue('設計メモ');
    await userEvent.clear(input);
    await userEvent.type(input, '打ちかけ{Escape}');
    await expect(args.onRename).not.toHaveBeenCalled();
    await waitFor(async () => {
      await expect(input).toHaveValue('設計メモ');
    });
  },
};

/**
 * 保存に失敗したとき。打った文字はそのまま残る。
 *
 * 失敗の作り方に注意: `fn().mockRejectedValue(...)` は使えない（Storybook が story ごとに
 * fn() を reset するため）。`fn(実装)` の形で渡す。
 */
export const 失敗しても消えない: Story = {
  args: {
    onRename: fn(async () => {
      throw new Error('保存に失敗しました');
    }),
  },
  play: async ({ canvasElement }) => {
    const input = within(canvasElement).getByDisplayValue('設計メモ');
    await userEvent.clear(input);
    await userEvent.type(input, '新しい題名{Enter}');
    await waitFor(async () => {
      await expect(input).toHaveValue('新しい題名');
    });
  },
};
