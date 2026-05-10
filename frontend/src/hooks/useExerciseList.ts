import { useCallback, useEffect, useState, useMemo } from 'react';
import ExerciseRepository from '../repositories/ExerciseRepository';
import { MasterExerciseWithStatus } from '../types';

// 言語フィルタを localStorage に保存するキー。 ページ離脱 / リロードしても
// 前回選んだ言語を復元できるよう、 ブラウザに永続化する。
const LANGUAGE_STORAGE_KEY = 'frestyle:exercise-list:language';

// localStorage が許す言語コード集合 (許容されない値は default に戻す)。
const VALID_LANGUAGES = new Set(['', 'php', 'go', 'bash', 'docker']);

function loadStoredLanguage(fallback: string): string {
  if (typeof window === 'undefined') return fallback;
  try {
    const v = window.localStorage.getItem(LANGUAGE_STORAGE_KEY);
    if (v !== null && VALID_LANGUAGES.has(v)) return v;
  } catch {
    // SSR / プライベートブラウジング 等で localStorage が使えない場合は fallback。
  }
  return fallback;
}

function persistLanguage(value: string) {
  if (typeof window === 'undefined') return;
  try {
    window.localStorage.setItem(LANGUAGE_STORAGE_KEY, value);
  } catch {
    // 容量制限 / 拒否設定 で書き込めない場合は黙って無視。
  }
}

/**
 * useExerciseList — 演習問題リストページの状態管理フック。
 *
 * 一覧 API は current user のステータス（solved / in_progress / 未着手）と
 * 全ユーザ合計の集計（提出数 / 正答ユーザ数）を含むので、 リスト UI で
 * バッジ + 解答率を直接表示できる。
 *
 * 言語フィルタは UI 側の `language` state に応じて再 fetch する。
 * 空文字なら全言語を返す。 selectedLanguage は localStorage で 永続化するので、
 * 演習詳細ページから戻っても 前回の選択が復元される。
 */
export function useExerciseList(initialLanguage: string = 'php') {
  const [language, setLanguageState] = useState(() => loadStoredLanguage(initialLanguage));

  const setLanguage = useCallback((next: string) => {
    setLanguageState(next);
    persistLanguage(next);
  }, []);
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
