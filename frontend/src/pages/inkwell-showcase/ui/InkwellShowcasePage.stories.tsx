import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withRouter } from '../../../../.storybook/decorators';
import InkwellShowcasePage from './InkwellShowcasePage';

/**
 * inkwell（触った感じを確かめるための部品の一群）を全部並べて見る画面。
 *
 * Storybook で 1 つずつ見るのとは目的が違う。こちらは**同じ面の上に並べたときに
 * 揃って見えるか**を確かめるためのもの（影の段・余白・書体が部品どうしで食い違わないか）。
 */
const meta = {
  title: 'pages/inkwell-showcase/InkwellShowcasePage',
  component: InkwellShowcasePage,
  parameters: { layout: 'fullscreen' },
  decorators: [withRouter],
} satisfies Meta<typeof InkwellShowcasePage>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 既定。 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { level: 1 })).toBeVisible();
  },
};
