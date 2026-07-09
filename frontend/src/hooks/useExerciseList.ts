import { useCallback, useEffect, useRef, useState, useMemo } from 'react';
import ExerciseRepository from '../repositories/ExerciseRepository';
import { EXERCISE_LANGUAGES } from '../constants/exerciseLanguages';
import { MasterExerciseWithStatus } from '../types';

const LANGUAGE_STORAGE_KEY = 'frestyle:exercise-list:language';
// '' = すべて。選択肢の定義(EXERCISE_LANGUAGES)から導出して二重管理を防ぐ(FRESTYLE-101)。
const VALID_LANGUAGES = new Set(['', ...EXERCISE_LANGUAGES.map((l) => l.key)]);
const PAGE_SIZE = 20;

function loadStoredLanguage(fallback: string): string {
  if (typeof window === 'undefined') return fallback;
  try {
    const v = window.localStorage.getItem(LANGUAGE_STORAGE_KEY);
    if (v !== null && VALID_LANGUAGES.has(v)) return v;
  } catch {
    // SSR / プライベートブラウジング等で localStorage が使えない場合は fallback。
  }
  return fallback;
}

function persistLanguage(value: string) {
  if (typeof window === 'undefined') return;
  try {
    window.localStorage.setItem(LANGUAGE_STORAGE_KEY, value);
  } catch {
    // 容量制限 / 拒否設定で書き込めない場合は黙って無視。
  }
}

/**
 * useExerciseList — 演習問題リストページの状態管理フック。
 *
 * スクロール型ページネーション（無限スクロール）対応。
 * 言語フィルタを切り替えると蓄積リストをリセットしてページ先頭から再取得する。
 * loadMore() を呼ぶと次ページを取得して items に追記する。
 */
export function useExerciseList(initialLanguage: string = 'php') {
  const [language, setLanguageState] = useState(() => loadStoredLanguage(initialLanguage));

  const setLanguage = useCallback((next: string) => {
    setLanguageState(next);
    persistLanguage(next);
  }, []);

  const [items, setItems] = useState<MasterExerciseWithStatus[]>([]);
  const [hasNext, setHasNext] = useState(false);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 言語変更時にリストをリセットして最初のページを再取得する。
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    setItems([]);
    setOffset(0);
    setHasNext(false);

    ExerciseRepository.listExercises(language || undefined, 0, PAGE_SIZE)
      .then((page) => {
        if (cancelled) return;
        // page.items が null/undefined のとき（旧バックエンドが配列を直接返す等）でもクラッシュしない。
        setItems(Array.isArray(page?.items) ? page.items : []);
        setHasNext(page?.hasNext ?? false);
        setOffset(PAGE_SIZE);
      })
      .catch(() => {
        if (!cancelled) setError('演習問題の取得に失敗しました');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => { cancelled = true; };
  }, [language]);

  // 次ページを追加取得する（IntersectionObserver から呼ばれる）。
  const loadMore = useCallback(() => {
    if (loadingMore || !hasNext) return;
    setLoadingMore(true);

    ExerciseRepository.listExercises(language || undefined, offset, PAGE_SIZE)
      .then((page) => {
        const newItems = Array.isArray(page?.items) ? page.items : [];
        setItems((prev) => [...prev, ...newItems]);
        setHasNext(page?.hasNext ?? false);
        setOffset((prev) => prev + PAGE_SIZE);
      })
      .catch(() => {
        setError('追加取得に失敗しました');
      })
      .finally(() => {
        setLoadingMore(false);
      });
  }, [language, offset, hasNext, loadingMore]);

  // カテゴリのユニーク一覧（出現順）。
  const categories = useMemo(
    () => Array.from(new Set(items.map((e) => e.category))),
    [items],
  );

  // sentinelRef を渡し、要素が viewport に入ったら loadMore を呼ぶ。
  const sentinelRef = useRef<HTMLDivElement | null>(null);
  useEffect(() => {
    const el = sentinelRef.current;
    if (!el || !hasNext) return;

    const obs = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) loadMore();
      },
      { rootMargin: '200px' },
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [hasNext, loadMore]);

  return {
    language,
    setLanguage,
    exercises: items,
    categories,
    loading,
    loadingMore,
    hasNext,
    error,
    sentinelRef,
    loadMore,
  };
}
