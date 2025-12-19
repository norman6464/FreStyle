import { useNavigate } from 'react-router-dom';
import { useSelector, useDispatch } from 'react-redux';
import { setAuthData, clearAuthData } from '../store/authSlice';

export default function MemberItem({ id, name, roomId }) {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const accessToken = useSelector((state) => state.auth.accessToken);
  const email = useSelector((state) => state.auth.email);

  const displayErrorAndRedirect = (message) => {
    console.error('❌ エラー発生:', message);
    // navigate('/');
  };

  const handleClick = async () => {
    try {
      // --- ① 既存ルームがある場合 ---
      if (roomId) {
        console.log(`✅ 既存ルームあり: roomId = ${roomId}`);
        navigate(`/chat/users/${roomId}`);
        return;
      }

      console.log(`🆕 新規ルーム作成リクエスト送信: userId = ${id}`);

      // --- ② APIリクエスト ---
      let res = await fetch(`${API_BASE_URL}/api/chat/users/${id}/create`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken}`,
        },
        credentials: 'include',
      });

      // --- ③ トークン期限切れの場合、リフレッシュを試みる ---
      if (res.status === 401) {
        console.warn('⚠️ アクセストークン期限切れ。リフレッシュ試行中...');

        const refreshRes = await fetch(
          `${API_BASE_URL}/api/auth/cognito/refresh-token`,
          {
            method: 'POST',
            credentials: 'include', // Cookie送信
          }
        );

        if (!refreshRes.ok) {
          console.error('❌ リフレッシュ失敗。ログインへリダイレクト。');
          dispatch(clearAuthData());
          navigate('/login');
          return;
        }

        const refreshData = await refreshRes.json();
        const newAccessToken = refreshData.accessToken;

        if (!newAccessToken) {
          console.error('❌ 新アクセストークンが取得できませんでした。');
          dispatch(clearAuthData());
          navigate('/login');
          return;
        }

        // Reduxに保存
        dispatch(setAuthData({ accessToken: newAccessToken }));
        console.log('✅ 新アクセストークン取得成功。リクエスト再試行。');

        // --- 再試行 ---
        res = await fetch(`${API_BASE_URL}/api/chat/users/${id}/create`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${newAccessToken}`,
          },
          credentials: 'include',
        });
      }

      // --- ④ エラーハンドリング ---
      if (!res.ok) {
        const errText = await res.text();
        throw new Error(`ルーム作成に失敗しました: ${errText}`);
      }

      const data = await res.json();
      console.log('🆗 ルーム作成成功:', data);

      // --- ⑤ 新しいルームへ遷移 ---
      if (data.roomId) {
        console.log(`➡️ チャット画面へ遷移: /chat/users/${data.roomId}`);
        navigate(`/chat/users/${data.roomId}`);
      } else {
        displayErrorAndRedirect('APIレスポンスに roomId が含まれていません');
      }
    } catch (err) {
      displayErrorAndRedirect(`ルーム作成中にエラー発生: ${err.message}`);
    }
  };

  return (
    <div
      onClick={handleClick}
      className="flex items-center justify-between bg-white p-5 rounded-xl shadow-md hover:shadow-2xl hover:bg-gradient-to-r hover:from-primary-50 hover:to-secondary-50 transition-all duration-300 cursor-pointer border border-gray-100 hover:border-primary-300 transform hover:scale-105 hover:translate-y-[-2px] group"
    >
      <div className="flex items-center flex-1 min-w-0">
        <div className="w-12 h-12 bg-gradient-primary rounded-full flex items-center justify-center text-white font-bold text-lg mr-4 flex-shrink-0 shadow-md group-hover:shadow-lg transform group-hover:scale-110 transition-all duration-300">
          {name.charAt(0).toUpperCase()}
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-gray-800 text-base font-semibold break-words">
            {name}
          </p>
          <p className="text-gray-500 text-xs mt-1">チャットを開始</p>
        </div>
      </div>
      <div className="text-gray-300 group-hover:text-primary-400 transition-colors ml-4 flex-shrink-0">
        <svg
          className="w-6 h-6 transform group-hover:translate-x-1 transition-transform"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 5l7 7-7 7"
          />
        </svg>
      </div>
    </div>
  );
}
