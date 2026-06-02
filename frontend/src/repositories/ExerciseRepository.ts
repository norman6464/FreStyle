import api from '../lib/axios';
import { EXERCISES, CODE } from '../constants/apiRoutes';
import {
  MasterExerciseWithStatus,
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
  async listExercises(language?: string): Promise<MasterExerciseWithStatus[]> {
    const url = language
      ? `${EXERCISES.list}?language=${encodeURIComponent(language)}`
      : EXERCISES.list;
    const res = await api.get<MasterExerciseWithStatus[]>(url);
    return res.data;
  },

  async getDetail(slug: string): Promise<MasterExerciseDetail> {
    const res = await api.get<MasterExerciseDetail>(EXERCISES.bySlug(slug));
    return res.data;
  },

  async execute(code: string, language = 'php'): Promise<CodeExecutionResult> {
    const res = await api.post<CodeExecutionResult>(CODE.execute, { code, language });
    return res.data;
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
