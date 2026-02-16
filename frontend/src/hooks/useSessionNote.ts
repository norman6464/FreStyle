import { useState, useCallback, useEffect } from 'react';
import { SessionNoteRepository } from '../repositories/SessionNoteRepository';

export function useSessionNote(sessionId: number | null) {
  const [note, setNote] = useState<string>('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!sessionId) {
      setNote('');
      setLoading(false);
      return;
    }
    let cancelled = false;
    SessionNoteRepository.get(sessionId).then((saved) => {
      if (!cancelled) {
        setNote(saved?.note || '');
        setLoading(false);
      }
    });
    return () => { cancelled = true; };
  }, [sessionId]);

  const saveNote = useCallback(async (text: string) => {
    if (!sessionId) return;
    await SessionNoteRepository.save(sessionId, text);
    setNote(text);
  }, [sessionId]);

  return { note, saveNote, loading };
}
