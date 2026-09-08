import { useState } from 'react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';
import KbSuggestEditButton from './KbSuggestEditButton';

/**
 * commenter だけに見える「変更を提案する」トグル。押すたびに呼び出し側（KbPage）の
 * ドラフトモードが反転する。ここではボタン単体の見た目・押した回数だけを見本にする
 * （実際にドラフトモードへ入る様子は KbPage.stories.tsx 側で確かめる）。
 */
const meta = {
  title: 'pages/kb/KbSuggestEditButton',
  component: KbSuggestEditButton,
  parameters: { layout: 'padded' },
} satisfies Meta<typeof KbSuggestEditButton>;

export default meta;
type Story = StoryObj<typeof meta>;

function Controlled() {
  const [active, setActive] = useState(false);
  return <KbSuggestEditButton active={active} onToggle={() => setActive((prev) => !prev)} />;
}

/** 閉じている状態。 */
export const 閉じている: Story = {
  args: { active: false, onToggle: () => {} },
  play: async ({ canvasElement }) => {
    const button = within(canvasElement).getByRole('button', { name: '変更を提案する' });
    await expect(button).toBeVisible();
    await expect(button).toHaveAttribute('aria-expanded', 'false');
  },
};

/** ドラフトモード中（aria-expanded で示す）。 */
export const ドラフトモード中: Story = {
  args: { active: true, onToggle: () => {} },
  play: async ({ canvasElement }) => {
    await expect(
      within(canvasElement).getByRole('button', { name: '変更を提案する' }),
    ).toHaveAttribute('aria-expanded', 'true');
  },
};

/** 押すたびに開閉が反転する。 */
export const 押すと反転する: Story = {
  args: { active: false, onToggle: () => {} },
  render: () => <Controlled />,
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const button = canvas.getByRole('button', { name: '変更を提案する' });
    await expect(button).toHaveAttribute('aria-expanded', 'false');

    await userEvent.click(button);
    await expect(button).toHaveAttribute('aria-expanded', 'true');

    await userEvent.click(button);
    await expect(button).toHaveAttribute('aria-expanded', 'false');
  },
};
