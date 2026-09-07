import { render, screen, waitFor, fireEvent, act, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { AxiosError, AxiosHeaders } from 'axios';
import KbPage from '../KbPage';
import { emitKbTreeEvent } from '@/entities/kb';
import type { CommentAnchor, CommentBadgeCounts, EditorCommand } from '@/shared/ui/RichTextEditor';

function blockIdConflictError(): AxiosError {
  return new AxiosError('Conflict', 'ERR_BAD_REQUEST', undefined, undefined, {
    status: 409,
    statusText: 'Conflict',
    headers: {},
    config: { headers: new AxiosHeaders() },
    data: { error: 'block_id_conflict' },
  });
}

const hoisted = vi.hoisted(() => ({
  resolvePage: vi.fn(),
  replaceContent: vi.fn(),
  renamePage: vi.fn(),
  setPageIcon: vi.fn(),
  clearPageIcon: vi.fn(),
  uploadPageImage: vi.fn(),
  issuePageImageDownloadURL: vi.fn(),
  setPageCover: vi.fn(),
  clearPageCover: vi.fn(),
  createPage: vi.fn(),
  listPageGrants: vi.fn(),
  listGrantablePrincipals: vi.fn(),
  listCommentThreads: vi.fn(),
  createCommentThread: vi.fn(),
  addComment: vi.fn(),
  resolveCommentThread: vi.fn(),
  reopenCommentThread: vi.fn(),
  listPageVersions: vi.fn(),
  getPageVersion: vi.fn(),
  createPageVersion: vi.fn(),
  restorePageVersion: vi.fn(),
  fetchWorkspaces: vi.fn(),
  fetchSpaces: vi.fn(),
  fetchPageTree: vi.fn(),
  getLastVisitedPageId: vi.fn(),
  emit: vi.fn(),
  showToast: vi.fn(),
  navigate: vi.fn(),
  useParams: vi.fn(() => ({ pageId: 'p1' }) as { pageId?: string }),
  editorProps: {
    current: null as null | {
      value?: { type: 'doc'; content: unknown[] };
      editable?: boolean;
      extraSlashCommands?: EditorCommand[];
      onImageUpload?: (file: File) => Promise<string>;
      resolveImageSrc?: (src: string) => Promise<string>;
      onRequestComment?: (anchor: CommentAnchor) => void;
      canComment?: boolean;
      commentBadgeCounts?: CommentBadgeCounts;
      onCommentBadgeClick?: (blockId: string) => void;
    },
  },
}));

vi.mock('@/entities/kb', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/entities/kb')>();
  return {
    ...actual,
    KbRepository: {
      resolvePage: hoisted.resolvePage,
      replaceContent: hoisted.replaceContent,
      renamePage: hoisted.renamePage,
      setPageIcon: hoisted.setPageIcon,
      clearPageIcon: hoisted.clearPageIcon,
      uploadPageImage: hoisted.uploadPageImage,
      issuePageImageDownloadURL: hoisted.issuePageImageDownloadURL,
      setPageCover: hoisted.setPageCover,
      clearPageCover: hoisted.clearPageCover,
      createPage: hoisted.createPage,
      listPageGrants: hoisted.listPageGrants,
      listGrantablePrincipals: hoisted.listGrantablePrincipals,
      listCommentThreads: hoisted.listCommentThreads,
      createCommentThread: hoisted.createCommentThread,
      addComment: hoisted.addComment,
      resolveCommentThread: hoisted.resolveCommentThread,
      reopenCommentThread: hoisted.reopenCommentThread,
      listPageVersions: hoisted.listPageVersions,
      getPageVersion: hoisted.getPageVersion,
      createPageVersion: hoisted.createPageVersion,
      restorePageVersion: hoisted.restorePageVersion,
      fetchWorkspaces: hoisted.fetchWorkspaces,
      fetchSpaces: hoisted.fetchSpaces,
      fetchPageTree: hoisted.fetchPageTree,
    },
    getLastVisitedPageId: hoisted.getLastVisitedPageId,
    // スパイしつつ実物へ転送する（購読側の配線もこのテストの検査対象のため）。
    emitKbTreeEvent: (event: Parameters<typeof actual.emitKbTreeEvent>[0]) => {
      hoisted.emit(event);
      actual.emitKbTreeEvent(event);
    },
  };
});

vi.mock('@/shared/lib/hooks/useToast', () => ({
  useToast: () => ({ showToast: hoisted.showToast, toasts: [], removeToast: vi.fn() }),
}));

vi.mock('react-router-dom', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-router-dom')>();
  return { ...actual, useNavigate: () => hoisted.navigate, useParams: () => hoisted.useParams() };
});

// サイドバーは自前のテストで検証済み。ここでは画面の配線だけを見る。
vi.mock('@/widgets/kb-sidebar', () => ({
  KbSidebar: () => <nav aria-label="サイドバーの偽物" />,
}));

// エディタは重い（tiptap 実体）ので、渡された props を捕まえる薄い偽物に差し替える。
// /page の run は本物の createSubpage を通る（そこが配線の検査対象）。
vi.mock('@/shared/ui/RichTextEditor', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/shared/ui/RichTextEditor')>();
  return {
    ...actual,
    RichTextEditor: (props: {
      value?: { type: 'doc'; content: unknown[] };
      editable?: boolean;
      extraSlashCommands?: EditorCommand[];
      onImageUpload?: (file: File) => Promise<string>;
      resolveImageSrc?: (src: string) => Promise<string>;
      onChange?: (doc: { type: 'doc'; content: unknown[] }) => void;
      onRequestComment?: (anchor: CommentAnchor) => void;
      canComment?: boolean;
      commentBadgeCounts?: CommentBadgeCounts;
      onCommentBadgeClick?: (blockId: string) => void;
    }) => {
      hoisted.editorProps.current = props;
      return <div data-testid="editor" />;
    },
  };
});

