import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import KbBacklogSection from './KbBacklogSection';
import { withRouter } from '../../../../.storybook/decorators';
import type { KbSpace } from '@/entities/kb';

const spaces: KbSpace[] = [
  { id: 's-1', key: 'frestyle', name: 'frestyle', visibility: 'workspace', createdAt: '2026-01-01T00:00:00Z' },
  { id: 's-2', key: 'devsync', name: 'devsync', visibility: 'workspace', createdAt: '2026-01-01T00:00:00Z' },
];

const meta = {
  title: 'widgets/kb-sidebar/KbBacklogSection',
  component: KbBacklogSection,
  parameters: { layout: 'padded' },
  args: { workspaceSlug: 'acme', spaces },
  decorators: [withRouter],
} satisfies Meta<typeof KbBacklogSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const バックログ節あり: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('バックログ')).toBeInTheDocument();
    await expect(canvas.getByText('frestyle')).toBeInTheDocument();
    await expect(canvas.getByText('devsync')).toBeInTheDocument();
  },
};

export const スペース0件: Story = {
  args: { spaces: [] },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).queryByText('バックログ')).toBeNull();
  },
};
