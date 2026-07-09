/**
 * 演習の言語フィルタ定義。
 *
 * key は backend の master_exercises.language 値、label は表示名。
 * 選択肢(ExerciseListPage のチップ)と有効値の検証(useExerciseList の localStorage 復元)の
 * 単一情報源にして二重管理を防ぐ(FRESTYLE-101)。
 * 'linux' という独立値は存在せず bash に統合されている(表示名で併記)。
 */
export interface ExerciseLanguageDef {
  key: string;
  label: string;
}

export const EXERCISE_LANGUAGES: ExerciseLanguageDef[] = [
  { key: 'php', label: 'PHP' },
  { key: 'go', label: 'Go' },
  { key: 'git', label: 'Git' },
  { key: 'bash', label: 'Bash / Linux' },
  { key: 'docker', label: 'Docker' },
];
