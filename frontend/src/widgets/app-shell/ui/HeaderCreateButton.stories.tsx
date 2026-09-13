import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';
import { withApi, withRouter, withToast } from '../../../../.storybook/decorators';
import { setCurrentKbSpace } from '@/entities/kb';
import HeaderCreateButton from './HeaderCreateButton';

const BUTTON_CLASS =
  'inline-flex items-center gap-1.5 rounded-md bg-brand-600 px-3 py-1.5 text-sm font-semibold text-white';

/**
 * ヘッダーの「作成」（段3・PR-3）。今いるスペースが分かるときだけ出す
 * （createPage が spaceId を必須とするため）。
 */
const meta = {
  title: 'widgets/app-shell/HeaderCreateButton',
  component: HeaderCreateButton,
  args: { className: BUTTON_CLASS },
  decorators: [withRouter, withToast],
} satisfies Meta<typeof HeaderCreateButton>;

export default meta;
type Story = StoryObj<typeof meta>;

/** ナレッジ以外の画面（今いるスペースが分からない）では何も出さない。 */
export const ナレッジ以外の画面では出さない: Story = {
  play: async ({ canvasElement }) => {
    setCurrentKbSpace(null);
    await expect(canvasElement).toBeEmptyDOMElement();
  },
};

/** 今いるスペースが分かっているとき。押すとそのスペース直下にページを作って開く。 */
export const 押すとページを作る: Story = {
  decorators: [
    withApi({
      '/kb/workspaces/w-3f2a9c/spaces/s-1/pages': {
        id: 'p-new',
        spaceId: 's-1',
        title: '無題',
        createdByUserId: 1,
        createdAt: '2026-09-13T00:00:00Z',
        updatedAt: '2026-09-13T00:00:00Z',
      },
    }),
  ],
  play: async ({ canvasElement }) => {
    setCurrentKbSpace({ workspaceSlug: 'w-3f2a9c', spaceId: 's-1' });
    const canvas = within(canvasElement);
    const button = await canvas.findByRole('button', { name: /作成/ });
    await userEvent.click(button);
  },
};
