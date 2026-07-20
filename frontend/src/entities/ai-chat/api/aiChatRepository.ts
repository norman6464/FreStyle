import apiClient from '@/shared/api/axios';
import { AI_CHAT } from '@/shared/config/apiRoutes';
import type { AiSession, AiMessage } from '../model/types';

/**
 * AI チャットリポジトリ。
 *
 * 旧版にあった addMessage（HTTP）/ rephrase / getScoreCard / getScoreHistory は
 * 廃止。メッセージ送受信は WebSocket（PR-C で SSE へ置換予定）。スコア機能は撤去。
 */

export interface CreateSessionRequest {
  title: string;
}

export interface UpdateSessionTitleRequest {
  title: string;
}

export interface IssueAttachmentUploadUrlRequest {
  filename: string;
  contentType: string;
  sizeBytes: number;
}

export interface AttachmentUploadUrlResponse {
  uploadUrl: string;
  key: string;
  expiresIn: number;
}

class AiChatRepository {
  async getSessions(): Promise<AiSession[]> {
    const response = await apiClient.get(AI_CHAT.sessions);
    return response.data;
  }

  async getSession(sessionId: number): Promise<AiSession> {
    const response = await apiClient.get(AI_CHAT.session(sessionId));
    return response.data;
  }

  async createSession(request: CreateSessionRequest): Promise<AiSession> {
    const response = await apiClient.post(AI_CHAT.sessions, request);
    return response.data;
  }

  async updateSessionTitle(sessionId: number, request: UpdateSessionTitleRequest): Promise<AiSession> {
    const response = await apiClient.put(AI_CHAT.session(sessionId), request);
    return response.data;
  }

  async deleteSession(sessionId: number): Promise<void> {
    await apiClient.delete(AI_CHAT.session(sessionId));
  }

  async getMessages(sessionId: number): Promise<AiMessage[]> {
    const response = await apiClient.get(AI_CHAT.sessionMessages(sessionId));
    return response.data;
  }

  /** AI チャット添付ファイル用の S3 PUT presigned URL を取得する。 */
  async issueAttachmentUploadUrl(
    request: IssueAttachmentUploadUrlRequest
  ): Promise<AttachmentUploadUrlResponse> {
    const response = await apiClient.post(AI_CHAT.attachmentUploadUrl, request);
    return response.data;
  }
}

export default new AiChatRepository();
