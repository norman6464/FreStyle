import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
import KbCommentComposer from './KbCommentComposer';

/**
 * コメント 1 件ぶんの入力欄（複数行 + 送信）。
 *
 * 本文はリッチな装飾を持たない「1 段落ぶんのプレーンテキスト」として送る
 * （`[{type:'text', text}]` の形。本文の RichTextEditor ほどの入力は今回のスコープ外）。
 *
 * **失敗しても入力は消さない。** 消すと書き直しになるうえ、何が悪かったのか分からない
 * （KbPageTitle・KbPageIconPicker と同じ約束）。
 */
const meta = {
  title: 'pages/kb/KbCommentComposer',
  component: KbCommentComposer,
  parameters: { layout: 'padded' },
  args: {
    placeholder: 'コメントを書く…',
    onSubmit: fn(async () => {}),
  },
  decorators: [
    (Story) => (
      <div className="w-96">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof KbCommentComposer>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 書いて送る。送った本文はプレーンテキストをインラインノード 1 個に包んだ形。 */
export const 入力して送信: Story = {
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    const textarea = canvas.getByPlaceholderText('コメントを書く…');
    const send = canvas.getByRole('button', { name: '送信' });
    await expect(send).toBeDisabled();

    await userEvent.type(textarea, 'これはどういう意味ですか？');
    await expect(send).toBeEnabled();

    await userEvent.click(send);
    await expect(args.onSubmit).toHaveBeenCalledWith([
      { type: 'text', text: 'これはどういう意味ですか？' },
    ]);
    // 送信が成功すると入力欄は空に戻る。
    await waitFor(async () => {
      await expect(textarea).toHaveValue('');
    });
  },
};

/** 空・空白のみでは送信できない。 */
export const 空では送信不可: Story = {
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    const textarea = canvas.getByPlaceholderText('コメントを書く…');
    const send = canvas.getByRole('button', { name: '送信' });
    await expect(send).toBeDisabled();

    await userEvent.type(textarea, '   ');
    await expect(send).toBeDisabled();
    await expect(args.onSubmit).not.toHaveBeenCalled();
  },
};

/** 送信に失敗したとき。エラーを出し、入力は消さない（書き直させない）。 */
export const 失敗時にエラー表示し入力を保持: Story = {
  args: {
    onSubmit: fn(async () => {
      throw new Error('network error');
    }),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const textarea = canvas.getByPlaceholderText('コメントを書く…');
    const send = canvas.getByRole('button', { name: '送信' });

    await userEvent.type(textarea, '消えてほしくない下書き');
    await userEvent.click(send);

    await expect(await canvas.findByRole('alert')).toHaveTextContent('送信できませんでした');
    // 入力は消さない。
    await expect(textarea).toHaveValue('消えてほしくない下書き');
    // 押し直せる（disabled のままにしない）。
    await expect(send).toBeEnabled();
  },
};
