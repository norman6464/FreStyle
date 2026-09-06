import type { Meta, StoryObj } from '@storybook/react-vite';
import { AcademicCapIcon } from '@heroicons/react/24/outline';
import { expect, within } from 'storybook/test';
import Button from './Button';
import PageIntro from './PageIntro';

/**
 * 画面のいちばん上に置く見出し。
 *
 * どのページでも同じ形（題名・説明・操作）で並べることで、「この画面で何ができるか」を
 * 毎回同じ場所から読めるようにする。
 *
 * `headingLevel` は 1 ページに h1 が 1 つだけになるように選ぶ。画面の主役なら 1、
 * 画面の中の区画なら 2。
 */
const meta = {
  title: 'shared/PageIntro',
  component: PageIntro,
  parameters: { layout: 'padded' },
  decorators: [
    (Story) => (
      <div className="max-w-3xl">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof PageIntro>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 題名だけ。 */
export const 題名だけ: Story = {
  args: { title: '演習一覧' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { level: 1 })).toHaveTextContent(
      '演習一覧',
    );
  },
};

/** 説明を添える。この画面で何ができるかを 1〜2 行で。 */
export const 説明つき: Story = {
  args: {
    title: '演習一覧',
    description: '言語を選んで、手を動かしながら学びます。詰まったら解説を読めます。',
  },
};

/** アイコンつき。 */
export const アイコンつき: Story = {
  args: {
    title: '演習一覧',
    description: '言語を選んで、手を動かしながら学びます。',
    icon: <AcademicCapIcon className="h-6 w-6" aria-hidden="true" />,
  },
};

/** 右側に操作を置く。狭い画面では下に回り込む。 */
export const 操作つき: Story = {
  args: {
    title: 'ナレッジ',
    description: 'チームで共有する文書です。',
    actions: (
      <>
        <Button variant="secondary" size="sm">
          並べ替え
        </Button>
        <Button size="sm">ページを作る</Button>
      </>
    ),
  },
};

/** 画面の中の区画として使うとき（h2）。 */
export const 区画の見出し: Story = {
  args: { title: '最近見たページ', headingLevel: 2 },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { level: 2 })).toBeVisible();
  },
};
