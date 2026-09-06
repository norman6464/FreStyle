import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import ExecutionResultTable from './ExecutionResultTable';

/**
 * 「提出する」の前に一度だけ動かしてみたときの結果。
 *
 * 緑は**期待する出力と一致したときだけ**に取ってある。エラーなく終わっただけで緑にすると、
 * 「動いた＝正解」と色の印象で思い込ませてしまうため。エラーは出ていないが出力が違うときは
 * 琥珀色にして、**何行目がどう違うか**まで出す（自力で目視で探させない）。
 *
 * ここでの一致はこの画面の 1 件との下見であって、正誤はサーバー側の採点で決まる。
 */
const meta = {
  title: 'entities/exercise/ExecutionResultTable',
  component: ExecutionResultTable,
  parameters: { layout: 'padded' },
  decorators: [
    (Story) => (
      <div className="max-w-2xl">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof ExecutionResultTable>;

export default meta;
type Story = StoryObj<typeof meta>;

/** エラーなく動き、期待どおりの出力が出たとき。緑。 */
export const 一致した: Story = {
  args: {
    result: { stdout: '6\n', stderr: '', exitCode: 0 },
    expected: '6',
    submitError: null,
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('実行成功・期待する出力と一致')).toBeVisible();
  },
};

/** エラーは無いが出力が違うとき。琥珀色 + どこが違うかを出す。 */
export const 出力が違う: Story = {
  args: {
    result: { stdout: '5\n', stderr: '', exitCode: 0 },
    expected: '6',
    submitError: null,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('実行成功（エラーなし）・期待する出力と不一致')).toBeVisible();
    // 「1 行目が違う」まで出す。探させない。
    await expect(canvas.getByText(/1 行目が異なります/)).toBeVisible();
  },
};

/** 行数そのものが足りないとき。「この行がありません」と出す。 */
export const 行が足りない: Story = {
  args: {
    result: { stdout: '1\n2\n', stderr: '', exitCode: 0 },
    expected: '1\n2\n3',
    submitError: null,
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/3 行目が異なります/)).toBeVisible();
  },
};

/** 途中で落ちたとき。赤 + エラー本文を専用の行に出す。 */
export const 実行エラー: Story = {
  args: {
    result: {
      stdout: '',
      stderr: 'panic: runtime error: integer divide by zero',
      exitCode: 2,
    },
    expected: '6',
    submitError: null,
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/実行エラー（exit 2）/)).toBeVisible();
  },
};

/** Go のコンパイルで落ちたとき。ラベルが「コンパイル時エラーメッセージ」に変わる。 */
export const コンパイルエラー: Story = {
  args: {
    result: {
      stdout: '',
      stderr: './main.go:6:2: undefined: fmt.Printline',
      exitCode: 1,
    },
    expected: '6',
    submitError: null,
    language: 'go',
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('コンパイル時エラーメッセージ')).toBeVisible();
  },
};

/** 動いたが、何も出力していないとき。次にやることを添える。 */
export const 出力がない: Story = {
  args: {
    result: { stdout: '', stderr: '', exitCode: 0 },
    expected: '6',
    submitError: null,
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/まだ出力がありません/)).toBeVisible();
  },
};

/** 期待する出力が決まっていない演習。比べようがないので中立に出す。 */
export const 比べる相手がない: Story = {
  args: {
    result: { stdout: 'なにか\n', stderr: '', exitCode: 0 },
    expected: '',
    submitError: null,
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('実行成功（エラーなし）')).toBeVisible();
  },
};

/** そもそも送れなかったとき（通信の失敗など）。表ごと出さず、理由だけを出す。 */
export const 送れなかった: Story = {
  args: {
    result: null,
    expected: '6',
    submitError: '実行に失敗しました。しばらく待ってからもう一度お試しください。',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText(/実行に失敗しました/)).toBeVisible();
    await expect(canvas.queryByRole('table')).toBeNull();
  },
};

/** まだ一度も動かしていないとき。何も描かない。 */
export const 何も出ない: Story = {
  args: { result: null, expected: '6', submitError: null },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).queryByRole('table')).toBeNull();
  },
};