const resolved = (canEdit: boolean, canManage = false, canComment = true) => ({
  workspaceSlug: 'w-3f2a9c',
  workspaceName: '開発チーム',
  ancestors: [{ id: 'anc-1', title: '親ページの親' }],
  page: {
    id: 'p1',
    spaceId: 's1',
    title: '親ページ',
    createdByUserId: 1,
    createdAt: '2026-08-01T00:00:00Z',
    updatedAt: '2026-08-01T00:00:00Z',
  },
  doc: { type: 'doc', content: [] },
  canEdit,
  canManage,
  canComment,
});

/** /page の run に渡す最小のエディタ（createSubpage が使う形だけ）。 */
function fakeEditor() {
  return {
    chain: () => ({ focus: () => ({ insertContent: () => ({ run: () => {} }) }) }),
  } as never;
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={['/kb/p1']}>
      <KbPage />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  hoisted.editorProps.current = null;
  hoisted.useParams.mockReturnValue({ pageId: 'p1' });
  hoisted.getLastVisitedPageId.mockReturnValue(null);
  hoisted.resolvePage.mockResolvedValue(resolved(true));
  hoisted.listPageGrants.mockResolvedValue([]);
  hoisted.listGrantablePrincipals.mockResolvedValue([]);
  hoisted.listCommentThreads.mockResolvedValue([]);
  hoisted.listPageVersions.mockResolvedValue([]);
});

describe('KbPage の配線', () => {
  it('題名の改名に失敗したら知らせを出し、入力は消えない', async () => {
    hoisted.renamePage.mockRejectedValue(new Error('403'));
    renderPage();
    const input = await screen.findByRole('textbox', { name: 'ページの題名' });

    fireEvent.change(input, { target: { value: '新しい題名' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() =>
      expect(hoisted.showToast).toHaveBeenCalledWith('error', '題名を変更できませんでした'),
    );
    // 再 throw が KbPageTitle まで届いている＝入力が保たれている。
    expect(input).toHaveValue('新しい題名');
  });

  it('/page で子を作って開く。失敗したら知らせを出し、遷移しない', async () => {
    renderPage();
    await screen.findByTestId('editor');
    const commands = hoisted.editorProps.current?.extraSlashCommands;
    expect(commands?.map((c) => c.id)).toEqual(['page']);

    // 成功: 作ったページへ遷移。
    hoisted.createPage.mockResolvedValue({
      id: 'child-1',
      spaceId: 's1',
      parentId: 'p1',
      title: '無題',
      createdByUserId: 1,
      createdAt: '2026-08-28T00:00:00Z',
      updatedAt: '2026-08-28T00:00:00Z',
    });
    await act(async () => {
      commands![0].run(fakeEditor());
    });
    await waitFor(() => expect(hoisted.navigate).toHaveBeenCalledWith('/kb/child-1'));

    // 失敗: 知らせを出し、遷移しない。
    hoisted.navigate.mockClear();
    hoisted.createPage.mockRejectedValue(new Error('403'));
    await act(async () => {
      commands![0].run(fakeEditor());
    });
    await waitFor(() =>
      expect(hoisted.showToast).toHaveBeenCalledWith('error', '子ページを作成できませんでした'),
    );
    expect(hoisted.navigate).not.toHaveBeenCalled();
  });

  it('パンくずにワークスペース名・閲覧できる祖先・現在のページが並ぶ', async () => {
    renderPage();
    const nav = await screen.findByRole('navigation', { name: 'ページの場所' });

    expect(within(nav).getByText('開発チーム')).toBeInTheDocument();
    // 祖先はリンク（押すとそのページへ）。現在のページはリンクにしない。
    expect(within(nav).getByRole('link', { name: '親ページの親' })).toHaveAttribute(
      'href',
      '/kb/anc-1',
    );
    expect(within(nav).getByText('親ページ')).toBeInTheDocument();
    expect(within(nav).queryByRole('link', { name: '親ページ' })).not.toBeInTheDocument();
  });

  it('ancestors の無い旧応答でも画面は落ちない（デプロイ順の防御）', async () => {
    const legacy = { ...resolved(true) } as Record<string, unknown>;
    delete legacy.ancestors;
    delete legacy.workspaceName;
    hoisted.resolvePage.mockResolvedValue(legacy);
    renderPage();

    const nav = await screen.findByRole('navigation', { name: 'ページの場所' });
    // ワークスペース名が無ければ slug で代用する。
    expect(within(nav).getByText('w-3f2a9c')).toBeInTheDocument();
  });

  it('祖先が空でもパンくずは壊れない（根ページ・穴だけの経路）', async () => {
    hoisted.resolvePage.mockResolvedValue({ ...resolved(true), ancestors: [] });
    renderPage();
    const nav = await screen.findByRole('navigation', { name: 'ページの場所' });

    expect(within(nav).getByText('開発チーム')).toBeInTheDocument();
    expect(within(nav).queryByRole('link')).not.toBeInTheDocument();
  });

  it('自分か祖先が削除されたら一覧へ戻る（サーバー応答の祖先で判定する）', async () => {
    renderPage();
    await screen.findByRole('navigation', { name: 'ページの場所' });

    // 祖先（anc-1）が消えたら CASCADE で自分も消えている — 一覧へ戻る。
    // 祖先はサーバー応答から取るので、サイドバーの現役の木に載っていない
    // （アーカイブ済みの）ページを開いていても判定できる。
    //
    // 購読が張られるより先に emit すると、イベントは誰にも届かないまま捨てられる
    // （replay を持たない pub-sub のため）。届くまで emit を繰り返して、購読が
    // 生きていることをここで確定させる。これを先にやらないと、下の「無関係なら
    // 動かない」が「イベントが届いていないだけ」で通る空検証になる。
    await waitFor(() => {
      emitKbTreeEvent({ type: 'page-deleted', pageId: 'anc-1' });
      expect(hoisted.navigate).toHaveBeenCalledWith('/kb');
    });

    // 購読が生きていると分かったうえで、無関係なページの削除では動かないことを見る。
    hoisted.navigate.mockClear();
    act(() => {
      emitKbTreeEvent({ type: 'page-deleted', pageId: 'unrelated' });
    });
    expect(hoisted.navigate).not.toHaveBeenCalled();

    // 自分自身の削除でも戻る。
    act(() => {
      emitKbTreeEvent({ type: 'page-deleted', pageId: 'p1' });
    });
    await waitFor(() => expect(hoisted.navigate).toHaveBeenCalledWith('/kb'));
  });

  it('開いているワークスペースが削除されたら一覧へ戻る（配下ごと消えるため）', async () => {
    renderPage();
    await screen.findByRole('navigation', { name: 'ページの場所' });

    // resolved() の workspaceSlug と一致する削除では戻る。
    // 上のテストと同じ理由で、届くまで emit を繰り返して購読を確定させてから
    // 「無関係なら動かない」を見る（順序を逆にすると空検証になる）。
    await waitFor(() => {
      emitKbTreeEvent({ type: 'workspace-deleted', workspaceSlug: 'w-3f2a9c' });
      expect(hoisted.navigate).toHaveBeenCalledWith('/kb');
    });

    hoisted.navigate.mockClear();
    act(() => {
      emitKbTreeEvent({ type: 'workspace-deleted', workspaceSlug: 'unrelated' });
    });
    expect(hoisted.navigate).not.toHaveBeenCalled();
  });

  it('編集できないページでは /page を渡さない（読むだけの人にメニューを見せない）', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(false));
    renderPage();
    await screen.findByTestId('editor');

    expect(hoisted.editorProps.current?.extraSlashCommands).toBeUndefined();
    // 題名も入力欄ではなく見出しで出る。
    expect(screen.getByRole('heading', { name: '親ページ' })).toBeInTheDocument();
  });
});

