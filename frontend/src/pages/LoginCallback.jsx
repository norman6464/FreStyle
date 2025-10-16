import { useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { setAuthData } from '../store/authSlice';

export default function LoginCallback() {
  // ReduxのaccessTokenを取得する
  const accessToken = useSelector((state) => state.auth.accessToken);

  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const dispatch = useDispatch();

  const code = searchParams.get('code');
  const error = searchParams.get('error');

  useEffect(() => {
    // すでにログイン済みなら、認証処理をスキップしてホームへ
    if (accessToken) {
      console.log('ログイン済み');
      navigate('/');
      return;
    }

    console.log('認可コードフロー開始');
    if (error) {
      alert('認証エラーが発生しました。' + error);
      navigate('/login');
      return;
    }

    if (code) {
      fetch('http://localhost:8080/api/auth/cognito/callback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' }, // 空白は開けない
        body: JSON.stringify({ code }),
      })
        .then((res) => {
          console.log(res.ok);
          if (!res.ok) throw new Error('認証に失敗しました。');
          return res.json();
        })
        .then((data) => {
          const token = data.accessToken;
          console.log(data);
          console.log(token);
          dispatch(setAuthData(data));
          navigate('/');
        })
        .catch(() => {
          console.log(2);
          alert('認証に失敗しました。');
          navigate('/login');
        });
    } else {
      // codeパラメーターがなければログイン画面へ戻す
      navigate('/login');
    }
  }, [code, error, dispatch, navigate, accessToken]);

  return <div className="text-center mt-20">認証処理中...</div>;
}
