/**
 * コースの「主に扱う言語・技術」の選択肢（コース作成/編集フォーム用。FRESTYLE-114）。
 * key は courses.language に保存される値。バッジの配色は languageBadgeClasses が
 * key から引く（未知の値でも無彩色で表示は壊れない = ここは選択肢の提示のみ）。
 */
export interface CourseLanguageDef {
  key: string;
  label: string;
}

export const COURSE_LANGUAGES: CourseLanguageDef[] = [
  { key: 'go', label: 'Go' },
  { key: 'php', label: 'PHP' },
  { key: 'javascript', label: 'JavaScript' },
  { key: 'typescript', label: 'TypeScript' },
  { key: 'web', label: 'Web (HTML/HTTP)' },
  { key: 'git', label: 'Git' },
  { key: 'docker', label: 'Docker' },
  { key: 'linux', label: 'Linux' },
  { key: 'postgresql', label: 'PostgreSQL' },
  { key: 'sql', label: 'SQL' },
  { key: 'terraform', label: 'Terraform' },
  { key: 'aws', label: 'AWS' },
  { key: 'openapi', label: 'OpenAPI' },
];
