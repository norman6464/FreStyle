import { CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline';
import { CodeExecutionResult } from '../../types';

interface Props {
  result: CodeExecutionResult | null;
  expected: string;
  submitError: string | null;
}

/**
 * 提出前 動作 確認 (= 単発 実行) の 結果 を テーブル 形式 で 表示。
 * stdout / stderr / 期待 出力 を 並列 で 見せて、 末尾 で 「期待値 一致 / 不一致」 を バナー 表示。
 *
 * 「実行 ステータス」 は exit code に 応じて 色 + アイコン を 付ける:
 *   - exit 0 → 緑 + ✓ (Success) で 「コンパイル / 実行 は 通った」 を 視覚 化
 *   - exit != 0 → 赤 + ✗ (Failure) で 「コンパイル or 実行 で 失敗 した」 を 強調
 */
export default function ExecutionResultTable({ result, expected, submitError }: Props) {
  if (submitError) {
    return (
      <div className="rounded-md border border-red-500/30 bg-red-500/5 p-3 text-xs text-red-400">
        {submitError}
      </div>
    );
  }
  if (!result) return null;

  const isSuccess = result.exitCode === 0;

  return (
    <div className="rounded-lg border border-surface-3 bg-surface-1 overflow-hidden">
      <table className="w-full text-xs">
        <tbody>
          <tr className="border-b border-surface-3">
            <th className="text-left px-4 py-2 font-medium text-[var(--color-text-muted)] bg-surface-2 w-1/3">
              実行結果ステータス
            </th>
            <td className="px-4 py-2">
              {isSuccess ? (
                <span className="inline-flex items-center gap-1 font-semibold text-green-400">
                  <CheckCircleIcon className="w-4 h-4" />
                  Success
                </span>
              ) : (
                <span className="inline-flex items-center gap-1 font-semibold text-red-400">
                  <XCircleIcon className="w-4 h-4" />
                  Failure (exit {result.exitCode})
                </span>
              )}
            </td>
          </tr>
          <tr className="border-b border-surface-3">
            <th className="text-left px-4 py-2 font-medium text-[var(--color-text-muted)] bg-surface-2 align-top">
              提出コードのアウトプット
            </th>
            <td className="px-4 py-2">
              <pre className="whitespace-pre-wrap break-words font-mono text-[var(--color-text-primary)]">
                {result.stdout || '(なし)'}
              </pre>
              {result.stderr && (
                <pre className="mt-2 whitespace-pre-wrap break-words font-mono text-red-400">
                  {result.stderr}
                </pre>
              )}
            </td>
          </tr>
          <tr className="border-b border-surface-3">
            <th className="text-left px-4 py-2 font-medium text-[var(--color-text-muted)] bg-surface-2 align-top">
              期待する出力
            </th>
            <td className="px-4 py-2">
              <pre className="whitespace-pre-wrap break-words font-mono text-[var(--color-text-primary)]">{expected}</pre>
            </td>
          </tr>
        </tbody>
      </table>
      {/* 採点はサーバ側で行う(フロントでローカル判定しない)。ここは実行結果のプレビューのみ。 */}
      <div className="px-4 py-2 text-xs text-[var(--color-text-muted)] border-t border-surface-3">
        ここは実行プレビューです。正誤は「提出する」でサーバ側採点されます。
      </div>
    </div>
  );
}
