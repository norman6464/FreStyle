import { useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import axios from 'axios';
import { setAuthData } from '../store/authSlice';
import authRepository from '../repositories/AuthRepository';
import { consumeInvitationToken } from '../lib/invitationToken';
import { classifyApiError } from '../utils/classifyApiError';

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
          navigate('/', { state: { toast: 'ログインしました' } });
        })
        .catch((err) => {
          // backend が PR-OIDC-Gate で返す 403 invitation_required を識別して
          // 専用メッセージを表示する。それ以外は従来どおり「認証に失敗しました」。
          if (axios.isAxiosError(err) && err.response?.status === 403) {
            const data = err.response.data as { error?: string; message?: string } | undefined;
            if (data?.error === 'invitation_required') {
              navigate('/login', {
                state: {
                  toast:
                    data.message ||
                    'FreStyle のご利用には管理者からの招待が必要です。',
                },
              });
              return;
            }
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
