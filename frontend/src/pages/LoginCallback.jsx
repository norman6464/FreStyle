import { useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { setAuthData } from '../store/authSlice';

export default function LoginCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const code = searchParams.get('code');
  const error = searchParams.get('error');

  useEffect(() => {
    if (error) {
      alert('認証エラーが発生しました。' + error);
      navigate('/login');
      return;
    }

    if (code) {
      fetch(`${API_BASE_URL}/api/auth/cognito/callback`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' }, // 空白は開けない
        body: JSON.stringify({ code }),
        credentials: 'include',
      })
        .then((res) => {
          if (!res.ok) {
            throw new Error('認証に失敗しました。');
          }
          return res.json();
        })
        .then(() => {
          dispatch(setAuthData());

          navigate('/');
        })
        .catch((err) => {
          alert('認証に失敗しました。');
          navigate('/login');
        });
    } else {
      navigate('/login');
    }
  }, [code, error, dispatch, navigate]);

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gradient-to-br from-gray-50 to-primary-50">
      <div className="flex flex-col items-center space-y-6">
        {/* ローディングスピナー */}
        <div className="relative w-20 h-20">
          <div className="absolute inset-0 rounded-full border-4 border-gray-200"></div>
          <div className="absolute inset-0 rounded-full border-4 border-transparent border-t-primary-600 border-r-primary-600 animate-spin"></div>
        </div>
        {/* ローディングテキスト */}
        <div className="text-center">
          <h3 className="text-2xl font-bold text-gray-800 mb-2">
            ログイン中...
          </h3>
          <p className="text-gray-600">お待たせしています</p>
        </div>
      </div>
    </div>
  );
}
