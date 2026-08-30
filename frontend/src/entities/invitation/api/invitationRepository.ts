import apiClient from '@/shared/api/axios';
import { INVITATIONS } from '@/shared/config/apiRoutes';

/**
 * 招待マジックリンク受諾フロー用のリポジトリ。
 *
 * 認証不要の公開エンドポイント (`GET /api/v2/invitations/accept/:token`) を呼び出し、
 * 受諾画面が「招待先の company / role / displayName」を表示するためのデータを取得する。
 *
 * email は意図的に含まれない（token 漏洩時の被害局所化）。
 */

export interface ValidatedInvitation {
  role: string;
  displayName: string;
  workspaceId?: string;
  companyName: string;
}

/**
 * backend の生レスポンス。表示名のキーが `name` で、画面が使う `displayName` と揃っていない。
 * ここで一度だけ読み替え、画面側は `displayName` だけを見る。
 */
interface InvitationValidateResponse {
  role: string;
  name: string;
  workspaceId?: string;
  companyName: string;
}

class InvitationRepository {
  /**
   * 招待 token を検証する。
   *
   * 該当なし / 期限切れ / 受諾済 / canceled は backend から 404 で返り、
   * axios が AxiosError を throw するので呼び出し側で catch して「無効なリンク」と表示する。
   */
  async validateToken(token: string): Promise<ValidatedInvitation> {
    const response = await apiClient.get<InvitationValidateResponse>(
      INVITATIONS.validateToken(token)
    );
    const { name, ...rest } = response.data;
    return { ...rest, displayName: name };
  }
}

export default new InvitationRepository();
