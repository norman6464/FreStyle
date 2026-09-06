import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, waitFor, within } from 'storybook/test';
import Avatar from './Avatar';

/**
 * 人を表す丸い絵。
 *
 * 写真があれば写真、無ければ**名前の頭文字**を出す。写真の読み込みに失敗したときも
 * 頭文字に戻る — ブラウザ既定の「壊れた画像」マークが出ると、そこだけ崩れて見えるため。
 */
const meta = {
  title: 'shared/Avatar',
  component: Avatar,
  parameters: { layout: 'centered' },
  args: { name: '川野 拓馬' },
} satisfies Meta<typeof Avatar>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 写真が無いとき。名前の 1 文字目が出る。 */
export const 頭文字: Story = {
  args: {},
};

/** 写真があるとき。 */
export const 写真つき: Story = {
  args: {
    // 外部を見に行かないよう、絵は story の中に埋め込む（オフラインでも同じ絵になる）。
    src:
      'data:image/svg+xml;utf8,' +
      encodeURIComponent(
        '<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96">' +
          '<rect width="96" height="96" fill="#7c6f64"/>' +
          '<circle cx="48" cy="38" r="18" fill="#f2e5bc"/>' +
          '<circle cx="48" cy="86" r="30" fill="#f2e5bc"/>' +
          '</svg>',
      ),
  },
};

/** 大きさ 4 段。 */
export const 大きさ: Story = {
  render: () => (
    <div className="flex items-end gap-4">
      <Avatar name="Sato" size="sm" />
      <Avatar name="Suzuki" size="md" />
      <Avatar name="Takahashi" size="lg" />
      <Avatar name="Tanaka" size="xl" />
    </div>
  ),
};

/** 名前が空でも崩れない。「?」を出す。 */
export const 名前が空: Story = {
  args: { name: '' },
};

/** 写真が読めなかったとき。頭文字に戻る。 */
export const 写真が読めないとき: Story = {
  args: { src: '/存在しない画像.png' },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    // onError が走って頭文字に差し替わるまで待つ。
    await waitFor(async () => {
      await expect(canvas.getByText('川')).toBeVisible();
    });
  },
};
