import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, waitFor, within } from 'storybook/test';
import KbBacklogPage from './KbBacklogPage';
import { routerWithParam, withApi, withToast, type ApiStubs } from '../../../../.storybook/decorators';

const workspaces = [{ slug: 'acme', name: '開発チーム', createdAt: '2026-01-01T00:00:00Z', canManage: true }];
const spaces = [{ id: 's-1', key: 'frestyle', name: 'frestyle', visibility: 'workspace', createdAt: '2026-01-01T00:00:00Z' }];

const status = (over: Record<string, unknown>) => ({
  id: 'st-1',
  workspaceId: 'w-1',
  spaceId: 's-1',
  name: 'To Do',
  category: 'todo',
  color: '#5b6b7a',
  position: 'a0',
  isInitial: true,
  createdAt: '2026-09-08T00:00:00Z',
  updatedAt: '2026-09-08T00:00:00Z',
  activeTicketCount: 1,
  ...over,
});

const type = (over: Record<string, unknown>) => ({
  id: 'ty-1',
  workspaceId: 'w-1',
  spaceId: 's-1',
  name: '開発タスク',
  hierarchyLevel: 0,
  color: '#2563eb',
  position: 'a0',
  isDefault: true,
  createdAt: '2026-09-08T00:00:00Z',
  updatedAt: '2026-09-08T00:00:00Z',
  activeTicketCount: 1,
  ...over,
});

const ticket = (over: Record<string, unknown>) => ({
  id: 't-1',
  workspaceId: 'w-1',
  spaceId: 's-1',
  number: 457,
  typeId: 'ty-1',
  statusId: 'st-1',
  title: '段1: チケットの骨格（9表）',
  doc: { type: 'doc', content: [] },
  priority: 1,
  position: 'a0',
  createdByUserId: 1,
  createdAt: '2026-09-08T00:00:00Z',
  updatedAt: '2026-09-09T00:00:00Z',
  ...over,
});

function baseApi(over: ApiStubs = {}): ApiStubs {
  return {
    '/kb/workspaces/acme/spaces/s-1/ticket-statuses': { statuses: [status({})] },
    '/kb/workspaces/acme/spaces/s-1/ticket-types': { types: [type({})] },
    '/kb/workspaces/acme/spaces/s-1/tickets': { tickets: [ticket({})] },
    '/kb/workspaces/acme/spaces/s-1/pages': { pages: [], hasHiddenChildren: false },
    '/kb/workspaces/acme/spaces': spaces,
    '/kb/workspaces': workspaces,
    ...over,
  };
}

const meta = {
  title: 'pages/backlog/KbBacklogPage',
  component: KbBacklogPage,
  parameters: { layout: 'fullscreen' },
  decorators: [withToast, routerWithParam('/kb/backlog/:spaceId', '/kb/backlog/s-1')],
} satisfies Meta<typeof KbBacklogPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ふつう: Story = {
  decorators: [withApi(baseApi())],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('FRESTYLE-457')).toBeInTheDocument();
    });
  },
};

export const 未有効化: Story = {
  decorators: [
    withApi(
      baseApi({
        '/kb/workspaces/acme/spaces/s-1/ticket-statuses': { statuses: [] },
        '/kb/workspaces/acme/spaces/s-1/ticket-types': { types: [] },
      }),
    ),
  ],
  play: async ({ canvasElement }) => {
    await waitFor(async () => {
      await expect(
        within(canvasElement).getByText('このスペースではチケットを使っていません'),
      ).toBeInTheDocument();
    });
  },
};

export const 状態タブへ切り替える: Story = {
  decorators: [withApi(baseApi())],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('FRESTYLE-457')).toBeInTheDocument();
    });
    await userEvent.click(canvas.getByRole('tab', { name: '状態' }));
    await waitFor(async () => {
      await expect(canvas.getByRole('tab', { name: '状態' })).toHaveAttribute('aria-selected', 'true');
    });
  },
};
