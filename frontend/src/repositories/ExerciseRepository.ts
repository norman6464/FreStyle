import api from '../lib/axios';
import { EXERCISES, CODE } from '../constants/apiRoutes';
import { MasterExercise, CodeExecutionResult } from '../types';

/**
 * 運営マスタ演習問題 + コード実行 API のラッパー。
 *
 * 旧 `PhpRepository` を「言語非依存」に汎用化したもの。当面 PHP 教材しか
 * 公開していないので listExercises は language="php" を query で渡す呼び出しが
 * 大半だが、API は language を引数で受け取って多言語対応の準備をしている。
 */
const ExerciseRepository = {
  /**
   * 公開中の演習問題一覧。`language` で絞り込み（空文字なら全言語）。
   */
  async listExercises(language?: string): Promise<MasterExercise[]> {
    const url = language
      ? `${EXERCISES.list}?language=${encodeURIComponent(language)}`
      : EXERCISES.list;
    const res = await api.get<MasterExercise[]>(url);
    return res.data;
  },

  async getExercise(id: number): Promise<MasterExercise> {
    const res = await api.get<MasterExercise>(EXERCISES.byId(id));
    return res.data;
  },

  async execute(code: string, language = 'php'): Promise<CodeExecutionResult> {
    const res = await api.post<CodeExecutionResult>(CODE.execute, { code, language });
    return res.data;
  },
};

export default ExerciseRepository;
