import { useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { setAuthData } from '../store/authSlice';
import AuthLayout from '../components/AuthLayout';

export default function LoginCallback() {
  // ReduxのaccessTokenを取得する
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  const code = searchParams.get('code');
  const error = searchParams.get('error');

  useEffect(() => {
    if (error) {
      alert('認証エラーが発生しました。' + error);
      console.log('認証エラーが発生しました。' + error);

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
          if (!res.ok) throw new Error('認証に失敗しました。');
          return res.json();
        })
        .then((data) => {
          dispatch(setAuthData(data));

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
    <AuthLayout>
      <p className="text-center">読み込み中...</p>
    </AuthLayout>
  );
}
