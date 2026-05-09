import { useEffect, useState, useMemo } from 'react';
import ExerciseRepository from '../repositories/ExerciseRepository';
import { MasterExerciseWithStatus } from '../types';

/**
 * useExerciseList — 演習問題リストページの状態管理フック。
 *
 * 一覧 API は current user のステータス（solved / in_progress / 未着手）と
 * 全ユーザ合計の集計（提出数 / 正答ユーザ数）を含むので、 リスト UI で
 * バッジ + 解答率を直接表示できる。
 *
 * 言語フィルタは UI 側の `language` state に応じて再 fetch する。
 * 空文字なら全言語を返す（PHP 以外の問題が将来追加されたとき用）。
 */
export function useExerciseList(initialLanguage: string = 'php') {
  const [language, setLanguage] = useState(initialLanguage);
  const [exercises, setExercises] = useState<MasterExerciseWithStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    ExerciseRepository.listExercises(language || undefined)
      .then((list) => {
        if (!cancelled) setExercises(list);
      })
      .catch(() => {
        if (!cancelled) setError('演習問題の取得に失敗しました');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [language]);

  const categories = useMemo(
    () => Array.from(new Set(exercises.map((e) => e.category))),
    [exercises],
  );

  return {
    language,
    setLanguage,
    exercises,
    categories,
    loading,
    error,
  };
}
