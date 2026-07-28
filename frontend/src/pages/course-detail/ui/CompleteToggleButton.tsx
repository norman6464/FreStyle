import { CheckCircleIcon as CheckCircleSolidIcon } from '@heroicons/react/24/solid';

/** 教材の完了 / 未完了を切り替えるトグルボタン。`large` で本文末尾用の大きめ表示にする。 */
export default function CompleteToggleButton({
  completed,
  onToggle,
  large = false,
}: {
  completed: boolean;
  onToggle: (done: boolean) => void;
  large?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={() => onToggle(!completed)}
      aria-pressed={completed}
      // whitespace-nowrap / shrink-0: 隣に長いタイトルの「次の章へ」ボタンが並んでも縮まず、
      // ラベルが「完了に / する」と 2 行に折り返さないようにする(FRESTYLE-188)。
      className={`inline-flex shrink-0 whitespace-nowrap items-center justify-center gap-1.5 rounded-lg font-medium transition-colors ${
        large ? 'px-5 pt-[9px] pb-[11px] text-sm' : 'px-3 pt-[5px] pb-[7px] text-sm'
      } ${
        completed
          ? 'bg-green-500/15 text-green-500 hover:bg-green-500/25'
          : 'bg-brand-500 text-white hover:bg-brand-600'
      }`}
    >
      <CheckCircleSolidIcon className="w-4 h-4" />
      {completed ? '完了済み' : '完了にする'}
    </button>
  );
}
