import { describe, it, expect, vi } from 'vitest';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import type { Editor } from '@tiptap/react';
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

  it('固定ツールバーを持たない（インライン表示）', () => {
    render(<RichTextEditor value={emptyRichDoc()} />);
    // 固定/可視のツールバーを持たないことだけを担保する。
    // 選択時のバブルメニューは BubbleMenu が「表示時にだけ」中身を DOM へ接続するため、
    // 未選択の jsdom では role=toolbar は hidden:true でも見つからない（＝未マウント）。
    // バブル書式バーの表示・操作は FormatMenuBar の単体テストと e2e（jsdom 外）で担保する。
    expect(screen.queryByRole('toolbar')).not.toBeInTheDocument();
    expect(screen.queryByRole('toolbar', { hidden: true })).not.toBeInTheDocument();
  });

  it('editable=false では本文が編集不可になる', () => {
    const { container } = render(<RichTextEditor value={headingDoc} editable={false} />);
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
    expect(screen.getByRole('textbox', { name: 'メモ本文' })).toBeInTheDocument();
  });

  it('初期描画では onChange を呼ばない（読み込み直後に未保存へ落ちない）', () => {
    const onChange = vi.fn();
    render(<RichTextEditor value={emptyRichDoc()} onChange={onChange} />);
    expect(onChange).not.toHaveBeenCalled();
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

  it('onCreate で生成直後の editor を渡す', async () => {
    const onCreate = vi.fn();
    render(<RichTextEditor value={emptyRichDoc()} onCreate={onCreate} />);
    await waitFor(() => expect(onCreate).toHaveBeenCalledTimes(1));
    // tiptap の editor 実体（chain を持つ）が渡ること。
    expect(typeof onCreate.mock.calls[0][0]?.chain).toBe('function');
  });

  // 制御コンポーネントの中核契約: 編集で onChange が新しい doc を伴って発火する（自動保存の起点）。
  // onCreate 経由で得た editor に対して編集を発火し、dedup 条件が反転して onChange が
  // 止まる退行を捕捉する。
  it('編集すると onChange に更新後の doc（type=doc）が渡る', async () => {
    const onChange = vi.fn();
    let editor: Editor | null = null;
    render(
      <RichTextEditor
        value={emptyRichDoc()}
        onChange={onChange}
        onCreate={(created) => {
          editor = created;
        }}
      />,
    );
    await waitFor(() => expect(editor).not.toBeNull());
    act(() => {
      editor!.commands.insertContent('追記テキスト');
    });
    await waitFor(() => expect(onChange).toHaveBeenCalled());
    const doc = onChange.mock.calls.at(-1)?.[0] as RichDocContent;
    expect(doc.type).toBe('doc');
    expect(JSON.stringify(doc)).toContain('追記テキスト');
  });

  // エコー抑止: 外部から value を差し替えた同期は onChange を再発火しない（常時「未保存」への退行防止）。
  it('外部 value の差し替えでは onChange を発火しない（エコー抑止）', async () => {
    const onChange = vi.fn();
    const { rerender } = render(<RichTextEditor value={emptyRichDoc()} onChange={onChange} />);
    const next: RichDocContent = {
      type: 'doc',
      content: [{ type: 'paragraph', content: [{ type: 'text', text: '外部から差し替え' }] }],
    };
    rerender(<RichTextEditor value={next} onChange={onChange} />);
    expect(await screen.findByText('外部から差し替え')).toBeInTheDocument();
    expect(onChange).not.toHaveBeenCalled();
  });
});

describe('doc の同一性はキー順に依らない', () => {
  it('キー順が違うだけの value を読み込んでも onChange を発火しない（開いただけで保存させない）', async () => {
    // サーバーはページ参照の題名解決で doc を作り直し、キーがアルファベット順で返る。
    // 素の文字列比較だとマウント時のエコーが「変更」と誤判定され、閲覧しただけの人が
    // 本文の全置換 PUT を発行してしまう。
    const alphabetized = JSON.parse(
      '{"content":[{"content":[{"text":"本文","type":"text"}],"type":"paragraph"}],"type":"doc"}',
    );
    const onChange = vi.fn();
    render(<RichTextEditor value={alphabetized} onChange={onChange} />);

    await waitFor(() => expect(screen.getByRole('textbox', { name: '本文' })).toBeInTheDocument());
    // マウント直後のエコーを流し切っても発火しない。
    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(onChange).not.toHaveBeenCalled();
  });
});

