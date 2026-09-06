import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import FormFieldError from './FormFieldError';

/**
 * 入力欄の下に出す 1 行のエラー文。
 *
 * `id` は `<入力欄の name>-error` の形で付く。入力欄側の `aria-describedby` がこの id を
 * 指すことで、読み上げソフトが「この欄のエラー」として読む。
 * `error` が無いときは**何も描かない** — 空の場所を確保すると、行が跳ねて読みにくい。
 */
const meta = {
  title: 'shared/FormFieldError',
  component: FormFieldError,
  parameters: { layout: 'centered' },
} satisfies Meta<typeof FormFieldError>;

export default meta;
type Story = StoryObj<typeof meta>;

/** エラーがあるとき。 */
export const 表示: Story = {
  args: { name: 'email', error: 'メールアドレスを入力してください' },
  play: async ({ canvasElement }) => {
    const alert = within(canvasElement).getByRole('alert');
    await expect(alert).toHaveAttribute('id', 'email-error');
  },
};

/** エラーが無いとき。何も出ない（場所も取らない）。 */
export const 何も出ない: Story = {
  args: { name: 'email' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).queryByRole('alert')).toBeNull();
  },
};
