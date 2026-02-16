import apiClient from '../lib/axios';
import { AiSession, AiMessage, ScoreCard, ScoreHistoryItem } from '../types';

/**
 * AI Chatリポジトリ
 *
 * <p>役割:</p>
 * <ul>
 *   <li>AI Chat関連のAPI呼び出しを抽象化</li>
 *   <li>セッション、メッセージ、スコアカード管理</li>
 * </ul>
 *
 * <p>インフラ層（Infrastructure Layer）:</p>
 * <ul>
 *   <li>外部APIとの通信を担当</li>
 * </ul>
 */

export interface CreateSessionRequest {
  title: string;
  relatedRoomId?: number | null;
}

export interface UpdateSessionTitleRequest {
  title: string;
}

export interface AddMessageRequest {
  content: string;
  role: 'user' | 'assistant';
}

export interface RephraseRequest {
  originalMessage: string;
  scene?: string | null;
}

class AiChatRepository {
  /**
   * セッション一覧を取得
   */
  async getSessions(): Promise<AiSession[]> {
    const response = await apiClient.get('/api/chat/ai/sessions');
    return response.data;
  }

  /**
   * セッション詳細を取得
   */
  async getSession(sessionId: number): Promise<AiSession> {
    const response = await apiClient.get(`/api/chat/ai/sessions/${sessionId}`);
    return response.data;
  }

  /**
   * 新規セッションを作成
   */
  async createSession(request: CreateSessionRequest): Promise<AiSession> {
    const response = await apiClient.post('/api/chat/ai/sessions', request);
    return response.data;
  }

  /**
   * セッションタイトルを更新
   */
  async updateSessionTitle(sessionId: number, request: UpdateSessionTitleRequest): Promise<AiSession> {
    const response = await apiClient.put(`/api/chat/ai/sessions/${sessionId}`, request);
    return response.data;
  }

  /**
   * セッションを削除
   */
  async deleteSession(sessionId: number): Promise<void> {
    await apiClient.delete(`/api/chat/ai/sessions/${sessionId}`);
  }

  /**
   * セッション内のメッセージ一覧を取得
   */
  async getMessages(sessionId: number): Promise<AiMessage[]> {
    const response = await apiClient.get(`/api/chat/ai/sessions/${sessionId}/messages`);
    return response.data;
  }

  /**
   * メッセージを追加
   */
  async addMessage(sessionId: number, request: AddMessageRequest): Promise<AiMessage> {
    const response = await apiClient.post(`/api/chat/ai/sessions/${sessionId}/messages`, request);
    return response.data;
  }

  /**
   * メッセージの言い換え提案を取得
   */
  async rephrase(request: RephraseRequest): Promise<{ result: string }> {
    const response = await apiClient.post('/api/chat/ai/rephrase', request);
    return response.data;
  }

  /**
   * セッションのスコアカードを取得
   */
  async getScoreCard(sessionId: number): Promise<ScoreCard> {
    const response = await apiClient.get(`/api/scores/sessions/${sessionId}`);
    return response.data;
  }

  /**
   * スコア履歴を取得
   */
  async getScoreHistory(): Promise<ScoreHistoryItem[]> {
    const response = await apiClient.get('/api/scores/history');
    return response.data;
  }
}

export default new AiChatRepository();
