import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import RichTextEditor from '../RichTextEditor';
import SaveStatusIndicator from '../SaveStatusIndicator';
import { emptyRichDoc, isRichDoc, type RichDocContent } from '../emptyRichDoc';
import type { JSONContent } from '@tiptap/react';

/** collectNodeTypes は doc ツリーに現れる全ノード type の集合を返す（書式コマンドの効果検証用）。 */
function collectNodeTypes(node: JSONContent, acc: Set<string> = new Set()): Set<string> {
  if (node.type) acc.add(node.type);
  node.content?.forEach((child) => collectNodeTypes(child, acc));
  return acc;
}

/** findNode は doc ツリーを深さ優先で走査し、最初に見つかった指定 type のノードを返す。 */
function findNode(node: JSONContent, type: string): JSONContent | undefined {
  if (node.type === type) return node;
  for (const child of node.content ?? []) {
    const found = findNode(child, type);
    if (found) return found;
  }
  return undefined;
}

const headingDoc: RichDocContent = {
  type: 'doc',
  content: [
    { type: 'heading', attrs: { level: 2 }, content: [{ type: 'text', text: '見出しテスト' }] },
    { type: 'paragraph', content: [{ type: 'text', text: '本文テキスト' }] },
  ],
};

describe('emptyRichDoc / isRichDoc', () => {
  it('emptyRichDoc は空の doc（段落1つ）を返す', () => {
    expect(emptyRichDoc()).toEqual({ type: 'doc', content: [{ type: 'paragraph' }] });
  });

  it('isRichDoc は type=doc の object のみ true', () => {
    expect(isRichDoc({ type: 'doc', content: [] })).toBe(true);
    expect(isRichDoc({ type: 'paragraph' })).toBe(false);
    expect(isRichDoc(null)).toBe(false);
    expect(isRichDoc('doc')).toBe(false);
    expect(isRichDoc([])).toBe(false);
  });
});

describe('SaveStatusIndicator', () => {
  it('idle は何も描画しない', () => {
    const { container } = render(<SaveStatusIndicator status="idle" />);
    expect(container).toBeEmptyDOMElement();
  });

  it('各状態のラベルと色を表示する', () => {
    const { rerender } = render(<SaveStatusIndicator status="unsaved" />);
    expect(screen.getByText('未保存')).toHaveClass('text-amber-500');

    rerender(<SaveStatusIndicator status="saving" />);
    expect(screen.getByText('保存中...')).toBeInTheDocument();

    rerender(<SaveStatusIndicator status="saved" />);
    expect(screen.getByText('保存済み')).toHaveClass('text-emerald-500');
  });
});

