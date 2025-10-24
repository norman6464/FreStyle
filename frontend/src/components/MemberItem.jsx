import { useNavigate } from 'react-router-dom';

export default function MemberItem({ id, name, roomId, token }) {
  const navigate = useNavigate();

  const handleClick = async () => {
    try {
      // ① すでにルームが存在する場合は、そのままチャットへ遷移
      if (roomId) {
        console.log(`✅ 既存ルームあり: roomId = ${roomId}`);
        navigate(`/chat/${roomId}`);
        return;
      }

      // ② なければ新しくルームを作成する
      console.log(`🆕 新規ルーム作成: userId = ${id}`);

      const res = await fetch(
        `http://localhost:8080/api/chat/members/${id}/create`,
        {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ id }),
        }
      );

      if (res.status === 401) {
        navigate('/login');
        return;
      }

      if (!res.ok) {
        throw new Error('ルーム作成に失敗しました');
      }

      const data = await res.json();
      console.log('🆗 ルーム作成完了:', data);

      // API から新しい roomId を受け取った前提
      if (data.roomId) {
        navigate(`/chat/${data.roomId}`);
      } else {
        console.error('❌ roomId が返ってきませんでした');
      }
    } catch (err) {
      console.error('❌ エラー:', err);
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
