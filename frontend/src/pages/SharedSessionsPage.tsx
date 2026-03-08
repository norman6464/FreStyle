import { useSharedSessions } from '../hooks/useSharedSessions';
import { useNavigate } from 'react-router-dom';
import Loading from '../components/Loading';
import { SharedSession } from '../types';

function SharedSessionCard({ session }: { session: SharedSession }) {
  const navigate = useNavigate();

  return (
    <div
      className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4 space-y-2 cursor-pointer hover:border-blue-400 transition-colors"
      onClick={() => navigate(`/chat/ask-ai/${session.sessionId}`)}
      data-testid={`shared-session-${session.id}`}
    >
      <div className="flex items-center gap-2">
        <div className="w-7 h-7 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden flex-shrink-0">
          {session.userIconUrl ? (
            <img src={session.userIconUrl} alt={session.username} className="w-full h-full object-cover" />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-xs text-gray-500">
              {session.username.charAt(0)}
            </div>
          )}
        </div>
        <span className="text-sm font-medium">{session.username}</span>
        <span className="text-xs text-gray-400 ml-auto">
          {new Date(session.createdAt).toLocaleDateString('ja-JP')}
        </span>
      </div>
      <h3 className="font-semibold">{session.sessionTitle}</h3>
      {session.description && (
        <p className="text-sm text-gray-500 dark:text-gray-400">{session.description}</p>
      )}
    </div>
  );
}

export default function SharedSessionsPage() {
  const { sessions, loading, error } = useSharedSessions();

  return (
    <div className="max-w-lg mx-auto px-4 py-6 space-y-4">
      <h1 className="text-xl font-bold">みんなの会話</h1>
      <p className="text-sm text-gray-500">コミュニティで共有されたAI会話セッション</p>

      {loading && <Loading message="共有セッションを読み込み中..." />}
      {error && <p className="text-red-500 text-center">{error}</p>}

      {!loading && !error && sessions.length === 0 && (
        <p className="text-gray-400 text-center py-8">まだ共有されたセッションはありません</p>
      )}

      {!loading && !error && (
        <div className="space-y-3">
          {sessions.map((session) => (
            <SharedSessionCard key={session.id} session={session} />
          ))}
        </div>
      )}
    </div>
  );
}
