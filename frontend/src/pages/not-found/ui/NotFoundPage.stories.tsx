import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withRouter } from '../../../../.storybook/decorators';
import NotFoundPage from './NotFoundPage';

/**
 * 存在しない URL に来たときの受け皿。
 *
 * 以前はこの受け皿が無く、打ち間違いや古いリンクで**真っ白な画面**になり、戻る手段も
 * 無いまま離脱していた。
 *
 * 案内の出し分けは、ログイン済みかどうかを示す目印（Cookie）だけで決める。この画面は
 * 行き先を案内するだけで権限を判定しないので、それで足りる。サーバーに問い合わせると
 * 表示が遅れ、待たずに手元の状態を読むと未確定の既定値（未ログイン扱い）で描いてしまう。
 *
 * SPA なので HTTP の 404 は返せない。せめて検索エンジンには「登録しないでほしい」と伝える。
 */
const meta = {
  title: 'pages/not-found/NotFoundPage',
  component: NotFoundPage,
  parameters: { layout: 'fullscreen' },
  decorators: [withRouter],
} satisfies Meta<typeof NotFoundPage>;

export default meta;
type Story = StoryObj<typeof meta>;

/** ログインしていない人が来たとき。トップとログインの両方を案内する。 */
export const 未ログイン: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('heading', { name: 'ページが見つかりません' })).toBeVisible();
    await expect(canvas.getByRole('link', { name: 'トップへ戻る' })).toBeVisible();
    await expect(canvas.getByRole('link', { name: 'ログイン' })).toBeVisible();
  },
};

/**
 * ログイン済みの人が来たとき。ホームへ戻る 1 つだけにする。
 *
 * 目印の Cookie でログイン済みかを見るので、story でもその Cookie を置いてから描く。
 */
export const ログイン済み: Story = {
  decorators: [
    (Story) => {
      document.cookie = 'fs_signed_in=1; path=/';
      return <Story />;
    },
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('link', { name: 'ホームへ戻る' })).toBeVisible();
    await expect(canvas.queryByRole('link', { name: 'ログイン' })).toBeNull();
    // 後の story に持ち越さないよう消しておく。
    document.cookie = 'fs_signed_in=; path=/; max-age=0';
  },
};
