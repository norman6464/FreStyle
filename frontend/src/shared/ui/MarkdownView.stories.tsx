import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import MarkdownView from './MarkdownView';

/**
 * Markdown の文字列を、アプリのトーンに揃えて描く。AI の返事の表示に使う。
 *
 * 見出し・表・リンク・コードの見た目をここで統一している。リンクは必ず別のタブで開き、
 * 開いた先からこのページを操作できないようにしてある（`rel="noopener"`）。
 *
 * `isStreaming` は、AI が書いている**途中**だけ true にする。文が現れるたびに
 * ふわっと出す演出が入る。書き終わったら false に戻して、素の Markdown に戻す。
 */
const meta = {
  title: 'shared/MarkdownView',
  component: MarkdownView,
  parameters: { layout: 'padded' },
  decorators: [
    (Story) => (
      // アプリ本体と同じく prose（文章向けの体裁）の中で見る。
      <div className="prose prose-sm max-w-2xl text-[var(--color-text-primary)]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof MarkdownView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 段落と強調だけの、いちばん短い形。 */
export const 文章: Story = {
  args: {
    content: 'Go の **スライス** は、配列の一部を指す軽い入れものです。中身は共有されます。',
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('スライス')).toBeVisible();
  },
};

/** 見出し・箇条書き・番号つき。 */
export const 見出しと箇条書き: Story = {
  args: {
    content: [
      '## はじめに',
      '',
      'この演習では次のことを学びます。',
      '',
      '- スライスの作り方',
      '- 追加のしかた',
      '- 落とし穴',
      '',
      '### 進め方',
      '',
      '1. コードを読む',
      '2. 手を動かす',
      '3. 結果を確かめる',
    ].join('\n'),
  },
};

/** コードの囲み。言語名とコピーボタンが付く。 */
export const コード: Story = {
  args: {
    content: ['次のように書きます。', '', '```go', 's := []int{1, 2, 3}', 's = append(s, 4)', '```'].join(
      '\n',
    ),
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('button', { name: 'コードをコピー' })).toBeVisible();
  },
};

/** 文の中のコード（囲みではない短いもの）。 */
export const 文中のコード: Story = {
  args: { content: '`append` は新しいスライスを返します。元のスライスは変わりません。' },
};

/** 表。横に長いときは表の中だけが流れる。 */
export const 表: Story = {
  args: {
    content: [
      '| 関数 | すること | 戻り値 |',
      '| --- | --- | --- |',
      '| `len` | 長さを返す | int |',
      '| `cap` | 容量を返す | int |',
      '| `append` | 末尾に足す | 新しいスライス |',
    ].join('\n'),
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('table')).toBeVisible();
  },
};

/** リンク。別のタブで開く。 */
export const リンク: Story = {
  args: { content: '詳しくは [Go の公式ドキュメント](https://go.dev/doc/) を読んでください。' },
  play: async ({ canvasElement }) => {
    const link = within(canvasElement).getByRole('link', { name: 'Go の公式ドキュメント' });
    await expect(link).toHaveAttribute('target', '_blank');
    // 開いた先からこのページを触られないようにする。
    await expect(link).toHaveAttribute('rel', 'noopener noreferrer');
  },
};

/** AI が書いている途中。文が現れるたびにふわっと出る。 */
export const 書いている途中: Story = {
  args: {
    isStreaming: true,
    content: 'スライスは配列の一部を指します。中身は共有されるので、片方を変えると',
  },
};
