import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withRouter } from '../../../../.storybook/decorators';
import MenuPage from './MenuPage';

/**
 * ログインしたあと最初に出るホーム。
 *
 * 機能への入口を、区画（学習・ツールなど）ごとにカードで並べる。ここも通信をしないので、
 * 待ち時間も失敗も無い — 最初に見る画面が読み込み中から始まらないようにしてある。
 */
const meta = {
  title: 'pages/home/MenuPage',
  component: MenuPage,
  parameters: { layout: 'fullscreen' },
  decorators: [
    withRouter,
    (Story) => (
      <div className="bg-surface">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof MenuPage>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 既定。 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    await expect(
      within(canvasElement).getByRole('heading', { name: 'FreStyle へようこそ' }),
    ).toBeVisible();
  },
};

/** 狭い画面。カードが 1 列になる。 */
export const 狭い画面: Story = {
  globals: { viewport: { value: 'mobile1', isRotated: false } },
};
