import { useState, useEffect, useRef } from 'react';
import MessageBubbleAi from '../components/MessageBubbleAi';
import MessageInput from '../components/MessageInput';
import ConfirmModal from '../components/ConfirmModal';
import ScoreCardComponent from '../components/ScoreCard';
import PracticeResultSummary from '../components/PracticeResultSummary';
import SecondaryPanel from '../components/layout/SecondaryPanel';
import { useNavigate, useLocation, useParams } from 'react-router-dom';
import { AiMessage, AiSession, ScoreCard } from '../types';
import { useAuth } from '../hooks/useAuth';
import { useAiChat } from '../hooks/useAiChat';
import { useWebSocket } from '../hooks/useWebSocket';

export default function AskAiPage() {
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const [currentSessionId, setCurrentSessionId] = useState<number | null>(null);
  const [initialPromptSent, setInitialPromptSent] = useState(false);
  const [deleteModal, setDeleteModal] = useState<{ isOpen: boolean; sessionId: number | null }>({
    isOpen: false,
    sessionId: null
  });
  const [editingSessionId, setEditingSessionId] = useState<number | null>(null);
  const [editingTitle, setEditingTitle] = useState('');

  const messagesEndRef = useRef<HTMLDivElement | null>(null);

  const navigate = useNavigate();
  const location = useLocation();
  const { sessionId: urlSessionId } = useParams<{ sessionId: string }>();

  const { getCurrentUser, user } = useAuth();
  const {
    sessions,
    messages,
    scoreCard,
    fetchSessions,
    fetchMessages,
    deleteSession,
    updateSessionTitle
  } = useAiChat();

  const locationState = location.state as {
    initialPrompt?: string;
    fromChatFeedback?: boolean;
    scene?: string;
    sessionType?: string;
    scenarioId?: number;
    scenarioName?: string;
  } | null;

  const initialPrompt = locationState?.initialPrompt;
  const fromChatFeedback = locationState?.fromChatFeedback || false;
  const scene = locationState?.scene || null;
  const sessionType = locationState?.sessionType || 'normal';
  const scenarioId = locationState?.scenarioId || null;
  const scenarioName = locationState?.scenarioName || null;
  const isPracticeMode = sessionType === 'practice';

  // WebSocket接続
  const { subscribe, publish } = useWebSocket({
    url: `${API_BASE_URL}/ws/ai-chat`,
    userId: user?.id || null,
    onConnect: () => {
      if (currentSessionId) {
        subscribeToSession(currentSessionId);
      }

      if (initialPrompt && !initialPromptSent) {
        handleSend(initialPrompt);
        setInitialPromptSent(true);
      }
    },
  });

  // ユーザー情報取得
  useEffect(() => {
    getCurrentUser();
  }, [getCurrentUser]);

  // セッション一覧取得
  useEffect(() => {
    if (user?.id) {
      fetchSessions();
    }
  }, [user?.id, fetchSessions]);

  // URLパラメータのセッションID変更時
  useEffect(() => {
    if (urlSessionId) {
      setCurrentSessionId(parseInt(urlSessionId));
    }
  }, [urlSessionId]);

  // セッション内のメッセージ履歴取得
  useEffect(() => {
    if (currentSessionId) {
      fetchMessages(currentSessionId);
    }
  }, [currentSessionId, fetchMessages]);

  // WebSocket購読設定
  useEffect(() => {
    if (!user?.id) return;

    // 新規セッション作成の購読
    subscribe(`/topic/ai-chat/user/${user.id}/session`, (message) => {
      const newSession = JSON.parse(message.body);
      setCurrentSessionId(newSession.id);
    });

    // スコアカード受信の購読
    subscribe(`/topic/ai-chat/user/${user.id}/scorecard`, (message) => {
      // スコアカードはuseAiChatフックで管理
    });

    // セッション削除通知の購読
    subscribe(`/topic/ai-chat/user/${user.id}/session-deleted`, (message) => {
      const data = JSON.parse(message.body);
      if (currentSessionId === data.sessionId) {
        setCurrentSessionId(null);
      }
    });
  }, [user?.id, subscribe, currentSessionId]);

  // セッション購読
  const subscribeToSession = (sessionId: number): void => {
    subscribe(`/topic/ai-chat/session/${sessionId}`, (message) => {
      // メッセージはuseAiChatフックで管理
    });
  };

  // セッション変更時の購読
  useEffect(() => {
    if (currentSessionId) {
      subscribeToSession(currentSessionId);
    }
  }, [currentSessionId]);

  // メッセージ最下部へスクロール
  const scrollToBottom = (): void => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // メッセージ削除処理
  const handleDeleteMessage = (messageId: number): void => {
    if (confirm('このメッセージを削除しますか？')) {
      // メッセージ削除はローカル状態のみ更新（サーバー側の削除は未実装）
    }
  };

  // 新規セッション作成
  const handleNewSession = (): void => {
    setCurrentSessionId(null);
    navigate('/chat/ask-ai');
  };

  // セッション選択
  const handleSelectSession = (sessionId: number): void => {
    setCurrentSessionId(sessionId);
    navigate(`/chat/ask-ai/${sessionId}`);
  };

  // セッション削除
  const handleDeleteSession = (sessionId: number): void => {
    setDeleteModal({ isOpen: true, sessionId });
  };

  const confirmDeleteSession = async (): Promise<void> => {
    const sessionId = deleteModal.sessionId;
    setDeleteModal({ isOpen: false, sessionId: null });

    if (sessionId) {
      const success = await deleteSession(sessionId);
      if (success && currentSessionId === sessionId) {
        setCurrentSessionId(null);
        navigate('/chat/ask-ai');
      }
    }
  };

  const cancelDeleteSession = (): void => {
    setDeleteModal({ isOpen: false, sessionId: null });
  };

  // セッションタイトル編集
  const handleStartEditTitle = (e: React.MouseEvent, session: AiSession): void => {
    e.stopPropagation();
    setEditingSessionId(session.id);
    setEditingTitle(session.title || '');
  };

  const handleSaveTitle = async (sessionId: number): Promise<void> => {
    if (!editingTitle.trim()) {
      setEditingSessionId(null);
      return;
    }

    const success = await updateSessionTitle(sessionId, { title: editingTitle.trim() });
    if (success) {
      setEditingSessionId(null);
    }
  };

  const handleCancelEditTitle = (): void => {
    setEditingSessionId(null);
    setEditingTitle('');
  };

  // メッセージ送信
  const handleSend = async (text: string): Promise<void> => {
    const payload: Record<string, unknown> = {
      userId: user?.id,
      sessionId: currentSessionId,
      content: text,
      role: 'user',
      fromChatFeedback: fromChatFeedback,
    };

    if (scene) {
      payload.scene = scene;
    }

    if (isPracticeMode && scenarioId) {
      payload.sessionType = 'practice';
      payload.scenarioId = scenarioId;
    }

    publish('/app/ai-chat/send', payload);
  };

  return (
    <div className="flex h-full">
      {/* セカンダリパネル: セッション一覧 */}
      <SecondaryPanel
        title="セッション"
        headerContent={
          <button
            onClick={handleNewSession}
            className="w-full bg-primary-500 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-primary-600 transition-colors flex items-center justify-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            新しいチャット
          </button>
        }
      >
        <div className="p-2 space-y-0.5">
          {sessions.map((session) => (
            <div
              key={session.id}
              className={`group flex items-center justify-between px-3 py-2.5 rounded-lg cursor-pointer transition-colors ${
                currentSessionId === session.id
                  ? 'bg-primary-50 text-primary-700'
                  : 'hover:bg-primary-50'
              }`}
              onClick={() => editingSessionId !== session.id && handleSelectSession(session.id)}
            >
              <div className="flex-1 min-w-0">
                {editingSessionId === session.id ? (
                  <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
                    <input
                      type="text"
                      value={editingTitle}
                      onChange={(e) => setEditingTitle(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') handleSaveTitle(session.id);
                        if (e.key === 'Escape') handleCancelEditTitle();
                      }}
                      className="flex-1 text-xs px-2 py-1 border border-primary-300 rounded focus:outline-none focus:ring-1 focus:ring-primary-400"
                      autoFocus
                    />
                    <button onClick={() => handleSaveTitle(session.id)} className="p-0.5 hover:bg-green-100 rounded">
                      <svg className="w-3.5 h-3.5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    </button>
                    <button onClick={handleCancelEditTitle} className="p-0.5 hover:bg-slate-200 rounded">
                      <svg className="w-3.5 h-3.5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                ) : (
                  <>
                    <p className="text-sm font-medium truncate">{session.title || '新しいチャット'}</p>
                    <p className="text-[11px] text-slate-500">
                      {session.createdAt ? new Date(session.createdAt).toLocaleDateString('ja-JP') : ''}
                    </p>
                  </>
                )}
              </div>
              {editingSessionId !== session.id && (
                <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    onClick={(e) => handleStartEditTitle(e, session)}
                    className="p-1 hover:bg-blue-100 rounded"
                    title="タイトルを編集"
                  >
                    <svg className="w-3.5 h-3.5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDeleteSession(session.id);
                    }}
                    className="p-1 hover:bg-rose-100 rounded"
                    title="削除"
                  >
                    <svg className="w-3.5 h-3.5 text-rose-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      </SecondaryPanel>

      {/* メインコンテンツ */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* 練習モードヘッダー */}
        {isPracticeMode && (
          <div className="bg-white border-b border-slate-200 px-4 py-3 flex items-center justify-between">
            <div>
              <h2 className="text-sm font-semibold text-slate-800">
                {scenarioName || '練習モード'}
              </h2>
              <p className="text-xs text-slate-500">AIが相手役を演じます</p>
            </div>
            <button
              onClick={() => handleSend('練習を終了して、今回の会話全体に対するフィードバックとスコアカードをお願いします。')}
              className="bg-rose-500 text-white text-xs px-3 py-1.5 rounded-lg hover:bg-rose-600 transition-colors"
            >
              練習終了
            </button>
          </div>
        )}

        {/* メッセージエリア */}
        <div className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
          {messages.length === 0 && (
            <div className="flex flex-col items-center justify-center h-full text-center">
              <div className="bg-slate-100 rounded-full p-4 mb-4">
                <svg className="w-8 h-8 text-slate-400" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                </svg>
              </div>
              <h3 className="text-base font-semibold text-slate-700 mb-1">
                AIアシスタントへようこそ
              </h3>
              <p className="text-sm text-slate-500 max-w-xs">
                質問や相談を何でも聞いてください
              </p>
            </div>
          )}
          {messages.map((msg) => (
            <div key={msg.id} className="max-w-3xl mx-auto w-full">
              <MessageBubbleAi
                {...msg}
                type={msg.isSender ? 'text' : 'bot'}
                onDelete={handleDeleteMessage}
              />
            </div>
          ))}

          {scoreCard && (
            <div className="max-w-3xl mx-auto w-full">
              <ScoreCardComponent scoreCard={scoreCard} />
              {isPracticeMode && (
                <PracticeResultSummary scoreCard={scoreCard} scenarioName={scenarioName || '練習'} />
              )}
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* 入力欄 */}
        <div className="bg-white border-t border-slate-200 p-4">
          <div className="max-w-3xl mx-auto w-full">
            <MessageInput onSend={handleSend} />
          </div>
        </div>
      </div>

      {/* セッション削除確認モーダル */}
      <ConfirmModal
        isOpen={deleteModal.isOpen}
        title="セッションを削除"
        message="このセッションを削除しますか？チャット履歴もすべて削除されます。この操作は取り消せません。"
        confirmText="削除する"
        cancelText="キャンセル"
        onConfirm={confirmDeleteSession}
        onCancel={cancelDeleteSession}
        isDanger={true}
      />
    </div>
  );
}
