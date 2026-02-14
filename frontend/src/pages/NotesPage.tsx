import { useState, useEffect } from 'react';
import SecondaryPanel from '../components/layout/SecondaryPanel';
import NoteListItem from '../components/NoteListItem';
import NoteEditor from '../components/NoteEditor';
import EmptyState from '../components/EmptyState';
import { DocumentTextIcon, PlusIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import { useNotes } from '../hooks/useNotes';
import { useNoteEditor } from '../hooks/useNoteEditor';

export default function NotesPage() {
  const [mobilePanelOpen, setMobilePanelOpen] = useState(false);
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
    setMobilePanelOpen(false);
  };

  const handleSelectNote = (noteId: string) => {
    selectNote(noteId);
    setMobilePanelOpen(false);
  };

  const handleDeleteNote = async (noteId: string) => {
    await deleteNote(noteId);
  };

  return (
    <div className="flex h-full">
      <SecondaryPanel
        title="ノート"
        mobileOpen={mobilePanelOpen}
        onMobileClose={() => setMobilePanelOpen(false)}
        headerContent={
          <div className="space-y-2">
            <div className="relative">
              <MagnifyingGlassIcon className="w-4 h-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" />
              <input
                type="text"
                placeholder="ノートを検索..."
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
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary-500" />
            </div>
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
            onClick={() => setMobilePanelOpen(true)}
            className="p-1.5 hover:bg-surface-2 rounded transition-colors"
            aria-label="ノート一覧を開く"
          >
            <svg className="w-5 h-5 text-[var(--color-text-muted)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <span className="ml-2 text-xs text-[var(--color-text-muted)]">ノート一覧</span>
        </div>

        {selectedNote ? (
          <NoteEditor
            title={editTitle}
            content={editContent}
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
    </div>
  );
}
