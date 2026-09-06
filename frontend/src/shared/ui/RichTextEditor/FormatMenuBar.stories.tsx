import { EditorContent, useEditor } from '@tiptap/react';
import type { Editor } from '@tiptap/react';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, waitFor, within } from 'storybook/test';
import FormatMenuBar from './FormatMenuBar';
import { createEditorExtensions } from './editorExtensions';
import './richTextEditor.css';

/**
 * 太字・斜体・見出し・リストなどを並べたボタンの列。
 *
 * ふだんは画面上部に固定せず、**文字を選んだときだけ浮かぶ吹き出し**（BubbleFormatMenu）の
 * 中身として出る。ここでは中身だけを取り出して見ている。
 *
 * ボタンを押しても本文の選択が外れないようにしてある。外れると「どこに書式を掛けるのか」が
 * 分からなくなるため。
 *
 * この部品はエディタ本体（tiptap）が無いと動かないので、見本では小さなエディタを一緒に作っている。
 */
/*
 * この部品は単体では立てられない（エディタ本体が要る）。args ではなく render の中で
 * 作るので、meta も satisfies ではなく注釈で受けて args を任意にする。
 */
const meta: Meta<typeof FormatMenuBar> = {
  title: 'shared/RichTextEditor/FormatMenuBar',
  component: FormatMenuBar,
  parameters: { layout: 'padded' },
};

export default meta;
type Story = StoryObj<typeof FormatMenuBar>;

const SAMPLE = {
  type: 'doc',
  content: [
    { type: 'paragraph', content: [{ type: 'text', text: 'ここの文字を選んで書式を変えます。' }] },
  ],
};

/** 見本用の小さなエディタ。書式の効き目を目で確かめられるよう、本文も一緒に出す。 */
function MenuBarHarness({ selectAll = false }: { selectAll?: boolean }) {
  const editor = useEditor({
    extensions: createEditorExtensions({}),
    content: SAMPLE,
    editorProps: {
      attributes: { role: 'textbox', 'aria-multiline': 'true', 'aria-label': '本文' },
    },
    onCreate: ({ editor: created }) => {
      if (selectAll) created.commands.selectAll();
    },
  });

  if (!editor) return null;

  return (
    <div className="max-w-2xl space-y-3">
      <div className="rte-bubble inline-flex">
        <FormatMenuBar editor={editor as Editor} />
      </div>
      <div className="rounded border border-surface-3 p-3">
        <EditorContent editor={editor} />
      </div>
    </div>
  );
}

/** 何も選んでいないとき。 */
export const 既定: Story = {
  render: () => <MenuBarHarness />,
  play: async ({ canvasElement }) => {
    await expect(
      await within(canvasElement).findByRole('toolbar', { name: '書式メニュー' }),
    ).toBeVisible();
  },
};

/** 文字を選んだ状態。ここから太字などを掛けられる。 */
export const 選んだ状態: Story = {
  render: () => <MenuBarHarness selectAll />,
};

/** 太字を押すと、選んだところが太くなる。 */
export const 太字にする: Story = {
  render: () => <MenuBarHarness selectAll />,
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const bold = await canvas.findByRole('button', { name: '太字' });
    await userEvent.click(bold);
    await waitFor(async () => {
      // 押されている状態が、色だけでなく属性でも伝わる。
      await expect(bold).toHaveAttribute('aria-pressed', 'true');
    });
    await expect(canvasElement.querySelector('strong')).not.toBeNull();
  },
};
