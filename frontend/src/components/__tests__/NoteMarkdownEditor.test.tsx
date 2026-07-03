import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
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
    fireEvent.click(screen.getByRole('button', { name: 'プレビュー' }));
    // Markdown の見出しがレンダリングされ、 textarea は消える
    expect(screen.getByRole('heading', { name: '見出し' })).toBeInTheDocument();
    expect(document.querySelector('textarea')).toBeNull();
  });

  it('Edit タブに戻ると textarea が再表示される', () => {
    renderEditor();
    fireEvent.click(screen.getByRole('button', { name: 'プレビュー' }));
    fireEvent.click(screen.getByRole('button', { name: '編集' }));
    expect(getMarkdownTextarea().value).toBe('# 見出し\n\n本文');
  });

  it('textarea を変更すると onContentChange が呼ばれる', () => {
    const onContentChange = vi.fn();
    renderEditor({ onContentChange });
    fireEvent.change(getMarkdownTextarea(), { target: { value: '新しい内容' } });
    expect(onContentChange).toHaveBeenCalledWith('新しい内容');
  });

  it('空コンテンツの Preview は案内文を出す', () => {
    renderEditor({ content: '   ' });
    fireEvent.click(screen.getByRole('button', { name: 'プレビュー' }));
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

  // --- 画像アップロード（onImageUpload 指定時） ---

  function fileInput(): HTMLInputElement {
    const el = document.querySelector('input[type="file"]');
    if (!el) throw new Error('file input not found');
    return el as HTMLInputElement;
  }

  it('onImageUpload 未指定なら画像挿入ボタンを出さない', () => {
    renderEditor();
    expect(screen.queryByLabelText('画像を挿入')).not.toBeInTheDocument();
    expect(document.querySelector('input[type="file"]')).toBeNull();
  });

  it('画像を選ぶとアップロードして ![](url) を挿入する', async () => {
    const onImageUpload = vi.fn().mockResolvedValue('https://cdn/x.png');
    const onContentChange = vi.fn();
    renderEditor({ onImageUpload, onContentChange, content: 'A' });

    const file = new File(['x'], 'diagram.png', { type: 'image/png' });
    fireEvent.change(fileInput(), { target: { files: [file] } });

    await waitFor(() => expect(onImageUpload).toHaveBeenCalledWith(file));
    await waitFor(() =>
      expect(onContentChange).toHaveBeenCalledWith(expect.stringContaining('![diagram](https://cdn/x.png)')),
    );
  });

  it('画像以外のファイルはアップロードしない', async () => {
    const onImageUpload = vi.fn().mockResolvedValue('https://cdn/x.png');
    const onContentChange = vi.fn();
    renderEditor({ onImageUpload, onContentChange });
    const file = new File(['x'], 'a.txt', { type: 'text/plain' });
    fireEvent.change(fileInput(), { target: { files: [file] } });
    // 非同期を一巡させてから、 アップロード/挿入されていないことを確認（type ガードで弾く）。
    await new Promise((r) => setTimeout(r, 0));
    expect(onImageUpload).not.toHaveBeenCalled();
    expect(onContentChange).not.toHaveBeenCalled();
  });

  it('textarea への画像ドロップでアップロードして挿入する', async () => {
    const onImageUpload = vi.fn().mockResolvedValue('https://cdn/y.png');
    const onContentChange = vi.fn();
    renderEditor({ onImageUpload, onContentChange, content: 'A' });
    const file = new File(['x'], 'arch.png', { type: 'image/png' });
    fireEvent.drop(getMarkdownTextarea(), { dataTransfer: { files: [file] } });
    await waitFor(() => expect(onImageUpload).toHaveBeenCalledWith(file));
    await waitFor(() =>
      expect(onContentChange).toHaveBeenCalledWith(expect.stringContaining('![arch](https://cdn/y.png)')),
    );
  });
});
