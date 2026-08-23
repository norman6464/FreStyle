/** CodeBlockLanguage は言語メニューの 1 項目。id は lowlight（highlight.js）の言語名。 */
export interface CodeBlockLanguage {
  id: string;
  /** メニューとバッジに出す表示名。 */
  label: string;
}

/**
 * CODE_BLOCK_LANGUAGES は言語メニューに出す一覧（よく使う順 → その他はアルファベット順）。
 * すべて lowlight（common）に登録済みの言語で、先頭の plaintext はハイライトなし。
 */
export const CODE_BLOCK_LANGUAGES: CodeBlockLanguage[] = [
  { id: 'plaintext', label: 'プレーンテキスト' },
  { id: 'sql', label: 'SQL' },
  { id: 'typescript', label: 'TypeScript' },
  { id: 'javascript', label: 'JavaScript' },
  { id: 'go', label: 'Go' },
  { id: 'python', label: 'Python' },
  { id: 'bash', label: 'Bash' },
  { id: 'json', label: 'JSON' },
  { id: 'yaml', label: 'YAML' },
  { id: 'xml', label: 'HTML / XML' },
  { id: 'css', label: 'CSS' },
  { id: 'markdown', label: 'Markdown' },
  { id: 'diff', label: 'Diff' },
  { id: 'java', label: 'Java' },
  { id: 'kotlin', label: 'Kotlin' },
  { id: 'swift', label: 'Swift' },
  { id: 'php', label: 'PHP' },
  { id: 'ruby', label: 'Ruby' },
  { id: 'rust', label: 'Rust' },
  { id: 'c', label: 'C' },
  { id: 'cpp', label: 'C++' },
  { id: 'csharp', label: 'C#' },
  { id: 'graphql', label: 'GraphQL' },
  { id: 'ini', label: 'INI / TOML' },
  { id: 'makefile', label: 'Makefile' },
  { id: 'scss', label: 'SCSS' },
  { id: 'shell', label: 'Shell' },
  { id: 'lua', label: 'Lua' },
  { id: 'perl', label: 'Perl' },
  { id: 'r', label: 'R' },
];

/** languageLabel は言語 id の表示名を返す（一覧に無い id はそのまま表示）。 */
export function languageLabel(id: string | null | undefined): string {
  if (!id) return 'プレーンテキスト';
  return CODE_BLOCK_LANGUAGES.find((language) => language.id === id)?.label ?? id;
}

/** filterLanguages は検索語（表示名・id の部分一致、大文字小文字無視）で一覧を絞り込む。 */
export function filterLanguages(query: string): CodeBlockLanguage[] {
  const normalized = query.trim().toLowerCase();
  if (normalized === '') return CODE_BLOCK_LANGUAGES;
  return CODE_BLOCK_LANGUAGES.filter(
    (language) =>
      language.id.toLowerCase().includes(normalized) ||
      language.label.toLowerCase().includes(normalized),
  );
}
