import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useEditor } from '@tiptap/react';
import FormatMenuBar from '../FormatMenuBar';
import { createEditorExtensions } from '../editorExtensions';
import { emptyRichDoc } from '../emptyRichDoc';

// FormatMenuBar は実 editor を必要とするので、useEditor で用意した editor を渡す薄いハーネスで包む。
function Harness() {
  const editor = useEditor({
    extensions: createEditorExtensions(),
    content: emptyRichDoc(),
  });
  if (!editor) return null;
  return <FormatMenuBar editor={editor} />;
}

describe('FormatMenuBar', () => {
  it('マーク＋ブロック変換のボタンを出す（挿入・履歴系は出さない）', () => {
    render(<Harness />);
    expect(screen.getByRole('toolbar', { name: '書式メニュー' })).toBeInTheDocument();
    for (const name of ['太字', '斜体', '下線', '打ち消し線', 'インラインコード', '見出し1', '箇条書き', '引用', 'コードブロック']) {
      expect(screen.getByRole('button', { name })).toBeInTheDocument();
    }
    // 水平線（insert）・元に戻す（history）はバブルメニューには出さない。
    expect(screen.queryByRole('button', { name: '水平線' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '元に戻す' })).not.toBeInTheDocument();
  });

  it.each(['太字', '斜体', '見出し1', '箇条書き', 'コードブロック'])(
    '「%s」クリックで aria-pressed が true になる',
    async (name) => {
      render(<Harness />);
      const button = screen.getByRole('button', { name });
      expect(button).toHaveAttribute('aria-pressed', 'false');
      fireEvent.click(button);
      await waitFor(() => expect(button).toHaveAttribute('aria-pressed', 'true'));
    },
  );
});
