import type { Meta, StoryObj } from '@storybook/react-vite';
import FaviconIcon from './FaviconIcon';

/**
 * アプリの印（favicon）を、他のアイコンと同じ形で置けるようにしたもの。
 *
 * `EmptyState` などは「`className` を受け取る部品」としてアイコンを求めるが、favicon は
 * 画像ファイルなのでそのままでは渡せない。その差を埋めるだけの薄い包み。
 *
 * 絵は飾りなので読み上げソフトからは隠してある。
 */
const meta = {
  title: 'shared/icons/FaviconIcon',
  component: FaviconIcon,
  parameters: { layout: 'centered' },
} satisfies Meta<typeof FaviconIcon>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 既定。 */
export const 既定: Story = {
  args: { className: 'h-10 w-10' },
};

/** 大きさは className で決める。 */
export const 大きさ: Story = {
  render: () => (
    <div className="flex items-end gap-4">
      <FaviconIcon className="h-4 w-4" />
      <FaviconIcon className="h-8 w-8" />
      <FaviconIcon className="h-16 w-16" />
    </div>
  ),
};
