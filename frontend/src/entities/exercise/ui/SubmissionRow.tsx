import { CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline';
import { ExerciseSubmission } from '../model/types';
import { pad } from '../lib/exerciseFormat';

/**
 * 提出 履歴 1 件 を 1 行 で 表示。
 * 日時 + 合格 / 不合格 アイコン + ラベル。 直近 10 件 を 一覧 表示 する 用途。
 */
export default function SubmissionRow({ submission }: { submission: ExerciseSubmission }) {
  const date = new Date(submission.submittedAt);
  const stamp = `${date.getFullYear()}/${pad(date.getMonth() + 1)}/${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
  return (
    <li className="flex items-center gap-2 text-xs px-2 py-1 rounded border border-surface-3 bg-surface-2">
      {submission.isCorrect
        ? <CheckCircleIcon className="w-4 h-4 text-green-400 flex-shrink-0" />
        : <XCircleIcon className="w-4 h-4 text-red-400 flex-shrink-0" />}
      <span className="font-mono text-[var(--color-text-primary)]">{stamp}</span>
      <span className={submission.isCorrect ? 'text-green-400' : 'text-red-400'}>
        {submission.isCorrect ? '合格' : '不合格'}
      </span>
    </li>
  );
}
