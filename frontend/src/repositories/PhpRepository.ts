import api from '../lib/axios';
import { PHP } from '../constants/apiRoutes';
import { PhpExercise, CodeExecutionResult } from '../types';

/** PHP 演習・コード実行の API ラッパー。axios の直接利用はここに集約する。 */
const PhpRepository = {
  async listExercises(): Promise<PhpExercise[]> {
    const res = await api.get<PhpExercise[]>(PHP.exercises);
    return res.data;
  },

  async getExercise(id: number): Promise<PhpExercise> {
    const res = await api.get<PhpExercise>(PHP.exercise(id));
    return res.data;
  },

  async execute(code: string, language = 'php'): Promise<CodeExecutionResult> {
    const res = await api.post<CodeExecutionResult>(PHP.execute, { code, language });
    return res.data;
  },
};

export default PhpRepository;
