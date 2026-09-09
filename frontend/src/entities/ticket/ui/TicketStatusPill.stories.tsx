import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import TicketStatusPill from './TicketStatusPill';

const meta = {
  title: 'entities/ticket/TicketStatusPill',
  component: TicketStatusPill,
  parameters: { layout: 'padded' },
} satisfies Meta<typeof TicketStatusPill>;

export default meta;
type Story = StoryObj<typeof meta>;

export const 未着手: Story = {
  args: { name: 'To Do', color: '#5b6b7a', category: 'todo' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('To Do')).toBeInTheDocument();
  },
};

export const 進行中_選べる: Story = {
  args: { name: '開発', color: '#a0661a', category: 'in_progress', showChevron: true },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('開発')).toBeInTheDocument();
  },
};

export const 完了: Story = {
  args: { name: 'リリース', color: '#2f6b47', category: 'done' },
};
