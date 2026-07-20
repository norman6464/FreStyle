import { useEffect } from 'react';
import { SecondaryPanel } from '@/widgets/secondary-panel';
import NoteListItem from '../components/NoteListItem';
import NoteMarkdownEditor from '../components/NoteMarkdownEditor';
import EmptyState from '../components/EmptyState';
import ConfirmModal from '../components/ConfirmModal';
import Loading from '../components/Loading';
import NoteSortMenu from '../components/NoteSortMenu';
import { DocumentTextIcon, PlusIcon, MagnifyingGlassIcon, Bars3Icon, QuestionMarkCircleIcon } from '@heroicons/react/24/outline';
import { Link } from 'react-router-dom';
import { useNotes } from '../hooks/useNotes';
import { useNoteEditor } from '../hooks/useNoteEditor';
import { useMobilePanelState } from '../hooks/useMobilePanelState';
import { useToast } from '../hooks/useToast';
import { useNoteKeyboardShortcuts } from '../hooks/useNoteKeyboardShortcuts';

export default function NotesPage() {
  const { showToast } = useToast();
  const { isOpen: mobilePanelOpen, open: openMobilePanel, close: closeMobilePanel } = useMobilePanelState();
  const {
    notes,
    filteredNotes,
    selectedNoteId,
    selectedNote,
    loading,
    searchQuery,
    setSearchQuery,
    fetchNotes,
    createNote,
    updateNote,
    selectNote,
    togglePin,
    deleteTargetId,
    requestDelete,
    confirmDelete,
    cancelDelete,
    noteSort,
    setNoteSort,
  } = useNotes();

  useEffect(() => {
    fetchNotes();
  }, [fetchNotes]);

  const { editTitle, editContent, saveStatus, handleTitleChange, handleContentChange, forceSave } =
    useNoteEditor(selectedNoteId, selectedNote, updateNote);

  const handleCreateNote = async () => {
    const note = await createNote('無題');
    if (note) {
      showToast('success', 'ノートを作成しました');
    } else {
      showToast('error', 'ノートの作成に失敗しました');
    }
    closeMobilePanel();
  };

  const handleConfirmDelete = async () => {
    await confirmDelete();
    showToast('success', 'ノートを削除しました');
  };

  useNoteKeyboardShortcuts({
    onCreateNote: handleCreateNote,
    onForceSave: forceSave,
  });

  const handleSelectNote = (noteId: number) => {
    selectNote(noteId);
    closeMobilePanel();
  };

  return (
    <div className="flex h-full">
      <SecondaryPanel
        title="ノート"
        badge={`${notes.length}件`}
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
            <NoteSortMenu selected={noteSort} onChange={setNoteSort} />
            <button
              onClick={handleCreateNote}
              className="w-full bg-brand-500 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-brand-600 transition-colors flex items-center justify-center gap-2"
            >
              <PlusIcon className="w-4 h-4" />
              新しいノート
            </button>
            <Link
              to="/notes/markdown-help"
              className="w-full flex items-center justify-center gap-1.5 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] py-1 transition-colors"
            >
              <QuestionMarkCircleIcon className="w-3.5 h-3.5" />
              Markdown 記法ヘルプ
            </Link>
          </div>
        }
      >
        <div className="p-2 space-y-0.5">
          {loading && notes.length === 0 ? (
            <Loading className="py-8" />
          ) : filteredNotes.length === 0 ? (
            <div className="py-12">
              <EmptyState
                icon={DocumentTextIcon}
                title={searchQuery ? '該当するノートがありません' : 'ノートがありません'}
                description={searchQuery ? '検索条件を変更してみてください' : '新しいノートを作成しましょう'}
                action={!searchQuery ? { label: '新しいノート', onClick: handleCreateNote } : undefined}
              />
            </div>
          ) : (
            filteredNotes.map((note) => (
              <NoteListItem
                key={note.id}
                noteId={note.id}
                title={note.title}
                content={note.content}
                updatedAt={note.updatedAt}
                isPinned={note.isPinned}
                isActive={selectedNoteId === note.id}
                onSelect={handleSelectNote}
                onDelete={requestDelete}
                onTogglePin={togglePin}
              />
            ))
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

        {selectedNote ? (
          <NoteMarkdownEditor
            title={editTitle}
            content={editContent}
            saveStatus={saveStatus}
            onTitleChange={handleTitleChange}
            onContentChange={handleContentChange}
          />
        ) : (
          <EmptyState
            icon={DocumentTextIcon}
            title="ノートを選択してください"
            description="左のリストからノートを選択するか、新しいノートを作成しましょう。"
            action={{ label: '新しいノート', onClick: handleCreateNote }}
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