describe('常設ツールバー', () => {
  it('toolbar を渡すと書式ボタン列が常に出る（編集できるときだけ）', async () => {
    render(<RichTextEditor value={emptyRichDoc()} toolbar />);
    await waitFor(() => expect(screen.getByRole('textbox', { name: '本文' })).toBeInTheDocument());

    expect(screen.getByRole('toolbar', { name: '書式メニュー' })).toBeInTheDocument();
  });

  it('読み取り専用ではツールバーを出さない（押せない操作を見せない）', async () => {
    render(<RichTextEditor value={emptyRichDoc()} toolbar editable={false} />);
    await waitFor(() => expect(screen.getByRole('textbox', { name: '本文' })).toBeInTheDocument());

    expect(screen.queryByRole('toolbar', { name: '書式メニュー' })).not.toBeInTheDocument();
  });
});

const PAGE_UUID = '01a045ef-35de-7e9d-b637-84a5eb6fad77';

/** 内部（相対・絶対）と外部のリンクを 1 段落に並べた doc。 */
function docWithLinks(): RichDocContent {
  return {
    type: 'doc',
    content: [
      {
        type: 'paragraph',
        content: [
          {
            type: 'text',
            text: '内部リンク',
            marks: [{ type: 'link', attrs: { href: `/p/${PAGE_UUID}` } }],
          },
          { type: 'text', text: ' / ' },
          {
            type: 'text',
            text: '共有URL',
            marks: [
              { type: 'link', attrs: { href: `${window.location.origin}/p/${PAGE_UUID}` } },
            ],
          },
          { type: 'text', text: ' / ' },
          {
            type: 'text',
            text: '外部リンク',
            marks: [{ type: 'link', attrs: { href: 'https://example.com/docs' } }],
          },
        ],
      },
    ],
  };
}

