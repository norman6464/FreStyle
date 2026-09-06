import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withRouter } from '../../../../.storybook/decorators';
import BackLink from './BackLink';

/**
 * 演習の詳細から一覧へ戻るリンク。
 *
 * 行き先も文言も固定で、受け取るものは無い。詳細画面が 2 種類（ふつうの演習と Q&A 形式）
 * あり、どちらの見出しにも同じものを置くために切り出してある。
 */
const meta = {
  title: 'entities/exercise/BackLink',
  component: BackLink,
  parameters: { layout: 'centered' },
  decorators: [withRouter],
} satisfies Meta<typeof BackLink>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 既定（唯一の形）。 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    await expect(
      within(canvasElement).getByRole('link', { name: '問題一覧に戻る' }),
    ).toHaveAttribute('href', '/code-editor');
  },
};
