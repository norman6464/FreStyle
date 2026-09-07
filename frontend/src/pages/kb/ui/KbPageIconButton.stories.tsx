import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import KbPageIconButton from './KbPageIconButton';

/**
 * ページ見出しの左に置くアイコンの表示・変更口。4 状態を持つ
 * （読むだけ / 書ける × 未設定 / 設定済み）。
 *
 * 「アイコンを追加」は md 以上ではホバー・フォーカスでだけ現れる作りだが、
 * story は素の DOM として描くのでこの見本では常に見えている
 * （`group-hover` は実際に親をホバーしないと効かない CSS の話で、
 * DOM に居ることそのものは常に確かめられる）。
 */
const meta = {
  title: 'pages/kb/KbPageIconButton',
  component: KbPageIconButton,
  parameters: { layout: 'centered' },
  args: {
    icon: null,
    canEdit: true,
    onChange: fn(async () => {}),
  },
  decorators: [
    // group-hover の対象になる祖先を再現しておく（実装上の前提を story にも持ち込む）。
    (Story) => (
      <div className="group p-8">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof KbPageIconButton>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 読むだけ + 未設定。何も出さない。 */
export const 読むだけ_未設定: Story = {
  args: { canEdit: false, icon: null },
  play: async ({ canvasElement }) => {
    await expect(canvasElement.querySelector('button')).toBeNull();
    await expect(canvasElement.querySelector('[role="img"]')).toBeNull();
  },
};

/** 読むだけ + 設定済み。絵文字を img で出すだけで、押せない。 */
export const 読むだけ_設定済み: Story = {
  args: { canEdit: false, icon: { type: 'emoji', value: '📘' } },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const img = canvas.getByRole('img', { name: 'ページのアイコン' });
    await expect(img).toHaveTextContent('📘');
    await expect(canvasElement.querySelector('button')).toBeNull();
  },
};

/** 書ける + 未設定。「アイコンを追加」。押すとピッカーが開く。 */
export const 書ける_未設定: Story = {
  args: { canEdit: true, icon: null },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const button = canvas.getByRole('button', { name: 'アイコンを追加' });
    await expect(button).toHaveAttribute('aria-expanded', 'false');

    await userEvent.click(button);
    await expect(button).toHaveAttribute('aria-expanded', 'true');
    await expect(canvas.getByRole('dialog', { name: 'ページのアイコンを選ぶ' })).toBeInTheDocument();
    // 未設定なので「外す」は無い。
    await expect(canvas.queryByRole('button', { name: 'アイコンを外す' })).toBeNull();
  },
};

/**
 * 書ける + 未設定。開いた直後に Escape を押すとピッカーが閉じる
 * （document への mousedown リスナーだけでは、直後の Escape で閉じない回帰の固定）。
 */
export const 書ける_未設定_Escapeで閉じる: Story = {
  args: { canEdit: true, icon: null },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const button = canvas.getByRole('button', { name: 'アイコンを追加' });

    await userEvent.click(button);
    await expect(button).toHaveAttribute('aria-expanded', 'true');

    await userEvent.keyboard('{Escape}');
    await expect(button).toHaveAttribute('aria-expanded', 'false');
    await expect(canvas.queryByRole('dialog')).toBeNull();
  },
};

/** 書ける + 設定済み。絵文字そのものが aria-expanded なボタンになる。 */
export const 書ける_設定済み: Story = {
  args: { canEdit: true, icon: { type: 'emoji', value: '📘' } },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const button = canvas.getByRole('button', { name: 'ページのアイコンを変更' });
    await expect(button).toHaveTextContent('📘');
    await expect(button).toHaveAttribute('aria-expanded', 'false');

    await userEvent.click(button);
    await expect(button).toHaveAttribute('aria-expanded', 'true');
    // 設定済みなので「外す」が出る。
    await expect(canvas.getByRole('button', { name: 'アイコンを外す' })).toBeInTheDocument();
  },
};
