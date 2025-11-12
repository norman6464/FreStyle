import { useNavigate } from 'react-router-dom';
import { useSelector, useDispatch } from 'react-redux';
import { setAuthData, clearAuthData } from '../store/authSlice';

export default function MemberItem({ id, name, roomId }) {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const accessToken = useSelector((state) => state.auth.accessToken);
  const email = useSelector((state) => state.auth.email);

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
          `${API_BASE_URL}/api/auth/cognito/refresh-token?email=${encodeURIComponent(
            email
          )}`,
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
        console.error('❌ APIレスポンスに roomId が含まれていません:', data);
        alert('ルーム作成は成功しましたが、roomIdが取得できませんでした。');
      }
    } catch (err) {
      console.error('❌ ルーム作成中にエラー発生:', err);
      alert('ルーム作成中にエラーが発生しました。');
      navigate('/');
    }
  };

  return (
    <div
      onClick={handleClick}
      className="flex items-center bg-white p-3 rounded shadow hover:bg-gray-100 transition cursor-pointer"
    >
      <div className="w-10 h-10 bg-blue-400 rounded-full flex items-center justify-center text-white font-bold mr-4">
        {name.charAt(0).toUpperCase()}
      </div>
      <span className="text-lg font-medium">{name}</span>
    </div>
  );
}
