import { useEffect, useRef } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { SecondaryPanel } from '@/widgets/secondary-panel';
import EmptyState from '@/shared/ui/EmptyState';
import ConfirmModal from '@/shared/ui/ConfirmModal';
import Loading from '@/shared/ui/Loading';
import { RichTextEditor, SaveStatusIndicator } from '@/shared/ui/RichTextEditor';
import { ImageUploadRepository } from '@/entities/user';
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
  // URL（/notes/:noteId）と選択状態を双方向に同期する（AI チャットの /chat/ask-ai/:sessionId と同じ流儀）。
  const { noteId } = useParams<{ noteId: string }>();
  const navigate = useNavigate();
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

  // URL 起点で選択を変えたときに立てる旗。選択→URL の Effect が同じ変更を URL へ
  // 書き戻さないようにする（戻る/進むで古い selectedId を書き戻さないための番人）。
  const urlDrivenSelectRef = useRef(false);

  // URL → 選択: /notes/:noteId で開いたとき（またはブラウザの戻る/進む）その id を選択する。
  // 取得前に選択しておくことで、取得後の先頭自動選択に上書きされない。
  // /notes（id なし）へ遷移してきた場合は、選択中ノートの URL へ replace して表示と URL の不一致を残さない。
  useEffect(() => {
    if (noteId) {
      if (noteId !== selectedId) {
        urlDrivenSelectRef.current = true;
        selectDocument(noteId);
      }
    } else if (selectedId) {
      navigate(`/notes/${selectedId}`, { replace: true });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- URL(noteId) 起点のみで動かす（selectedId での再実行は選択→URL 側が担う）
  }, [noteId, selectDocument, navigate]);

  // 選択 → URL: 選択が変わったら URL を追随させる。/notes（id なし）からの自動選択は
  // 履歴を汚さないよう replace、ユーザー操作による選択替えは push で戻る/進むを効かせる。
  useEffect(() => {
    if (urlDrivenSelectRef.current) {
      // URL 起点の選択変更は URL が既に正なので書き戻さない。
      urlDrivenSelectRef.current = false;
      return;
    }
    if (selectedId && selectedId !== noteId) {
      navigate(`/notes/${selectedId}`, { replace: noteId === undefined });
    } else if (!selectedId && noteId) {
      // 最後のノートを削除した等で選択が消えたら一覧の素の URL へ戻す。
      navigate('/notes', { replace: true });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- 選択(selectedId) 起点のみで動かす（noteId での再実行は URL→選択側が担う）
  }, [selectedId, navigate]);

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

  // 画像は既存の presign 基盤（ノート/教材で共有）で S3 へ上げ、公開 URL を本文へ挿入する。
  const handleImageUpload = async (file: File) => {
    try {
      return await ImageUploadRepository.upload(file);
    } catch (error) {
      showToast('error', '画像のアップロードに失敗しました');
      throw error;
    }
  };

  return (
    <div className="flex h-full">
      <SecondaryPanel
        title="ノート"
        badge={`${documents.length}件`}
        peekable
        storageKey="frestyle.panel.notes"
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
            // 枠のないインライン文書。タイトルと本文を同じ中央カラムに載せて 1 つの文書として見せる。
            <div className="flex flex-1 flex-col min-h-0 overflow-y-auto">
              <div className="mx-auto w-full max-w-3xl px-6 py-10">
                <div className="mb-3 flex items-start justify-between gap-3">
                  <input
                    type="text"
                    value={editTitle}
                    onChange={(e) => handleTitleChange(e.target.value)}
                    placeholder="無題"
                    aria-label="ノートのタイトル"
                    className="min-w-0 flex-1 bg-transparent text-3xl font-bold text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none md:text-4xl"
                  />
                  <div className="shrink-0 pt-2">
                    <SaveStatusIndicator status={saveStatus} />
                  </div>
                </div>
                <RichTextEditor
                  key={selectedId}
                  value={editDoc}
                  onChange={handleDocChange}
                  onImageUpload={handleImageUpload}
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
