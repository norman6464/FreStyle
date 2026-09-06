import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import Button from './Button';

/**
 * アプリ共通の押しボタン。
 *
 * **見た目の種類（variant）で「押すとどうなるか」を伝える。**
 * 迷ったら `primary` を 1 画面に 1 つだけ置く。並べると、どれが本命か分からなくなる。
 *
 * - `primary` … その画面でいちばんしてほしいこと（保存する・作る）
 * - `secondary` … それ以外の選択肢（戻る・あとで）
 * - `ghost` … 枠も地色も無い、控えめな操作（閉じる・切り替え）
 * - `danger` … 取り消せないこと（削除する）
 */
const meta = {
  title: 'shared/Button',
  component: Button,
  parameters: { layout: 'centered' },
  args: { children: 'ボタン', onClick: fn() },
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

/** いちばんしてほしいこと。1 画面に 1 つ。 */
export const 主役: Story = {
  args: { variant: 'primary', children: '保存する' },
};

/** 主役の隣に置く、もう一方の選択肢。 */
export const 控えめ: Story = {
  args: { variant: 'secondary', children: 'あとで' },
};

/** 枠も地色も無い形。ツールバーなど、並べても騒がしくならない。 */
export const 地味: Story = {
  args: { variant: 'ghost', children: '閉じる' },
};

/** 取り消せない操作。赤は「戻れない」の合図なので、それ以外に使わない。 */
export const 危険: Story = {
  args: { variant: 'danger', children: '削除する' },
};

/** 大きさは 3 段。並べて比べる。 */
export const 大きさ: Story = {
  render: () => (
    <div className="flex items-center gap-3">
      <Button size="sm">小 (sm)</Button>
      <Button size="md">中 (md)</Button>
      <Button size="lg">大 (lg)</Button>
    </div>
  ),
};

/** 4 種類を一度に見る。 */
export const 種類ぜんぶ: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      <Button variant="primary">主役</Button>
      <Button variant="secondary">控えめ</Button>
      <Button variant="ghost">地味</Button>
      <Button variant="danger">危険</Button>
    </div>
  ),
};

/**
 * 処理中。ぐるぐるが出て、**押せなくなる**。
 *
 * 押せたままにすると二重に送信されるので、`loading` は自動で `disabled` も兼ねる。
 * 読み上げソフトにも伝わるよう `aria-busy` が付く。
 */
export const 処理中: Story = {
  args: { loading: true, children: '保存中…' },
  play: async ({ canvasElement }) => {
    const button = within(canvasElement).getByRole('button');
    await expect(button).toBeDisabled();
    await expect(button).toHaveAttribute('aria-busy', 'true');
  },
};

/** 押せないとき。薄くなり、指の形も変わる。 */
export const 押せない: Story = {
  args: { disabled: true, children: '保存する' },
};

/** 横幅いっぱい。ログイン画面のように、縦に積むときに使う。 */
export const 幅いっぱい: Story = {
  args: { fullWidth: true, children: 'ログイン' },
  decorators: [
    (Story) => (
      <div className="w-72">
        <Story />
      </div>
    ),
  ],
};

/** 押すと onClick が呼ばれる。 */
export const 押したとき: Story = {
  args: { children: '押してみる' },
  play: async ({ args, canvasElement }) => {
    await userEvent.click(within(canvasElement).getByRole('button'));
    await expect(args.onClick).toHaveBeenCalledTimes(1);
  },
};
