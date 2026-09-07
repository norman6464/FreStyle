import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import KbPageIconPicker from './KbPageIconPicker';

/**
 * ページアイコンを選ぶ小さな格子 + 自由入力。
 *
 * 成功したら閉じ、**失敗では開いたままにする**（何が悪かったのか分からないまま
 * 消えるのを避ける — KbInlineRename と同じ約束）。知らせ（トースト）はここでは
 * 出さない。呼び出し側（KbPage）が出す。
 */
const meta = {
  title: 'pages/kb/KbPageIconPicker',
  component: KbPageIconPicker,
  parameters: { layout: 'centered' },
  args: {
    current: { type: 'emoji', value: '📘' },
    onSelect: fn(async () => {}),
    onClear: fn(async () => {}),
    onClose: fn(),
  },
} satisfies Meta<typeof KbPageIconPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 格子の 1 つを押すと、その絵文字で確定して閉じる。 */
export const 一覧から選ぶ: Story = {
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'アイコンを 📙 にする' }));
    await expect(args.onSelect).toHaveBeenCalledWith({ type: 'emoji', value: '📙' });
    await expect(args.onClose).toHaveBeenCalled();
  },
};

/** 一覧に無い絵文字も自由入力から設定できる。 */
export const 自由入力で決める: Story = {
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.type(canvas.getByLabelText('絵文字を入力'), '🐙');
    await userEvent.click(canvas.getByRole('button', { name: '入力した絵文字をアイコンに設定' }));
    await expect(args.onSelect).toHaveBeenCalledWith({ type: 'emoji', value: '🐙' });
    await expect(args.onClose).toHaveBeenCalled();
  },
};

/** 2 文字以上は「1 文字だけ入力してください」で弾く。API は呼ばない。 */
export const 二文字以上は受け付けない: Story = {
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.type(canvas.getByLabelText('絵文字を入力'), 'ab');
    await userEvent.click(canvas.getByRole('button', { name: '入力した絵文字をアイコンに設定' }));
    await expect(canvas.getByRole('alert')).toHaveTextContent('1 文字だけ入力してください');
    await expect(args.onSelect).not.toHaveBeenCalled();
    await expect(args.onClose).not.toHaveBeenCalled();
  },
};

/** 設定済みのときだけ出る「外す」。押すと onClear が飛び、閉じる。 */
export const 外す: Story = {
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'アイコンを外す' }));
    await expect(args.onClear).toHaveBeenCalled();
    await expect(args.onClose).toHaveBeenCalled();
  },
};

/** 未設定（current が null）では「外す」を出さない。外すものが無いため。 */
export const 未設定では外すが無い: Story = {
  args: { current: null },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.queryByRole('button', { name: 'アイコンを外す' })).toBeNull();
  },
};

/** Escape で閉じる（何も選ばずに）。 */
export const Escapeで閉じる: Story = {
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    // ダイアログの onKeyDown は要素の内側からの bubble で拾うので、
    // 中の入力にフォーカスを置いてから打つ（KbSearchDialog と同じ形）。
    await userEvent.type(canvas.getByLabelText('絵文字を入力'), '{Escape}');
    await expect(args.onClose).toHaveBeenCalled();
    await expect(args.onSelect).not.toHaveBeenCalled();
  },
};

/**
 * 失敗しても閉じない。
 *
 * `fn().mockRejectedValue(...)` は使えない（Storybook が story ごとに fn() を
 * reset するため）。`fn(実装)` の形で渡す。
 */
export const 失敗しても閉じない: Story = {
  args: {
    onSelect: fn(async () => {
      throw new Error('invalid_icon');
    }),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'アイコンを 📙 にする' }));
    await expect(args.onSelect).toHaveBeenCalled();
    await expect(args.onClose).not.toHaveBeenCalled();
    // ダイアログはまだ在る。
    await expect(canvas.getByRole('dialog')).toBeInTheDocument();
  },
};
