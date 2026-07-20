import { useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { setAuthData } from '@/entities/user';
import authRepository from '@/entities/user/api/authRepository';
import { consumeInvitationToken } from '../lib/invitationToken';
import { classifyApiError, getApiError } from '@/shared/lib/classifyApiError';

export function useLoginCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const dispatch = useDispatch();
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
      authRepository
        .callback(code, invitationToken)
        .then(() => {
          dispatch(setAuthData());
          navigate('/');
        })
        .catch((err) => {
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
