import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import TicketKeyBadge from './TicketKeyBadge';

const meta = {
  title: 'entities/ticket/TicketKeyBadge',
  component: TicketKeyBadge,
  parameters: { layout: 'padded' },
} satisfies Meta<typeof TicketKeyBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const 既定: Story = {
  args: { spaceKey: 'FRESTYLE', number: 457 },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('FRESTYLE-457')).toBeInTheDocument();
  },
};

export const ハイフンを含むスペースキー: Story = {
  args: { spaceKey: 'my-app', number: 12 },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('MY-APP-12')).toBeInTheDocument();
  },
};
