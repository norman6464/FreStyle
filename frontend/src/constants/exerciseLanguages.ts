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
  /** 言語バッジの配色。淡色背景 + 濃色文字 + 枠でライトテーマでも読める。 */
  badgeClass: string;
}

// 各言語の一般的なイメージカラーに寄せる
// （Docker=青 / Go=シアン / PHP=藍紫 / Git=橙 / Bash・Linux=スレート / JavaScript=黄 / TypeScript=青）。
export const EXERCISE_LANGUAGES: ExerciseLanguageDef[] = [
  { key: 'php', label: 'PHP', badgeClass: 'bg-indigo-500/15 text-indigo-700 border-indigo-500/30' },
  { key: 'go', label: 'Go', badgeClass: 'bg-cyan-500/15 text-cyan-700 border-cyan-500/30' },
  { key: 'javascript', label: 'JavaScript', badgeClass: 'bg-yellow-500/15 text-yellow-700 border-yellow-500/30' },
  { key: 'typescript', label: 'TypeScript', badgeClass: 'bg-blue-500/15 text-blue-700 border-blue-500/30' },
  { key: 'git', label: 'Git', badgeClass: 'bg-orange-500/15 text-orange-700 border-orange-500/30' },
  { key: 'bash', label: 'Bash / Linux', badgeClass: 'bg-slate-500/15 text-slate-700 border-slate-500/30' },
  { key: 'docker', label: 'Docker', badgeClass: 'bg-sky-500/15 text-sky-700 border-sky-500/30' },
];

/** language 値（大文字小文字を無視）から定義を引く。未知は undefined。 */
export function findExerciseLanguage(language: string): ExerciseLanguageDef | undefined {
  const key = language.toLowerCase();
  return EXERCISE_LANGUAGES.find((l) => l.key === key);
}
