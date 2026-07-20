import { useEffect, useMemo, useState } from 'react';
import { ExerciseRepository } from '@/entities/exercise';
import { EXERCISE_LANGUAGES } from '@/entities/exercise';
import type { ExerciseLanguageSummary } from '@/entities/exercise';

/** 言語選択カード 1 枚ぶんの表示データ。 */
export interface ExerciseLanguageCard extends ExerciseLanguageSummary {
  /** 表示名（未知の言語は key をそのまま出す）。 */
  label: string;
}

/**
 * useExerciseLanguageSummary — コード学習の言語選択カード用データ（FRESTYLE-152）。
 *
 * `GET /exercises/summary`（言語ごとの問題数 + current user の正解済み件数）を取得し、
 * 表示名を [EXERCISE_LANGUAGES] から解決して返す。
 *
 * 並び順は EXERCISE_LANGUAGES の定義順（学習の入口として見せたい順）を優先し、
 * そこに無い言語（教材が先に増えたケース）は後ろに言語名順で続ける。
 * 問題が 0 件の言語は API が返さないのでカードにも出ない。
 */
export function useExerciseLanguageSummary() {
  const [summaries, setSummaries] = useState<ExerciseLanguageSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    ExerciseRepository.listLanguageSummary()
      .then((rows) => {
        if (!cancelled) setSummaries(rows);
      })
      .catch(() => {
        if (!cancelled) setError('演習の取得に失敗しました');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const cards = useMemo<ExerciseLanguageCard[]>(() => {
    const order = new Map(EXERCISE_LANGUAGES.map((l, i) => [l.key, i]));
    const labels = new Map(EXERCISE_LANGUAGES.map((l) => [l.key, l.label]));

    return [...summaries]
      .sort((a, b) => {
        const ai = order.get(a.language);
        const bi = order.get(b.language);
        if (ai !== undefined && bi !== undefined) return ai - bi;
        if (ai !== undefined) return -1;
        if (bi !== undefined) return 1;
        return a.language.localeCompare(b.language);
      })
      .map((s) => ({ ...s, label: labels.get(s.language) ?? s.language }));
  }, [summaries]);

  return { cards, loading, error };
}