describe('KbPage のアイコン・最終編集', () => {
  it('アイコンを付ける: 一覧から選ぶと保存され、頭部の絵文字に変わる', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(true));
    hoisted.setPageIcon.mockResolvedValue({
      ...resolved(true).page,
      icon: { type: 'emoji', value: '📘' },
    });
    renderPage();

    fireEvent.click(await screen.findByRole('button', { name: 'アイコンを追加' }));
    const dialog = await screen.findByRole('dialog', { name: 'ページのアイコンを選ぶ' });
    fireEvent.click(within(dialog).getByRole('button', { name: 'アイコンを 📘 にする' }));

    await waitFor(() =>
      expect(hoisted.setPageIcon).toHaveBeenCalledWith('w-3f2a9c', 'p1', {
        type: 'emoji',
        value: '📘',
      }),
    );
    // 保存が成功すると、頭部の行が絵文字のボタンに差し替わる。
    expect(await screen.findByRole('button', { name: 'ページのアイコンを変更' })).toBeInTheDocument();
    expect(document.querySelector('[data-icon="emoji"]')).toHaveTextContent('📘');
  });

  it('アイコンを外す', async () => {
    hoisted.resolvePage.mockResolvedValue({
      ...resolved(true),
      page: { ...resolved(true).page, icon: { type: 'emoji', value: '📘' } },
    });
    hoisted.clearPageIcon.mockResolvedValue({ ...resolved(true).page, icon: null });
    renderPage();

    fireEvent.click(await screen.findByRole('button', { name: 'ページのアイコンを変更' }));
    const dialog = await screen.findByRole('dialog', { name: 'ページのアイコンを選ぶ' });
    fireEvent.click(within(dialog).getByRole('button', { name: 'アイコンを外す' }));

    await waitFor(() => expect(hoisted.clearPageIcon).toHaveBeenCalledWith('w-3f2a9c', 'p1'));
    expect(await screen.findByRole('button', { name: 'アイコンを追加' })).toBeInTheDocument();
  });

  it('アイコンの変更に失敗したら知らせを出し、ピッカーは開いたまま', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(true));
    hoisted.setPageIcon.mockRejectedValue(new Error('invalid_icon'));
    renderPage();

    fireEvent.click(await screen.findByRole('button', { name: 'アイコンを追加' }));
    const dialog = await screen.findByRole('dialog', { name: 'ページのアイコンを選ぶ' });
    fireEvent.click(within(dialog).getByRole('button', { name: 'アイコンを 📘 にする' }));

    await waitFor(() =>
      expect(hoisted.showToast).toHaveBeenCalledWith('error', 'アイコンを変更できませんでした'),
    );
    // 失敗時は閉じない — 何が悪かったのか分からないまま消えるのを避ける。
    expect(screen.getByRole('dialog', { name: 'ページのアイコンを選ぶ' })).toBeInTheDocument();
  });

  it('最終編集が名前・日時つきで出る（名前が引けなければ不明なユーザー）', async () => {
    hoisted.resolvePage.mockResolvedValue({
      ...resolved(true),
      lastEditedBy: { userId: 1, name: '' },
      lastEditedAt: '2026-09-06T10:30:00',
    });
    renderPage();

    expect(await screen.findByText(/最終編集 不明なユーザー/)).toBeInTheDocument();
  });

  it('読むだけの人にはアイコンは img として出て、押せない', async () => {
    hoisted.resolvePage.mockResolvedValue({
      ...resolved(false),
      page: { ...resolved(false).page, icon: { type: 'emoji', value: '📘' } },
    });
    renderPage();

    const img = await screen.findByRole('img', { name: 'ページのアイコン' });
    expect(img).toHaveTextContent('📘');
    expect(screen.queryByRole('button', { name: 'アイコンを追加' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ページのアイコンを変更' })).not.toBeInTheDocument();
  });
});

describe('KbPage の画像配線（本文の画像アップロード・遅延解決）', () => {
  it('onImageUpload が配線され、KbRepository.uploadPageImage を呼ぶ（編集できるとき）', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(true));
    hoisted.uploadPageImage.mockResolvedValue('kb/w-3f2a9c/p1/1.bin');
    renderPage();
    await screen.findByTestId('editor');

    const onImageUpload = hoisted.editorProps.current?.onImageUpload;
    expect(onImageUpload).toBeTypeOf('function');

    const file = new File(['x'], 'a.png', { type: 'image/png' });
    const key = await onImageUpload!(file);

    expect(hoisted.uploadPageImage).toHaveBeenCalledWith('w-3f2a9c', 'p1', file);
    // durable な保存形式は key であって公開 URL ではない。
    expect(key).toBe('kb/w-3f2a9c/p1/1.bin');
  });

  it('編集できないページには onImageUpload を渡さない', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(false));
    renderPage();
    await screen.findByTestId('editor');

    expect(hoisted.editorProps.current?.onImageUpload).toBeUndefined();
  });

  it('resolveImageSrc が配線され、"kb/" の src を KbRepository.issuePageImageDownloadURL で解決する', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(true));
    hoisted.issuePageImageDownloadURL.mockResolvedValue({
      url: 'https://s3.example.com/get?sig=1',
      expiresIn: 600,
    });
    renderPage();
    await screen.findByTestId('editor');

    const resolveImageSrc = hoisted.editorProps.current?.resolveImageSrc;
    expect(resolveImageSrc).toBeTypeOf('function');

    const url = await resolveImageSrc!('kb/w-3f2a9c/p1/1.bin');

    expect(hoisted.issuePageImageDownloadURL).toHaveBeenCalledWith(
      'w-3f2a9c',
      'p1',
      'kb/w-3f2a9c/p1/1.bin',
    );
    expect(url).toBe('https://s3.example.com/get?sig=1');
  });

  it('resolveImageSrc は "kb/" で始まらない src には触れない', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(true));
    renderPage();
    await screen.findByTestId('editor');

    const resolveImageSrc = hoisted.editorProps.current?.resolveImageSrc;
    await expect(resolveImageSrc!('https://cdn.example.com/a.png')).resolves.toBe(
      'https://cdn.example.com/a.png',
    );
    expect(hoisted.issuePageImageDownloadURL).not.toHaveBeenCalled();
  });
});

