import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
import NoteCreateForm from './NoteCreateForm';

/**
 * ワークスペース / スペースを作る入力欄。
 *
 * 見本もサイドバーと同じ幅（256px）の区画に入れて見る。全幅で並べると入力とボタンの
 * 詰まり具合が実物と変わり、「見本どおりか」の判断ができない。
 *
 * 受け取るのは名前だけ（URL に出る短い名前はサーバーが決める）。
 */
const meta = {
  title: 'note-sidebar/NoteCreateForm',
  component: NoteCreateForm,
  parameters: { layout: 'centered' },
  args: { what: 'スペース', onCreate: fn() },
  decorators: [
    (Story) => (
      // 実際の置き場所（サイドバー幅の、枠で囲った区画）。
      <div className="w-64 rounded-md border border-surface-3 bg-surface">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof NoteCreateForm>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 開いた直後。名前が空のあいだは作らせない（サーバーも空名を弾く）。 */
export const 名前が空: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('button', { name: 'スペースを作る' })).toBeDisabled();
  },
};

/** 名前を入れたところ。ここで初めてボタンが押せるようになる。 */
export const 名前を入れたところ: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.type(canvas.getByRole('textbox', { name: 'スペースの名前' }), '設計チーム');
    await expect(canvas.getByRole('button', { name: 'スペースを作る' })).toBeEnabled();
  },
};

/**
 * 送信中。**決着しない約束**を返させて、その見た目のまま止める
 * （解決する約束だと一瞬で通り過ぎ、絵として確かめられない）。
 */
export const 保存中: Story = {
  args: { onCreate: fn(() => new Promise<void>(() => {})) },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.type(canvas.getByRole('textbox', { name: 'スペースの名前' }), '設計チーム');
    await userEvent.click(canvas.getByRole('button', { name: 'スペースを作る' }));
    // 送信中は押せない（二度押しで同じスペースが 2 つできるのを防ぐ）。
    await waitFor(async () => {
      await expect(canvas.getByRole('button', { name: '作成中…' })).toBeDisabled();
    });
  },
};
