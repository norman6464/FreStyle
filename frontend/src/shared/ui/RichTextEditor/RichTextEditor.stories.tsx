import { useState } from 'react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, waitFor, within } from 'storybook/test';
import RichTextEditor from './RichTextEditor';
import { emptyRichDoc, type RichDocContent } from './emptyRichDoc';

/**
 * 本文エディタの見本。ノート画面（/p）での使われ方をそのまま再現する。
 *
 * ここの story は addon-vitest で**そのままテストとして実行される**
 * （play の assertion が落ちると CI が落ちる）。見た目の検証と回帰テストを兼ねる。
 */
const meta = {
  title: 'shared/RichTextEditor',
  component: RichTextEditor,
  parameters: { layout: 'fullscreen' },
} satisfies Meta<typeof RichTextEditor>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 見出し・箇条書き・番号付きを含む見本の本文（リストの印の色の確認も兼ねる）。 */
const sampleDoc = (): RichDocContent => ({
  type: 'doc',
  content: [
    {
      type: 'paragraph',
      content: [{ type: 'text', text: '10日間のインターンがやっと終わりました。' }],
    },
    {
      type: 'bulletList',
      content: [
        {
          type: 'listItem',
          content: [
            {
              type: 'paragraph',
              content: [{ type: 'text', text: 'TipTap のリッチテキストエディタを使ったこと' }],
            },
          ],
        },
        {
          type: 'listItem',
          content: [
            {
              type: 'paragraph',
              content: [{ type: 'text', text: 'oss-DB の勉強も間に挟んだよ' }],
            },
          ],
        },
      ],
    },
    {
      type: 'orderedList',
      content: [
        {
          type: 'listItem',
          content: [
            {
              type: 'paragraph',
              content: [{ type: 'text', text: '番号付きの項目（印は本文と同じ色）' }],
            },
          ],
        },
      ],
    },
  ],
});

/**
 * ノート画面の形の再現。ツールバーは**ヘッダー直下の sticky バー**に置かれ、
 * 題名より上に出る（Confluence と同じ並び: ツールバー → 題名 → 本文）。
 * エディタには置き場所（toolbarContainer）だけを渡す。
 */
function NotePageLayout() {
  const [toolbarHost, setToolbarHost] = useState<HTMLDivElement | null>(null);
  return (
    <div style={{ height: '100vh', overflowY: 'auto' }}>
      <div className="sticky top-0 z-10 border-b border-surface-3 bg-surface">
        <div ref={setToolbarHost} className="mx-auto w-full max-w-3xl px-6 py-1.5" data-testid="toolbar-host" />
      </div>
      <div className="mx-auto w-full max-w-3xl px-6 py-10">
        <h1 className="mb-4 text-3xl font-bold">サイボウズインターン終わったー</h1>
        <RichTextEditor value={sampleDoc()} editable toolbar toolbarContainer={toolbarHost} />
      </div>
    </div>
  );
}

export const 題名の上のツールバー: Story = {
  // render で全体を差し替えるので args は使わないが、型の上では必須（value を満たしておく）。
  args: { value: emptyRichDoc() },
  render: () => <NotePageLayout />,
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    // ツールバーが出るまで待つ（editor の初期化は非同期）。
    const toolbar = await canvas.findByRole('toolbar', { name: '書式メニュー' });
    const title = canvas.getByRole('heading', { level: 1 });
    const host = canvas.getByTestId('toolbar-host');
    // ツールバーは sticky バー（host）の中に入っている（ポータルが効いている）。
    await expect(host.contains(toolbar)).toBe(true);
    // DOM 上でツールバーが題名より前 = 画面で題名より上に出る。
    await expect(
      Boolean(toolbar.compareDocumentPosition(title) & Node.DOCUMENT_POSITION_FOLLOWING),
    ).toBe(true);
  },
};

/** 置き場所を渡さないときの既定: 本文の直上に出る（後方互換の形）。 */
export const 本文の直上のツールバー: Story = {
  args: { value: sampleDoc(), editable: true, toolbar: true },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await canvas.findByRole('toolbar', { name: '書式メニュー' });
  },
};

/** 読み取り専用ではツールバーを出さない（押せない操作を見せない）。 */
export const 読み取り専用: Story = {
  args: { value: sampleDoc(), editable: false, toolbar: true },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    // 本文は出る。
    await waitFor(async () => {
      await expect(canvas.getByText(/インターンがやっと終わりました/)).toBeInTheDocument();
    });
    await expect(canvas.queryByRole('toolbar', { name: '書式メニュー' })).not.toBeInTheDocument();
  },
};

/** 空の本文（プレースホルダの確認）。 */
export const 空の本文: Story = {
  args: { value: emptyRichDoc(), editable: true, toolbar: true },
};
