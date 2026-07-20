import { useCallback, useEffect, useState } from 'react';
import { ExerciseRepository } from '@/entities/exercise';
import type { CodeExecutionResult, ExerciseSubmission, ExerciseSubmitResult, MasterExerciseDetail } from '@/entities/exercise';

/**
 * useExerciseDetail — 詳細ページの状態管理フック。
 *
 * 機能:
 *   - 問題本体 + 入出力例の取得（PR-V）
 *   - コード編集 + 実行（出力プレビュー、採点はしない）
 *   - 提出（テストケース全件採点 + 履歴に保存）（PR-W）
 *   - current user の提出履歴の取得 + 採点後に再 fetch
 */
export function useExerciseDetail(slug: string | undefined) {
  const [detail, setDetail] = useState<MasterExerciseDetail | null>(null);
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [running, setRunning] = useState(false);
  const [executionResult, setExecutionResult] = useState<CodeExecutionResult | null>(null);
  // エディタ入場時の事前ウォームアップが完了したか（UI に「実行環境 準備完了」を出す用）。
  const [warmupReady, setWarmupReady] = useState(false);

  const [submitting, setSubmitting] = useState(false);
  const [submitResult, setSubmitResult] = useState<ExerciseSubmitResult | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const [submissions, setSubmissions] = useState<ExerciseSubmission[]>([]);
  const [submissionsLoading, setSubmissionsLoading] = useState(false);

  const refreshSubmissions = useCallback(async (s: string) => {
    setSubmissionsLoading(true);
    try {
      const rows = await ExerciseRepository.listSubmissions(s);
      setSubmissions(rows);
    } catch {
      // 履歴取得は失敗してもページ自体は使えるので、 silent fail 扱い。
    } finally {
      setSubmissionsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!slug) {
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    setError(null);
    setWarmupReady(false);
    ExerciseRepository.getDetail(slug)
      .then((d) => {
        if (cancelled) return;
        setDetail(d);
        setCode(d.exercise.starterCode ?? '');
        // 言語が確定した時点で実行環境を warm にしておく（最初の Run を即時化）。
        // 失敗してもエディタは使えるので fire-and-forget の silent fail 扱い。
        ExerciseRepository.warmup(d.exercise.language)
          .then(() => {
            if (!cancelled) setWarmupReady(true);
          })
          .catch(() => {
            /* warmup 失敗は無視（実行自体は可能） */
          });
      })
      .catch(() => {
        if (!cancelled) setError('演習問題の取得に失敗しました');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    refreshSubmissions(slug);
    return () => {
      cancelled = true;
    };
  }, [slug, refreshSubmissions]);

  const runCode = useCallback(async () => {
    if (!detail || !code.trim() || running) return;
    setRunning(true);
    setExecutionResult(null);
    try {
      const out = await ExerciseRepository.execute(code, detail.exercise.language);
      setExecutionResult(out);
    } catch {
      setExecutionResult({
        stdout: '',
        stderr: 'コードの実行に失敗しました。サーバーエラーが発生しました。',
        exitCode: 1,
      });
    } finally {
      setRunning(false);
    }
  }, [detail, code, running]);

  const submitCode = useCallback(async () => {
    if (!slug || !detail || !code.trim() || submitting) return;
    setSubmitting(true);
    setSubmitResult(null);
    setSubmitError(null);
    try {
      const out = await ExerciseRepository.submit(slug, code);
      setSubmitResult(out);
      await refreshSubmissions(slug);
    } catch {
      setSubmitError('提出に失敗しました');
    } finally {
      setSubmitting(false);
    }
  }, [slug, detail, code, submitting, refreshSubmissions]);

  const resetCode = useCallback(() => {
    if (detail) {
      setCode(detail.exercise.starterCode ?? '');
      setExecutionResult(null);
      setSubmitResult(null);
      setSubmitError(null);
    }
  }, [detail]);

  return {
    detail,
    code,
    setCode,
    loading,
    error,
    running,
    executionResult,
    warmupReady,
    submitting,
    submitResult,
    submitError,
    submissions,
    submissionsLoading,
    runCode,
    submitCode,
    resetCode,
  };
}
