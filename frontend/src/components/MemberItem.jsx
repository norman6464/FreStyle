import { useNavigate } from 'react-router-dom';

export default function MemberItem({ id, name, roomId, token }) {
  const navigate = useNavigate();

  const handleClick = async () => {
    try {
      // --- ① 既存ルームがある場合 ---
      if (roomId) {
        console.log(`✅ 既存ルームあり: roomId = ${roomId}`);
        navigate(`/chat/users/${roomId}`);
        return;
      }

      // --- ② ルームが存在しない場合は新規作成 ---
      console.log(`🆕 新規ルーム作成リクエスト送信: userId = ${id}`);

      const res = await fetch(
        `http://localhost:8080/api/chat/users/${id}/create`,
        {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      );

      // --- 認証切れ ---
      if (res.status === 401) {
        navigate('/login');
        return;
      }

      // --- 失敗処理 ---
      if (!res.ok) {
        const errText = await res.text();
        throw new Error(`ルーム作成に失敗しました: ${errText}`);
      }

      const data = await res.json();
      console.log('🆗 ルーム作成成功:', data);

      // --- ③ 新しいルームIDが返ってきたらチャットへ遷移 ---
      if (data.roomId) {
        console.log(`➡️ チャット画面へ遷移: /chat/${data.roomId}`);
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
