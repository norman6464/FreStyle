import apiClient from '../lib/axios';

/**
 * 練習モードリポジトリ
 *
 * <p>役割:</p>
 * <ul>
 *   <li>練習シナリオ関連のAPI呼び出しを抽象化</li>
 * </ul>
 *
 * <p>インフラ層（Infrastructure Layer）:</p>
 * <ul>
 *   <li>外部APIとの通信を担当</li>
 * </ul>
 */

export interface PracticeScenario {
  id: number;
  name: string;
  description: string;
  category: string;
  roleName: string;
  difficulty: string;
  systemPrompt: string;
}

export interface CreatePracticeSessionRequest {
  scenarioId: number;
}

class PracticeRepository {
  /**
   * シナリオ一覧を取得
   */
  async getScenarios(): Promise<PracticeScenario[]> {
    const response = await apiClient.get('/api/practice/scenarios');
    return response.data;
  }

  /**
   * シナリオ詳細を取得
   */
  async getScenario(scenarioId: number): Promise<PracticeScenario> {
    const response = await apiClient.get(`/api/practice/scenarios/${scenarioId}`);
    return response.data;
  }

  /**
   * 練習セッションを作成
   */
  async createPracticeSession(request: CreatePracticeSessionRequest): Promise<any> {
    const response = await apiClient.post('/api/practice/sessions', request);
    return response.data;
  }
}

export default new PracticeRepository();
