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

// エディタスタブに渡された props（onImageUpload 検証用）と画像アップロード mock を hoist する。
const hoisted = vi.hoisted(() => ({
  rteProps: { current: null as { onImageUpload?: (file: File) => Promise<string> } | null },
  upload: vi.fn(),
}));

vi.mock('@/entities/user', () => ({
  ImageUploadRepository: { upload: (...args: unknown[]) => hoisted.upload(...args) },
}));

// tiptap を jsdom に載せないため、エディタは軽量スタブに差し替える（props は捕捉する）。
vi.mock('@/shared/ui/RichTextEditor', () => ({
  RichTextEditor: (props: { onImageUpload?: (file: File) => Promise<string> }) => {
    hoisted.rteProps.current = props;
    return <div data-testid="rich-text-editor" />;
  },
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

const baseDocuments: ReturnType<typeof useDocuments> = {
  documents: [],
  filteredDocuments: [],
  selectedId: null,
  loading: false,
  error: null,
  searchQuery: '',
  setSearchQuery: vi.fn(),
  sort: 'default',
  setSort: vi.fn(),
  fetchDocuments: vi.fn(),
  createDocument: vi.fn(),
  deleteDocument: vi.fn(),
  selectDocument: vi.fn(),
  syncSummary: vi.fn(),
  deleteTargetId: null,
  requestDelete: vi.fn(),
  confirmDelete: vi.fn(),
  cancelDelete: vi.fn(),
};

const baseEditor: ReturnType<typeof useDocumentEditor> = {
  editTitle: '',
  editDoc: { type: 'doc', content: [] },
  saveStatus: 'idle',
  loadingDoc: false,
  loadError: false,
  handleTitleChange: vi.fn(),
  handleDocChange: vi.fn(),
  forceSave: vi.fn(),
  reload: vi.fn(),
};

function setup(docOverrides: Partial<ReturnType<typeof useDocuments>> = {}, editorOverrides: Partial<ReturnType<typeof useDocumentEditor>> = {}) {
  vi.mocked(useDocuments).mockReturnValue({ ...baseDocuments, ...docOverrides });
  vi.mocked(useDocumentEditor).mockReturnValue({ ...baseEditor, ...editorOverrides });
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

  it('useDocumentEditor に selectedId と onSynced/onConflict を配線する', () => {
    setup({ selectedId: 'a' });
    renderPage();
    const call = vi.mocked(useDocumentEditor).mock.calls.at(-1);
    expect(call?.[0]).toBe('a');
    expect(call?.[1]?.onSynced).toBe(baseDocuments.syncSummary);
    // onConflict を実行すると競合トーストが出る（この PR の主要機能）。
    call?.[1]?.onConflict?.();
    expect(mockShowToast).toHaveBeenCalledWith('info', expect.stringContaining('最新版'));
  });

  it('文書が無いとき空表示を出す', () => {
    setup({ documents: [], filteredDocuments: [] });
    renderPage();
    expect(screen.getAllByText('ノートがありません').length).toBeGreaterThan(0);
  });

  it('取得失敗時は「取得に失敗」を出し、空表示にはしない', () => {
    setup({ documents: [], filteredDocuments: [], error: 'ノートの取得に失敗しました' });
    renderPage();
    expect(screen.getAllByText('ノートの取得に失敗しました').length).toBeGreaterThan(0);
    expect(screen.queryByText('ノートがありません')).not.toBeInTheDocument();
  });

  it('一覧を描画し、選択で selectDocument を呼ぶ', () => {
    const selectDocument = vi.fn();
    const docs = [summary({ id: 'a', title: 'メモA' }), summary({ id: 'b', title: 'メモB' })];
    setup({ documents: docs, filteredDocuments: docs, selectDocument });
    renderPage();
    expect(screen.getAllByText('メモA').length).toBeGreaterThan(0);
    fireEvent.click(screen.getAllByRole('button', { name: 'ノート「メモA」を選択' })[0]);
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

  it('削除アイコンで requestDelete、確認で成功トースト', async () => {
    const requestDelete = vi.fn();
    const confirmDelete = vi.fn().mockResolvedValue(true);
    const docs = [summary({ id: 'a', title: 'メモA' })];
    setup({ documents: docs, filteredDocuments: docs, requestDelete, deleteTargetId: 'a', confirmDelete });
    renderPage();
    fireEvent.click(screen.getAllByRole('button', { name: 'ノート「メモA」を削除' })[0]);
    expect(requestDelete).toHaveBeenCalledWith('a');
    fireEvent.click(screen.getByRole('button', { name: '削除' }));
    await waitFor(() => expect(mockShowToast).toHaveBeenCalledWith('success', 'ノートを削除しました'));
  });

  it('削除に失敗するとエラートーストを出す', async () => {
    const confirmDelete = vi.fn().mockResolvedValue(false);
    const docs = [summary({ id: 'a', title: 'メモA' })];
    setup({ documents: docs, filteredDocuments: docs, deleteTargetId: 'a', confirmDelete });
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: '削除' }));
    await waitFor(() => expect(mockShowToast).toHaveBeenCalledWith('error', 'ノートの削除に失敗しました'));
  });

  it('選択中はタイトル入力と本文エディタと保存状態を表示する', () => {
    setup({ selectedId: 'a' }, { editTitle: 'タイトルX', saveStatus: 'saved' });
    renderPage();
    expect(screen.getByLabelText('ノートのタイトル')).toHaveValue('タイトルX');
    expect(screen.getByTestId('rich-text-editor')).toBeInTheDocument();
    expect(screen.getByTestId('save-status')).toHaveTextContent('saved');
  });

  it('doc 読み込み中はローディング(role=status)を出す', () => {
    setup({ selectedId: 'a' }, { loadingDoc: true });
    renderPage();
    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.queryByLabelText('ノートのタイトル')).not.toBeInTheDocument();
  });

  it('本文取得に失敗すると再読み込み UI を出し、エディタは出さない', () => {
    const reload = vi.fn();
    setup({ selectedId: 'a' }, { loadError: true, reload });
    renderPage();
    expect(screen.getByText('本文の取得に失敗しました')).toBeInTheDocument();
    expect(screen.queryByLabelText('ノートのタイトル')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '再読み込み' }));
    expect(reload).toHaveBeenCalled();
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

  it('RichTextEditor に画像アップロード（ImageUploadRepository）を配線する', async () => {
    hoisted.upload.mockResolvedValue('https://cdn.example.com/a.png');
    setup({ selectedId: 'a' });
    renderPage();
    const onImageUpload = hoisted.rteProps.current?.onImageUpload;
    expect(onImageUpload).toBeTypeOf('function');
    await expect(onImageUpload!(new File(['x'], 'a.png', { type: 'image/png' }))).resolves.toBe(
      'https://cdn.example.com/a.png',
    );
    expect(hoisted.upload).toHaveBeenCalled();
  });

  it('画像アップロード失敗でエラートーストを出す', async () => {
    hoisted.upload.mockRejectedValue(new Error('boom'));
    setup({ selectedId: 'a' });
    renderPage();
    const onImageUpload = hoisted.rteProps.current?.onImageUpload;
    await expect(onImageUpload!(new File(['x'], 'a.png', { type: 'image/png' }))).rejects.toThrow();
    expect(mockShowToast).toHaveBeenCalledWith('error', '画像のアップロードに失敗しました');
  });
});
