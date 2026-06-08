import { CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline';
import { CodeExecutionResult } from '../../types';

interface Props {
  result: CodeExecutionResult | null;
  expected: string;
  submitError: string | null;
}

/**
 * 提出前 動作 確認 (= 単発 実行) の 結果 を テーブル 形式 で 表示。
 * stdout / stderr / 期待 出力 を 並列 で 見せる。 採点 (正誤) は 行わ ず、 末尾 で
 * 「ここ は 実行 プレビュー」 と 明示 する (正誤 は サーバ 側 採点)。
 *
 * 「実行 ステータス」 は exit code に 応じた 「実行 が 通った / エラー」 を 表す だけ で、
 * 「答え が 正しい か どうか」 とは 無関係 (= ここ を 正誤 と 誤解 され ない 文言 に する):
 *   - exit 0 → 緑 + ✓ 「実行 成功 (エラー なし)」
 *   - exit != 0 → 赤 + ✗ 「実行 エラー (exit N)」
 * さらに exit 0 でも stdout が 空 の とき は 「まだ 出力 が ない」 ヒント を 出し、
 * 「出力 ゼロ なのに Success」 で 正解 と 誤解 され る の を 防ぐ。
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
  const hasNoOutput = isSuccess && !result.stdout;

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
                  実行成功（エラーなし）
                </span>
              ) : (
                <span className="inline-flex items-center gap-1 font-semibold text-red-400">
                  <XCircleIcon className="w-4 h-4" />
                  実行エラー（exit {result.exitCode}）
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
              {hasNoOutput && (
                <p className="mt-2 text-amber-500">
                  まだ出力がありません。<code>echo</code> や <code>print</code> などで結果を出力すると、ここに表示されます。
                </p>
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
        「実行成功」はコードがエラーなく動いたという意味で、正誤の判定ではありません。正誤は「提出する」で複数のテストケースによりサーバ側採点されます。
      </div>
    </div>
  );
}
