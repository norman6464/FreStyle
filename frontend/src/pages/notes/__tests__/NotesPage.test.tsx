import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ReactElement } from 'react';
import NotesPage from '../ui/NotesPage';
import { useDocuments } from '../model/useDocuments';
import { useDocumentEditor } from '../model/useDocumentEditor';
import type { RichDocumentSummary } from '@/entities/document';

vi.mock('../model/useDocuments');
vi.mock('../model/useDocumentEditor');

const mockShowToast = vi.fn();
vi.mock('@/shared/lib/hooks/useToast', () => ({
  useToast: () => ({ showToast: mockShowToast, toasts: [], removeToast: vi.fn() }),
}));

// tiptap を jsdom に載せないため、エディタは軽量スタブに差し替える。
vi.mock('@/shared/ui/RichTextEditor', () => ({
  RichTextEditor: () => <div data-testid="rich-text-editor" />,
  SaveStatusIndicator: ({ status }: { status: string }) => <span data-testid="save-status">{status}</span>,
  emptyRichDoc: () => ({ type: 'doc', content: [{ type: 'paragraph' }] }),
}));

function summary(partial: Partial<RichDocumentSummary> & { id: string }): RichDocumentSummary {
  return {
    id: partial.id,
    ownerId: partial.ownerId ?? 7,
    kind: 'note',
    title: partial.title ?? `ノート${partial.id}`,
    isPublic: false,
    schemaVersion: 1,
    revision: partial.revision ?? 1,
    createdAt: partial.createdAt ?? '2026-01-01T00:00:00Z',
    updatedAt: partial.updatedAt ?? '2026-01-02T00:00:00Z',
  };
}

const baseDocuments = {
  documents: [] as RichDocumentSummary[],
  filteredDocuments: [] as RichDocumentSummary[],
  selectedId: null as string | null,
  loading: false,
  error: null as string | null,
  searchQuery: '',
  setSearchQuery: vi.fn(),
  sort: 'default' as const,
  setSort: vi.fn(),
  fetchDocuments: vi.fn(),
  createDocument: vi.fn(),
  deleteDocument: vi.fn(),
  selectDocument: vi.fn(),
  syncSummary: vi.fn(),
  deleteTargetId: null as string | null,
  requestDelete: vi.fn(),
  confirmDelete: vi.fn(),
  cancelDelete: vi.fn(),
};

const baseEditor = {
  editTitle: '',
  editDoc: { type: 'doc', content: [] },
  saveStatus: 'idle' as const,
  loadingDoc: false,
  handleTitleChange: vi.fn(),
  handleDocChange: vi.fn(),
  forceSave: vi.fn(),
};

function setup(docOverrides = {}, editorOverrides = {}) {
  vi.mocked(useDocuments).mockReturnValue({ ...baseDocuments, ...docOverrides } as never);
  vi.mocked(useDocumentEditor).mockReturnValue({ ...baseEditor, ...editorOverrides } as never);
}

const renderPage = (ui: ReactElement = <NotesPage />) => render(ui);

describe('NotesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setup();
  });

  it('マウント時に fetchDocuments を呼ぶ', () => {
    const fetchDocuments = vi.fn();
    setup({ fetchDocuments });
    renderPage();
    expect(fetchDocuments).toHaveBeenCalled();
  });

  // SecondaryPanel はモバイル/デスクトップで内容を二重描画するため、パネル内要素は getAllBy で拾う。
  it('文書が無いとき空表示を出す', () => {
    setup({ documents: [], filteredDocuments: [] });
    renderPage();
    expect(screen.getAllByText('ノートがありません').length).toBeGreaterThan(0);
  });

  it('一覧を描画し、選択で selectDocument を呼ぶ', () => {
    const selectDocument = vi.fn();
    const docs = [summary({ id: 'a', title: 'メモA' }), summary({ id: 'b', title: 'メモB' })];
    setup({ documents: docs, filteredDocuments: docs, selectDocument });
    renderPage();
    expect(screen.getAllByText('メモA').length).toBeGreaterThan(0);
    fireEvent.click(screen.getAllByLabelText('ノート「メモA」を選択')[0]);
    expect(selectDocument).toHaveBeenCalledWith('a');
  });

  it('新しいノートボタンで createDocument を呼び成功トーストを出す', async () => {
    const createDocument = vi.fn().mockResolvedValue(summary({ id: 'new' }));
    setup({ createDocument });
    renderPage();
    fireEvent.click(screen.getAllByRole('button', { name: '新しいノート' })[0]);
    await waitFor(() => expect(createDocument).toHaveBeenCalledWith('無題'));
    await waitFor(() => expect(mockShowToast).toHaveBeenCalledWith('success', 'ノートを作成しました'));
  });

  it('削除アイコンで requestDelete、確認モーダルで confirmDelete', async () => {
    const requestDelete = vi.fn();
    const confirmDelete = vi.fn().mockResolvedValue(undefined);
    const docs = [summary({ id: 'a', title: 'メモA' })];
    setup({ documents: docs, filteredDocuments: docs, requestDelete, deleteTargetId: 'a', confirmDelete });
    renderPage();
    fireEvent.click(screen.getAllByLabelText('ノート「メモA」を削除')[0]);
    expect(requestDelete).toHaveBeenCalledWith('a');
    // deleteTargetId='a' なので確認モーダルが出ている。
    fireEvent.click(screen.getByRole('button', { name: '削除' }));
    await waitFor(() => expect(confirmDelete).toHaveBeenCalled());
  });

  it('選択中はタイトル入力と本文エディタと保存状態を表示する', () => {
    setup({ selectedId: 'a' }, { editTitle: 'タイトルX', saveStatus: 'saved' });
    renderPage();
    expect(screen.getByLabelText('ノートのタイトル')).toHaveValue('タイトルX');
    expect(screen.getByTestId('rich-text-editor')).toBeInTheDocument();
    expect(screen.getByTestId('save-status')).toHaveTextContent('saved');
  });

  it('doc 読み込み中はローディングを出す', () => {
    setup({ selectedId: 'a' }, { loadingDoc: true });
    renderPage();
    expect(screen.queryByLabelText('ノートのタイトル')).not.toBeInTheDocument();
  });

  it('未選択のときは選択を促す空表示', () => {
    setup({ selectedId: null });
    renderPage();
    expect(screen.getByText('ノートを選択してください')).toBeInTheDocument();
  });

  it('タイトル入力で handleTitleChange を呼ぶ', () => {
    const handleTitleChange = vi.fn();
    setup({ selectedId: 'a' }, { editTitle: 'X', handleTitleChange });
    renderPage();
    fireEvent.change(screen.getByLabelText('ノートのタイトル'), { target: { value: 'XY' } });
    expect(handleTitleChange).toHaveBeenCalledWith('XY');
  });
});
