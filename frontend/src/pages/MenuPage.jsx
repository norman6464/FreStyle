import { useNavigate } from 'react-router-dom';

export default function MenuPage() {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-100 px-4">
      <div className="w-full max-w-md space-y-6">
        <div
          onClick={() => navigate('/chat')}
          className="bg-white shadow-md rounded-lg p-6 cursor-pointer hover:bg-gray-50 transition"
        >
          <h2 className="text-xl font-bold mb-2">チャットを開く</h2>
          <p className="text-gray-600">
            過去の会話を確認したり、新しチャットを始める
          </p>
        </div>

        <div
          onClick={() => navigate('/ask-ai')}
          className="bg-white shadow-md rounded-lg p-6 cursor-pointer hover:bg-gray-50 transition"
        >
          <h2 className="text-xl font-bold mb-2">AIに聞いてみる</h2>
          <p className="text-gray-600">AIに質問して素早く答えを得る</p>
        </div>
      </div>
    </div>
  );
}
