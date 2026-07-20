/**
 * 演習の言語フィルタ定義。
 *
 * key は backend の master_exercises.language 値、label は表示名。
 * 選択肢(ExerciseListPage のチップ)と有効値の検証(useExerciseList の localStorage 復元)の
 * 単一情報源にして二重管理を防ぐ(FRESTYLE-101)。
 * 'linux' という独立値は存在せず bash に統合されている(表示名で併記)。
 */
import { LANGUAGE_BADGE_CLASSES } from '@/shared/config/languageBadgeClasses';

export interface ExerciseLanguageDef {
  key: string;
  label: string;
  /** 言語バッジの配色。色の正本は languageBadgeClasses（コースの言語バッジと共用）。 */
  badgeClass: string;
}

export const EXERCISE_LANGUAGES: ExerciseLanguageDef[] = [
  { key: 'php', label: 'PHP', badgeClass: LANGUAGE_BADGE_CLASSES.php },
  { key: 'go', label: 'Go', badgeClass: LANGUAGE_BADGE_CLASSES.go },
  { key: 'javascript', label: 'JavaScript', badgeClass: LANGUAGE_BADGE_CLASSES.javascript },
  { key: 'typescript', label: 'TypeScript', badgeClass: LANGUAGE_BADGE_CLASSES.typescript },
  { key: 'git', label: 'Git', badgeClass: LANGUAGE_BADGE_CLASSES.git },
  { key: 'bash', label: 'Bash / Linux', badgeClass: LANGUAGE_BADGE_CLASSES.bash },
  { key: 'docker', label: 'Docker', badgeClass: LANGUAGE_BADGE_CLASSES.docker },
];

