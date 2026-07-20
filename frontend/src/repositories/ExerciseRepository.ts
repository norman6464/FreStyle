import api from '@/shared/api/axios';
import { EXERCISES, CODE } from '@/shared/config/apiRoutes';
import {
  ExercisePage,
  ExerciseLanguageSummary,
  MasterExerciseDetail,
  CodeExecutionResult,
  ExerciseSubmitResult,
  ExerciseSubmission,
} from '../types';

/**
 * 運営マスタ演習問題 + コード実行 + 提出 API のラッパー。
 *
 * 詳細取得は slug ベース URL（PR-V）。
 * 一覧 API は current user の status / 全体集計を含む `MasterExerciseWithStatus[]` を返す（PR-W）。
 * 提出 API はテストケース全件を採点して履歴に保存する。
 */
const ExerciseRepository = {
  async listExercises(language?: string, offset = 0, limit = 20): Promise<ExercisePage> {
    const params = new URLSearchParams();
    if (language) params.set('language', language);
    params.set('offset', String(offset));
    params.set('limit', String(limit));
    const url = `${EXERCISES.list}?${params.toString()}`;
    const res = await api.get<ExercisePage>(url);
    return res.data;
  },

  /** 言語別の問題数 / 正解済み件数を取得する（言語選択カード用・FRESTYLE-152）。 */
  async listLanguageSummary(): Promise<ExerciseLanguageSummary[]> {
    const res = await api.get<ExerciseLanguageSummary[]>(EXERCISES.summary);
    return Array.isArray(res.data) ? res.data : [];
  },

  async getDetail(slug: string): Promise<MasterExerciseDetail> {
    const res = await api.get<MasterExerciseDetail>(EXERCISES.bySlug(slug));
    return res.data;
  },

  async execute(code: string, language = 'php'): Promise<CodeExecutionResult> {
    const res = await api.post<CodeExecutionResult>(CODE.execute, { code, language });
    return res.data;
  },

  /**
   * エディタ入場時に実行環境を事前ウォームアップする（Go はコンパイルキャッシュ、php/bash は no-op）。
   * 実行時に立ち上げるのではなく、入場時に warm にしておき最初の Run を即時化する。
   */
  async warmup(language: string): Promise<void> {
    await api.post(CODE.warmup, { language });
  },

  async submit(slug: string, code: string): Promise<ExerciseSubmitResult> {
    const res = await api.post<ExerciseSubmitResult>(EXERCISES.submit(slug), { code });
    return res.data;
  },

  async listSubmissions(slug: string): Promise<ExerciseSubmission[]> {
    const res = await api.get<ExerciseSubmission[]>(EXERCISES.submissions(slug));
    return res.data;
  },
};

export default ExerciseRepository;
