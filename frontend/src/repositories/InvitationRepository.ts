import apiClient from '../lib/axios';
import { INVITATIONS } from '../constants/apiRoutes';

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
  companyId: number;
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
    const response = await apiClient.get<ValidatedInvitation>(INVITATIONS.validateToken(token));
    return response.data;
  }
}

export default new InvitationRepository();