describe('RichTextEditor', () => {
  it('value の内容を描画する', async () => {
    render(<RichTextEditor value={headingDoc} />);
    expect(await screen.findByText('見出しテスト')).toBeInTheDocument();
    expect(screen.getByText('本文テキスト')).toBeInTheDocument();
  });

  it('editable=true でツールバー（主要ボタン）を表示する', () => {
    render(<RichTextEditor value={emptyRichDoc()} />);
    expect(screen.getByRole('toolbar', { name: '書式ツールバー' })).toBeInTheDocument();
    for (const name of ['太字', '斜体', '見出し1', '箇条書き', 'コードブロック', '元に戻す']) {
      expect(screen.getByRole('button', { name })).toBeInTheDocument();
    }
  });

  it('editable=false ではツールバーを出さず、本文が編集不可になる', () => {
    const { container } = render(<RichTextEditor value={headingDoc} editable={false} />);
    expect(screen.queryByRole('toolbar')).not.toBeInTheDocument();
    const pm = container.querySelector('.ProseMirror');
    expect(pm).not.toBeNull();
    expect(pm).toHaveAttribute('contenteditable', 'false');
  });

  it('saveStatus を渡すと保存状態を表示する', () => {
    render(<RichTextEditor value={emptyRichDoc()} saveStatus="saved" />);
    expect(screen.getByText('保存済み')).toBeInTheDocument();
  });

  it('ariaLabel が編集領域のアクセシブルネームになる', () => {
    render(<RichTextEditor value={emptyRichDoc()} ariaLabel="メモ本文" />);
    // role=textbox とアクセシブルネームの両方を検証する（CSS セレクタでは a11y ツリーを見ない）。
    expect(screen.getByRole('textbox', { name: 'メモ本文' })).toBeInTheDocument();
  });

  it('初期描画では onChange を呼ばない（読み込み直後に未保存へ落ちない）', () => {
    const onChange = vi.fn();
    render(<RichTextEditor value={emptyRichDoc()} onChange={onChange} />);
    expect(onChange).not.toHaveBeenCalled();
  });

  // ブロック系コマンドは空段落を変換するので、onChange の doc JSON に対象ノードが現れることを検証する。
  it.each([
    ['見出し1', 'heading'],
    ['見出し2', 'heading'],
    ['見出し3', 'heading'],
    ['箇条書き', 'bulletList'],
    ['番号付きリスト', 'orderedList'],
    ['引用', 'blockquote'],
    ['コードブロック', 'codeBlock'],
    ['水平線', 'horizontalRule'],
  ])('「%s」操作で doc JSON に %s ノードが現れる', async (buttonName, nodeType) => {
    const onChange = vi.fn();
    render(<RichTextEditor value={emptyRichDoc()} onChange={onChange} />);
    fireEvent.click(screen.getByRole('button', { name: buttonName }));
    await waitFor(() => expect(onChange).toHaveBeenCalled());
    const doc = onChange.mock.calls.at(-1)?.[0] as RichDocContent;
    expect(doc.type).toBe('doc');
    expect(collectNodeTypes(doc)).toContain(nodeType);
  });

  it('「見出し1」操作で level:1 の heading になる', async () => {
    const onChange = vi.fn();
    render(<RichTextEditor value={emptyRichDoc()} onChange={onChange} />);
    fireEvent.click(screen.getByRole('button', { name: '見出し1' }));
    await waitFor(() => expect(onChange).toHaveBeenCalled());
    const doc = onChange.mock.calls.at(-1)?.[0] as RichDocContent;
    const heading = findNode(doc, 'heading');
    expect(heading?.attrs?.level).toBe(1);
  });

  // マーク系コマンドは（選択が無くても）記憶マークとして有効になり、active（aria-pressed）が立つ。
  it.each(['太字', '斜体', '下線', '打ち消し線', 'インラインコード'])(
    '「%s」操作で aria-pressed が true になる',
    async (buttonName) => {
      render(<RichTextEditor value={emptyRichDoc()} />);
      const button = screen.getByRole('button', { name: buttonName });
      expect(button).toHaveAttribute('aria-pressed', 'false');
      fireEvent.click(button);
      await waitFor(() => expect(button).toHaveAttribute('aria-pressed', 'true'));
    },
  );

  it('非トグル操作（水平線・元に戻す・やり直す）には aria-pressed を付けない', () => {
    render(<RichTextEditor value={emptyRichDoc()} />);
    for (const name of ['水平線', '元に戻す', 'やり直す']) {
      expect(screen.getByRole('button', { name })).not.toHaveAttribute('aria-pressed');
    }
  });

  it('undo / redo は初期状態で無効', () => {
    render(<RichTextEditor value={emptyRichDoc()} />);
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'やり直す' })).toBeDisabled();
  });

  it('外部から value が変わると本文が差し替わる', async () => {
    const { rerender } = render(<RichTextEditor value={emptyRichDoc()} />);
    const next: RichDocContent = {
      type: 'doc',
      content: [{ type: 'paragraph', content: [{ type: 'text', text: '新しい本文' }] }],
    };
    rerender(<RichTextEditor value={next} />);
    expect(await screen.findByText('新しい本文')).toBeInTheDocument();
  });

  it('editable を後から false にすると編集不可になる', async () => {
    const { rerender, container } = render(<RichTextEditor value={emptyRichDoc()} editable />);
    expect(container.querySelector('.ProseMirror')).toHaveAttribute('contenteditable', 'true');
    rerender(<RichTextEditor value={emptyRichDoc()} editable={false} />);
    await waitFor(() =>
      expect(container.querySelector('.ProseMirror')).toHaveAttribute('contenteditable', 'false'),
    );
  });

  it('onImageUpload 指定時だけ画像ボタンを出す', () => {
    const { rerender } = render(<RichTextEditor value={emptyRichDoc()} />);
    expect(screen.queryByRole('button', { name: '画像を挿入' })).not.toBeInTheDocument();
    rerender(<RichTextEditor value={emptyRichDoc()} onImageUpload={vi.fn()} />);
    expect(screen.getByRole('button', { name: '画像を挿入' })).toBeInTheDocument();
  });

  it('ファイル選択で onImageUpload を呼び、返却 URL とファイル名を img の src/alt に保存する', async () => {
    const onImageUpload = vi.fn().mockResolvedValue('https://cdn.example.com/a.png');
    const onChange = vi.fn();
    render(<RichTextEditor value={emptyRichDoc()} onImageUpload={onImageUpload} onChange={onChange} />);
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(['x'], 'a.png', { type: 'image/png' });
    fireEvent.change(input, { target: { files: [file] } });
    await waitFor(() => expect(onImageUpload).toHaveBeenCalledWith(file));
    // 描画された img の src=返却 URL / alt=ファイル名 を検証する。
    const img = await screen.findByRole('img', { name: 'a.png' });
    expect(img).toHaveAttribute('src', 'https://cdn.example.com/a.png');
    const last = onChange.mock.calls.at(-1)?.[0] as RichDocContent | undefined;
    expect(last && collectNodeTypes(last)).toContain('image');
  });

  // NOTE: 貼り付け/ドロップは editorProps.handlePaste/handleDrop から同じ uploadImageFiles を呼ぶ。
  // ProseMirror の paste/drop は jsdom の elementFromPoint 未実装のため実地シミュレーションが不安定なので、
  // 共有ロジックはファイル選択経路（上のテスト）で src/alt まで固定し、ここでは扱わない。

  it('画像以外のファイルは onImageUpload を呼ばない', () => {
    const onImageUpload = vi.fn().mockResolvedValue('x');
    render(<RichTextEditor value={emptyRichDoc()} onImageUpload={onImageUpload} />);
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    fireEvent.change(input, { target: { files: [new File(['x'], 'a.txt', { type: 'text/plain' })] } });
    expect(onImageUpload).not.toHaveBeenCalled();
  });

  it('アップロード完了前にアンマウントされたら画像を挿入しない（別文書への誤挿入防止）', async () => {
    let resolveUpload: (url: string) => void = () => {};
    const onImageUpload = vi.fn(
      () =>
        new Promise<string>((resolve) => {
          resolveUpload = resolve;
        }),
    );
    const onChange = vi.fn();
    const { unmount } = render(
      <RichTextEditor value={emptyRichDoc()} onImageUpload={onImageUpload} onChange={onChange} />,
    );
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    fireEvent.change(input, { target: { files: [new File(['x'], 'a.png', { type: 'image/png' })] } });
    await waitFor(() => expect(onImageUpload).toHaveBeenCalled());
    unmount(); // エディタ破棄（別ノートへ切替相当）
    resolveUpload('https://cdn.example.com/late.png');
    await Promise.resolve();
    await Promise.resolve();
    const insertedImage = onChange.mock.calls.some((call) =>
      collectNodeTypes(call[0] as RichDocContent).has('image'),
    );
    expect(insertedImage).toBe(false);
  });
});