describe('リンクのクリック', () => {
  it('内部ページリンクは編集中でもクリックでアプリ内遷移する', async () => {
    const navigateToPage = vi.fn();
    render(
      <RichTextEditor value={docWithLinks()} editable onNavigateToPage={navigateToPage} />,
    );
    await waitFor(() => expect(screen.getByText('内部リンク')).toBeInTheDocument());

    fireEvent.click(screen.getByText('内部リンク'));

    expect(navigateToPage).toHaveBeenCalledWith(`/p/${PAGE_UUID}`);
  });

  it('同一オリジンの絶対 URL（共有 URL の貼り付け）もアプリ内遷移に畳む', async () => {
    const navigateToPage = vi.fn();
    render(
      <RichTextEditor value={docWithLinks()} editable onNavigateToPage={navigateToPage} />,
    );
    await waitFor(() => expect(screen.getByText('共有URL')).toBeInTheDocument());

    fireEvent.click(screen.getByText('共有URL'));

    expect(navigateToPage).toHaveBeenCalledWith(`/p/${PAGE_UUID}`);
  });

  it('外部リンクは新しいタブで開く（rel で opener を渡さない）', async () => {
    const open = vi.spyOn(window, 'open').mockReturnValue(null);
    const navigateToPage = vi.fn();
    render(
      <RichTextEditor value={docWithLinks()} editable onNavigateToPage={navigateToPage} />,
    );
    await waitFor(() => expect(screen.getByText('外部リンク')).toBeInTheDocument());

    fireEvent.click(screen.getByText('外部リンク'));

    expect(open).toHaveBeenCalledWith('https://example.com/docs', '_blank', 'noopener,noreferrer');
    expect(navigateToPage).not.toHaveBeenCalled();
    open.mockRestore();
  });

  it('文字を選んだだけのクリックでは開かない（リンクの文言を選べる）', async () => {
    // リンクをドラッグで選ぶと mouseup のときに click も発火する。ここで開くと
    // リンクの文言を打ち直す・リンクを外す、といった編集がマウスでできなくなる。
    const navigateToPage = vi.fn();
    const open = vi.spyOn(window, 'open').mockReturnValue(null);
    render(
      <RichTextEditor value={docWithLinks()} editable onNavigateToPage={navigateToPage} />,
    );
    await waitFor(() => expect(screen.getByText('内部リンク')).toBeInTheDocument());

    const selection = window.getSelection();
    const range = document.createRange();
    range.selectNodeContents(screen.getByText('内部リンク'));
    selection?.removeAllRanges();
    selection?.addRange(range);
    fireEvent.click(screen.getByText('内部リンク'));

    expect(navigateToPage).not.toHaveBeenCalled();
    expect(open).not.toHaveBeenCalled();
    selection?.removeAllRanges();
    open.mockRestore();
  });

  it('編集中の Shift+クリックは選択を伸ばす操作なので、新しいタブを開かない', async () => {
    const navigateToPage = vi.fn();
    const open = vi.spyOn(window, 'open').mockReturnValue(null);
    render(
      <RichTextEditor value={docWithLinks()} editable onNavigateToPage={navigateToPage} />,
    );
    await waitFor(() => expect(screen.getByText('内部リンク')).toBeInTheDocument());

    fireEvent.click(screen.getByText('内部リンク'), { shiftKey: true });

    expect(open).not.toHaveBeenCalled();
    expect(navigateToPage).toHaveBeenCalledWith(`/p/${PAGE_UUID}`);
    open.mockRestore();
  });

  it('読み取り専用の Shift+クリックは新しいタブで開く（選択を伸ばす操作が無い面）', async () => {
    const open = vi.spyOn(window, 'open').mockReturnValue(null);
    render(<RichTextEditor value={docWithLinks()} editable={false} onNavigateToPage={vi.fn()} />);
    await waitFor(() => expect(screen.getByText('内部リンク')).toBeInTheDocument());

    fireEvent.click(screen.getByText('内部リンク'), { shiftKey: true });

    expect(open).toHaveBeenCalled();
    open.mockRestore();
  });

  it('修飾キー付きのクリックは内部リンクでも新しいタブで開く', async () => {
    const open = vi.spyOn(window, 'open').mockReturnValue(null);
    const navigateToPage = vi.fn();
    render(
      <RichTextEditor value={docWithLinks()} editable onNavigateToPage={navigateToPage} />,
    );
    await waitFor(() => expect(screen.getByText('内部リンク')).toBeInTheDocument());

    fireEvent.click(screen.getByText('内部リンク'), { metaKey: true });

    expect(open).toHaveBeenCalled();
    expect(navigateToPage).not.toHaveBeenCalled();
    open.mockRestore();
  });

  it('読み取り専用でもクリックでアプリ内遷移する（素の全画面リロードにしない）', async () => {
    const navigateToPage = vi.fn();
    render(
      <RichTextEditor value={docWithLinks()} editable={false} onNavigateToPage={navigateToPage} />,
    );
    await waitFor(() => expect(screen.getByText('内部リンク')).toBeInTheDocument());

    fireEvent.click(screen.getByText('内部リンク'));

    expect(navigateToPage).toHaveBeenCalledWith(`/p/${PAGE_UUID}`);
  });
});

describe('focusSignal（題名で Enter → 本文へ）', () => {
  it('合図が増えたら本文へフォーカスが移る', async () => {
    const { rerender } = render(
      <RichTextEditor value={emptyRichDoc()} editable focusSignal={0} />,
    );
    const textbox = await screen.findByRole('textbox', { name: '本文' });
    expect(document.activeElement).not.toBe(textbox);

    rerender(<RichTextEditor value={emptyRichDoc()} editable focusSignal={1} />);

    await waitFor(() => expect(document.activeElement).toBe(textbox));
  });

  it('マウント時の値では動かない（ページを開いただけで本文が奪わない）', async () => {
    render(<RichTextEditor value={emptyRichDoc()} editable focusSignal={5} />);
    const textbox = await screen.findByRole('textbox', { name: '本文' });

    expect(document.activeElement).not.toBe(textbox);
  });
});
