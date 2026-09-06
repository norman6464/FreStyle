import { useState } from 'react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import CodeEditor from './CodeEditor';

/**
 * 演習でコードを書くところ（Monaco エディタ）。
 *
 * **重い部品**なので、アプリでは `@/shared/ui` の入口から出していない。使う画面が
 * 必要になった時点で読み込む（そうしないと、全ページがエディタ一式を抱え込む）。
 *
 * 既定では中でスクロールせず、行数に合わせて**縦に伸びる**。ページ側でスクロールする
 * ほうが、書いている途中でスクロール枠が二重になる混乱を避けられるため。
 */
const meta = {
  title: 'shared/CodeEditor',
  component: CodeEditor,
  parameters: {
    layout: 'padded',
    // Monaco は自前で入力欄や装飾用の DOM を大量に作る。その中身は
    // このアプリで直せないので、a11y の自動検査からは外す（囲み側は検査する）。
    a11y: { test: 'off' },
  },
  decorators: [
    (Story) => (
      <div className="max-w-3xl">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof CodeEditor>;

export default meta;
type Story = StoryObj<typeof meta>;

const GO_SAMPLE = `package main

import "fmt"

func main() {
	fmt.Println("こんにちは")
}`;

/** 打てる形。行を足すと下に伸びる。 */
export const 書ける: Story = {
  args: { value: GO_SAMPLE, language: 'go', onChange: () => {} },
  render: (args) => {
    function Interactive() {
      const [value, setValue] = useState(args.value);
      return <CodeEditor {...args} value={value} onChange={setValue} />;
    }
    return <Interactive />;
  },
};

/** 読むだけ（模範解答の表示など）。 */
export const 読むだけ: Story = {
  args: { value: GO_SAMPLE, language: 'go', readOnly: true, onChange: () => {} },
};

/** 実行して失敗したとき。該当の行に印と、乗せると理由が出る。 */
export const エラーの行に印: Story = {
  args: {
    value: GO_SAMPLE,
    language: 'go',
    onChange: () => {},
    errorMarkers: [{ line: 6, message: 'undefined: fmt.Printline' }],
  },
};

/** 高さを固定したいとき。中でスクロールする。 */
export const 高さ固定: Story = {
  args: {
    value: Array.from({ length: 40 }, (_, i) => `line ${i + 1}`).join('\n'),
    language: 'plaintext',
    autoGrow: false,
    height: '240px',
    onChange: () => {},
  },
};
