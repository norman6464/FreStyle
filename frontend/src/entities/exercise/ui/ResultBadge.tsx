import { CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline';

/**
 * 採点結果 (全テストケース合格 / 不合格) を 1 行 で 表示 する バッジ。
 * 主ビュー / QA ビュー の ヘッダー で 共有。
 */
export default function ResultBadge({ isCorrect }: { isCorrect: boolean }) {
  if (isCorrect) {
    return (
      <span className="flex-shrink-0 flex items-center gap-1 text-xs px-2 py-1 rounded-full bg-green-500/15 text-green-400 border border-green-500/30">
        <CheckCircleIcon className="w-4 h-4" /> 全テストケース合格
      </span>
    );
  }
  return (
    <span className="flex-shrink-0 flex items-center gap-1 text-xs px-2 py-1 rounded-full bg-red-500/15 text-red-400 border border-red-500/30">
      <XCircleIcon className="w-4 h-4" /> 不合格
    </span>
  );
}
