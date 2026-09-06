import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import Loading from './Loading';

/**
 * 読み込み中を示すぐるぐる。
 *
 * 読み上げソフトには `role="status"` と「読み込み中」で伝わる。画面に何も出さずに
 * 待たせると「固まった」と誤解されるので、待ち時間があるところには必ず置く。
 */
const meta = {
  title: 'shared/Loading',
  component: Loading,
  parameters: { layout: 'centered' },
} satisfies Meta<typeof Loading>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 既定の大きさ。 */
export const 既定: Story = {
  args: {},
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('status')).toHaveAccessibleName('読み込み中');
  },
};

/** 大きさ 3 段を並べて比べる。 */
export const 大きさ: Story = {
  render: () => (
    <div className="flex items-center gap-8">
      <Loading size="small" />
      <Loading size="medium" />
      <Loading size="large" />
    </div>
  ),
};

/** 待たせる理由を添える。長く待たせるときほど効く。 */
export const 文言つき: Story = {
  args: { size: 'large', message: '演習を読み込んでいます…' },
};

/**
 * 画面いっぱい。**他の操作をさせたくない**ときだけ使う。
 *
 * `fixed inset-0` で全面を覆うので、story も画面ぜんぶを使って見る。
 */
export const 画面いっぱい: Story = {
  args: { fullscreen: true, message: '準備しています…' },
  parameters: { layout: 'fullscreen' },
};
