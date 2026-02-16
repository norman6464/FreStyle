import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

interface DeleteModal {
  isOpen: boolean;
  sessionId: number | null;
}

interface UseAiSessionProps {
  deleteSession: (sessionId: number) => Promise<boolean>;
  updateSessionTitle: (sessionId: number, data: { title: string }) => Promise<boolean>;
}

export function useAiSession({ deleteSession, updateSessionTitle }: UseAiSessionProps) {
  const navigate = useNavigate();
  const [currentSessionId, setCurrentSessionId] = useState<number | null>(null);
  const [deleteModal, setDeleteModal] = useState<DeleteModal>({ isOpen: false, sessionId: null });
  const [editingSessionId, setEditingSessionId] = useState<number | null>(null);
  const [editingTitle, setEditingTitle] = useState('');

  const handleNewSession = useCallback(() => {
    setCurrentSessionId(null);
    navigate('/chat/ask-ai');
  }, [navigate]);

  const handleSelectSession = useCallback((sessionId: number) => {
    setCurrentSessionId(sessionId);
    navigate(`/chat/ask-ai/${sessionId}`);
  }, [navigate]);

  const handleDeleteSession = useCallback((sessionId: number) => {
    setDeleteModal({ isOpen: true, sessionId });
  }, []);

  const confirmDeleteSession = useCallback(async () => {
    const sessionId = deleteModal.sessionId;
    setDeleteModal({ isOpen: false, sessionId: null });

    if (sessionId) {
      const success = await deleteSession(sessionId);
      if (success && currentSessionId === sessionId) {
        setCurrentSessionId(null);
        navigate('/chat/ask-ai');
      }
    }
  }, [deleteModal.sessionId, deleteSession, currentSessionId, navigate]);

  const cancelDeleteSession = useCallback(() => {
    setDeleteModal({ isOpen: false, sessionId: null });
  }, []);

  const handleStartEditTitle = useCallback((session: { id: number; title?: string }) => {
    setEditingSessionId(session.id);
    setEditingTitle(session.title || '');
  }, []);

  const handleSaveTitle = useCallback(async (sessionId: number) => {
    if (!editingTitle.trim()) {
      return;
    }

    const success = await updateSessionTitle(sessionId, { title: editingTitle.trim() });
    if (success) {
      setEditingSessionId(null);
    }
  }, [editingTitle, updateSessionTitle]);

  const handleCancelEditTitle = useCallback(() => {
    setEditingSessionId(null);
    setEditingTitle('');
  }, []);

  return {
    currentSessionId,
    setCurrentSessionId,
    deleteModal,
    editingSessionId,
    editingTitle,
    setEditingTitle,
    handleNewSession,
    handleSelectSession,
    handleDeleteSession,
    confirmDeleteSession,
    cancelDeleteSession,
    handleStartEditTitle,
    handleSaveTitle,
    handleCancelEditTitle,
  };
}
