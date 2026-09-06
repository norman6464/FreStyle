import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import LanguageBadge from './LanguageBadge';

/**
 * 言語・技術の名前を、識別しやすい色つきの小さな札で見せる。
 *
 * 表記は**先頭だけ大文字**に整える（`TYPESCRIPT` は圧が強い、というのが由来）。
 * 機械では直せないもの（`cpp` → `C++`）だけ、対応表で上書きしている。
 * 知らない言語が来ても色を無彩色に落として崩れない。
 */
const meta = {
  title: 'shared/LanguageBadge',
  component: LanguageBadge,
  parameters: { layout: 'centered' },
} satisfies Meta<typeof LanguageBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 1 つだけ。 */
export const 既定: Story = {
  args: { language: 'go' },
  play: async ({ canvasElement }) => {
    // 入力は小文字でも、表示は先頭だけ大文字になる。
    await expect(within(canvasElement).getByText('Go')).toBeVisible();
  },
};

/** 機械では直せない表記。対応表で上書きする。 */
export const 表記の上書き: Story = {
  args: { language: 'cpp' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('C++')).toBeVisible();
  },
};

/** 等幅で出す（詳細画面の見出し脇など）。 */
export const 等幅: Story = {
  args: { language: 'typescript', mono: true },
};

/** 知らない言語。色は無彩色に落ちるが、文字は読める。 */
export const 未知の言語: Story = {
  args: { language: 'brainfuck' },
};

/** よく使うものを並べて、色の見分けを確かめる。 */
export const 並べたところ: Story = {
  args: { language: 'go' },
  render: () => (
    <div className="flex flex-wrap gap-2">
      {['go', 'typescript', 'javascript', 'python', 'java', 'php', 'ruby', 'rust', 'cpp', 'docker', 'terraform', 'sql'].map(
        (language) => (
          <LanguageBadge key={language} language={language} />
        ),
      )}
    </div>
  ),
};
