import { useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { clearAuth } from '../store/authSlice';
import { ChatBubbleLeftIcon, UserPlusIcon } from '@heroicons/react/24/solid';

interface MemberItemProps {
  id: number;
  name: string;
  roomId?: number;
  email: string;
}

export default function MemberItem({ id, name, roomId, email }: MemberItemProps) {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  const displayErrorAndRedirect = (message: string) => {
    console.error('❌ エラー発生:', message);
  };

  const handleClick = async () => {
    try {
      if (roomId) {
        console.log(`✅ 既存ルームあり: roomId = ${roomId}`);
        navigate(`/chat/users/${roomId}`);
        return;
      }

      console.log(`🆕 新規ルーム作成リクエスト送信: userId = ${id}`);

      let res = await fetch(`${API_BASE_URL}/api/chat/users/${id}/create`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });

      if (res.status === 401) {
        console.warn('⚠️ アクセストークン期限切れ。リフレッシュ試行中...');

        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          {
            method: 'POST',
            credentials: 'include',
          }
        );

        if (!refreshRes.ok) {
          console.error('❌ リフレッシュ失敗。ログインへリダイレクト。');
          dispatch(clearAuth());
          navigate('/login');
          return;
        }

        const refreshData = await refreshRes.json();
        console.log('✅ 新アクセストークン取得成功。リクエスト再試行。');

        res = await fetch(`${API_BASE_URL}/api/chat/users/${id}/create`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          credentials: 'include',
        });
      }

      if (!res.ok) {
        const errText = await res.text();
        throw new Error(`ルーム作成に失敗しました: ${errText}`);
      }

      const data = await res.json();
      console.log('🆗 ルーム作成成功:', data);

      if (data.roomId) {
        console.log(`➡️ チャット画面へ遷移: /chat/users/${data.roomId}`);
        navigate(`/chat/users/${data.roomId}`);
      } else {
        displayErrorAndRedirect('APIレスポンスに roomId が含まれていません');
      }
    } catch (err) {
      displayErrorAndRedirect(`ルーム作成中にエラー発生: ${(err as Error).message}`);
    }
  };

  return (
    <div
      onClick={handleClick}
      className="bg-white rounded-2xl cursor-pointer overflow-hidden group border border-slate-200 hover:bg-primary-50 transition-colors duration-150 mb-3"
    >
      <div className="flex items-center p-4">
        <div className="w-14 h-14 bg-primary-500 rounded-full flex items-center justify-center text-white font-bold text-xl flex-shrink-0">
          {name?.charAt(0)?.toUpperCase() || '?'}
        </div>
        
        <div className="flex-1 ml-4 min-w-0">
          <p className="text-base font-bold text-slate-800 truncate">
            {name || 'Unknown'}
          </p>
          <p className="text-sm text-slate-500 truncate">
            {email}
          </p>
        </div>
        
        <div className="flex-shrink-0 ml-3">
          {roomId ? (
            <div className="bg-primary-50 text-primary-600 px-3 py-1.5 rounded-full flex items-center gap-1.5 group-hover:bg-primary-100 transition-colors">
              <ChatBubbleLeftIcon className="w-4 h-4" />
              <span className="text-xs font-semibold">チャット</span>
            </div>
          ) : (
            <div className="bg-green-50 text-green-600 px-3 py-1.5 rounded-full flex items-center gap-1.5 group-hover:bg-green-100 transition-colors">
              <UserPlusIcon className="w-4 h-4" />
              <span className="text-xs font-semibold">追加</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
