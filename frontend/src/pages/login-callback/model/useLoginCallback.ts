import { useEffect } from 'react';
import { useAppDispatch } from '@/shared/lib/store';
import { useSearchParams, useNavigate } from 'react-router-dom';

import { setAuthData } from '@/entities/user';
import { AuthRepository as authRepository } from '@/entities/user';
import { consumeAuthFlowState } from '@/features/auth';
import { consumeInvitationToken } from '@/shared/lib/invitationToken';
import { setAuthHint } from '@/shared/lib/authHint';
import { classifyApiError, getApiError } from '@/shared/lib/classifyApiError';

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

    // 招待マジックリンク経由なら、AcceptInvitationPage が置いた token を渡す（使い切り）。
    const invitationToken = consumeInvitationToken();
    // 認可コードの交換が済んだか。済んでいれば認証は成立しているので、以降の失敗で
    // ログイン画面へ戻してはいけない（ログインできたのに戻される、という挙動になる）。
    let exchanged = false;

    authRepository
      .callback({
        code,
        codeVerifier: flow.codeVerifier,
        nonce: flow.nonce,
        invitationToken,
      })
      .then(() => {
        exchanged = true;
        // callback の応答はロールを含まない。ロールを確定させずに遷移すると、
        // ダッシュボードが「管理者ではない」と誤判定して学習者向け画面を一瞬描画する。
        return authRepository.probeCurrentUser();
      })
      .then((me) => {
        dispatch(setAuthData({ isAdmin: !!me.isAdmin, role: me.role }));
        setAuthHint();
        navigate('/');
      })
      .catch((err) => {
        if (exchanged) {
          // ロールを引けなかっただけ。フル読み込みで AuthInitializer に確定させる
          // （SPA 内 navigate では再取得されず、ロール未確定のまま留まるため）。
          // replace で遷移し、使用済みの認可コードを含む URL を履歴に残さない。
          setAuthHint();
          window.location.replace('/');
          return;
        }
        const { status, serverCode, serverMessage } = getApiError(err);
        if (status === 403 && serverCode === 'invitation_required') {
          navigate('/login', {
            state: {
              toast: serverMessage || 'FreStyle のご利用には管理者からの招待が必要です。',
            },
          });
          return;
        }
        navigate('/login', {
          state: { toast: classifyApiError(err, '認証に失敗しました') },
        });
      });
  }, [code, returnedState, error, dispatch, navigate]);
}
