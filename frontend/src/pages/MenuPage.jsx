import { useNavigate } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { use, useState, useEffect } from 'react';
import HamburgerMenu from '../components/HamburgerMenu';

export default function MenuPage() {
  const navigate = useNavigate();
  const message = useSelector((state) => state.flash?.message);
  const accessToken = useSelector((state) => state.auth.accessToken);

  useEffect(() => {
    if (!accessToken) {
      navigate('/login');
    }
  }, [accessToken]);

  return (
    <>
      {message && <p className="text-green-600 text-center">{message}</p>}
      <HamburgerMenu title="メニュー" />
      <div className="min-h-screen flex flex-col items-center justify-center bg-gray-100 px-4">
        <div className="w-full max-w-md space-y-6">
          <div
            onClick={() => navigate('/profile/me')}
            className="bg-white shadow-md rounded-lg p-6 cursor-pointer hover:bg-gray-50 transition"
          >
            <h2 className="text-xl font-bold mb-2">プロフィールを編集</h2>
            <p className="text-gray-600">ユーザー情報の編集をする</p>
          </div>
          <div
            onClick={() => navigate('/chat/users')}
            className="bg-white shadow-md rounded-lg p-6 cursor-pointer hover:bg-gray-50 transition"
          >
            <h2 className="text-xl font-bold mb-2">ユーザー検索</h2>
            <p className="text-gray-600">
              メールアドレスで追加し、チャットを開始する
            </p>
          </div>

          <div
            onClick={() => navigate('/chat/ask-ai')}
            className="bg-white shadow-md rounded-lg p-6 cursor-pointer hover:bg-gray-50 transition"
          >
            <h2 className="text-xl font-bold mb-2">AIに聞いてみる</h2>
            <p className="text-gray-600">AIに質問して素早く答えを得る</p>
          </div>
        </div>
      </div>
    </>
  );
}
