import { act, render, screen, waitFor, fireEvent, within } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import NoteSidebar from '../NoteSidebar';
import { emitNoteTreeEvent } from '@/entities/note';
import type { NotePage, NotePageTree, NoteSpace, NoteWorkspace } from '@/entities/note';

const hoisted = vi.hoisted(() => ({
  fetchWorkspaces: vi.fn(),
  fetchSpaces: vi.fn(),
  fetchPageTree: vi.fn(),
  createPage: vi.fn(),
  renamePage: vi.fn(),
  archivePage: vi.fn(),
  unarchivePage: vi.fn(),
  movePage: vi.fn(),
  createWorkspace: vi.fn(),
  createSpace: vi.fn(),
  renameSpace: vi.fn(),
  searchPages: vi.fn(),
  showToast: vi.fn(),
}));

// トーストは検査の対象。**失敗したときだけ知らせが出ること**を確かめるために捕まえる。
vi.mock('@/shared/lib/hooks/useToast', () => ({
  useToast: () => ({ showToast: hoisted.showToast, toasts: [], removeToast: vi.fn() }),
}));

vi.mock('@/entities/note', async () => {
  // ツリーの平坦化と祖先探索は本物を使う（そこは entities 側でテスト済みで、
  // ここで別実装を差し込むと「サイドバーは通るが本番は壊れる」形になる）。
  const actual = await vi.importActual<typeof import('@/entities/note')>(
    '@/entities/note',
  );
  return {
    ...actual,
    NoteRepository: {
      fetchWorkspaces: hoisted.fetchWorkspaces,
      fetchSpaces: hoisted.fetchSpaces,
      fetchPageTree: hoisted.fetchPageTree,
      createPage: hoisted.createPage,
      renamePage: hoisted.renamePage,
      archivePage: hoisted.archivePage,
      unarchivePage: hoisted.unarchivePage,
      movePage: hoisted.movePage,
      createWorkspace: hoisted.createWorkspace,
      createSpace: hoisted.createSpace,
      renameSpace: hoisted.renameSpace,
      searchPages: hoisted.searchPages,
    },
  };
});

function workspace(slug: string, name = slug): NoteWorkspace {
  return { slug, name, createdAt: '2026-08-01T00:00:00Z' };
}

function space(id: string, name = id): NoteSpace {
  return { id, key: id, name, createdAt: '2026-08-01T00:00:00Z' };
}

