import { InboxIcon, ClipboardDocumentCheckIcon } from '@heroicons/react/24/outline';
import { MasterExerciseExample } from '../../types';

interface Props {
  index: number;
  total: number;
  example: MasterExerciseExample;
}

/**
 * 入出力例 1 件 を 表示 する カード。
 * 例 が 複数 ある 場合 は 番号 を 付けて 「入力 1 / 期待 出力 1」 のように 表示、
 * 1 件 のみ なら 番号 なし で シンプル に。
 */
export default function ExampleBlock({ index, total, example }: Props) {
  const suffix = total > 1 ? ` ${index}` : '';
  const inputDisplay = example.inputText.length > 0 ? example.inputText : 'ありません。';
  return (
    <div className="space-y-3">
      <div className="space-y-1.5 pb-4 border-b border-surface-3 last:border-b-0 last:pb-0">
        <div className="flex items-center gap-2 text-sm text-[var(--color-text-primary)] font-semibold">
          <InboxIcon className="w-4 h-4 text-[var(--color-text-muted)]" />
          入力される値{suffix}
        </div>
        <pre className="whitespace-pre-wrap break-words text-sm text-[var(--color-text-primary)]">
          {inputDisplay}
        </pre>
        {example.inputText.length === 0 && (
          <p className="text-xs text-[var(--color-text-muted)] leading-relaxed pt-1">
            入力値最終行の末尾に改行が 1 つ入ります。<br />
            文字列は標準入力から渡されます。
          </p>
        )}
      </div>

      <div className="space-y-1.5">
        <div className="flex items-center gap-2 text-sm text-[var(--color-text-primary)] font-semibold">
          <ClipboardDocumentCheckIcon className="w-4 h-4 text-[var(--color-text-muted)]" />
          期待する出力{suffix}
        </div>
        <pre className="whitespace-pre-wrap break-words text-sm text-[var(--color-text-primary)] bg-surface-2 border border-surface-3 rounded px-3 py-2 font-mono">
          {example.expectedOutput}
        </pre>
      </div>
    </div>
  );
}
