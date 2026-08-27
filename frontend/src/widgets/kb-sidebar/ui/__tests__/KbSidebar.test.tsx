import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import KbSidebar from '../KbSidebar';
import type { KbPage, KbPageTree, KbSpace, KbWorkspace } from '@/entities/knowledge-base';

const hoisted = vi.hoisted(() => ({
  fetchWorkspaces: vi.fn(),
  fetchSpaces: vi.fn(),
  fetchPageTree: vi.fn(),
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
    position: 'a0',
    title,
    createdByUserId: 1,
    createdAt: '2026-08-01T00:00:00Z',
    updatedAt: '2026-08-01T00:00:00Z',
  };
}

function tree(
  nodes: { id: string; title?: string; hidden?: boolean; children?: string[] }[],
  hiddenAtRoot = false,
): KbPageTree {
  return {
    pages: nodes.map((n) => ({
      page: page(n.id, n.title),
      hasHiddenChildren: n.hidden ?? false,
      children: (n.children ?? []).map((childId) => ({
        page: page(childId),
        hasHiddenChildren: false,
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
});

describe('KbSidebar', () => {
  it('先頭のスペースを開いて木を出す', async () => {
    renderSidebar();

    expect(await screen.findByText('設計メモ')).toBeInTheDocument();
    expect(hoisted.fetchPageTree).toHaveBeenCalledWith('acme', 'space-1');
  });

  it('開いていないスペースの木は取りに行かない', async () => {
    hoisted.fetchSpaces.mockResolvedValue([space('space-1', '開発部'), space('space-2', '営業部')]);

    renderSidebar();
    await screen.findByText('設計メモ');

    // スペースの数だけ要求が飛ぶ（N+1）と、開いていない分が丸ごと無駄になる。
    expect(hoisted.fetchPageTree).toHaveBeenCalledTimes(1);
    expect(hoisted.fetchPageTree).not.toHaveBeenCalledWith('acme', 'space-2');
  });

  it('スペースを開いたときに初めて取りに行く', async () => {
    hoisted.fetchSpaces.mockResolvedValue([space('space-1', '開発部'), space('space-2', '営業部')]);
    renderSidebar();
    await screen.findByText('設計メモ');

    fireEvent.click(screen.getByRole('button', { name: '営業部' }));

    await waitFor(() => expect(hoisted.fetchPageTree).toHaveBeenCalledWith('acme', 'space-2'));
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