describe('KbPage のカバー画像', () => {
  /** カバー画像用の隠しファイル入力は「カバー画像を追加/変更」ボタンと同じ行に居る。 */
  function coverFileInput(button: HTMLElement): HTMLInputElement {
    const input = button.parentElement?.querySelector('input[type="file"]');
    if (!input) throw new Error('カバー画像の file input が見つかりません');
    return input as HTMLInputElement;
  }

  it('ファイルを選ぶとアップロードして設定する（未設定 → 設定済みの表示に変わる）', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(true));
    hoisted.uploadPageImage.mockResolvedValue('kb/w-3f2a9c/p1/1.bin');
    hoisted.setPageCover.mockResolvedValue({
      page: resolved(true).page,
      cover: { type: 'file', url: 'https://s3.example.com/get?sig=1' },
    });
    renderPage();

    const addButton = await screen.findByRole('button', { name: 'カバー画像を追加' });
    const file = new File(['x'], 'cover.png', { type: 'image/png' });
    fireEvent.change(coverFileInput(addButton), { target: { files: [file] } });

    await waitFor(() =>
      expect(hoisted.uploadPageImage).toHaveBeenCalledWith('w-3f2a9c', 'p1', file),
    );
    await waitFor(() =>
      expect(hoisted.setPageCover).toHaveBeenCalledWith('w-3f2a9c', 'p1', 'kb/w-3f2a9c/p1/1.bin'),
    );
    expect(await screen.findByRole('button', { name: 'カバー画像を変更' })).toBeInTheDocument();
    expect(await screen.findByRole('button', { name: 'カバー画像を外す' })).toBeInTheDocument();
  });

  it('カバー画像を外す', async () => {
    hoisted.resolvePage.mockResolvedValue({
      ...resolved(true),
      cover: { type: 'file', url: 'https://s3.example.com/get?sig=1' },
    });
    hoisted.clearPageCover.mockResolvedValue({ page: resolved(true).page, cover: null });
    renderPage();

    const removeButton = await screen.findByRole('button', { name: 'カバー画像を外す' });
    fireEvent.click(removeButton);

    await waitFor(() => expect(hoisted.clearPageCover).toHaveBeenCalledWith('w-3f2a9c', 'p1'));
    expect(await screen.findByRole('button', { name: 'カバー画像を追加' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'カバー画像を外す' })).not.toBeInTheDocument();
  });

  it('アップロードに失敗したら知らせを出す', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(true));
    hoisted.uploadPageImage.mockRejectedValue(new Error('boom'));
    renderPage();

    const addButton = await screen.findByRole('button', { name: 'カバー画像を追加' });
    const file = new File(['x'], 'cover.png', { type: 'image/png' });
    fireEvent.change(coverFileInput(addButton), { target: { files: [file] } });

    await waitFor(() =>
      expect(hoisted.showToast).toHaveBeenCalledWith('error', 'カバー画像を変更できませんでした'),
    );
    expect(hoisted.setPageCover).not.toHaveBeenCalled();
  });

  it('p1でアップロード中にp2へ移動したら、p1向けのカバー設定は送らない', async () => {
    let resolveUpload: ((key: string) => void) | undefined;
    hoisted.resolvePage.mockImplementation((pageId: string) =>
      Promise.resolve({ ...resolved(true), page: { ...resolved(true).page, id: pageId } }),
    );
    hoisted.uploadPageImage.mockImplementation(
      () =>
        new Promise<string>((resolve) => {
          resolveUpload = resolve;
        }),
    );
    const view = renderPage();

    const addButton = await screen.findByRole('button', { name: 'カバー画像を追加' });
    const file = new File(['x'], 'cover.png', { type: 'image/png' });
    fireEvent.change(coverFileInput(addButton), { target: { files: [file] } });
    await waitFor(() => expect(hoisted.uploadPageImage).toHaveBeenCalledWith('w-3f2a9c', 'p1', file));

    // アップロードが終わる前に p2 へ移動する。
    hoisted.useParams.mockReturnValue({ pageId: 'p2' });
    view.rerender(
      <MemoryRouter initialEntries={['/kb/p2']}>
        <KbPage />
      </MemoryRouter>,
    );
    await screen.findByRole('button', { name: 'カバー画像を追加' });

    // p1 向けのアップロードがいま完了しても、p2 の cover API へは送らない。
    await act(async () => {
      resolveUpload?.('kb/w-3f2a9c/p1/1.bin');
    });
    expect(hoisted.setPageCover).not.toHaveBeenCalled();
    expect(hoisted.showToast).not.toHaveBeenCalled();
  });

  // block_id_conflict は再送しても直らない失敗なので、「未保存」の表示だけでは
  // 原因が伝わらない。再読み込みを促す通知を出す（CodeRabbit 指摘）。
  it('本文保存がblock_id_conflictで失敗したら再読み込みを促す', async () => {
    hoisted.replaceContent.mockRejectedValue(blockIdConflictError());
    renderPage();
    await screen.findByTestId('editor');

    // findByTestId のポーリングと fake timers が競合しないよう、初期描画が
    // 落ち着いてから fake timers に切り替える（useKbPageDoc.test.ts と違い、
    // ここは実 DOM の非同期待ち（findBy*）を経由するため）。
    vi.useFakeTimers();
    try {
      act(() => {
        hoisted.editorProps.current?.onChange?.({ type: 'doc', content: [] });
      });
      await act(async () => {
        vi.advanceTimersByTime(1000);
      });
      await act(async () => {
        await Promise.resolve();
        await Promise.resolve();
      });

      expect(hoisted.showToast).toHaveBeenCalledWith(
        'error',
        '他の変更と競合したため本文を保存できませんでした。ページを再読み込みしてやり直してください。',
      );
    } finally {
      vi.useRealTimers();
    }
  });

  it('カバー画像は読むだけの人には出さない', async () => {
    hoisted.resolvePage.mockResolvedValue({
      ...resolved(false),
      cover: { type: 'file', url: 'https://s3.example.com/get?sig=1' },
    });
    renderPage();

    await screen.findByRole('heading', { name: '親ページ' });
    expect(screen.queryByRole('button', { name: 'カバー画像を追加' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'カバー画像を変更' })).not.toBeInTheDocument();
  });
});

