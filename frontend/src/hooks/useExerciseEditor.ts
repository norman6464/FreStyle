import { useState, useCallback, useEffect } from 'react';
import ExerciseRepository from '../repositories/ExerciseRepository';
import { MasterExercise, CodeExecutionResult } from '../types';

/**
 * useExerciseEditor — コード演習ページの状態管理フック。
 *
 * 演習選択・コード編集・実行結果・ヒント表示を管理する。
 * 当面は PHP 教材しか公開していないため、デフォルトで `language="php"` を
 * 指定して `master_exercises` から PHP 問題のみを取得する。
 *
 * 旧 `usePhpEditor` を MasterExercise (言語非依存) に切替えたもの。
 */
export function useExerciseEditor(language: string = 'php') {
  const [exercises, setExercises] = useState<MasterExercise[]>([]);
  const [selectedExercise, setSelectedExercise] = useState<MasterExercise | null>(null);
  const [code, setCode] = useState('');
  const [result, setResult] = useState<CodeExecutionResult | null>(null);
  const [running, setRunning] = useState(false);
  const [showHint, setShowHint] = useState(false);
  const [loadingExercises, setLoadingExercises] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    ExerciseRepository.listExercises(language)
      .then((list) => {
        setExercises(list);
        if (list.length > 0) {
          setSelectedExercise(list[0]);
          setCode(list[0].starterCode ?? '');
        }
      })
      .catch(() => setError('演習問題の取得に失敗しました'))
      .finally(() => setLoadingExercises(false));
  }, [language]);

  const selectExercise = useCallback((exercise: MasterExercise) => {
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
      // 選択中の問題と同じ言語で実行する（無ければ初期化時の language）。
      const lang = selectedExercise?.language ?? language;
      const out = await ExerciseRepository.execute(code, lang);
      setResult(out);
    } catch {
      setResult({ stdout: '', stderr: 'コードの実行に失敗しました。サーバーエラーが発生しました。', exitCode: 1 });
    } finally {
      setRunning(false);
    }
  }, [code, running, selectedExercise, language]);

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
