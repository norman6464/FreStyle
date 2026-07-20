import { useCallback, useEffect, useRef, useState, useMemo } from 'react';
import ExerciseRepository from '../repositories/ExerciseRepository';
import { MasterExerciseWithStatus } from '../types';

const PAGE_SIZE = 20;

/**
 * useExerciseList — 演習問題リストページの状態管理フック。
 *
 * 対象言語は **引数（= URL の `/code-editor/lang/:language`）が正**（FRESTYLE-152）。
 * 以前はページ内のチップで切り替えて localStorage に保存していたが、言語選択カード画面を
 * 入口にしたため、選択状態は URL だけが持つ（戻る / 共有 / リロードでもぶれない）。
 *
 * スクロール型ページネーション（無限スクロール）対応。
 * language が変わると蓄積リストをリセットしてページ先頭から再取得する。
 * loadMore() を呼ぶと次ページを取得して items に追記する。
 */
export function useExerciseList(language: string) {
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
