import { useState, useCallback } from 'react';
import { SessionNoteRepository } from '../repositories/SessionNoteRepository';

export function useSessionNote(sessionId: number | null) {
  const [note, setNote] = useState<string>(() => {
    if (!sessionId) return '';
    const saved = SessionNoteRepository.get(sessionId);
    return saved?.note || '';
  });

  const saveNote = useCallback((text: string) => {
    if (!sessionId) return;
    SessionNoteRepository.save(sessionId, text);
    setNote(text);
  }, [sessionId]);

  return { note, saveNote };
}
