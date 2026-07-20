/*
 * shared/ui の Public API。
 *
 * FSD の公式仕様どおり、ワイルドカード（`export *`）は使わず名前付きで再 export する。
 * 何が外に出ているかがこのファイルだけで分かる状態を保つため。
 */

// --- プリミティブ ---
export { default as Button } from './Button';
export type { ButtonProps, ButtonVariant, ButtonSize } from './Button';
export { default as InputField } from './InputField';
export { default as TextareaField } from './TextareaField';
export { default as LinkText } from './LinkText';
export { default as Loading } from './Loading';
export { default as Avatar } from './Avatar';

// --- フォーム補助 ---
export { default as FormFieldError } from './FormFieldError';
export { default as FormMessage } from './FormMessage';

// --- 画面の枠・状態表示 ---
export { default as ActionCard } from './ActionCard';
export { default as ConfirmModal } from './ConfirmModal';
export { default as EmptyState } from './EmptyState';
export { default as PageIntro } from './PageIntro';
export { default as StepIndicator } from './StepIndicator';
export { default as Toast } from './Toast';

// --- 初心者向けの補助 UI ---
export { default as FilterChip } from './FilterChip';
export type { FilterChipProps } from './FilterChip';
export { default as FirstTimeWelcome } from './FirstTimeWelcome';
export { default as GlossaryTerm } from './GlossaryTerm';
export { default as GuidedHint } from './GuidedHint';
export { default as HelpTooltip } from './HelpTooltip';

// --- テキスト・Markdown 周り ---
export { default as LineCount } from './LineCount';
export { default as WordCount } from './WordCount';
export { default as ReadingTime } from './ReadingTime';
export { default as MarkdownTableOfContents } from './MarkdownTableOfContents';

/*
 * CodeEditor は **意図的にこの barrel から出さない**。
 *
 * 中身は monaco-editor（数百 KB）で、演習ページだけが
 * `lazyWithReload(() => import('@/shared/ui/CodeEditor'))` で遅延ロードしている。
 * ここで re-export すると `@/shared/ui` を import した全ページが monaco を巻き込み、
 * コード分割が壊れる（実際 CoursesListPage / HelpPage のテストが monaco の
 * document.queryCommandSupported で落ちて発覚した）。深いパスで直接 import すること。
 */

/*
 * 言語バッジ / アイコンは entity ではなく shared に置く。
 * コース（courses.language）と演習（master_exercises.language）の両方が使うため、
 * どちらかの entity に置くと同一レイヤーの Slice 間 import になり FSD 違反になる。
 * 中身も devicon スラッグと Tailwind クラスの対応表で、FreStyle 固有のルールではない。
 */
export { default as LanguageBadge } from './LanguageBadge';
export { default as LanguageIcon } from './LanguageIcon';
