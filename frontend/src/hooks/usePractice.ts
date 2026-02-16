import { useState, useCallback } from 'react';
import { classifyApiError } from '../utils/classifyApiError';
import PracticeRepository, {
  PracticeScenario,
  PracticeSession,
  CreatePracticeSessionRequest,
} from '../repositories/PracticeRepository';

/**
 * 練習モードフック
 *
 * <p>役割:</p>
 * <ul>
 *   <li>練習シナリオ・セッション管理</li>
 *   <li>PracticeRepositoryを使用してAPI呼び出し</li>
 * </ul>
 *
 * <p>Hooks層（Presentation Layer - Business Logic）:</p>
 * <ul>
 *   <li>コンポーネントからビジネスロジックを分離</li>
 *   <li>Repository層を使用してAPI呼び出し</li>
 * </ul>
 */
export const usePractice = () => {
  const [scenarios, setScenarios] = useState<PracticeScenario[]>([]);
  const [currentScenario, setCurrentScenario] = useState<PracticeScenario | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /**
   * シナリオ一覧を取得
   */
  const fetchScenarios = useCallback(async (): Promise<void> => {
    setLoading(true);
    setError(null);

    try {
      const data = await PracticeRepository.getScenarios();
      setScenarios(data);
    } catch (err) {
      setError(classifyApiError(err, 'シナリオ一覧の取得に失敗しました。'));
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * シナリオ詳細を取得
   */
  const fetchScenario = useCallback(async (scenarioId: number): Promise<PracticeScenario | null> => {
    setLoading(true);
    setError(null);

    try {
      const data = await PracticeRepository.getScenario(scenarioId);
      setCurrentScenario(data);
      return data;
    } catch (err) {
      setError(classifyApiError(err, 'シナリオ詳細の取得に失敗しました。'));
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * 練習セッションを作成
   */
  const createPracticeSession = useCallback(
    async (request: CreatePracticeSessionRequest): Promise<PracticeSession | null> => {
      setLoading(true);
      setError(null);

      try {
        const data = await PracticeRepository.createPracticeSession(request);
        return data;
      } catch (err) {
        setError(classifyApiError(err, '練習セッションの作成に失敗しました。'));
        return null;
      } finally {
        setLoading(false);
      }
    },
    []
  );

  return {
    scenarios,
    currentScenario,
    loading,
    error,
    fetchScenarios,
    fetchScenario,
    createPracticeSession,
  };
};