describe('KbPage の共有', () => {
  it('権限を変えられないページには共有ボタンを出さない', async () => {
    // 押しても 404 が返るだけのボタンは、権限が無いことすら伝えない。
    hoisted.resolvePage.mockResolvedValue(resolved(true, false));
    renderPage();

    await screen.findByText('親ページ');
    expect(screen.queryByRole('button', { name: '共有' })).not.toBeInTheDocument();
  });

  it('開くまで権限は取りに行かない', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(true, true));
    renderPage();

    const share = await screen.findByRole('button', { name: '共有' });
    // 開いていないパネルのために、ページを開くたび 2 本引かない。
    // 片方だけ先読みに戻る退行を拾えるよう、両方を見る。
    expect(hoisted.listPageGrants).not.toHaveBeenCalled();
    expect(hoisted.listGrantablePrincipals).not.toHaveBeenCalled();

    fireEvent.click(share);
    await waitFor(() => expect(hoisted.listPageGrants).toHaveBeenCalledWith('w-3f2a9c', 'p1'));
    expect(hoisted.listGrantablePrincipals).toHaveBeenCalledWith('w-3f2a9c', 'p1');
    expect(await screen.findByRole('region', { name: '共有' })).toBeInTheDocument();
  });

  it('閉じるボタンでパネルが消える', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(true, true));
    renderPage();

    fireEvent.click(await screen.findByRole('button', { name: '共有' }));
    const panel = await screen.findByRole('region', { name: '共有' });
    fireEvent.click(within(panel).getByLabelText('共有を閉じる'));

    await waitFor(() => expect(screen.queryByRole('region', { name: '共有' })).not.toBeInTheDocument());
  });
});

