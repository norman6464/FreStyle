import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import RichTextEditor from '../RichTextEditor';
import SaveStatusIndicator from '../SaveStatusIndicator';
import { emptyRichDoc, isRichDoc, type RichDocContent } from '../emptyRichDoc';

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

  it('太字ボタンで aria-pressed がトグルする', async () => {
    render(<RichTextEditor value={emptyRichDoc()} />);
    const bold = screen.getByRole('button', { name: '太字' });
    expect(bold).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(bold);
    await waitFor(() => expect(bold).toHaveAttribute('aria-pressed', 'true'));
  });

  it('saveStatus を渡すと保存状態を表示する', () => {
    render(<RichTextEditor value={emptyRichDoc()} saveStatus="saved" />);
    expect(screen.getByText('保存済み')).toBeInTheDocument();
  });

  it('ariaLabel が編集領域に付く', () => {
    const { container } = render(<RichTextEditor value={emptyRichDoc()} ariaLabel="メモ本文" />);
    expect(container.querySelector('[aria-label="メモ本文"]')).not.toBeNull();
  });

  it('onChange は編集で doc JSON を返す（初期描画では呼ばれない）', async () => {
    const onChange = vi.fn();
    render(<RichTextEditor value={emptyRichDoc()} onChange={onChange} />);
    // 初期描画では onUpdate は発火しない。
    expect(onChange).not.toHaveBeenCalled();
    // ツールバー操作（コマンド）で内容が変わると onChange が doc JSON を返す。
    fireEvent.click(screen.getByRole('button', { name: '箇条書き' }));
    await waitFor(() => expect(onChange).toHaveBeenCalled());
    const last = onChange.mock.calls.at(-1)?.[0];
    expect(last).toMatchObject({ type: 'doc' });
  });

  it('全ての書式ボタンを操作してもクラッシュしない（各コマンド配線）', () => {
    render(<RichTextEditor value={emptyRichDoc()} />);
    const names = [
      '太字',
      '斜体',
      '下線',
      '打ち消し線',
      'インラインコード',
      '見出し1',
      '見出し2',
      '見出し3',
      '箇条書き',
      '番号付きリスト',
      '引用',
      'コードブロック',
      '水平線',
    ];
    for (const name of names) {
      fireEvent.click(screen.getByRole('button', { name }));
    }
    expect(screen.getByRole('toolbar')).toBeInTheDocument();
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
});
