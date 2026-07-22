import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useToast } from '@/shared/lib/hooks/useToast';

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
  // セッション操作(新規 / 削除 / タイトル変更)の結果をトーストで知らせる。
  // 他画面(教材・ノート・プロフィール等)と同じ通知体験に揃える(FRESTYLE-151)。
  const { showToast } = useToast();
  const [currentSessionId, setCurrentSessionId] = useState<number | null>(null);
  const [deleteModal, setDeleteModal] = useState<DeleteModal>({ isOpen: false, sessionId: null });
  const [editingSessionId, setEditingSessionId] = useState<number | null>(null);
  const [editingTitle, setEditingTitle] = useState('');

  const handleNewSession = useCallback(() => {
    setCurrentSessionId(null);
    navigate('/chat/ask-ai');
    showToast('success', '新しいチャットを開始しました');
  }, [navigate, showToast]);

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
      if (success) {
        showToast('success', 'セッションを削除しました');
        if (currentSessionId === sessionId) {
          setCurrentSessionId(null);
          navigate('/chat/ask-ai');
        }
      } else {
        showToast('error', 'セッションの削除に失敗しました');
      }
    }
  }, [deleteModal.sessionId, deleteSession, currentSessionId, navigate, showToast]);

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
      showToast('success', 'タイトルを変更しました');
    } else {
      showToast('error', 'タイトルの変更に失敗しました');
    }
  }, [editingTitle, updateSessionTitle, showToast]);

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