describe('KbPage のコメント', () => {
  const thread = (id: string, resolvedAt: string | null = null, blockId?: string) => ({
    id,
    createdBy: { userId: 1, name: '田中 太郎' },
    resolvedAt,
    resolvedBy: resolvedAt ? { userId: 2, name: '鈴木 花子' } : null,
    createdAt: '2026-09-01T00:00:00Z',
    comments: [
      {
        id: `${id}-c1`,
        author: { userId: 1, name: '田中 太郎' },
        body: [{ type: 'text', text: 'これは何ですか？' }],
        createdAt: '2026-09-01T00:00:00Z',
        updatedAt: '2026-09-01T00:00:00Z',
      },
    ],
    ...(blockId ? { blockId } : {}),
  });

  it('パネルの開閉に関わらず常に取得する。未解決件数がコメントボタンとエディタ側の両方に出る', async () => {
    hoisted.listCommentThreads.mockResolvedValue([
      thread('t1', null, 'block-1'),
      thread('t2', '2026-09-02T00:00:00Z', 'block-2'),
    ]);
    renderPage();
    await screen.findByTestId('editor');

    // パネルを開く前から取得している（バッジ表示のため）。
    await waitFor(() => expect(hoisted.listCommentThreads).toHaveBeenCalledWith('w-3f2a9c', 'p1'));

    // 未解決 1 件（もう 1 件は解決済み）でコメントボタンにバッジが出る。
    expect(await screen.findByRole('button', { name: 'コメント (未解決 1 件)' })).toBeInTheDocument();
    // エディタ側へも、解決済みを除いたブロックIDごとの件数が渡る。
    await waitFor(() =>
      expect(hoisted.editorProps.current?.commentBadgeCounts).toEqual({ 'block-1': 1 }),
    );

    // まだパネル自体は開いていない。
    expect(screen.queryAllByText('未解決（1）').length).toBe(0);

    fireEvent.click(screen.getByRole('button', { name: 'コメント (未解決 1 件)' }));
    await waitFor(() => expect(screen.getAllByText('未解決（1）').length).toBeGreaterThan(0));
    expect(screen.getAllByText('解決済み（1）').length).toBeGreaterThan(0);
  });

  it('もう一度押すと閉じる', async () => {
    renderPage();
    const toggle = await screen.findByRole('button', { name: 'コメント' });

    fireEvent.click(toggle);
    await waitFor(() => expect(hoisted.listCommentThreads).toHaveBeenCalled());
    expect(screen.getAllByText('まだコメントはありません。').length).toBeGreaterThan(0);

    fireEvent.click(toggle);
    await waitFor(() =>
      expect(screen.queryAllByText('まだコメントはありません。').length).toBe(0),
    );
  });

  it('コメント権限が無ければ読めるが、作成フォームは出ない', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(true, false, false));
    renderPage();

    fireEvent.click(await screen.findByRole('button', { name: 'コメント' }));
    await waitFor(() => expect(hoisted.listCommentThreads).toHaveBeenCalled());

    expect(screen.queryAllByPlaceholderText('コメントを書く…').length).toBe(0);
  });

  it('選択範囲からコメントを作る一連の流れ: onRequestComment → パネルが開いて引用文が出る → 送信で錨込みのスレッドを作る', async () => {
    const anchor: CommentAnchor = { blockId: 'block-1', anchorFrom: 6, anchorTo: 11, quote: '選んだ文' };
    hoisted.createCommentThread.mockResolvedValue({
      id: 't-new',
      createdBy: { userId: 1, name: '田中 太郎' },
      resolvedAt: null,
      resolvedBy: null,
      createdAt: '2026-09-01T00:00:00Z',
      comments: [],
      ...anchor,
    });
    renderPage();
    await screen.findByTestId('editor');

    // バブルメニューの「コメント」相当（本物の tiptap 選択は RichTextEditor 自体のテストで
    // 担保済み。ここでは KbPage が onRequestComment を正しく受け止めるかを見る）。
    act(() => {
      hoisted.editorProps.current?.onRequestComment?.(anchor);
    });

    // パネルが（閉じていたのに）開き、引用文が見える。page-level の作成フォームは隠れる。
    await waitFor(() => expect(screen.getAllByText('選択範囲へコメント').length).toBeGreaterThan(0));
    expect(screen.getAllByText('選んだ文').length).toBeGreaterThan(0);
    expect(screen.queryAllByText('新しいスレッドを作成').length).toBe(0);

    fireEvent.change(screen.getAllByPlaceholderText('コメントを書く…')[0], {
      target: { value: 'ここは要修正です' },
    });
    fireEvent.click(screen.getAllByRole('button', { name: '送信' })[0]);

    await waitFor(() =>
      expect(hoisted.createCommentThread).toHaveBeenCalledWith(
        'w-3f2a9c',
        'p1',
        [{ type: 'text', text: 'ここは要修正です' }],
        anchor,
      ),
    );
    // 送信できたら「選択範囲へコメント」フォームが閉じる（pendingAnchor がクリアされる）。
    await waitFor(() => expect(screen.queryAllByText('選択範囲へコメント').length).toBe(0));
  });

  it('バッジをクリックするとパネルが開き、該当スレッドまでスクロールする', async () => {
    hoisted.listCommentThreads.mockResolvedValue([thread('t1', null, 'block-1')]);
    const scrollIntoView = vi.spyOn(Element.prototype, 'scrollIntoView').mockImplementation(() => {});
    renderPage();
    await screen.findByTestId('editor');
    await waitFor(() =>
      expect(hoisted.editorProps.current?.commentBadgeCounts).toEqual({ 'block-1': 1 }),
    );

    act(() => {
      hoisted.editorProps.current?.onCommentBadgeClick?.('block-1');
    });

    await waitFor(() => expect(screen.getAllByText('未解決（1）').length).toBeGreaterThan(0));
    // scrollIntoView は Element.prototype に spy を張っているので、呼ばれたことだけでは
    // 「何か別の要素」が呼んだ可能性を排除できない。呼び出し元（this）が実際に
    // comment-thread-t1 要素であることまで確かめる — CodeRabbit 指摘。
    await waitFor(() => expect(scrollIntoView).toHaveBeenCalled());
    const calledOn = scrollIntoView.mock.instances[0] as unknown as HTMLElement;
    expect(calledOn.id).toBe('comment-thread-t1');
    scrollIntoView.mockRestore();
  });
});

