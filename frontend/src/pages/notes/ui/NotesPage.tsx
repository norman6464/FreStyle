import { useEffect } from 'react';
import { SecondaryPanel } from '@/widgets/secondary-panel';
import EmptyState from '@/shared/ui/EmptyState';
import ConfirmModal from '@/shared/ui/ConfirmModal';
import Loading from '@/shared/ui/Loading';
import { RichTextEditor, SaveStatusIndicator } from '@/shared/ui/RichTextEditor';
import NoteSortMenu from './NoteSortMenu';
import DocumentListItem from './DocumentListItem';
import {
  DocumentTextIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  Bars3Icon,
  ExclamationTriangleIcon,
} from '@heroicons/react/24/outline';
import { useDocuments } from '../model/useDocuments';
import { useDocumentEditor } from '../model/useDocumentEditor';
import { useMobilePanelState } from '@/shared/lib/hooks/useMobilePanelState';
import { useToast } from '@/shared/lib/hooks/useToast';
import { useNoteKeyboardShortcuts } from '../model/useNoteKeyboardShortcuts';

export default function NotesPage() {
  const { showToast } = useToast();
  const { isOpen: mobilePanelOpen, open: openMobilePanel, close: closeMobilePanel } = useMobilePanelState();
  const {
    documents,
    filteredDocuments,
    selectedId,
    loading,
    error,
    searchQuery,
    setSearchQuery,
    sort,
    setSort,
    fetchDocuments,
    createDocument,
    selectDocument,
    syncSummary,
    deleteTargetId,
    requestDelete,
    confirmDelete,
    cancelDelete,
  } = useDocuments();

  useEffect(() => {
    fetchDocuments();
  }, [fetchDocuments]);

  const {
    editTitle,
    editDoc,
    saveStatus,
    loadingDoc,
    loadError,
    handleTitleChange,
    handleDocChange,
    forceSave,
    reload,
  } = useDocumentEditor(selectedId, {
    onSynced: syncSummary,
    onConflict: () => showToast('info', '他の場所で更新されたため、最新版を読み込みました'),
  });

  const handleCreate = async () => {
    const created = await createDocument('無題');
    showToast(created ? 'success' : 'error', created ? 'ノートを作成しました' : 'ノートの作成に失敗しました');
    closeMobilePanel();
  };

  const handleConfirmDelete = async () => {
    const ok = await confirmDelete();
    showToast(ok ? 'success' : 'error', ok ? 'ノートを削除しました' : 'ノートの削除に失敗しました');
  };

  useNoteKeyboardShortcuts({ onCreateNote: handleCreate, onForceSave: forceSave });

  const handleSelect = (id: string) => {
    selectDocument(id);
    closeMobilePanel();
  };

  return (
    <div className="flex h-full">
      <SecondaryPanel
        title="ノート"
        badge={`${documents.length}件`}
        mobileOpen={mobilePanelOpen}
        onMobileClose={closeMobilePanel}
        headerContent={
          <div className="space-y-2">
            <div className="relative">
              <MagnifyingGlassIcon className="w-4 h-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" />
              <input
                type="text"
                placeholder="ノートを検索..."
                aria-label="ノートを検索"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-8 pr-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none focus:border-brand-400 transition-colors"
              />
            </div>
            <NoteSortMenu selected={sort} onChange={setSort} />
            <button
              onClick={handleCreate}
              className="w-full bg-brand-500 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-brand-600 transition-colors flex items-center justify-center gap-2"
            >
              <PlusIcon className="w-4 h-4" />
              新しいノート
            </button>
          </div>
        }
      >
        <div className="p-2 space-y-0.5">
          {loading && documents.length === 0 ? (
            <Loading className="py-8" />
          ) : error && documents.length === 0 ? (
            // 取得失敗を「ノートがありません」と誤表示しない（既存ノートの喪失と誤解させない）。
            <div className="py-12">
              <EmptyState
                icon={ExclamationTriangleIcon}
                title="ノートの取得に失敗しました"
                description="時間をおいて再読み込みしてください。"
                action={{ label: '再読み込み', onClick: fetchDocuments }}
              />
            </div>
          ) : filteredDocuments.length === 0 ? (
            <div className="py-12">
              <EmptyState
                icon={DocumentTextIcon}
                title={searchQuery ? '該当するノートがありません' : 'ノートがありません'}
                description={searchQuery ? '検索条件を変更してみてください' : '新しいノートを作成しましょう'}
                action={!searchQuery ? { label: '新しいノート', onClick: handleCreate } : undefined}
              />
            </div>
          ) : (
            <ul className="space-y-0.5">
              {filteredDocuments.map((doc) => (
                <DocumentListItem
                  key={doc.id}
                  id={doc.id}
                  title={doc.title}
                  updatedAt={doc.updatedAt}
                  isActive={selectedId === doc.id}
                  onSelect={handleSelect}
                  onDelete={requestDelete}
                />
              ))}
            </ul>
          )}
        </div>
      </SecondaryPanel>

      <div className="flex-1 flex flex-col min-w-0">
        {/* モバイルヘッダー */}
        <div className="md:hidden bg-surface-1 border-b border-surface-3 px-4 py-2 flex items-center">
          <button
            onClick={openMobilePanel}
            className="p-1.5 hover:bg-surface-2 rounded transition-colors"
            aria-label="ノート一覧を開く"
          >
            <Bars3Icon className="w-5 h-5 text-[var(--color-text-muted)]" />
          </button>
          <span className="ml-2 text-xs text-[var(--color-text-muted)]">ノート一覧</span>
        </div>

        {selectedId ? (
          loadingDoc ? (
            <Loading className="py-16" />
          ) : loadError ? (
            // 本文取得の失敗時は空の編集可能エディタを出さない（誤った 409 を誘発しないため）。
            <EmptyState
              icon={ExclamationTriangleIcon}
              title="本文の取得に失敗しました"
              description="通信状況を確認して再読み込みしてください。"
              action={{ label: '再読み込み', onClick: reload }}
            />
          ) : (
            <div className="flex flex-1 flex-col min-h-0 overflow-y-auto">
              <div className="flex items-center gap-3 px-6 pt-6">
                <input
                  type="text"
                  value={editTitle}
                  onChange={(e) => handleTitleChange(e.target.value)}
                  placeholder="無題"
                  aria-label="ノートのタイトル"
                  className="min-w-0 flex-1 bg-transparent text-2xl font-bold text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none"
                />
                <SaveStatusIndicator status={saveStatus} />
              </div>
              <div className="px-4 pb-10 pt-2">
                <RichTextEditor
                  value={editDoc}
                  onChange={handleDocChange}
                  ariaLabel="ノート本文"
                  placeholder="本文を入力…"
                />
              </div>
            </div>
          )
        ) : (
          <EmptyState
            icon={DocumentTextIcon}
            title="ノートを選択してください"
            description="左のリストからノートを選択するか、新しいノートを作成しましょう。"
            action={{ label: '新しいノート', onClick: handleCreate }}
          />
        )}
      </div>

      <ConfirmModal
        isOpen={deleteTargetId !== null}
        message="このノートを削除しますか？"
        onConfirm={handleConfirmDelete}
        onCancel={cancelDelete}
      />
    </div>
  );
}
