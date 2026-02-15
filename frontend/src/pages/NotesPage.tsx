import { useEffect, useState } from 'react';
import SecondaryPanel from '../components/layout/SecondaryPanel';
import NoteListItem from '../components/NoteListItem';
import NoteEditor from '../components/NoteEditor';
import EmptyState from '../components/EmptyState';
import ConfirmModal from '../components/ConfirmModal';
import Loading from '../components/Loading';
import { DocumentTextIcon, PlusIcon, MagnifyingGlassIcon, Bars3Icon } from '@heroicons/react/24/outline';
import { useNotes } from '../hooks/useNotes';
import { useNoteEditor } from '../hooks/useNoteEditor';
import { useMobilePanelState } from '../hooks/useMobilePanelState';

export default function NotesPage() {
  const { isOpen: mobilePanelOpen, open: openMobilePanel, close: closeMobilePanel } = useMobilePanelState();
  const {
    notes,
    filteredNotes,
    selectedNoteId,
    loading,
    searchQuery,
    setSearchQuery,
    fetchNotes,
    createNote,
    updateNote,
    deleteNote,
    selectNote,
    togglePin,
  } = useNotes();

  useEffect(() => {
    fetchNotes();
  }, [fetchNotes]);

  const selectedNote = notes.find((n) => n.noteId === selectedNoteId) || null;

  const { editTitle, editContent, handleTitleChange, handleContentChange } =
    useNoteEditor(selectedNoteId, selectedNote, updateNote);

  const handleCreateNote = async () => {
    await createNote('無題');
    closeMobilePanel();
  };

  const handleSelectNote = (noteId: string) => {
    selectNote(noteId);
    closeMobilePanel();
  };

  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);

  const handleDeleteNote = (noteId: string) => {
    setDeleteTargetId(noteId);
  };

  const confirmDelete = async () => {
    if (deleteTargetId) {
      await deleteNote(deleteTargetId);
      setDeleteTargetId(null);
    }
  };

  return (
    <div className="flex h-full">
      <SecondaryPanel
        title="ノート"
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
                className="w-full pl-8 pr-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none focus:border-primary-500 transition-colors"
              />
            </div>
            <button
              onClick={handleCreateNote}
              className="w-full bg-primary-500 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-primary-600 transition-colors flex items-center justify-center gap-2"
            >
              <PlusIcon className="w-4 h-4" />
              新しいノート
            </button>
          </div>
        }
      >
        <div className="p-2 space-y-0.5">
          {loading && notes.length === 0 ? (
            <Loading className="py-8" />
          ) : filteredNotes.length === 0 ? (
            <div className="p-4 text-center text-xs text-[var(--color-text-muted)]">
              ノートがありません
            </div>
          ) : (
            filteredNotes.map((note) => (
              <NoteListItem
                key={note.noteId}
                noteId={note.noteId}
                title={note.title}
                content={note.content}
                updatedAt={note.updatedAt}
                isPinned={note.isPinned}
                isActive={selectedNoteId === note.noteId}
                onSelect={handleSelectNote}
                onDelete={handleDeleteNote}
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
          <NoteEditor
            title={editTitle}
            content={editContent}
            noteId={selectedNoteId}
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
        onConfirm={confirmDelete}
        onCancel={() => setDeleteTargetId(null)}
      />
    </div>
  );
}
