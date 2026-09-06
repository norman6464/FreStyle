import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import CodeBlock from './CodeBlock';

/**
 * AI の返事に出てくるコードの囲み。上に言語名とコピーボタンが付く。
 *
 * ふだんは Markdown の描画（MarkdownView）から自動で使われるので、直接置くことは少ない。
 * ここでは、Markdown が渡してくるのと同じ形
 * （`<code className="language-go">…</code>`）を手で渡して見ている。
 *
 * コピーはブラウザの設定や http 環境で断られることがある。そのときは**何も起きない**
 * （壊れた反応を見せるより、無反応のほうが安全という判断）。
 */
const meta = {
  title: 'shared/CodeBlock',
  component: CodeBlock,
  parameters: { layout: 'padded' },
  decorators: [
    (Story) => (
      <div className="max-w-2xl">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof CodeBlock>;

export default meta;
type Story = StoryObj<typeof meta>;

const GO_SAMPLE = `package main

import "fmt"

func main() {
	fmt.Println("こんにちは")
}`;

/** 言語が分かるとき。左上に言語名が出る。 */
export const 言語つき: Story = {
  args: { children: <code className="language-go">{GO_SAMPLE}</code> },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('go')).toBeVisible();
    await expect(canvas.getByRole('button', { name: 'コードをコピー' })).toBeVisible();
  },
};

/** 言語が分からないとき。「code」と出す。 */
export const 言語なし: Story = {
  args: { children: <code>{'echo "言語の指定がないとき"'}</code> },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('code')).toBeVisible();
  },
};

/** 横に長いコード。囲みの中だけが横に流れ、ページ全体は横に伸びない。 */
export const 横に長い: Story = {
  args: {
    children: (
      <code className="language-sql">
        {'SELECT p.id, p.title, s.name AS space_name, w.name AS workspace_name FROM pages p JOIN spaces s ON s.id = p.space_id JOIN workspaces w ON w.id = s.workspace_id ORDER BY p.updated_at DESC;'}
      </code>
    ),
  },
};

/** 縦に長いコード。 */
export const 縦に長い: Story = {
  args: {
    children: (
      <code className="language-typescript">
        {Array.from({ length: 20 }, (_, i) => `const line${i + 1} = ${i + 1};`).join('\n')}
      </code>
    ),
  },
};
