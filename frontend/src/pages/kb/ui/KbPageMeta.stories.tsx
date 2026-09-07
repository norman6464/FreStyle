import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import KbPageMeta from './KbPageMeta';

/**
 * 題名の下に出す「最終編集」の一行。
 *
 * lastEditedBy / lastEditedAt が無ければ何も出さない（旧応答・未保存のページの
 * どちらも該当し得るので、無いことを匂わせる空欄は置かない）。
 */
const meta = {
  title: 'pages/kb/KbPageMeta',
  component: KbPageMeta,
  parameters: { layout: 'padded' },
  args: {
    lastEditedBy: { userId: 1, name: '田中 太郎' },
    // 'Z' を付けない — 実行環境のタイムゾーンによらず、書いたとおりの時刻として読める。
    lastEditedAt: '2026-09-06T13:05:00',
  },
} satisfies Meta<typeof KbPageMeta>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 名前が引けたとき。 */
export const 名前あり: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/最終編集 田中 太郎 · 9\/6/)).toBeVisible();
  },
};

/** 名前が引けなかったとき。行は消さず「不明なユーザー」で埋める。 */
export const 名前が引けない: Story = {
  args: { lastEditedBy: { userId: 1, name: '' } },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/最終編集 不明なユーザー/)).toBeVisible();
  },
};

/** まだ一度も保存されていない（旧応答も同じ形）。何も出さない。 */
export const 最終編集が無い: Story = {
  args: { lastEditedBy: null, lastEditedAt: null },
  play: async ({ canvasElement }) => {
    await expect(canvasElement).toBeEmptyDOMElement();
  },
};
