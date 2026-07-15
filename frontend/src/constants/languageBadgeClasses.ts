/**
 * 言語・技術名 → バッジ配色の対応表。演習の言語バッジとコースの言語バッジで共用する
 * （同じ技術は画面をまたいで同じ色に見えるように、色の正本をここに一本化。FRESTYLE-114）。
 *
 * 淡色背景 + 濃色文字 + 枠のトーン。背景 /25 + 枠 /50 は「淡すぎて見えにくい」という
 * ユーザー要望によるコントラスト強化(FRESTYLE-112)。
 * 各技術の一般的なイメージカラーに寄せる。
 */
export const LANGUAGE_BADGE_CLASSES: Record<string, string> = {
  php: 'bg-indigo-500/25 text-indigo-700 border-indigo-500/50',
  go: 'bg-cyan-500/25 text-cyan-700 border-cyan-500/50',
  javascript: 'bg-yellow-500/25 text-yellow-700 border-yellow-500/50',
  typescript: 'bg-blue-500/25 text-blue-700 border-blue-500/50',
  git: 'bg-orange-500/25 text-orange-700 border-orange-500/50',
  bash: 'bg-slate-500/25 text-slate-700 border-slate-500/50',
  linux: 'bg-slate-500/25 text-slate-700 border-slate-500/50',
  docker: 'bg-sky-500/25 text-sky-700 border-sky-500/50',
  postgresql: 'bg-teal-500/25 text-teal-700 border-teal-500/50',
  sql: 'bg-teal-500/25 text-teal-700 border-teal-500/50',
  terraform: 'bg-violet-500/25 text-violet-700 border-violet-500/50',
  aws: 'bg-amber-500/25 text-amber-700 border-amber-500/50',
  openapi: 'bg-emerald-500/25 text-emerald-700 border-emerald-500/50',
  web: 'bg-rose-500/25 text-rose-700 border-rose-500/50',
};

/** 言語・技術名（大文字小文字を無視）から配色を引く。未知は undefined。 */
export function languageBadgeClass(language: string): string | undefined {
  return LANGUAGE_BADGE_CLASSES[language.toLowerCase()];
}