describe('KbPage の履歴', () => {
  const version = (seq: number, note: string | null = null) => ({
    seq,
    author: { userId: 1, name: '田中 太郎' },
    note,
    createdAt: '2026-09-01T09:00:00Z',
  });

  const versionDetail = (seq: number, note: string | null = null) => ({
    ...version(seq, note),
    doc: { type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text: '過去の内容' }] }] },
  });

  it('開くと一覧を取得する。行を選ぶと読み取り専用になり本文がその版へ切り替わる', async () => {
    hoisted.listPageVersions.mockResolvedValue([version(2), version(1, '初版')]);
    hoisted.getPageVersion.mockResolvedValue(versionDetail(1, '初版'));
    renderPage();
    await screen.findByTestId('editor');
    expect(hoisted.editorProps.current?.editable).toBe(true);

    fireEvent.click(await screen.findByRole('button', { name: '履歴' }));
    await waitFor(() => expect(hoisted.listPageVersions).toHaveBeenCalledWith('w-3f2a9c', 'p1'));

    const row = (await screen.findAllByRole('button', { name: /初版/ }))[0];
    fireEvent.click(row);

    await waitFor(() => expect(hoisted.getPageVersion).toHaveBeenCalledWith('w-3f2a9c', 'p1', 1));
    await waitFor(() => expect(hoisted.editorProps.current?.editable).toBe(false));
    expect(hoisted.editorProps.current?.value).toEqual(versionDetail(1, '初版').doc);
    // 版を表示中の帯が出る。
    expect(await screen.findByText(/の版を表示中/)).toBeInTheDocument();
  });

  it('版プレビュー中はコメントのバブルメニュー・件数バッジ関連の props を渡さない', async () => {
    hoisted.listPageVersions.mockResolvedValue([version(1, '初版')]);
    hoisted.getPageVersion.mockResolvedValue(versionDetail(1, '初版'));
    renderPage();
    await screen.findByTestId('editor');

    fireEvent.click(await screen.findByRole('button', { name: '履歴' }));
    fireEvent.click((await screen.findAllByRole('button', { name: /初版/ }))[0]);
    await waitFor(() => expect(hoisted.editorProps.current?.editable).toBe(false));

    // editable=false と組み合わせ、コメント関連 props はそもそも渡さない
    // （RichTextEditor 自身が (editable || canComment) の判定でバブルメニューを
    // 出さなくなるので、渡さないだけで十分安全 — 明示的に false を渡す必要は無い）。
    expect(hoisted.editorProps.current?.canComment).toBeUndefined();
    expect(hoisted.editorProps.current?.commentBadgeCounts).toBeUndefined();
    expect(hoisted.editorProps.current?.onCommentBadgeClick).toBeUndefined();
    expect(hoisted.editorProps.current?.onRequestComment).toBeUndefined();
  });

  it('この版に戻すを押すと確認ダイアログが出て、確定すると復元APIが呼ばれ、本文・最終編集が更新される', async () => {
    hoisted.listPageVersions.mockResolvedValue([version(1, '初版')]);
    hoisted.getPageVersion.mockResolvedValue(versionDetail(1, '初版'));
    hoisted.restorePageVersion.mockResolvedValue({
      doc: { type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text: '戻した内容' }] }] },
      builtAt: '2026-09-06T00:00:00Z',
      lastEditedBy: { userId: 2, name: '鈴木 花子' },
      lastEditedAt: '2026-09-06T00:00:00Z',
    });
    renderPage();
    await screen.findByTestId('editor');

    fireEvent.click(await screen.findByRole('button', { name: '履歴' }));
    fireEvent.click((await screen.findAllByRole('button', { name: /初版/ }))[0]);
    await waitFor(() => expect(hoisted.editorProps.current?.editable).toBe(false));

    // まだ確定していないので復元 API は呼ばれていない。
    fireEvent.click(await screen.findByRole('button', { name: 'この版に戻す' }));
    expect(hoisted.restorePageVersion).not.toHaveBeenCalled();

    const dialog = await screen.findByRole('dialog', { name: 'この版に戻しますか' });
    expect(dialog).toHaveTextContent('現在の内容は上書きされますが');
    fireEvent.click(within(dialog).getByRole('button', { name: 'この版に戻す' }));

    await waitFor(() =>
      expect(hoisted.restorePageVersion).toHaveBeenCalledWith('w-3f2a9c', 'p1', 1),
    );
    // 復元成功: プレビューを終えて編集可能な表示へ戻り、本文・最終編集が更新される。
    await waitFor(() => expect(hoisted.editorProps.current?.editable).toBe(true));
    expect(hoisted.editorProps.current?.value).toEqual({
      type: 'doc',
      content: [{ type: 'paragraph', content: [{ type: 'text', text: '戻した内容' }] }],
    });
    expect(await screen.findByText(/最終編集 鈴木 花子/)).toBeInTheDocument();
    expect(screen.queryByText(/の版を表示中/)).not.toBeInTheDocument();
  });

  it('復元に失敗したら知らせを出す。プレビューは終えない', async () => {
    hoisted.listPageVersions.mockResolvedValue([version(1, '初版')]);
    hoisted.getPageVersion.mockResolvedValue(versionDetail(1, '初版'));
    hoisted.restorePageVersion.mockRejectedValue(new Error('forbidden'));
    renderPage();
    await screen.findByTestId('editor');

    fireEvent.click(await screen.findByRole('button', { name: '履歴' }));
    fireEvent.click((await screen.findAllByRole('button', { name: /初版/ }))[0]);
    await waitFor(() => expect(hoisted.editorProps.current?.editable).toBe(false));

    fireEvent.click(await screen.findByRole('button', { name: 'この版に戻す' }));
    const dialog = await screen.findByRole('dialog', { name: 'この版に戻しますか' });
    fireEvent.click(within(dialog).getByRole('button', { name: 'この版に戻す' }));

    await waitFor(() =>
      expect(hoisted.showToast).toHaveBeenCalledWith('error', 'この版に戻せませんでした'),
    );
    expect(hoisted.editorProps.current?.editable).toBe(false);
  });

  it('現在の版に戻る（閉じる）を押すとプレビューを終え、パネルを閉じても帯は残る', async () => {
    hoisted.listPageVersions.mockResolvedValue([version(1, '初版')]);
    hoisted.getPageVersion.mockResolvedValue(versionDetail(1, '初版'));
    renderPage();
    await screen.findByTestId('editor');

    const historyToggle = await screen.findByRole('button', { name: '履歴' });
    fireEvent.click(historyToggle);
    fireEvent.click((await screen.findAllByRole('button', { name: /初版/ }))[0]);
    await waitFor(() => expect(hoisted.editorProps.current?.editable).toBe(false));

    // 履歴パネルを閉じても、帯とプレビューは残る（明示的な「現在の版に戻る」だけが退出手段）。
    fireEvent.click(historyToggle);
    expect(hoisted.editorProps.current?.editable).toBe(false);
    expect(screen.getByText(/の版を表示中/)).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '現在の版に戻る' }));
    await waitFor(() => expect(hoisted.editorProps.current?.editable).toBe(true));
    expect(screen.queryByText(/の版を表示中/)).not.toBeInTheDocument();
  });

  it('編集できないページでは「版を残す」フォームも「この版に戻す」ボタンも出ない', async () => {
    hoisted.resolvePage.mockResolvedValue(resolved(false));
    hoisted.listPageVersions.mockResolvedValue([version(1, '初版')]);
    hoisted.getPageVersion.mockResolvedValue(versionDetail(1, '初版'));
    renderPage();

    fireEvent.click(await screen.findByRole('button', { name: '履歴' }));
    await waitFor(() => expect(hoisted.listPageVersions).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '版を残す' })).not.toBeInTheDocument();

    fireEvent.click((await screen.findAllByRole('button', { name: /初版/ }))[0]);
    await waitFor(() => expect(hoisted.editorProps.current?.editable).toBe(false));
    expect(screen.queryByRole('button', { name: 'この版に戻す' })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '現在の版に戻る' })).toBeInTheDocument();
  });
});

