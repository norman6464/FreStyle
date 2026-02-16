import { useEffect, useCallback } from 'react';

interface UseNoteKeyboardShortcutsOptions {
  onCreateNote: () => void;
  onForceSave: () => void;
}

export function useNoteKeyboardShortcuts({ onCreateNote, onForceSave }: UseNoteKeyboardShortcutsOptions) {
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      const mod = e.ctrlKey || e.metaKey;
      if (!mod) return;

      if (e.key === 'n') {
        e.preventDefault();
        onCreateNote();
      } else if (e.key === 's') {
        e.preventDefault();
        onForceSave();
      }
    },
    [onCreateNote, onForceSave]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);
}
