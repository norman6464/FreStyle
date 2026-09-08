import type { SuggestionDiffLine } from '../lib/suggestionDiff';

export interface KbSuggestionDiffViewProps {
  lines: SuggestionDiffLine[];
}

const LINE_STYLE: Record<SuggestionDiffLine['type'], string> = {
  added: 'bg-green-500/15 text-green-800',
  removed: 'bg-red-500/15 text-red-800 line-through',
  unchanged: 'text-[var(--color-text-secondary)]',
};

/**
 * KbSuggestionDiffView は行単位の差分（追加=緑・削除=赤の打ち消し線・不変=地の文）を描く。
 *
 * 差分そのものの計算は持たない（pages/kb/lib/suggestionDiff.computeSuggestionDiff が担う）。
 * 空行は `&nbsp;` で埋め、行の高さが潰れて見えなくならないようにする。
 */
export default function KbSuggestionDiffView({ lines }: KbSuggestionDiffViewProps) {
  if (lines.length === 0) {
    return (
      <p className="text-xs leading-relaxed text-[var(--color-text-muted)]">差分はありません。</p>
    );
  }
  return (
    <div className="overflow-x-auto rounded border border-surface-3 bg-surface-1 p-2 font-mono text-xs leading-relaxed">
      {lines.map((line, index) => (
        <div key={index} className={LINE_STYLE[line.type]}>
          {line.text || ' '}
        </div>
      ))}
    </div>
  );
}
