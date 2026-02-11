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
    console.log('========== [LoginCallback] Callbackページ読み込み ==========');
    console.log('[LoginCallback] API_BASE_URL:', API_BASE_URL);
    console.log('[LoginCallback] code:', code ? code.substring(0, 20) + '...' : 'null');
    console.log('[LoginCallback] error:', error);

    if (error) {
      console.error('[LoginCallback] 認証エラー:', error);
      alert('認証エラーが発生しました。' + error);
      navigate('/login');
      return;
    }

    if (code) {
      const callbackUrl = `${API_BASE_URL}/api/auth/cognito/callback`;
      console.log('[LoginCallback] POSTリクエスト送信先:', callbackUrl);
      console.log('[LoginCallback] リクエスト設定: credentials=include, Content-Type=application/json');

      fetch(callbackUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code }),
        credentials: 'include',
      })
        .then((res) => {
          console.log('[LoginCallback] レスポンス受信:');
          console.log('   - status:', res.status);
          console.log('   - ok:', res.ok);
          console.log('   - statusText:', res.statusText);

          console.log('[LoginCallback] レスポンスヘッダー:');
          res.headers.forEach((value, key) => {
            console.log(`   - ${key}: ${value}`);
          });

          if (!res.ok) {
            throw new Error(`認証に失敗しました。Status: ${res.status}`);
          }
          return res.json();
        })
        .then((data) => {
          console.log('[LoginCallback] 認証成功:', data);
          dispatch(setAuthData());
          navigate('/');
        })
        .catch((err: Error) => {
          console.error('[LoginCallback] エラー発生:', err);
          console.error('[LoginCallback] エラータイプ:', err.name);
          console.error('[LoginCallback] エラーメッセージ:', err.message);

          if (err.message.includes('Failed to fetch') || err.name === 'TypeError') {
            console.error('[LoginCallback] ⚠️ CORSエラーの可能性があります');
            console.error('[LoginCallback] 確認事項:');
            console.error('   1. バックエンドのCORS設定でOriginが許可されているか');
            console.error('   2. ALB/CloudFrontがCORSヘッダーを削除していないか');
            console.error('   3. プリフライト(OPTIONS)リクエストが正常に処理されているか');
          }

          alert('認証に失敗しました。');
          navigate('/login');
        });
    } else {
      console.warn('[LoginCallback] codeパラメータがありません。/loginへリダイレクト');
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
