import { useEffect } from 'react';
import { useAppDispatch } from '@/shared/lib/store';
import { useSearchParams, useNavigate } from 'react-router-dom';

import { setAuthData } from '@/entities/user';
import { AuthRepository as authRepository } from '@/entities/user';
import { consumeAuthFlowState } from '@/features/auth';
import { setAuthHint } from '@/shared/lib/authHint';
import { classifyApiError } from '@/shared/lib/classifyApiError';

/**
 * 発行者のログイン画面からの戻りを処理する。
 *
 * 認可コードを交換する前に、**戻ってきた state が自分の作った値と一致するか**を確かめる。
 * 確かめないと、攻撃者が自分の認可コードを他人のブラウザに踏ませて、
 * 被害者を攻撃者のアカウントでログインさせられる（被害者が書いたものが
 * 攻撃者の手元に残る）。
 */
export function useLoginCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const code = searchParams.get('code');
  const returnedState = searchParams.get('state');
  const error = searchParams.get('error');

  useEffect(() => {
    if (error) {
      navigate('/login', { state: { toast: '認証エラーが発生しました' } });
      return;
    }
    if (!code) {
      navigate('/login');
      return;
    }

    // 認可を始めたときに置いた値を取り出す（使い切り。残すと同じ値で 2 回試せる）。
    const flow = consumeAuthFlowState();
    if (!flow) {
      navigate('/login', {
        state: { toast: 'ログインの手続きが見つかりませんでした。もう一度お試しください。' },
      });
      return;
    }
    if (!returnedState || returnedState !== flow.state) {
      navigate('/login', {
        state: { toast: 'ログインの検証に失敗しました。もう一度お試しください。' },
      });
      return;
    }

    authRepository
      .callback({
        code,
        codeVerifier: flow.codeVerifier,
        nonce: flow.nonce,
      })
      .then(() => {
        dispatch(setAuthData());
        setAuthHint();
        navigate('/');
      })
      .catch((err) => {
        navigate('/login', {
          state: { toast: classifyApiError(err, '認証に失敗しました') },
        });
      });
  }, [code, returnedState, error, dispatch, navigate]);
}
