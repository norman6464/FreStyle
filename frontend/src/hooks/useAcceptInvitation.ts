import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import axios from 'axios';
import invitationRepository, {
  ValidatedInvitation,
} from '../repositories/InvitationRepository';
import { saveInvitationToken, clearInvitationToken } from '../lib/invitationToken';

export type AcceptInvitationStatus =
  | { kind: 'loading' }
  | { kind: 'invalid' }
  | { kind: 'ready'; invitation: ValidatedInvitation }
  | { kind: 'error' };

/**
 * AcceptInvitationPage の状態管理。
 *
 * URL クエリ ?token=<UUID> を取り出し、backend (`GET /invitations/accept/:token`) を叩いて検証する。
 * 結果に応じて以下の状態に遷移する:
 *
 *   - loading : リクエスト中
 *   - invalid : token が空・404 (期限切れ・受諾済・存在しない)
 *   - ready   : 招待データ取得成功 → ボタンでログインへ進める
 *   - error   : ネットワーク/サーバ障害（再試行を促す）
 *
 * ready に入ったタイミングで token を sessionStorage に保存し、ログイン後の callback で
 * pickup できるようにする。token が invalid の場合は念のため既存値もクリアする。
 */
export function useAcceptInvitation() {
  const [params] = useSearchParams();
  const token = params.get('token') || '';
  const [status, setStatus] = useState<AcceptInvitationStatus>({ kind: 'loading' });

  useEffect(() => {
    if (!token) {
      clearInvitationToken();
      setStatus({ kind: 'invalid' });
      return;
    }
    let cancelled = false;
    invitationRepository
      .validateToken(token)
      .then((invitation) => {
        if (cancelled) return;
        saveInvitationToken(token);
        setStatus({ kind: 'ready', invitation });
      })
      .catch((err) => {
        if (cancelled) return;
        clearInvitationToken();
        // 404 は「無効/期限切れリンク」として明示。それ以外はネットワーク等の障害扱い。
        if (axios.isAxiosError(err) && err.response?.status === 404) {
          setStatus({ kind: 'invalid' });
          return;
        }
        setStatus({ kind: 'error' });
      });
    return () => {
      cancelled = true;
    };
  }, [token]);

  return { status };
}
