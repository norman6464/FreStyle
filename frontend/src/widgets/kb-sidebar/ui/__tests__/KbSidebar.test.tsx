import { act, render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import KbSidebar from '../KbSidebar';
import type { KbPage, KbPageTree, KbSpace, KbWorkspace } from '@/entities/knowledge-base';

const hoisted = vi.hoisted(() => ({
  fetchWorkspaces: vi.fn(),
  fetchSpaces: vi.fn(),
  fetchPageTree: vi.fn(),
  createPage: vi.fn(),
  renamePage: vi.fn(),
  archivePage: vi.fn(),
  unarchivePage: vi.fn(),
  movePage: vi.fn(),
  showToast: vi.fn(),
}));

// トーストは検査の対象。**失敗したときだけ知らせが出ること**を確かめるために捕まえる。
vi.mock('@/shared/lib/hooks/useToast', () => ({
  useToast: () => ({ showToast: hoisted.showToast, toasts: [], removeToast: vi.fn() }),
}));

vi.mock('@/entities/knowledge-base', async () => {
  // ツリーの平坦化と祖先探索は本物を使う（そこは entities 側でテスト済みで、
  // ここで別実装を差し込むと「サイドバーは通るが本番は壊れる」形になる）。
  const actual = await vi.importActual<typeof import('@/entities/knowledge-base')>(
    '@/entities/knowledge-base',
  );
  return {
    ...actual,
    KnowledgeBaseRepository: {
      fetchWorkspaces: hoisted.fetchWorkspaces,
      fetchSpaces: hoisted.fetchSpaces,
      fetchPageTree: hoisted.fetchPageTree,
      createPage: hoisted.createPage,
      renamePage: hoisted.renamePage,
      archivePage: hoisted.archivePage,
      unarchivePage: hoisted.unarchivePage,
      movePage: hoisted.movePage,
    },
  };
});

function workspace(slug: string, name = slug): KbWorkspace {
  return { slug, name, createdAt: '2026-08-01T00:00:00Z' };
}

function space(id: string, name = id): KbSpace {
  return { id, key: id, name, createdAt: '2026-08-01T00:00:00Z' };
}

function page(id: string, title = id): KbPage {
  return {
    id,
    spaceId: 'space-1',
    title,
    createdByUserId: 1,
    createdAt: '2026-08-01T00:00:00Z',
    updatedAt: '2026-08-01T00:00:00Z',
  };
}

function tree(
  nodes: {
    id: string;
    title?: string;
    hidden?: boolean;
    children?: string[];
    parentArchived?: boolean;
  }[],
  hiddenAtRoot = false,
): KbPageTree {
  return {
    pages: nodes.map((node) => ({
      page: page(node.id, node.title),
      hasHiddenChildren: node.hidden ?? false,
      parentArchived: node.parentArchived ?? false,
      children: (node.children ?? []).map((childId) => ({
        page: page(childId),
        hasHiddenChildren: false,
        // 一緒にアーカイブされた子は、その子だけを戻すことはできない。
        parentArchived: true,
        children: [],
      })),
    })),
    hasHiddenChildren: hiddenAtRoot,
  };
}

function renderSidebar(props: { workspaceSlug?: string; activePageId?: string } = {}) {
  return render(
    <MemoryRouter initialEntries={['/kb']}>
      <KbSidebar {...props} />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  hoisted.fetchWorkspaces.mockResolvedValue([workspace('acme', 'Acme 社')]);
  hoisted.fetchSpaces.mockResolvedValue([space('space-1', '開発部')]);
  hoisted.fetchPageTree.mockResolvedValue(tree([{ id: 'p1', title: '設計メモ' }]));
  hoisted.createPage.mockResolvedValue(page('new-1', '無題'));
  hoisted.renamePage.mockImplementation(async (_slug: string, id: string, title: string) =>
    page(id, title),
  );
  hoisted.archivePage.mockResolvedValue(undefined);
  hoisted.unarchivePage.mockResolvedValue(undefined);
  hoisted.movePage.mockResolvedValue(page('p1'));
});

/** 行の矩形を固定して、落とす位置（上端 / 中央 / 下端）を狙えるようにする。 */
function stubRowRect(row: HTMLElement) {
  row.getBoundingClientRect = () =>
    ({ top: 0, height: 100, left: 0, right: 0, bottom: 100, width: 100, x: 0, y: 0, toJSON: () => ({}) }) as DOMRect;
}

/**
 * dragRowOnto は from の行を to の行の指定位置へドラッグして落とす。
 *
 * dragOver / drop は **MouseEvent として作る**。fireEvent.drop(el, { clientY }) だと
 * jsdom に DragEvent が無いぶん素の Event に落ちて clientY が届かず、
 * 落下先が必ず「中央（子として）」になってしまう（実測で確認）。
 */
function dragRowOnto(fromTitle: string, toTitle: string, clientY: number) {
  const from = screen.getByRole('treeitem', { name: fromTitle });
  const to = screen.getByRole('treeitem', { name: toTitle });
  stubRowRect(to);
  fireEvent.dragStart(from, { dataTransfer: { setData: vi.fn(), effectAllowed: '' } });
  const withPosition = (type: string) => {
    const event = new MouseEvent(type, { bubbles: true, clientY });
    Object.defineProperty(event, 'dataTransfer', { value: { dropEffect: '' } });
    return event;
  };
  fireEvent(to, withPosition('dragover'));
  fireEvent(to, withPosition('drop'));
}

describe('KbSidebar', () => {
  it('先頭のスペースを開いて木を出す', async () => {
    renderSidebar();

    expect(await screen.findByText('設計メモ')).toBeInTheDocument();
    expect(hoisted.fetchPageTree).toHaveBeenCalledWith('acme', 'space-1', { archived: false });
  });

  it('開いていないスペースの木は取りに行かない', async () => {
    hoisted.fetchSpaces.mockResolvedValue([space('space-1', '開発部'), space('space-2', '営業部')]);

    renderSidebar();
    await screen.findByText('設計メモ');

    // スペースの数だけ要求が飛ぶ（N+1）と、開いていない分が丸ごと無駄になる。
    expect(hoisted.fetchPageTree).toHaveBeenCalledTimes(1);
    expect(hoisted.fetchPageTree).not.toHaveBeenCalledWith('acme', 'space-2', { archived: false });
  });

  it('スペースを開いたときに初めて取りに行く', async () => {
    hoisted.fetchSpaces.mockResolvedValue([space('space-1', '開発部'), space('space-2', '営業部')]);
    renderSidebar();
    await screen.findByText('設計メモ');

    fireEvent.click(screen.getByRole('button', { name: '営業部' }));

    await waitFor(() =>
      expect(hoisted.fetchPageTree).toHaveBeenCalledWith('acme', 'space-2', { archived: false }),
    );
  });

  it('子を持つページはフォルダ、持たないページは紙にする', async () => {
    hoisted.fetchPageTree.mockResolvedValue(
      tree([{ id: 'p1', title: '親ページ', children: ['p1-child'] }, { id: 'p2', title: '葉ページ' }]),
    );
    renderSidebar();
    await screen.findByText('親ページ');

    // リンクとして引く。役割とアクセシブルな名前の崩れも一緒に捕まえられる。
    const iconOf = (title: string) =>
      screen.getByRole('link', { name: title }).querySelector('[data-icon]')?.getAttribute('data-icon');

    expect(iconOf('親ページ')).toBe('page-group');
    expect(iconOf('葉ページ')).toBe('page');
  });

  it('伏せた子しか居ないページはフォルダにしない', async () => {
    // 開閉の三角が無いのにフォルダ、という食い違った行になるうえ、
    // 「この下に何かある」ことを形からも二重に漏らすことになる。
    hoisted.fetchPageTree.mockResolvedValue(tree([{ id: 'p1', title: '設計メモ', hidden: true }]));
    renderSidebar();
    await screen.findByText('設計メモ');

    const icon = screen.getByRole('link', { name: '設計メモ' }).querySelector('[data-icon]');
    expect(icon?.getAttribute('data-icon')).toBe('page');
  });

  it('伏せた子が在ることだけを出し、枚数も題名も出さない', async () => {
    hoisted.fetchPageTree.mockResolvedValue(tree([{ id: 'p1', title: '設計メモ', hidden: true }]));

    renderSidebar();

    expect(await screen.findByText('表示できないページがあります')).toBeInTheDocument();
  });

  it('伏せた子しか居ない行に開閉ボタンを出さない', async () => {
    // 開いても何も出ない行に三角を出すと、押しても反応しない行になる。
    hoisted.fetchPageTree.mockResolvedValue(tree([{ id: 'p1', title: '設計メモ', hidden: true }]));

    renderSidebar();
    await screen.findByText('設計メモ');

    expect(screen.queryByRole('button', { name: /設計メモ を開く/ })).not.toBeInTheDocument();
  });

  it('子を開くと次の段が出る', async () => {
    hoisted.fetchPageTree.mockResolvedValue(
      tree([{ id: 'p1', title: '設計メモ', children: ['p1-child'] }]),
    );
    renderSidebar();
    await screen.findByText('設計メモ');

    expect(screen.queryByText('p1-child')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '設計メモ を開く' }));

    expect(await screen.findByText('p1-child')).toBeInTheDocument();
  });

  it('伏せた印は、開いた段の子より後ろに出す', async () => {
    // 平らにした配列では親の要素が子より先に来るので、ページの行に混ぜると
    // 印が子より前に出る。読み順として最後の子の後ろが正しい。
    hoisted.fetchPageTree.mockResolvedValue(
      tree([{ id: 'p1', title: '設計メモ', hidden: true, children: ['p1-child'] }]),
    );
    renderSidebar();
    await screen.findByText('設計メモ');

    fireEvent.click(screen.getByRole('button', { name: '設計メモ を開く' }));
    await screen.findByText('p1-child');

    const text = document.body.textContent ?? '';
    expect(text.indexOf('p1-child')).toBeLessThan(text.indexOf('表示できないページがあります'));
  });

  it('閉じている段では伏せた印を出さない', async () => {
    // 見える子も出していないのに伏せた分だけ出すと、閉じているのに何か書いてある行になる。
    hoisted.fetchPageTree.mockResolvedValue(
      tree([{ id: 'p1', title: '設計メモ', hidden: true, children: ['p1-child'] }]),
    );

    renderSidebar();
    await screen.findByText('設計メモ');

    expect(screen.queryByText('表示できないページがあります')).not.toBeInTheDocument();
  });

  it('現在位置のページの祖先を自動で開く', async () => {
    // リンクを辿って開いたページが、閉じた枝の中に隠れたままにならないこと。
    hoisted.fetchPageTree.mockResolvedValue(
      tree([{ id: 'p1', title: '設計メモ', children: ['p1-child'] }]),
    );

    renderSidebar({ activePageId: 'p1-child' });

    expect(await screen.findByText('p1-child')).toBeInTheDocument();
  });

  it('木の取得に失敗したら理由を出して再試行できる', async () => {
    // 黙って空にすると「ページが 1 枚も無い」と見分けが付かず、消えたと思われる。
    hoisted.fetchPageTree.mockRejectedValueOnce(new Error('boom'));
    renderSidebar();

    expect(await screen.findByText('ページを読み込めませんでした')).toBeInTheDocument();

    hoisted.fetchPageTree.mockResolvedValue(tree([{ id: 'p1', title: '設計メモ' }]));
    fireEvent.click(screen.getByRole('button', { name: '再試行' }));

    expect(await screen.findByText('設計メモ')).toBeInTheDocument();
  });

  it('所属が無いときは、壊れているのではなく招待が要ると伝える', async () => {
    hoisted.fetchWorkspaces.mockResolvedValue([]);

    renderSidebar();

    expect(await screen.findByText(/招待してもらってください/)).toBeInTheDocument();
    expect(hoisted.fetchSpaces).not.toHaveBeenCalled();
  });

  it('ワークスペースの取得に失敗したら理由を出して再試行できる', async () => {
    // 一度だけ走る effect の中に閉じ込めると、失敗後は画面を再読み込みするしか手が無くなる。
    hoisted.fetchWorkspaces.mockRejectedValueOnce(new Error('boom'));
    renderSidebar();

    expect(await screen.findByText('ワークスペースを読み込めませんでした')).toBeInTheDocument();

    hoisted.fetchWorkspaces.mockResolvedValue([workspace('acme', 'Acme 社')]);
    fireEvent.click(screen.getByRole('button', { name: '再試行' }));

    expect(await screen.findByText('設計メモ')).toBeInTheDocument();
  });

  it('ワークスペースを切り替えたら、前のスペースを先に捨てる', async () => {
    // 残したまま新しい応答を待つと、その間だけ「前のワークスペースのスペース」と
    // 「新しい slug」が組み合わさって表示され、そこを開くと別ワークスペースの
    // spaceId で木を取りに行ってしまう。
    hoisted.fetchSpaces.mockResolvedValue([space('space-1', '開発部')]);

    const { rerender } = render(
      <MemoryRouter initialEntries={['/kb/acme']}>
        <KbSidebar workspaceSlug="acme" />
      </MemoryRouter>,
    );
    await screen.findByRole('button', { name: '開発部' });

    // 切り替え先の一覧は保留にして、待っている間の表示を確かめる。
    hoisted.fetchSpaces.mockImplementationOnce(() => new Promise(() => {}));
    rerender(
      <MemoryRouter initialEntries={['/kb/beta']}>
        <KbSidebar workspaceSlug="beta" />
      </MemoryRouter>,
    );

    await waitFor(() =>
      expect(screen.queryByRole('button', { name: '開発部' })).not.toBeInTheDocument(),
    );
  });

  it('行の読み上げ名を題名で固定する', async () => {
    // 名前の決まり方を行の中身に任せておくと、次に何かを足したときに黙って変わる。
    //
    // 計算後の名前で確かめても意味が無い（実測では aria-label の有無にかかわらず
    // 「設計メモ」になる）ので、**明示しているという事実そのもの**を固定する。
    renderSidebar();
    await screen.findByText('設計メモ');

    const row = screen.getByRole('treeitem', { name: '設計メモ' });

    expect(row).toHaveAttribute('aria-label', '設計メモ');
    expect(row).toHaveAttribute('aria-level', '1');
  });

  describe('作る・名前を変える', () => {
    it('スペースの ＋ でスペース直下に作る', async () => {
      renderSidebar();
      await screen.findByText('設計メモ');

      fireEvent.click(screen.getByRole('button', { name: '開発部 にページを追加' }));

      await waitFor(() =>
        expect(hoisted.createPage).toHaveBeenCalledWith('acme', 'space-1', {
          title: '無題',
          parentId: undefined,
        }),
      );
    });

    it('行の ＋ でその行の子として作る', async () => {
      renderSidebar();
      await screen.findByText('設計メモ');

      fireEvent.click(screen.getByRole('button', { name: '設計メモ の下にページを追加' }));

      await waitFor(() =>
        expect(hoisted.createPage).toHaveBeenCalledWith('acme', 'space-1', {
          title: '無題',
          parentId: 'p1',
        }),
      );
    });

    it('作ったら木を取り直す（並び順を決めるのはサーバー）', async () => {
      renderSidebar();
      await screen.findByText('設計メモ');
      const before = hoisted.fetchPageTree.mock.calls.length;

      fireEvent.click(screen.getByRole('button', { name: '開発部 にページを追加' }));

      await waitFor(() =>
        expect(hoisted.fetchPageTree.mock.calls.length).toBeGreaterThan(before),
      );
    });

    it('作成に失敗したら知らせを出す', async () => {
      // 轍: コース削除も教材の保存も「失敗したのに成功の表示」だった。
      hoisted.createPage.mockRejectedValueOnce(new Error('boom'));
      renderSidebar();
      await screen.findByText('設計メモ');

      fireEvent.click(screen.getByRole('button', { name: '開発部 にページを追加' }));

      await waitFor(() =>
        expect(hoisted.showToast).toHaveBeenCalledWith('error', 'ページを作成できませんでした'),
      );
      expect(hoisted.showToast).not.toHaveBeenCalledWith('success', expect.anything());
    });

    it('名前を変えると、その行だけが書き換わる', async () => {
      renderSidebar();
      await screen.findByText('設計メモ');

      fireEvent.click(screen.getByRole('button', { name: '設計メモ の操作' }));
      fireEvent.click(screen.getByRole('button', { name: '名前を変更' }));

      const input = screen.getByRole('textbox', { name: 'ページの題名' });
      fireEvent.change(input, { target: { value: '新しい名前' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      expect(await screen.findByText('新しい名前')).toBeInTheDocument();
      expect(hoisted.renamePage).toHaveBeenCalledWith('acme', 'p1', '新しい名前');
      // 木ごと取り直すと一瞬空になり、開いていた段も畳まれて見える。
      expect(hoisted.fetchPageTree).toHaveBeenCalledTimes(1);
    });

    it('名前の変更に失敗したら、知らせを出して入力欄を閉じない', async () => {
      // 閉じると、書いた文字は消えるのに元の題名が残り、保存されたのか分からなくなる。
      hoisted.renamePage.mockRejectedValueOnce(new Error('boom'));
      renderSidebar();
      await screen.findByText('設計メモ');

      fireEvent.click(screen.getByRole('button', { name: '設計メモ の操作' }));
      fireEvent.click(screen.getByRole('button', { name: '名前を変更' }));

      const input = screen.getByRole('textbox', { name: 'ページの題名' });
      fireEvent.change(input, { target: { value: '新しい名前' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      await waitFor(() =>
        expect(hoisted.showToast).toHaveBeenCalledWith('error', '名前を変更できませんでした'),
      );
      expect(screen.getByRole('textbox', { name: 'ページの題名' })).toHaveValue('新しい名前');
      expect(hoisted.showToast).not.toHaveBeenCalledWith('success', expect.anything());
    });

    it('Escape で取り消すと、サーバーへ投げない', async () => {
      renderSidebar();
      await screen.findByText('設計メモ');

      fireEvent.click(screen.getByRole('button', { name: '設計メモ の操作' }));
      fireEvent.click(screen.getByRole('button', { name: '名前を変更' }));

      const input = screen.getByRole('textbox', { name: 'ページの題名' });
      fireEvent.change(input, { target: { value: '書きかけ' } });
      fireEvent.keyDown(input, { key: 'Escape' });

      expect(await screen.findByText('設計メモ')).toBeInTheDocument();
      expect(hoisted.renamePage).not.toHaveBeenCalled();
    });

    it('空の題名はサーバーへ投げない', async () => {
      // サーバーも弾くが、往復させる意味が無い。
      renderSidebar();
      await screen.findByText('設計メモ');

      fireEvent.click(screen.getByRole('button', { name: '設計メモ の操作' }));
      fireEvent.click(screen.getByRole('button', { name: '名前を変更' }));

      const input = screen.getByRole('textbox', { name: 'ページの題名' });
      fireEvent.change(input, { target: { value: '   ' } });
      fireEvent.keyDown(input, { key: 'Enter' });

      await waitFor(() => expect(screen.getByText('設計メモ')).toBeInTheDocument());
      expect(hoisted.renamePage).not.toHaveBeenCalled();
    });
  });

  describe('アーカイブ', () => {
    const openArchive = () =>
      fireEvent.click(screen.getByRole('button', { name: 'アーカイブしたページを表示' }));

    it('切り替えると、同じスペースをアーカイブ済みで取り直す', async () => {
      // 別の口ではなく同じ口のスコープ。権限の見方は現役とまったく同じ。
      renderSidebar();
      await screen.findByText('設計メモ');

      openArchive();

      await waitFor(() =>
        expect(hoisted.fetchPageTree).toHaveBeenCalledWith('acme', 'space-1', { archived: true }),
      );
    });

    it('切り替えたら前のスコープの木を捨てる', async () => {
      // 残すと、切り替えた直後だけ前のスコープの木が見える。
      hoisted.fetchPageTree.mockResolvedValueOnce(tree([{ id: 'p1', title: '現役のページ' }]));
      hoisted.fetchPageTree.mockImplementationOnce(() => new Promise(() => {}));
      renderSidebar();
      await screen.findByText('現役のページ');

      openArchive();

      await waitFor(() => expect(screen.queryByText('現役のページ')).not.toBeInTheDocument());
    });

    it('アーカイブ済みでは作る・名前を変えるを出さない', async () => {
      renderSidebar();
      await screen.findByText('設計メモ');
      openArchive();
      await screen.findByText('設計メモ');

      expect(screen.queryByRole('button', { name: '開発部 にページを追加' })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: '設計メモ の操作' })).not.toBeInTheDocument();
      // 切り替え自体は残り、押すと現役へ戻る。
      expect(screen.getByRole('button', { name: '現役のページに戻る' })).toBeInTheDocument();
    });

    it('復帰は、アーカイブの根にだけ出す', async () => {
      // 親がまだアーカイブ中の行に出すと、押せるのに必ず断られるボタンになる。
      hoisted.fetchPageTree.mockResolvedValue(
        tree([{ id: 'p1', title: 'アーカイブの根', children: ['p1-child'] }]),
      );
      renderSidebar();
      await screen.findByText('アーカイブの根');
      openArchive();
      await screen.findByText('アーカイブの根');

      fireEvent.click(screen.getByRole('button', { name: 'アーカイブの根 を開く' }));
      await screen.findByText('p1-child');

      // 子（parentArchived: true）には出ない。
      expect(screen.getAllByRole('button', { name: '復帰' })).toHaveLength(1);
    });

    it('メニューからアーカイブすると、木を取り直す', async () => {
      // 消えるのは 1 枚とは限らない（子孫ごと消える）ので、手元で抜くとずれる。
      renderSidebar();
      await screen.findByText('設計メモ');
      const before = hoisted.fetchPageTree.mock.calls.length;

      fireEvent.click(screen.getByRole('button', { name: '設計メモ の操作' }));
      fireEvent.click(screen.getByRole('button', { name: 'アーカイブ' }));

      await waitFor(() => expect(hoisted.archivePage).toHaveBeenCalledWith('acme', 'p1'));
      await waitFor(() =>
        expect(hoisted.fetchPageTree.mock.calls.length).toBeGreaterThan(before),
      );
    });

    it('操作中に切り替えても、取り直しは新しいスコープで走る', async () => {
      // アーカイブの完了は await をまたぐ。その間に切り替えられたとき、書き換えを
      // 始めた時点のスコープで取りに行くと、**古い木が新しい表示に入る**。
      let finishArchive: () => void = () => {};
      hoisted.archivePage.mockImplementationOnce(
        () =>
          new Promise<void>((resolve) => {
            finishArchive = resolve;
          }),
      );
      renderSidebar();
      await screen.findByText('設計メモ');

      fireEvent.click(screen.getByRole('button', { name: '設計メモ の操作' }));
      fireEvent.click(screen.getByRole('button', { name: 'アーカイブ' }));
      await waitFor(() => expect(hoisted.archivePage).toHaveBeenCalled());

      // まだ終わっていないうちに切り替える。
      openArchive();
      await waitFor(() =>
        expect(hoisted.fetchPageTree).toHaveBeenCalledWith('acme', 'space-1', { archived: true }),
      );
      hoisted.fetchPageTree.mockClear();

      await act(async () => {
        finishArchive();
      });

      await waitFor(() => expect(hoisted.fetchPageTree).toHaveBeenCalled());
      for (const call of hoisted.fetchPageTree.mock.calls) {
        expect(call[2]).toEqual({ archived: true });
      }
    });

    it('アーカイブに失敗したら知らせを出す', async () => {
      hoisted.archivePage.mockRejectedValueOnce(new Error('boom'));
      renderSidebar();
      await screen.findByText('設計メモ');

      fireEvent.click(screen.getByRole('button', { name: '設計メモ の操作' }));
      fireEvent.click(screen.getByRole('button', { name: 'アーカイブ' }));

      await waitFor(() =>
        expect(hoisted.showToast).toHaveBeenCalledWith('error', 'アーカイブできませんでした'),
      );
      expect(hoisted.showToast).not.toHaveBeenCalledWith('success', expect.anything());
    });

    it('復帰に失敗したら知らせを出す', async () => {
      hoisted.unarchivePage.mockRejectedValueOnce(new Error('boom'));
      renderSidebar();
      await screen.findByText('設計メモ');
      openArchive();
      await screen.findByText('設計メモ');

      fireEvent.click(screen.getByRole('button', { name: '復帰' }));

      await waitFor(() =>
        expect(hoisted.showToast).toHaveBeenCalledWith('error', '復帰できませんでした'),
      );
    });
  });

  describe('ドラッグで動かす', () => {
    const twoRoots = () =>
      hoisted.fetchPageTree.mockResolvedValue(
        tree([{ id: 'p1', title: 'ページ A' }, { id: 'p2', title: 'ページ B' }]),
      );

    it('行の下端に落とすと、その直後の兄弟として送る', async () => {
      twoRoots();
      renderSidebar();
      await screen.findByText('ページ B');

      dragRowOnto('ページ A', 'ページ B', 90);

      await waitFor(() =>
        expect(hoisted.movePage).toHaveBeenCalledWith('acme', 'p1', {
          parentId: '',
          afterPageId: 'p2',
        }),
      );
    });

    it('行の上端に落とすと、その手前の兄弟として送る', async () => {
      twoRoots();
      renderSidebar();
      await screen.findByText('ページ B');

      dragRowOnto('ページ B', 'ページ A', 10);

      await waitFor(() =>
        expect(hoisted.movePage).toHaveBeenCalledWith('acme', 'p2', {
          parentId: '',
          beforePageId: 'p1',
        }),
      );
    });

    it('行の中央に落とすと、その行の子として送る', async () => {
      twoRoots();
      renderSidebar();
      await screen.findByText('ページ B');

      dragRowOnto('ページ B', 'ページ A', 50);

      await waitFor(() =>
        expect(hoisted.movePage).toHaveBeenCalledWith('acme', 'p2', { parentId: 'p1' }),
      );
    });

    it('返事を待たずに画面が動く', async () => {
      // ドラッグは即座に動かないと使えない。
      twoRoots();
      hoisted.movePage.mockImplementationOnce(() => new Promise(() => {}));
      renderSidebar();
      await screen.findByText('ページ B');

      dragRowOnto('ページ A', 'ページ B', 90);

      await waitFor(() => {
        const titles = screen.getAllByRole('treeitem').map((row) => row.getAttribute('aria-label'));
        expect(titles).toEqual(['ページ B', 'ページ A']);
      });
    });

    it('断られたら元の並びへ戻し、知らせを出す', async () => {
      // 戻せないと、画面と DB が食い違ったまま利用者が次の操作をする。
      twoRoots();
      hoisted.movePage.mockRejectedValueOnce(new Error('boom'));
      renderSidebar();
      await screen.findByText('ページ B');

      dragRowOnto('ページ A', 'ページ B', 90);

      await waitFor(() =>
        expect(hoisted.showToast).toHaveBeenCalledWith('error', '移動できませんでした'),
      );
      const titles = screen.getAllByRole('treeitem').map((row) => row.getAttribute('aria-label'));
      expect(titles).toEqual(['ページ A', 'ページ B']);
    });

    it('巻き戻しで木を取り直さない', async () => {
      // 取り直すと失敗が見えないまま画面だけ整い、間に別の操作が挟まると
      // どちらが正か分からなくなる。
      twoRoots();
      hoisted.movePage.mockRejectedValueOnce(new Error('boom'));
      renderSidebar();
      await screen.findByText('ページ B');
      const before = hoisted.fetchPageTree.mock.calls.length;

      dragRowOnto('ページ A', 'ページ B', 90);

      await waitFor(() => expect(hoisted.showToast).toHaveBeenCalled());
      expect(hoisted.fetchPageTree.mock.calls.length).toBe(before);
    });

    it('成功しても木を取り直さない', async () => {
      // どの兄弟の隣かで指定しているので、サーバーの並びとこちらの並びは割れない。
      twoRoots();
      renderSidebar();
      await screen.findByText('ページ B');
      const before = hoisted.fetchPageTree.mock.calls.length;

      dragRowOnto('ページ A', 'ページ B', 90);

      await waitFor(() => expect(hoisted.movePage).toHaveBeenCalled());
      expect(hoisted.fetchPageTree.mock.calls.length).toBe(before);
    });

    it('自分の子孫の中へは動かさない（サーバーへも投げない）', async () => {
      hoisted.fetchPageTree.mockResolvedValue(
        tree([{ id: 'p1', title: '親ページ', children: ['p1-child'] }]),
      );
      renderSidebar();
      await screen.findByText('親ページ');
      fireEvent.click(screen.getByRole('button', { name: '親ページ を開く' }));
      await screen.findByText('p1-child');

      dragRowOnto('親ページ', 'p1-child', 50);

      await waitFor(() => expect(hoisted.showToast).toHaveBeenCalled());
      expect(hoisted.movePage).not.toHaveBeenCalled();
    });

    it('子として落としたら、その段を開いて動かしたページを見せる', async () => {
      // 畳まれた段の中に入ると、成功したのに画面から消えたように見える。
      twoRoots();
      renderSidebar();
      await screen.findByText('ページ B');

      dragRowOnto('ページ B', 'ページ A', 50);

      await waitFor(() => expect(hoisted.movePage).toHaveBeenCalled());
      expect(screen.getByRole('treeitem', { name: 'ページ B' })).toHaveAttribute('aria-level', '2');
    });

    it('移動中に重ねて落としても、2 本目は投げない', async () => {
      // 重ねると、1 本目が失敗したときの戻し先が 2 本目の結果の上になり、
      // どちらが正か決められなくなる。
      twoRoots();
      hoisted.movePage.mockImplementationOnce(() => new Promise(() => {}));
      renderSidebar();
      await screen.findByText('ページ B');

      dragRowOnto('ページ A', 'ページ B', 90);
      await waitFor(() => expect(hoisted.movePage).toHaveBeenCalledTimes(1));

      dragRowOnto('ページ B', 'ページ A', 10);

      await waitFor(() => expect(hoisted.showToast).toHaveBeenCalled());
      expect(hoisted.movePage).toHaveBeenCalledTimes(1);
    });

    it('スコープが切り替わったあとの失敗では、新しい木を巻き戻さない', async () => {
      // 自分が描いた木がもう表示されていないなら、そちらのほうが新しい。
      twoRoots();
      let failMove: (reason: Error) => void = () => {};
      hoisted.movePage.mockImplementationOnce(
        () =>
          new Promise((_, reject) => {
            failMove = reject;
          }),
      );
      renderSidebar();
      await screen.findByText('ページ B');

      dragRowOnto('ページ A', 'ページ B', 90);
      await waitFor(() => expect(hoisted.movePage).toHaveBeenCalled());

      // アーカイブ済みへ切り替える（木が丸ごと別のものになる）。
      hoisted.fetchPageTree.mockResolvedValue(tree([{ id: 'z', title: 'アーカイブの行' }]));
      fireEvent.click(screen.getByRole('button', { name: 'アーカイブしたページを表示' }));
      await screen.findByText('アーカイブの行');

      await act(async () => {
        failMove(new Error('boom'));
      });

      expect(screen.getByRole('treeitem', { name: 'アーカイブの行' })).toBeInTheDocument();
      expect(screen.queryByRole('treeitem', { name: 'ページ A' })).not.toBeInTheDocument();
    });

    it('アーカイブ済みでは並べ替えを受け付けない', async () => {
      twoRoots();
      renderSidebar();
      await screen.findByText('ページ B');
      fireEvent.click(screen.getByRole('button', { name: 'アーカイブしたページを表示' }));
      await screen.findByText('ページ B');

      dragRowOnto('ページ A', 'ページ B', 90);

      expect(hoisted.movePage).not.toHaveBeenCalled();
    });
  });

  it('ワークスペースを切り替えると URL も変わる', async () => {
    // 状態だけ変えると、再読み込みや共有で別の場所が開く。
    hoisted.fetchWorkspaces.mockResolvedValue([workspace('acme', 'Acme 社'), workspace('beta', 'Beta 社')]);
    let path = '';
    function PathProbe() {
      path = useLocation().pathname;
      return null;
    }

    render(
      <MemoryRouter initialEntries={['/kb']}>
        <PathProbe />
        <Routes>
          <Route path="/kb" element={<KbSidebar />} />
          <Route path="/kb/:slug" element={<KbSidebar />} />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(await screen.findByRole('button', { name: /Acme 社/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Beta 社' }));

    await waitFor(() => expect(path).toBe('/kb/beta'));
  });
});
