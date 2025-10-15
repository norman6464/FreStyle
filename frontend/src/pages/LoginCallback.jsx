import { useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { accessTokenState } from '../recoil/authAtom';
import { useSetRecoilState } from 'recoil';

export default function LoginCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const setAccessToken = useSetRecoilState(accessTokenState);

  useEffect(() => {
    const code = searchParams.get('code');
    const error = searchParams.get('error');

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
          if (!res.ok) throw new Error('認証に失敗');
          return res.json();
        })
        .then((data) => {
          const token = data.access_token;
          setAccessToken(token); // Recoilに保存
          navigate('/');
        })
        .catch(() => {
          alert('認証に失敗しました。');
          navigate('/login');
        });
    } else {
      // codeパラメーターがなければログイン画面へ戻す
      navigate('/login');
    }
  }, [searchParams, navigate]);

  return <div className="text-center mt-20">認証処理中...</div>;
}
