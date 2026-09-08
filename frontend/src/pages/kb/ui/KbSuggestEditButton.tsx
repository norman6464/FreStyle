export interface KbSuggestEditButtonProps {
  /** ドラフトモード中か。押すたびに呼び出し側（KbPage）がこれを反転させる。 */
  active: boolean;
  onToggle: () => void;
}

/**
 * KbSuggestEditButton は commenter（閲覧+コメントはできるが編集はできない役割）だけに
 * 見える「変更を提案する」の入口。
 *
 * コメント・履歴・テンプレートの各ボタンと同じ場所・同じ見た目のトグルだが、開く先は
 * 浮かぶフォームではなく本文表示エリアそのもの — 押すと本文が編集可能なドラフトモードへ
 * 切り替わる（KbPage 側が draft.open を見て本文の描画を切り替える）。呼び出し側
 * （KbPage）が data.canComment && !data.canEdit のときだけ描画する。
 */
export default function KbSuggestEditButton({ active, onToggle }: KbSuggestEditButtonProps) {
  return (
    <button
      type="button"
      onClick={onToggle}
      aria-expanded={active}
      className="rounded border border-surface-3 px-2 py-1 text-xs text-[var(--color-text-secondary)] transition-colors hover:bg-surface-2"
    >
      変更を提案する
    </button>
  );
}
