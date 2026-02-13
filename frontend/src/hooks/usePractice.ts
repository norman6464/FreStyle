import { useState, useCallback } from 'react';
import PracticeRepository, {
  PracticeScenario,
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
      const errorMessage =
        err instanceof Error ? err.message : 'シナリオ一覧の取得に失敗しました。';
      setError(errorMessage);
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
      const errorMessage =
        err instanceof Error ? err.message : 'シナリオ詳細の取得に失敗しました。';
      setError(errorMessage);
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * 練習セッションを作成
   */
  const createPracticeSession = useCallback(
    async (request: CreatePracticeSessionRequest): Promise<any | null> => {
      setLoading(true);
      setError(null);

      try {
        const data = await PracticeRepository.createPracticeSession(request);
        return data;
      } catch (err) {
        const errorMessage =
          err instanceof Error ? err.message : '練習セッションの作成に失敗しました。';
        setError(errorMessage);
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
