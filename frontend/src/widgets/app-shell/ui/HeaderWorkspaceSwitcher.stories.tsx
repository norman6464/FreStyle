import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';
import { withApi, withRouter, withToast } from '../../../../.storybook/decorators';
import HeaderWorkspaceSwitcher from './HeaderWorkspaceSwitcher';

/**
 * ヘッダーからワークスペースを切り替える入口。
 *
 * 「いまどれを開いているか」はナレッジの画面側だけが持っていて、ヘッダーとは共有していない。
 * そのためここでは常に**どれも選ばれていない**見た目で、選んだ先だけを画面へ渡す。
 *
 * 所属が 1 つも無いときは**まるごと出さない**。切り替える先が無いのに切替の口を置いても、
 * 押した人は空の一覧を見るだけになる。
 */
const meta = {
  title: 'widgets/app-shell/HeaderWorkspaceSwitcher',
  component: HeaderWorkspaceSwitcher,
  parameters: { layout: 'padded' },
  decorators: [
    withRouter,
    withToast,
    (Story) => (
      // 実物はヘッダーの中。開いた一覧が入る高さを確保する。
      <div className="h-72 bg-surface-1 p-3">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof HeaderWorkspaceSwitcher>;

export default meta;
type Story = StoryObj<typeof meta>;

const workspaces = [
  { slug: 'w-3f2a9c', name: '開発チーム', createdAt: '2026-01-01T00:00:00Z', canManage: true },
  { slug: 'w-88ab21', name: '営業部', createdAt: '2026-02-01T00:00:00Z', canManage: false },
];

/** 所属があるとき。 */
export const 所属あり: Story = {
  decorators: [withApi({ '/kb/workspaces': workspaces })],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    // どれも選ばれていない状態で出る（開いているものはヘッダーからは分からない）。
    await expect(await canvas.findByText('ワークスペースを選択')).toBeVisible();
  },
};

/** 開いたところ。 */
export const 開いたところ: Story = {
  decorators: [withApi({ '/kb/workspaces': workspaces })],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(await canvas.findByRole('button', { name: /ワークスペースを選択/ }));
    await expect(canvas.getByRole('button', { name: '開発チーム' })).toBeVisible();
  },
};

/** どこにも所属していないとき。何も出さない。 */
export const 所属なし: Story = {
  decorators: [withApi({ '/kb/workspaces': [] })],
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).queryByRole('button')).toBeNull();
  },
};
