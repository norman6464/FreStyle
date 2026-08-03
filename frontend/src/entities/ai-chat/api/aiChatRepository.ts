import apiClient from '@/shared/api/axios';
import axios from 'axios';
import { toArray } from '@/shared/lib/toArray';
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
    return toArray<AiSession>(response.data);
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
    return toArray<AiMessage>(response.data);
  }

  /** AI チャット添付ファイル用の S3 PUT presigned URL を取得する。 */
  async issueAttachmentUploadUrl(
    request: IssueAttachmentUploadUrlRequest
  ): Promise<AttachmentUploadUrlResponse> {
    const response = await apiClient.post(AI_CHAT.attachmentUploadUrl, request);
    return response.data;
  }

  /**
   * 発行済みの presigned URL へファイル本体を直接 PUT する。
   *
   * 送信先は S3 で自分の API ではないため、apiClient（認証 Cookie と 401 の
   * リダイレクトを持つ）ではなく素の axios を使う。ProfileRepository.uploadToS3 と
   * 同じ方針。通信そのものは画面ではなくこの層に置く（FRESTYLE-22）。
   */
  async uploadAttachment(uploadUrl: string, file: File): Promise<void> {
    await axios.put(uploadUrl, file, {
      headers: { 'Content-Type': file.type },
    });
  }
}

export default new AiChatRepository();
