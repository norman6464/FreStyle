import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';
import { withApi, withRouter, withToast } from '../../../../.storybook/decorators';
import { setCurrentKbSpace } from '@/entities/kb';
import HeaderSpacesNav from './HeaderSpacesNav';

const NAV_CLASS = 'px-3 py-1.5 rounded-md text-sm font-medium text-[var(--color-text-tertiary)]';

/**
 * ヘッダーの「スペース ▾」（段3・PR-3）。
 *
 * 今いるスペース（entities/kb の共有値。KbPage 等が確定させる）が分かるときだけ
 * ドロップダウンになる。分からないときは素のリンク（/kb/spaces。入口解決）にする。
 */
const meta = {
  title: 'widgets/app-shell/HeaderSpacesNav',
  component: HeaderSpacesNav,
  args: { className: NAV_CLASS },
  decorators: [withRouter, withToast],
} satisfies Meta<typeof HeaderSpacesNav>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 今いるスペースが分からないとき（ナレッジ以外の画面）。 */
export const 今いるスペースが分からないとき: Story = {
  play: async ({ canvasElement }) => {
    setCurrentKbSpace(null);
    const canvas = within(canvasElement);
    const link = await canvas.findByRole('link', { name: 'スペース' });
    await expect(link).toHaveAttribute('href', '/kb/spaces');
  },
};

/** 今いるスペースが分かっていて、開くと一覧が出る。 */
export const 開くと一覧が出る: Story = {
  decorators: [
    withApi({
      '/kb/workspaces/w-3f2a9c/me/spaces': [
        { id: 's-1', name: '開発チーム', role: 'editor' },
        { id: 's-2', name: '営業定例', role: 'viewer' },
      ],
    }),
  ],
  play: async ({ canvasElement }) => {
    setCurrentKbSpace({ workspaceSlug: 'w-3f2a9c', spaceId: 's-1' });
    const canvas = within(canvasElement);
    const button = await canvas.findByRole('button', { name: /スペース/ });
    await userEvent.click(button);
    await expect(await canvas.findByRole('link', { name: '営業定例' })).toBeVisible();
    // 今いるスペースは強調表示。
    const current = canvas.getByRole('link', { name: '開発チーム' });
    await expect(current.className).toContain('font-semibold');
  },
};
