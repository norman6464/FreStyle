import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import NoteMarkdownEditor from '../NoteMarkdownEditor';
import { ToastProvider } from '../ToastProvider';

function renderEditor(props: Partial<React.ComponentProps<typeof NoteMarkdownEditor>> = {}) {
  const defaults: React.ComponentProps<typeof NoteMarkdownEditor> = {
    title: 'タイトル',
    content: '# 見出し\n\n本文',
    saveStatus: 'idle',
    onTitleChange: vi.fn(),
    onContentChange: vi.fn(),
  };
  return render(
    <ToastProvider>
      <NoteMarkdownEditor {...defaults} {...props} />
    </ToastProvider>,
  );
}

function getMarkdownTextarea(): HTMLTextAreaElement {
  const ta = document.querySelector('textarea');
  if (!ta) throw new Error('textarea not found');
  return ta;
}

describe('NoteMarkdownEditor', () => {
  it('Edit タブで textarea に Markdown を表示する', () => {
    renderEditor();
    expect(getMarkdownTextarea().value).toBe('# 見出し\n\n本文');
  });

  it('Preview タブをクリックすると Markdown レンダリングが表示される', () => {
    renderEditor();
    fireEvent.click(screen.getByRole('button', { name: 'Preview' }));
    // Markdown の見出しがレンダリングされ、 textarea は消える
    expect(screen.getByRole('heading', { name: '見出し' })).toBeInTheDocument();
    expect(document.querySelector('textarea')).toBeNull();
  });

  it('Edit タブに戻ると textarea が再表示される', () => {
    renderEditor();
    fireEvent.click(screen.getByRole('button', { name: 'Preview' }));
    fireEvent.click(screen.getByRole('button', { name: 'Edit' }));
    expect(getMarkdownTextarea().value).toBe('# 見出し\n\n本文');
  });

  it('textarea を変更すると onContentChange が呼ばれる', () => {
    const onContentChange = vi.fn();
    renderEditor({ onContentChange });
    fireEvent.change(getMarkdownTextarea(), { target: { value: '新しい内容' } });
    expect(onContentChange).toHaveBeenCalledWith('新しい内容');
  });

  it('TipTap JSON のレガシー content は Markdown に変換して表示される', () => {
    const tiptapJson = JSON.stringify({
      type: 'doc',
      content: [
        { type: 'heading', attrs: { level: 1 }, content: [{ type: 'text', text: '旧データ' }] },
      ],
    });
    const onContentChange = vi.fn();
    renderEditor({ content: tiptapJson, onContentChange });
    // textarea には変換後の Markdown が出る
    expect(getMarkdownTextarea().value).toBe('# 旧データ');
    // 自動的に親に通知して store を Markdown 化する
    expect(onContentChange).toHaveBeenCalledWith('# 旧データ');
  });

  it('空コンテンツの Preview は案内文を出す', () => {
    renderEditor({ content: '   ' });
    fireEvent.click(screen.getByRole('button', { name: 'Preview' }));
    expect(screen.getByText('プレビューするコンテンツがありません')).toBeInTheDocument();
  });

  // IME 変換中の Enter（日本語入力の確定）はフォーカス移動しない（誤って textarea に飛んで
  // 未確定文字列が紛れ込む事故の再発防止）。
  it('タイトルで IME 変換中の Enter は textarea にフォーカスを移さない', () => {
    renderEditor();
    const titleInput = screen.getByRole('textbox', { name: 'ノートのタイトル' });
    titleInput.focus();
    fireEvent.keyDown(titleInput, { key: 'Enter', isComposing: true });
    expect(document.activeElement).toBe(titleInput);
  });

  it('タイトルで通常の Enter は textarea にフォーカスを移す', () => {
    renderEditor();
    const titleInput = screen.getByRole('textbox', { name: 'ノートのタイトル' });
    titleInput.focus();
    fireEvent.keyDown(titleInput, { key: 'Enter' });
    expect(document.activeElement).toBe(getMarkdownTextarea());
  });
});
