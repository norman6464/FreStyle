import { render, screen, waitFor, fireEvent, act, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import NotePage from '../NotePage';
import { emitNoteTreeEvent } from '@/entities/note';
import type { EditorCommand } from '@/shared/ui/RichTextEditor';

const hoisted = vi.hoisted(() => ({
  resolvePage: vi.fn(),
  replaceContent: vi.fn(),
  renamePage: vi.fn(),
  createPage: vi.fn(),
  emit: vi.fn(),
  showToast: vi.fn(),
  navigate: vi.fn(),
  editorProps: { current: null as null | { extraSlashCommands?: EditorCommand[] } },
}));

vi.mock('@/entities/note', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/entities/note')>();
  return {
    ...actual,
    NoteRepository: {
      resolvePage: hoisted.resolvePage,
      replaceContent: hoisted.replaceContent,
      renamePage: hoisted.renamePage,
      createPage: hoisted.createPage,
    },
    // スパイしつつ実物へ転送する（購読側の配線もこのテストの検査対象のため）。
    emitNoteTreeEvent: (event: Parameters<typeof actual.emitNoteTreeEvent>[0]) => {
      hoisted.emit(event);
      actual.emitNoteTreeEvent(event);
    },
  };
});

vi.mock('@/shared/lib/hooks/useToast', () => ({
  useToast: () => ({ showToast: hoisted.showToast, toasts: [], removeToast: vi.fn() }),
}));

vi.mock('react-router-dom', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-router-dom')>();
  return { ...actual, useNavigate: () => hoisted.navigate, useParams: () => ({ pageId: 'p1' }) };
});

// サイドバーは自前のテストで検証済み。ここでは画面の配線だけを見る。
vi.mock('@/widgets/note-sidebar', () => ({
  NoteSidebar: () => <nav aria-label="サイドバーの偽物" />,
}));

// エディタは重い（tiptap 実体）ので、渡された props を捕まえる薄い偽物に差し替える。
// /page の run は本物の createSubpage を通る（そこが配線の検査対象）。
vi.mock('@/shared/ui/RichTextEditor', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/shared/ui/RichTextEditor')>();
  return {
    ...actual,
    RichTextEditor: (props: { extraSlashCommands?: EditorCommand[] }) => {
      hoisted.editorProps.current = props;
      return <div data-testid="editor" />;
    },
  };
});

const resolved = (canEdit: boolean) => ({
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
});

/** /page の run に渡す最小のエディタ（createSubpage が使う形だけ）。 */
function fakeEditor() {
  return {
    chain: () => ({ focus: () => ({ insertContent: () => ({ run: () => {} }) }) }),
  } as never;
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={['/p/p1']}>
      <NotePage />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  hoisted.editorProps.current = null;
  hoisted.resolvePage.mockResolvedValue(resolved(true));
});

describe('NotePage の配線', () => {
  it('題名の改名に失敗したら知らせを出し、入力は消えない', async () => {
    hoisted.renamePage.mockRejectedValue(new Error('403'));
    renderPage();
    const input = await screen.findByRole('textbox', { name: 'ページの題名' });

    fireEvent.change(input, { target: { value: '新しい題名' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    await waitFor(() =>
      expect(hoisted.showToast).toHaveBeenCalledWith('error', '題名を変更できませんでした'),
    );
    // 再 throw が NotePageTitle まで届いている＝入力が保たれている。
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
    await waitFor(() => expect(hoisted.navigate).toHaveBeenCalledWith('/p/child-1'));

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
      '/p/anc-1',
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

    // 無関係なページの削除では動かない。
    act(() => {
      emitNoteTreeEvent({ type: 'page-deleted', pageId: 'unrelated' });
    });
    expect(hoisted.navigate).not.toHaveBeenCalled();

    // 祖先（anc-1）が消えたら CASCADE で自分も消えている — 一覧へ戻る。
    // 祖先はサーバー応答から取るので、サイドバーの現役の木に載っていない
    // （アーカイブ済みの）ページを開いていても判定できる。
    act(() => {
      emitNoteTreeEvent({ type: 'page-deleted', pageId: 'anc-1' });
    });
    await waitFor(() => expect(hoisted.navigate).toHaveBeenCalledWith('/notes'));

    // 自分自身の削除でも戻る。
    hoisted.navigate.mockClear();
    act(() => {
      emitNoteTreeEvent({ type: 'page-deleted', pageId: 'p1' });
    });
    await waitFor(() => expect(hoisted.navigate).toHaveBeenCalledWith('/notes'));
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
