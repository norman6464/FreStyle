import { useState, useCallback, useEffect } from 'react';
import PhpRepository from '../repositories/PhpRepository';
import { PhpExercise, CodeExecutionResult } from '../types';

/**
 * usePhpEditor — PHP コード実行環境ページの状態管理フック。
 * 演習選択・コード編集・実行結果・ヒント表示を管理する。
 */
export function usePhpEditor() {
  const [exercises, setExercises] = useState<PhpExercise[]>([]);
  const [selectedExercise, setSelectedExercise] = useState<PhpExercise | null>(null);
  const [code, setCode] = useState('');
  const [result, setResult] = useState<CodeExecutionResult | null>(null);
  const [running, setRunning] = useState(false);
  const [showHint, setShowHint] = useState(false);
  const [loadingExercises, setLoadingExercises] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    PhpRepository.listExercises()
      .then((list) => {
        setExercises(list);
        if (list.length > 0) {
          setSelectedExercise(list[0]);
          setCode(list[0].starterCode ?? '');
        }
      })
      .catch(() => setError('演習問題の取得に失敗しました'))
      .finally(() => setLoadingExercises(false));
  }, []);

  const selectExercise = useCallback((exercise: PhpExercise) => {
    setSelectedExercise(exercise);
    setCode(exercise.starterCode ?? '');
    setResult(null);
    setShowHint(false);
  }, []);

  const runCode = useCallback(async () => {
    if (!code.trim() || running) return;
    setRunning(true);
    setResult(null);
    try {
      const out = await PhpRepository.execute(code);
      setResult(out);
    } catch {
      setResult({ stdout: '', stderr: 'コードの実行に失敗しました。サーバーエラーが発生しました。', exitCode: 1 });
    } finally {
      setRunning(false);
    }
  }, [code, running]);

  const resetCode = useCallback(() => {
    if (selectedExercise) {
      setCode(selectedExercise.starterCode ?? '');
      setResult(null);
    }
  }, [selectedExercise]);

  const categories = [...new Set(exercises.map((e) => e.category))];

  return {
    exercises,
    categories,
    selectedExercise,
    code,
    setCode,
    result,
    running,
    showHint,
    setShowHint,
    loadingExercises,
    error,
    selectExercise,
    runCode,
    resetCode,
  };
}
