import { useState } from 'react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, waitFor, within } from 'storybook/test';
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
      content: [{ type: 'text', text: '10 日間の研修がやっと終わりました。' }],
    },
    {
      type: 'bulletList',
      content: [
        {
          type: 'listItem',
          content: [
            {
              type: 'paragraph',
              content: [{ type: 'text', text: 'リッチテキストエディタを組み込んだこと' }],
            },
          ],
        },
        {
          type: 'listItem',
          content: [
            {
              type: 'paragraph',
              content: [{ type: 'text', text: 'データベースの勉強も間に挟んだ' }],
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
 * 題名より上に出る（並びは ツールバー → 題名 → 本文）。
 * エディタには置き場所（toolbarContainer）だけを渡す。
 */
function NotePageLayout() {
  const [toolbarHost, setToolbarHost] = useState<HTMLDivElement | null>(null);
  // 題名で Enter を押したら本文へ移る合図（ノート画面と同じ配線）。
  const [bodyFocusSignal, setBodyFocusSignal] = useState(0);
  return (
    <div style={{ height: '100vh', overflowY: 'auto' }}>
      <div className="sticky top-0 z-30 border-b border-surface-3 bg-surface">
        <div ref={setToolbarHost} className="mx-auto w-full max-w-3xl px-6 py-1.5" data-testid="toolbar-host" />
      </div>
      <div className="mx-auto w-full max-w-3xl px-6 py-10">
        <input
          aria-label="ページの題名"
          defaultValue="研修の振り返り"
          onKeyDown={(event) => {
            if (event.key !== 'Enter') return;
            event.preventDefault();
            setBodyFocusSignal((prev) => prev + 1);
          }}
          className="mb-4 w-full border-none bg-transparent p-0 text-3xl font-bold outline-none"
        />
        <RichTextEditor
          value={sampleDoc()}
          editable
          toolbar
          toolbarContainer={toolbarHost}
          focusSignal={bodyFocusSignal}
        />
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
    const title = canvas.getByRole('textbox', { name: 'ページの題名' });
    const host = canvas.getByTestId('toolbar-host');
    // ツールバーは sticky バー（host）の中に入っている（ポータルが効いている）。
    await expect(host.contains(toolbar)).toBe(true);
    // DOM 上でツールバーが題名より前 = 画面で題名より上に出る。
    await expect(
      Boolean(toolbar.compareDocumentPosition(title) & Node.DOCUMENT_POSITION_FOLLOWING),
    ).toBe(true);
  },
};

/**
 * 題名で Enter を押すと本文の先頭へ移る。
 *
 * **本物のブラウザで確かめる。** jsdom では、同じ document で何度もエディタを作った後の
 * `focus()` が実際には activeElement を動かさないことがあり、合図を無視する実装でも
 * 単体テストが通ってしまう（合図のガードを外して確認済み）。ここは story として
 * ブラウザで実行し、フォーカスが本当に移ることを見る。
 */
export const 題名でEnterすると本文へ移る: Story = {
  args: { value: emptyRichDoc() },
  render: () => <NotePageLayout />,
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const body = await canvas.findByRole('textbox', { name: '本文' });
    const title = canvas.getByRole('textbox', { name: 'ページの題名' });

    title.focus();
    await expect(document.activeElement).toBe(title);

    await userEvent.keyboard('{Enter}');

    await waitFor(async () => {
      await expect(body.contains(document.activeElement)).toBe(true);
    });
  },
};

/** ページを開いただけでは本文がフォーカスを奪わない（マウント時の合図では動かない）。 */
export const 開いただけでは本文にフォーカスしない: Story = {
  args: { value: sampleDoc(), editable: true, toolbar: true, focusSignal: 3 },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const body = await canvas.findByRole('textbox', { name: '本文' });
    await expect(body.contains(document.activeElement)).toBe(false);
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
      await expect(canvas.getByText(/研修がやっと終わりました/)).toBeInTheDocument();
    });
    await expect(canvas.queryByRole('toolbar', { name: '書式メニュー' })).not.toBeInTheDocument();
  },
};

/** 空の本文（プレースホルダの確認）。 */
export const 空の本文: Story = {
  args: { value: emptyRichDoc(), editable: true, toolbar: true },
};
