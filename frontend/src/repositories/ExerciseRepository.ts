import api from '../lib/axios';
import { EXERCISES, CODE } from '../constants/apiRoutes';
import { MasterExercise, MasterExerciseDetail, CodeExecutionResult } from '../types';

/**
 * 運営マスタ演習問題 + コード実行 API のラッパー。
 *
 * 詳細取得は paiza 風 URL に揃えるため `:slug` ベースに変更（PR-V）。
 * `getDetail` は問題本体に加えて入出力例（テストケース）の配列も返す。
 */
const ExerciseRepository = {
  async listExercises(language?: string): Promise<MasterExercise[]> {
    const url = language
      ? `${EXERCISES.list}?language=${encodeURIComponent(language)}`
      : EXERCISES.list;
    const res = await api.get<MasterExercise[]>(url);
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
};

export default ExerciseRepository;
