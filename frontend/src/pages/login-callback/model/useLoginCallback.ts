import { useEffect } from 'react';
import { useAppDispatch } from '@/shared/lib/store';
import { useSearchParams, useNavigate } from 'react-router-dom';

import { setAuthData } from '@/entities/user';
import { AuthRepository as authRepository } from '@/entities/user';
import { consumeInvitationToken } from '@/shared/lib/invitationToken';
import { setAuthHint } from '@/shared/lib/authHint';
import { classifyApiError, getApiError } from '@/shared/lib/classifyApiError';

export function useLoginCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const code = searchParams.get('code');
  const error = searchParams.get('error');

  useEffect(() => {
    if (error) {
      navigate('/login', { state: { toast: '認証エラーが発生しました' } });
      return;
    }

    if (code) {
      // 招待マジックリンク経由でログインしてきた場合、AcceptInvitationPage が保存した
      // sessionStorage の token を取り出して callback に渡す（使い切りで自動削除）。
      const invitationToken = consumeInvitationToken();
      // 認可コードの交換が済んだか。済んでいれば認証は成立しているので、以降の失敗で
      // ログイン画面へ戻してはいけない（ログインできたのに戻される、という挙動になる）。
      let exchanged = false;

      authRepository
        .callback(code, invitationToken)
        .then(() => {
          exchanged = true;
          // callback の応答は { success } のみでロールを含まない。ロールを確定させずに
          // 遷移すると、ダッシュボードが「管理者ではない」と誤判定して学習者向け画面を
          // 一瞬描画してしまう（FRESTYLE-233）。遷移前に /auth/me で確定させる。
          return authRepository.probeCurrentUser();
        })
        .then((me) => {
          dispatch(
            setAuthData({
              isAdmin: !!me.isAdmin,
              role: me.role ?? null,
              aiChatEnabledForTrainees: me.aiChatEnabledForTrainees ?? true,
            }),
          );
          setAuthHint();
          navigate('/dashboard');
        })
        .catch((err) => {
          if (exchanged) {
            // ロールを引けなかっただけ。フル読み込みで AuthInitializer に確定させる
            // （SPA 内 navigate では再取得されず、ロール未確定のまま留まるため）。
            setAuthHint();
            window.location.assign('/dashboard');
            return;
          }
          // backend が PR-OIDC-Gate で返す 403 invitation_required を識別して
          // 専用メッセージを表示する。それ以外は従来どおり「認証に失敗しました」。
          const { status, serverCode, serverMessage } = getApiError(err);
          if (status === 403 && serverCode === 'invitation_required') {
            navigate('/login', {
              state: {
                toast:
                  serverMessage ||
                  'FreStyle のご利用には管理者からの招待が必要です。',
              },
            });
            return;
          }
          navigate('/login', {
            state: { toast: classifyApiError(err, '認証に失敗しました') },
          });
        });
    } else {
      navigate('/login');
    }
  }, [code, error, dispatch, navigate]);
}