function page(id: string, title = id): NotePage {
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
): NotePageTree {
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
      <NoteSidebar {...props} />
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
  hoisted.createWorkspace.mockResolvedValue(workspace('new', '新しい会社'));
  hoisted.createSpace.mockResolvedValue(space('new-space', '新しい区画'));
  hoisted.renameSpace.mockImplementation(async (_slug: string, id: string, name: string) =>
    space(id, name),
  );
  hoisted.searchPages.mockResolvedValue([]);
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
function rowOf(title: string): HTMLElement {
  // 行そのものは役割を持たない（木として名乗らない）。題名のリンクから辿る。
  const link = screen.getByRole('link', { name: title });
  const row = link.parentElement;
  if (!row) throw new Error('row not found: ' + title);
  return row;
}

function dragRowOnto(fromTitle: string, toTitle: string, clientY: number) {
  const from = rowOf(fromTitle);
  const to = rowOf(toTitle);
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

describe('NoteSidebar', () => {
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

  describe('ここから始める', () => {
    it('所属が無いときは、行き止まりにせずワークスペースを作らせる', async () => {
      // 作る手段が無いと、API があってもサイドバーには永久にたどり着けない
      // （実際そうなっていた）。
      hoisted.fetchWorkspaces.mockResolvedValue([]);
      renderSidebar();
      await screen.findByText(/まだワークスペースがありません/);

      fireEvent.change(screen.getByLabelText('ワークスペースの名前'), {
        target: { value: 'Acme 社' },
      });
      fireEvent.click(screen.getByRole('button', { name: 'ワークスペースを作る' }));

      await waitFor(() =>
        // URL に出る slug はサーバーが自動採番する。フロントから送るのは名前だけ。
        expect(hoisted.createWorkspace).toHaveBeenCalledWith({ name: 'Acme 社' }),
      );
    });

    it('スペースが無いときも行き止まりにしない', async () => {
      // ワークスペースを作っただけではスペースは付いてこない。
      hoisted.fetchSpaces.mockResolvedValue([]);
      renderSidebar();
      await screen.findByText(/まだスペースがありません/);

      fireEvent.change(screen.getByLabelText('スペースの名前'), { target: { value: '開発部' } });
      fireEvent.click(screen.getByRole('button', { name: 'スペースを作る' }));

      await waitFor(() =>
        expect(hoisted.createSpace).toHaveBeenCalledWith('acme', { name: '開発部' }),
      );
    });

    it('URL に使う短い名前は人に入力させない（サーバーが自動採番する）', async () => {
      // 日本語の名前からは英数字が残らず、人に決めさせると先へ進めない。
      // 欄そのものを出さないことが仕様（うっかり復活したらこのテストが落ちる）。
      hoisted.fetchWorkspaces.mockResolvedValue([]);
      renderSidebar();
      await screen.findByText(/まだワークスペースがありません/);

      expect(screen.queryByLabelText(/短い名前/)).not.toBeInTheDocument();
    });

    it('作成に失敗したら知らせを出し、入力は消さない', async () => {
      // 消すと打ち直しになるうえ、何が悪かったのかも分からない。
      hoisted.fetchWorkspaces.mockResolvedValue([]);
      hoisted.createWorkspace.mockRejectedValueOnce(new Error('boom'));
      renderSidebar();
      await screen.findByText(/まだワークスペースがありません/);

      fireEvent.change(screen.getByLabelText('ワークスペースの名前'), {
        target: { value: 'Acme 社' },
      });
      fireEvent.click(screen.getByRole('button', { name: 'ワークスペースを作る' }));

      await waitFor(() =>
        expect(hoisted.showToast).toHaveBeenCalledWith(
          'error',
          'ワークスペースを作成できませんでした',
        ),
      );
      expect(screen.getByLabelText('ワークスペースの名前')).toHaveValue('Acme 社');
    });
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
        <NoteSidebar workspaceSlug="acme" />
      </MemoryRouter>,
    );
    await screen.findByRole('button', { name: '開発部' });

    // 切り替え先の一覧は保留にして、待っている間の表示を確かめる。
    hoisted.fetchSpaces.mockImplementationOnce(() => new Promise(() => {}));
    rerender(
      <MemoryRouter initialEntries={['/kb/beta']}>
        <NoteSidebar workspaceSlug="beta" />
      </MemoryRouter>,
    );

    await waitFor(() =>
      expect(screen.queryByRole('button', { name: '開発部' })).not.toBeInTheDocument(),
    );
  });

  it('木としては名乗らない（矢印キーを用意していないため）', async () => {
    // role="tree" を名乗ると矢印キーでの移動を約束することになるが、行の中に
    // リンクと操作ボタンが同居しているぶん、それは Tab とは別の操作体系になる。
    // 名乗りだけ残すのは嘘なので、名乗らないことを固定する。
    renderSidebar();
    await screen.findByText('設計メモ');

    // 木としては名乗らないので、段の深さは入れ子の ul が表す。
    // いま開いているページは aria-current が表す。
    expect(screen.getByRole('link', { name: '設計メモ' })).toBeInTheDocument();
    expect(screen.queryByRole('tree')).not.toBeInTheDocument();
    expect(screen.queryByRole('treeitem')).not.toBeInTheDocument();
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

  describe('キーボードだけで動かす', () => {
    const threeRoots = () =>
      hoisted.fetchPageTree.mockResolvedValue(
        tree([
          { id: 'p1', title: 'ページ A' },
          { id: 'p2', title: 'ページ B' },
          { id: 'p3', title: 'ページ C' },
        ]),
      );

    /** openMenu は行の ⋯ を押してメニューを開く（Tab で届く素のボタン）。 */
    const openMenu = (title: string) =>
      fireEvent.click(screen.getByRole('button', { name: `${title} の操作` }));

    it('上へ移動は、ひとつ上の兄弟の手前へ送る', async () => {
      threeRoots();
      renderSidebar();
      await screen.findByText('ページ B');

      openMenu('ページ B');
      fireEvent.click(screen.getByRole('button', { name: '上へ移動' }));

      await waitFor(() =>
        expect(hoisted.movePage).toHaveBeenCalledWith('acme', 'p2', {
          parentId: '',
          beforePageId: 'p1',
        }),
      );
    });

    it('下へ移動は、ひとつ下の兄弟の直後へ送る', async () => {
      threeRoots();
      renderSidebar();
      await screen.findByText('ページ B');

      openMenu('ページ B');
      fireEvent.click(screen.getByRole('button', { name: '下へ移動' }));

      await waitFor(() =>
        expect(hoisted.movePage).toHaveBeenCalledWith('acme', 'p2', {
          parentId: '',
          afterPageId: 'p3',
        }),
      );
    });

    it('ひとつ内側へは、ひとつ上の兄弟の子にする', async () => {
      threeRoots();
      renderSidebar();
      await screen.findByText('ページ B');

      openMenu('ページ B');
      fireEvent.click(screen.getByRole('button', { name: 'ひとつ内側へ' }));

      await waitFor(() =>
        expect(hoisted.movePage).toHaveBeenCalledWith('acme', 'p2', { parentId: 'p1' }),
      );
    });

    it('ひとつ外側へは、親の直後へ出す', async () => {
      hoisted.fetchPageTree.mockResolvedValue(
        tree([{ id: 'p1', title: '親ページ', children: ['p1-child'] }]),
      );
      renderSidebar();
      await screen.findByText('親ページ');
      fireEvent.click(screen.getByRole('button', { name: '親ページ を開く' }));
      await screen.findByText('p1-child');

      openMenu('p1-child');
      fireEvent.click(screen.getByRole('button', { name: 'ひとつ外側へ' }));

      await waitFor(() =>
        expect(hoisted.movePage).toHaveBeenCalledWith('acme', 'p1-child', {
          parentId: '',
          afterPageId: 'p1',
        }),
      );
    });

    it('動かせない向きは項目自体を出さない', async () => {
      // 押しても何も起きない項目を並べない。キーボードだけの人には
      // これが唯一の並べ替えの手段なので、押せるものだけを見せる。
      threeRoots();
      renderSidebar();
      await screen.findByText('ページ A');

      openMenu('ページ A');

      expect(screen.queryByRole('button', { name: '上へ移動' })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'ひとつ内側へ' })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'ひとつ外側へ' })).not.toBeInTheDocument();
      expect(screen.getByRole('button', { name: '下へ移動' })).toBeInTheDocument();
    });

    it('ドラッグと同じ経路を通るので、失敗の扱いも同じ', async () => {
      threeRoots();
      hoisted.movePage.mockRejectedValueOnce(new Error('boom'));
      renderSidebar();
      await screen.findByText('ページ B');

      openMenu('ページ B');
      fireEvent.click(screen.getByRole('button', { name: '上へ移動' }));

      await waitFor(() =>
        expect(hoisted.showToast).toHaveBeenCalledWith('error', '移動できませんでした'),
      );
      const titles = screen.getAllByRole('link').map((link) => link.textContent);
      expect(titles).toEqual(['ページ A', 'ページ B', 'ページ C']);
    });

    it('アーカイブ済みでは移動の項目を出さない', async () => {
      threeRoots();
      renderSidebar();
      await screen.findByText('ページ B');
      fireEvent.click(screen.getByRole('button', { name: 'アーカイブしたページを表示' }));
      await screen.findByText('ページ B');

      expect(screen.queryByRole('button', { name: 'ページ B の操作' })).not.toBeInTheDocument();
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
        const titles = screen.getAllByRole('link').map((link) => link.textContent);
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
      const titles = screen.getAllByRole('link').map((link) => link.textContent);
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
      // 段の深さは入れ子の ul が表す。B が A の中の一覧に入っていること。
      expect(screen.getByRole('list', { name: 'ページ A の中' })).toHaveTextContent('ページ B');
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

      expect(screen.getByRole('link', { name: 'アーカイブの行' })).toBeInTheDocument();
      expect(screen.queryByRole('link', { name: 'ページ A' })).not.toBeInTheDocument();
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

  it('ワークスペースを切り替えると /notes に戻り、木は切り替え先を指す', async () => {
    // URL はワークスペースを持たない（ページの URL は /p/{pageId} だけ）。
    // 開いていたページは前のワークスペースのものなので、本文画面には残さず一覧に戻す。
    hoisted.fetchWorkspaces.mockResolvedValue([workspace('acme', 'Acme 社'), workspace('beta', 'Beta 社')]);
    let path = '';
    function PathProbe() {
      path = useLocation().pathname;
      return null;
    }

    render(
      <MemoryRouter initialEntries={['/p/p1']}>
        <PathProbe />
        <Routes>
          <Route path="/notes" element={<NoteSidebar />} />
          <Route path="/p/:pageId" element={<NoteSidebar workspaceSlug="acme" />} />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(await screen.findByRole('button', { name: /Acme 社/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Beta 社' }));

    await waitFor(() => expect(path).toBe('/notes'));
    // 切り替え先のスペース一覧を取り直している（状態が新しいワークスペースを指す）。
    await waitFor(() => expect(hoisted.fetchSpaces).toHaveBeenCalledWith('beta'));
  });
});

describe('見た目の印（葉の点と開いたフォルダ）', () => {
  it('子が無い行は三角の位置に「・」が出て、子を持つ行には出ない', async () => {
    hoisted.fetchPageTree.mockResolvedValue(
      tree([{ id: 'parent', title: '親ページ', children: ['child-1'] }, { id: 'leaf', title: '葉ページ' }]),
    );
    renderSidebar();
    await screen.findByText('親ページ');

    const leafRow = screen.getByRole('link', { name: /葉ページ/ }).closest('div[draggable]') as HTMLElement;
    const parentRow = screen.getByRole('link', { name: /親ページ/ }).closest('div[draggable]') as HTMLElement;
    expect(leafRow.textContent).toContain('•');
    expect(parentRow.textContent).not.toContain('•');
  });

  it('子を持つ行のフォルダは、開くと開いた形のアイコンに変わる', async () => {
    hoisted.fetchPageTree.mockResolvedValue(
      tree([{ id: 'parent', title: '親ページ', children: ['child-1'] }]),
    );
    renderSidebar();
    await screen.findByText('親ページ');

    const row = () => screen.getByRole('link', { name: /親ページ/ }).closest('div[draggable]') as HTMLElement;
    expect(row().querySelector('[data-icon="page-group"]')).not.toBeNull();
    expect(row().querySelector('[data-icon="page-group-open"]')).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: '親ページ を開く' }));

    expect(row().querySelector('[data-icon="page-group-open"]')).not.toBeNull();
    expect(row().querySelector('[data-icon="page-group"]')).toBeNull();
  });
});

describe('題名で検索（モーダル）', () => {
  let currentPath = '';
  function PathProbe() {
    currentPath = useLocation().pathname;
    return null;
  }

  /** openSearch は木の描画を待ってから検索モーダルを開き、入力欄を返す。 */
  async function openSearch() {
    currentPath = '';
    render(
      <MemoryRouter initialEntries={['/notes']}>
        <PathProbe />
        <NoteSidebar />
      </MemoryRouter>,
    );
    await screen.findByText('設計メモ');
    fireEvent.click(screen.getByRole('button', { name: '検索' }));
    return screen.getByRole('combobox', { name: 'ページを題名で検索' });
  }

  it('入口を押すとモーダルが開き、入力にフォーカスが移る', async () => {
    const input = await openSearch();

    expect(screen.getByRole('dialog', { name: 'ページを検索' })).toBeInTheDocument();
    expect(input).toHaveFocus();
    // サイドバーの木はそのまま（検索が場所の面を奪わない）。
    expect(screen.getByRole('link', { name: /設計メモ/ })).toBeInTheDocument();
  });

  it('入力すると少し待ってからサーバーに問い合わせ、結果がスペースの見出し付きで出る', async () => {
    hoisted.searchPages.mockResolvedValue([
      { ...page('hit-1', 'Docker 手順'), spaceId: 'space-1' },
    ]);
    const input = await openSearch();

    fireEvent.change(input, { target: { value: 'docker' } });

    await waitFor(() => expect(hoisted.searchPages).toHaveBeenCalledWith('acme', 'docker'));
    expect(await screen.findByRole('option', { name: /Docker 手順/ })).toBeInTheDocument();
    expect(screen.getByRole('group', { name: '開発部' })).toBeInTheDocument();
  });

  it('結果を押すとそのページへ移り、モーダルが閉じる', async () => {
    hoisted.searchPages.mockResolvedValue([
      { ...page('hit-1', 'Docker 手順'), spaceId: 'space-1' },
    ]);
    const input = await openSearch();
    fireEvent.change(input, { target: { value: 'docker' } });
    const option = await screen.findByRole('option', { name: /Docker 手順/ });

    fireEvent.click(within(option).getByRole('button'));

    await waitFor(() =>
      expect(screen.queryByRole('dialog', { name: 'ページを検索' })).not.toBeInTheDocument(),
    );
    // 閉じるだけでなく、選んだページへ実際に移っている。
    expect(currentPath).toBe('/p/hit-1');
  });

  it('↑↓ で選び Enter で開ける（キーボードだけで完結する）', async () => {
    hoisted.searchPages.mockResolvedValue([
      { ...page('hit-1', '一番目'), spaceId: 'space-1' },
      { ...page('hit-2', '二番目'), spaceId: 'space-1' },
    ]);
    const input = await openSearch();
    fireEvent.change(input, { target: { value: '番目' } });
    await screen.findByRole('option', { name: /一番目/ });

    fireEvent.keyDown(input, { key: 'ArrowDown' });
    expect(screen.getByRole('option', { name: /二番目/ })).toHaveAttribute(
      'aria-selected',
      'true',
    );

    fireEvent.keyDown(input, { key: 'Enter' });
    await waitFor(() =>
      expect(screen.queryByRole('dialog', { name: 'ページを検索' })).not.toBeInTheDocument(),
    );
    expect(currentPath).toBe('/p/hit-2');
  });

  it('Escape と外側クリックで閉じる', async () => {
    const input = await openSearch();

    fireEvent.keyDown(input, { key: 'Escape' });
    expect(screen.queryByRole('dialog', { name: 'ページを検索' })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '検索' }));
    fireEvent.click(screen.getByTestId('note-search-overlay'));
    expect(screen.queryByRole('dialog', { name: 'ページを検索' })).not.toBeInTheDocument();
  });

  it('入力を消した後に届いた古い応答は表示しない（空の入力に結果が湧かない）', async () => {
    let resolveSlow: (pages: NotePage[]) => void = () => {};
    hoisted.searchPages.mockImplementationOnce(
      () => new Promise<NotePage[]>((resolve) => { resolveSlow = resolve; }),
    );
    const input = await openSearch();

    fireEvent.change(input, { target: { value: 'docker' } });
    await waitFor(() => expect(hoisted.searchPages).toHaveBeenCalledTimes(1));
    // 応答が返る前に入力を消す。
    fireEvent.change(input, { target: { value: '' } });

    await act(async () => {
      resolveSlow([{ ...page('late-hit', '遅れて届いた結果'), spaceId: 'space-1' }]);
    });
    expect(screen.queryByRole('option', { name: /遅れて届いた結果/ })).not.toBeInTheDocument();
  });

  it('0 件なら「一致するページがありません」と伝える', async () => {
    hoisted.searchPages.mockResolvedValue([]);
    const input = await openSearch();

    fireEvent.change(input, { target: { value: '存在しない題名' } });

    expect(await screen.findByText('一致するページがありません')).toBeInTheDocument();
  });

  it('古い検索の応答が、後から届いても新しい結果を上書きしない', async () => {
    // 1 回目の応答を保留し、2 回目が確定した後で解決する（遅い応答が追い越される形）。
    let resolveOld: (pages: NotePage[]) => void = () => {};
    hoisted.searchPages.mockImplementationOnce(
      () => new Promise<NotePage[]>((resolve) => { resolveOld = resolve; }),
    );
    hoisted.searchPages.mockResolvedValueOnce([
      { ...page('new-hit', '新しい結果'), spaceId: 'space-1' },
    ]);
    const input = await openSearch();

    fireEvent.change(input, { target: { value: 'ふるい' } });
    await waitFor(() => expect(hoisted.searchPages).toHaveBeenCalledTimes(1));
    fireEvent.change(input, { target: { value: '新しい' } });
    expect(await screen.findByRole('option', { name: /新しい結果/ })).toBeInTheDocument();

    // ここで 1 回目（古い方）が届く。世代番号で捨てられ、画面は新しい結果のまま。
    await act(async () => {
      resolveOld([{ ...page('old-hit', '古い結果'), spaceId: 'space-1' }]);
    });
    expect(screen.queryByRole('option', { name: /古い結果/ })).not.toBeInTheDocument();
    expect(screen.getByRole('option', { name: /新しい結果/ })).toBeInTheDocument();
  });

  it('検索に失敗したら再試行の導線を出し、押すともう一度問い合わせる', async () => {
    hoisted.searchPages.mockRejectedValueOnce(new Error('down'));
    hoisted.searchPages.mockResolvedValueOnce([
      { ...page('hit-1', 'Docker 手順'), spaceId: 'space-1' },
    ]);
    const input = await openSearch();

    fireEvent.change(input, { target: { value: 'docker' } });
    fireEvent.click(await screen.findByRole('button', { name: '再試行' }));

    expect(await screen.findByRole('option', { name: /Docker 手順/ })).toBeInTheDocument();
    expect(hoisted.searchPages).toHaveBeenCalledTimes(2);
  });
});

describe('スペースの見出しの操作', () => {
  it('⋯ の「スペースの名前を変更」で見出しが書き換わる', async () => {
    renderSidebar();
    await screen.findByText('設計メモ');

    fireEvent.click(screen.getByRole('button', { name: '開発部 の操作' }));
    fireEvent.click(screen.getByRole('button', { name: 'スペースの名前を変更' }));

    const input = screen.getByRole('textbox', { name: 'スペースの名前' });
    fireEvent.change(input, { target: { value: '技術部' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() => expect(hoisted.renameSpace).toHaveBeenCalledWith('acme', 'space-1', '技術部'));
    expect(await screen.findByText('技術部')).toBeInTheDocument();
    expect(hoisted.showToast).not.toHaveBeenCalled();
  });

  it('変更に失敗したら知らせが出て、見出しは元のまま', async () => {
    hoisted.renameSpace.mockRejectedValue(new Error('forbidden'));
    renderSidebar();
    await screen.findByText('設計メモ');

    fireEvent.click(screen.getByRole('button', { name: '開発部 の操作' }));
    fireEvent.click(screen.getByRole('button', { name: 'スペースの名前を変更' }));
    const input = screen.getByRole('textbox', { name: 'スペースの名前' });
    fireEvent.change(input, { target: { value: '技術部' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() =>
      expect(hoisted.showToast).toHaveBeenCalledWith('error', 'スペースの名前を変更できませんでした'),
    );
    // 入力欄は開いたまま・書いた文字も残る（閉じると、保存されたのか分からなくなる）。
    // ページの改名と同じ設計。
    const stillOpen = screen.getByRole('textbox', { name: 'スペースの名前' });
    expect(stillOpen).toHaveValue('技術部');
  });

  it('見出しの ＋ と ⋯ はホバーしなくても見えている（行の操作はホバーで現れる）', async () => {
    renderSidebar();
    await screen.findByText('設計メモ');

    const plus = screen.getByRole('button', { name: '開発部 にページを追加' });
    const menu = screen.getByRole('button', { name: '開発部 の操作' });
    // 行の操作（opacity-0 で隠れる）と違い、見出しの操作は常時表示のクラス構成。
    expect(plus.className).not.toContain('opacity-0');
    expect(menu.className).not.toContain('opacity-0');
  });

  it('行の ＋ には何が起きるかのツールチップが付いている', async () => {
    renderSidebar();
    await screen.findByText('設計メモ');

    const rowPlus = screen.getByRole('button', { name: '設計メモ の下にページを追加' });
    expect(rowPlus).toHaveAttribute('title', '中にページを作成');
  });
});

describe('スペースを追加する入口', () => {
  it('スペースが既にあっても「スペースを追加」から作れる', async () => {
    renderSidebar();
    await screen.findByRole('button', { name: '開発部' });

    fireEvent.click(screen.getByRole('button', { name: 'スペースを追加' }));
    hoisted.createSpace.mockResolvedValue(space('space-2', '営業部'));

    fireEvent.change(screen.getByLabelText('スペースの名前'), { target: { value: '営業部' } });
    fireEvent.click(screen.getByRole('button', { name: 'スペースを作る' }));

    await waitFor(() =>
      expect(hoisted.createSpace).toHaveBeenCalledWith('acme', { name: '営業部' }),
    );
    // 成功したらフォームは畳まれ、入口のボタンに戻る。
    await waitFor(() =>
      expect(screen.queryByLabelText('スペースの名前')).not.toBeInTheDocument(),
    );
    expect(screen.getByRole('button', { name: 'スペースを追加' })).toBeInTheDocument();
  });

  it('やめるでフォームを畳める（作らない）', async () => {
    renderSidebar();
    await screen.findByRole('button', { name: '開発部' });

    fireEvent.click(screen.getByRole('button', { name: 'スペースを追加' }));
    expect(screen.getByLabelText('スペースの名前')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'やめる' }));

    expect(screen.queryByLabelText('スペースの名前')).not.toBeInTheDocument();
    expect(hoisted.createSpace).not.toHaveBeenCalled();
  });

  it('失敗したら知らせを出し、入力は消さない', async () => {
    hoisted.createSpace.mockRejectedValue(new Error('forbidden'));
    renderSidebar();
    await screen.findByRole('button', { name: '開発部' });

    fireEvent.click(screen.getByRole('button', { name: 'スペースを追加' }));
    fireEvent.change(screen.getByLabelText('スペースの名前'), { target: { value: '営業部' } });
    fireEvent.click(screen.getByRole('button', { name: 'スペースを作る' }));

    await waitFor(() =>
      expect(hoisted.showToast).toHaveBeenCalledWith('error', 'スペースを作成できませんでした'),
    );
    expect(screen.getByLabelText('スペースの名前')).toHaveValue('営業部');
  });

  it('アーカイブ表示の間は入口を出さない（現役の木の操作なので）', async () => {
    renderSidebar();
    await screen.findByRole('button', { name: '開発部' });

    fireEvent.click(screen.getByRole('button', { name: 'アーカイブしたページを表示' }));

    await waitFor(() =>
      expect(screen.queryByRole('button', { name: 'スペースを追加' })).not.toBeInTheDocument(),
    );
  });
});

describe('ページ画面からの通知に木が追従する', () => {
  it('page-created で親を開き、そのスペースの木を取り直す', async () => {
    renderSidebar();
    // 先頭のスペースは自動で開き、木が読み込まれている。
    await screen.findByText('設計メモ');
    hoisted.fetchPageTree.mockClear();
    hoisted.fetchPageTree.mockResolvedValue(
      tree([{ id: 'p1', title: '設計メモ', children: ['p1-child'] }]),
    );

    act(() => {
      emitNoteTreeEvent({
        type: 'page-created',
        page: { ...page('p1-child', '無題'), parentId: 'p1' },
      });
    });

    await waitFor(() => expect(hoisted.fetchPageTree).toHaveBeenCalledWith(
      'acme',
      'space-1',
      { archived: false },
    ));
    // 取り直した木で子が見えている（親は自動で開いている）。
    expect(await screen.findByText('p1-child')).toBeInTheDocument();
  });

  it('page-renamed が未読込のスペース宛でも壊れない（何も起きない）', async () => {
    hoisted.fetchSpaces.mockResolvedValue([space('space-1', '開発部'), space('space-2', '営業部')]);
    renderSidebar();
    await screen.findByText('設計メモ');

    // space-2 は開いておらず木が無い。宛先の state が無くても落ちず、取りにも行かない。
    act(() => {
      emitNoteTreeEvent({
        type: 'page-renamed',
        page: { ...page('px', 'よそのページ'), spaceId: 'space-2' },
      });
    });

    expect(screen.queryByText('よそのページ')).not.toBeInTheDocument();
    expect(hoisted.fetchPageTree).toHaveBeenCalledTimes(1);
  });

  it('page-renamed で木の題名が差し替わる', async () => {
    renderSidebar();
    await screen.findByText('設計メモ');

    act(() => {
      emitNoteTreeEvent({ type: 'page-renamed', page: page('p1', '設計メモ v2') });
    });

    expect(await screen.findByText('設計メモ v2')).toBeInTheDocument();
    expect(screen.queryByText('設計メモ')).not.toBeInTheDocument();
  });
});

describe('ワークスペース切替ポップアップ', () => {
  let popPath = '';
  function PopPathProbe() {
    popPath = useLocation().pathname;
    return null;
  }

  it('所属が 1 つでも開け、追加の入口から名前だけで作れる', async () => {
    popPath = '';
    render(
      <MemoryRouter initialEntries={['/p/p1']}>
        <PopPathProbe />
        <NoteSidebar workspaceSlug="acme" activePageId="p1" />
      </MemoryRouter>,
    );
    await screen.findByText('設計メモ');

    // 1 件でも見出しではなくボタン（ポップアップに追加の入口があるため）。
    fireEvent.click(screen.getByRole('button', { name: /Acme 社/ }));
    hoisted.createWorkspace.mockResolvedValue(workspace('w-new', '新チーム'));

    fireEvent.click(screen.getByRole('button', { name: 'ワークスペースを追加' }));
    fireEvent.change(screen.getByLabelText('ワークスペースの名前'), {
      target: { value: '新チーム' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ワークスペースを作る' }));

    await waitFor(() =>
      expect(hoisted.createWorkspace).toHaveBeenCalledWith({ name: '新チーム' }),
    );
    // 成功したらポップアップは閉じ、一覧（/notes）へ戻る — 開いていた旧ワークスペースの
    // ページと、新ワークスペースを指すサイドバーが食い違ったまま残らないように。
    await waitFor(() =>
      expect(screen.queryByLabelText('ワークスペースの名前')).not.toBeInTheDocument(),
    );
    await waitFor(() => expect(popPath).toBe('/notes'));
  });

  it('日本語入力の変換キャンセルの Escape ではポップアップを閉じない（打ちかけの名前を守る）', async () => {
    renderSidebar();
    await screen.findByText('設計メモ');
    fireEvent.click(screen.getByRole('button', { name: /Acme 社/ }));
    fireEvent.click(screen.getByRole('button', { name: 'ワークスペースを追加' }));
    const input = screen.getByLabelText('ワークスペースの名前');
    fireEvent.change(input, { target: { value: '開発ちー' } });

    // 変換中の Escape（isComposing=true）は document のリスナーに届いても無視される。
    fireEvent.keyDown(input, { key: 'Escape', isComposing: true });
    expect(screen.getByLabelText('ワークスペースの名前')).toHaveValue('開発ちー');

    // 変換していない Escape では従来どおり閉じる。
    fireEvent.keyDown(document, { key: 'Escape' });
    await waitFor(() =>
      expect(screen.queryByLabelText('ワークスペースの名前')).not.toBeInTheDocument(),
    );
  });

  it('失敗したら知らせを出し、入力は消さない', async () => {
    hoisted.createWorkspace.mockRejectedValue(new Error('boom'));
    renderSidebar();
    await screen.findByText('設計メモ');

    fireEvent.click(screen.getByRole('button', { name: /Acme 社/ }));
    fireEvent.click(screen.getByRole('button', { name: 'ワークスペースを追加' }));
    fireEvent.change(screen.getByLabelText('ワークスペースの名前'), {
      target: { value: '新チーム' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ワークスペースを作る' }));

    await waitFor(() =>
      expect(hoisted.showToast).toHaveBeenCalledWith('error', 'ワークスペースを作成できませんでした'),
    );
    expect(screen.getByLabelText('ワークスペースの名前')).toHaveValue('新チーム');
  });

  it('スペース一覧に「チームスペース」の節見出しが付く', async () => {
    renderSidebar();
    await screen.findByText('設計メモ');

    expect(screen.getByText('チームスペース')).toBeInTheDocument();
  });
});