describe('KbPage の入口解決（素の /kb）', () => {
  function renderEntry() {
    hoisted.useParams.mockReturnValue({});
    return render(
      <MemoryRouter initialEntries={['/kb']}>
        <KbPage />
      </MemoryRouter>,
    );
  }

  it('直近に開いたページがあれば、そこへ即座に移る', async () => {
    hoisted.getLastVisitedPageId.mockReturnValue('p9');
    renderEntry();

    await waitFor(() => expect(hoisted.navigate).toHaveBeenCalledWith('/kb/p9', { replace: true }));
    expect(hoisted.fetchWorkspaces).not.toHaveBeenCalled();
  });

  it('閲覧履歴が無ければ、最初に見つかったページへ移る', async () => {
    hoisted.fetchWorkspaces.mockResolvedValue([{ slug: 'acme', name: '', createdAt: '', canManage: true }]);
    hoisted.fetchSpaces.mockResolvedValue([{ id: 'space-1', key: 's', name: '', visibility: 'workspace', createdAt: '' }]);
    hoisted.fetchPageTree.mockResolvedValue({
      pages: [{ page: { id: 'p2', spaceId: 'space-1', title: '設計メモ', createdByUserId: 1, createdAt: '', updatedAt: '' }, children: [], hasHiddenChildren: false, parentArchived: false }],
      hasHiddenChildren: false,
    });
    renderEntry();

    await waitFor(() => expect(hoisted.navigate).toHaveBeenCalledWith('/kb/p2', { replace: true }));
  });

  it('どこにも 1 枚も見つからなければ「まだページがありません」を出す', async () => {
    hoisted.fetchWorkspaces.mockResolvedValue([]);
    renderEntry();

    expect(await screen.findByText('まだページがありません')).toBeInTheDocument();
    expect(hoisted.navigate).not.toHaveBeenCalled();
  });
});
