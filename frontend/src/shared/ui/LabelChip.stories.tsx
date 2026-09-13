import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import LabelChip from './LabelChip';

/**
 * 名前 + 色だけのラベル chip。塗り方は色の明るさで決まる（labelPaint）。
 *
 * チケットのラベル（`pages/backlog/ui/TicketLabelChip`）とナレッジのページラベル
 * （`pages/kb/ui/KbPageMeta`）が共有する見た目の実体。
 */
const meta = {
  title: 'shared/LabelChip',
  component: LabelChip,
  args: { name: '不具合', color: '#1d4ed8' },
} satisfies Meta<typeof LabelChip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const 暗い色は地に敷いて白文字: Story = {
  play: async ({ canvasElement }) => {
    const chip = within(canvasElement).getByText('不具合');
    await expect(chip).toHaveStyle({ backgroundColor: 'rgb(29, 78, 216)', color: 'rgb(255, 255, 255)' });
  },
};

export const 明るい色は地に敷いて濃い文字: Story = {
  args: { name: '検索', color: '#dbeafe' },
  play: async ({ canvasElement }) => {
    const chip = within(canvasElement).getByText('検索');
    await expect(chip).toHaveStyle({ backgroundColor: 'rgb(219, 234, 254)', color: 'rgb(25, 25, 25)' });
  },
};

export const 中間の明るさは地に敷かず枠線に落とす: Story = {
  args: { name: '要調査', color: '#8b7355' },
  play: async ({ canvasElement }) => {
    const chip = within(canvasElement).getByText('要調査');
    await expect(chip).toHaveStyle({ borderColor: 'rgb(139, 115, 85)' });
    await expect(chip).not.toHaveStyle({ backgroundColor: 'rgb(139, 115, 85)' });
  },
};

export const 色が壊れていても消えない: Story = {
  args: { name: '壊れた色', color: 'not-a-color' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('壊れた色')).toBeInTheDocument();
  },
};
