import { CodeExecutionResult } from '../../types';
import { normalizeOutput } from '../../utils/exerciseFormat';

interface Props {
  result: CodeExecutionResult | null;
  expected: string;
  submitError: string | null;
}

/**
 * 提出前 動作 確認 (= 単発 実行) の 結果 を テーブル 形式 で 表示。
 * stdout / stderr / 期待 出力 を 並列 で 見せて、 末尾 で 「期待値 一致 / 不一致」 を バナー 表示。
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

  const status = result.exitCode === 0 ? 'Success' : `Failure (exit ${result.exitCode})`;
  const matched = result.exitCode === 0 && normalizeOutput(result.stdout) === normalizeOutput(expected);

  return (
    <div className="rounded-lg border border-surface-3 bg-surface-1 overflow-hidden">
      <table className="w-full text-xs">
        <tbody>
          <tr className="border-b border-surface-3">
            <th className="text-left px-4 py-2 font-medium text-[var(--color-text-muted)] bg-surface-2 w-1/3">
              実行結果ステータス
            </th>
            <td className="px-4 py-2 text-[var(--color-text-primary)]">{status}</td>
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
      <div
        className={`px-4 py-2 text-xs font-semibold ${
          matched
            ? 'bg-green-500/15 text-green-400 border-t border-green-500/30'
            : 'bg-red-500/15 text-red-400 border-t border-red-500/30'
        }`}
      >
        コード実行結果: {matched ? '◎ 期待出力と一致' : '✗ 期待出力と不一致'}
      </div>
    </div>
  );
}
