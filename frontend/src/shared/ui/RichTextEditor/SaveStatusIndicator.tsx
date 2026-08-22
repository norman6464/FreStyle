/**
 * SaveStatus はエディタ本文の保存状態。既存の Markdown エディタ（NoteMarkdownEditor）と
 * 同じ意味・同じ色に揃える。保存の実処理（debounce・PUT・楽観ロック）は画面側が持ち、
 * この部品は状態を受け取って表示するだけ（presentational）。
 */
export type SaveStatus = 'idle' | 'unsaved' | 'saving' | 'saved';

const SAVE_STATUS_CONFIG: Record<
  Exclude<SaveStatus, 'idle'>,
  { label: string; color: string }
> = {
  unsaved: { label: '未保存', color: 'text-amber-500' },
  saving: { label: '保存中...', color: 'text-[var(--color-text-muted)]' },
  saved: { label: '保存済み', color: 'text-emerald-500' },
};

/**
 * SaveStatusIndicator は保存状態のラベルを表示する。idle のときは何も描画しない。
 */
export default function SaveStatusIndicator({ status }: { status: SaveStatus }) {
  if (status === 'idle') {
    return null;
  }
  const { label, color } = SAVE_STATUS_CONFIG[status];
  return (
    <span className={`text-xs ${color}`} role="status" aria-label="保存状態">
      {label}
    </span>
  );
}
