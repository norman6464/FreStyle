/**
 * 言語・技術名 → バッジ配色の対応表。演習の言語バッジとコースの言語バッジで共用する
 * （同じ技術は画面をまたいで同じ色に見えるように、色の正本をここに一本化）。
 *
 * 淡色背景 + 濃色文字 + 枠のトーン。背景 /25 + 枠 /50 は「淡すぎて見えにくい」という
 * ユーザー要望によるコントラスト強化。
 * 各技術の一般的なイメージカラーに寄せる。
 *
 * 文字は 800 を使う。700 だと、淡色背景（500 を 25% で白に重ねた色）の上で 4.2〜4.35 にしか
 * ならず、小さな文字の基準 4.5:1 に僅かに届かない色が複数あった（story の a11y 検査が拾った）。
 */
export const LANGUAGE_BADGE_CLASSES: Record<string, string> = {
  php: 'bg-indigo-500/25 text-indigo-800 border-indigo-500/50',
  go: 'bg-cyan-500/25 text-cyan-800 border-cyan-500/50',
  javascript: 'bg-yellow-500/25 text-yellow-800 border-yellow-500/50',
  typescript: 'bg-blue-500/25 text-blue-800 border-blue-500/50',
  // HTML の定番オレンジは git と重複するため amber を使う（aws と同色だが演習には出ないため可）。
  html: 'bg-amber-500/25 text-amber-800 border-amber-500/50',
  git: 'bg-orange-500/25 text-orange-800 border-orange-500/50',
  java: 'bg-red-500/25 text-red-800 border-red-500/50',
  // Ruby の定番赤は java と重複するため rose を使う（web と同色だが web は演習には出ないため可）。
  ruby: 'bg-rose-500/25 text-rose-800 border-rose-500/50',
  // C++ の定番青は typescript と重複するため violet を使う（terraform と同色だが演習には出ないため可）。
  cpp: 'bg-violet-500/25 text-violet-800 border-violet-500/50',
  // C は emerald（openapi と同色だが演習には出ないため可）。
  c: 'bg-emerald-500/25 text-emerald-800 border-emerald-500/50',
  bash: 'bg-slate-500/25 text-slate-800 border-slate-500/50',
  linux: 'bg-slate-500/25 text-slate-800 border-slate-500/50',
  docker: 'bg-sky-500/25 text-sky-800 border-sky-500/50',
  postgresql: 'bg-teal-500/25 text-teal-800 border-teal-500/50',
  sql: 'bg-teal-500/25 text-teal-800 border-teal-500/50',
  terraform: 'bg-violet-500/25 text-violet-800 border-violet-500/50',
  aws: 'bg-amber-500/25 text-amber-800 border-amber-500/50',
  openapi: 'bg-emerald-500/25 text-emerald-800 border-emerald-500/50',
  web: 'bg-rose-500/25 text-rose-800 border-rose-500/50',
};

/**
 * 「先頭のみ大文字」の機械整形では正しく表せない表示名の上書き。
 * 例: cpp → "Cpp" になってしまうため "C++" を明示する。
 */
export const LANGUAGE_DISPLAY_OVERRIDES: Record<string, string> = {
  cpp: 'C++',
};

/** 言語・技術名（大文字小文字を無視）から配色を引く。未知・空は undefined。 */
export function languageBadgeClass(language: string): string | undefined {
  // API のデータ欠損等で実行時に falsy が来ても toLowerCase でクラッシュしないよう防御する。
  return language ? LANGUAGE_BADGE_CLASSES[language.toLowerCase()] : undefined;
}
