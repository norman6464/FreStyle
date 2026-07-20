import { CheckCircleIcon, ExclamationTriangleIcon, XCircleIcon } from '@heroicons/react/24/outline';
import { CodeExecutionResult } from '../model/types';

interface Props {
  result: CodeExecutionResult | null;
  expected: string;
  submitError: string | null;
  /** 演習の言語(エラーメッセージ行のラベル判定に使用。省略時は汎用ラベル)。 */
  language?: string;
}

/**
 * サーバ採点の normalizeOutput と同じ正規化（改行コード統一・各行の末尾空白除去・
 * 末尾の空行除去）。末尾改行の有無だけでプレビュー比較が不一致にならないようにする。
 */
function normalizeOutput(s: string): string {
  return s
    .replace(/\r\n?/g, '\n')
    .split('\n')
    .map((line) => line.replace(/[ \t]+$/, ''))
    .join('\n')
    .replace(/[ \t\n]+$/, '');
}

interface DiffLine {
  /** 1 始まりの行番号。 */
  line: number;
  /** 該当行。片方に行が存在しない（行数が違う）場合は undefined。 */
  actual?: string;
  expected?: string;
}

/**
 * 正規化済みの出力と期待出力を行単位で比較し、最初に異なる行を返す（一致なら null）。
 * 「どこが違うのか」を学習者が自力で見つけなくて済むようにするための情報(FRESTYLE-113)。
 */
function firstDiffLine(normalizedActual: string, normalizedExpected: string): DiffLine | null {
  if (normalizedActual === normalizedExpected) return null;
  const actualLines = normalizedActual.split('\n');
  const expectedLines = normalizedExpected.split('\n');
  const max = Math.max(actualLines.length, expectedLines.length);
  for (let i = 0; i < max; i++) {
    if (actualLines[i] !== expectedLines[i]) {
      return { line: i + 1, actual: actualLines[i], expected: expectedLines[i] };
    }
  }
  return null;
}

/**
 * 提出前 動作 確認 (= 単発 実行) の 結果 を テーブル 形式 で 表示。
 * stdout / stderr / 期待 出力 を 並列 で 見せる。
 *
 * 「実行 ステータス」 は 3 状態 に 分け、 緑 は 「期待 出力 と 一致 した とき だけ」 に 予約 する。
 * exit 0 なら 常に 緑 だと 「エラー なく 動いた ＝ 正解」 と 色 の 印象 で 誤解 され やすい ため
 * (FRESTYLE-111)。 正誤 の 確定 は 従来 どおり 提出 時 の サーバ 側 採点 (複数 テストケース):
 *   - exit 0 + 出力 一致   → 緑 + ✓ 「実行 成功・期待 する 出力 と 一致」
 *   - exit 0 + 出力 不一致 → 琥珀 + ⚠ 「実行 成功 (エラー なし)・期待 する 出力 と 不一致」
 *                            + 最初 に 異なる 行 の 提示 (どこ が 違う か を 自力 で 探さ せ ない / FRESTYLE-113)
 *   - exit != 0            → 赤 + ✗ 「実行 エラー (exit N)」
 * 期待 出力 が 空 の 演習 は 比較 でき ない ので 従来 の 中立 表示 に フォールバック。
 * さらに exit 0 でも stdout が 空 の とき は 「まだ 出力 が ない」 ヒント を 出す。
 */
export default function ExecutionResultTable({ result, expected, submitError, language }: Props) {
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
  // Go の実行失敗で stderr に main.go:N が出ているのはコンパイル段階のエラー
  // (コンパイルが通れば stdout が出得る)。ラベルを具体的にして原因の区別を助ける。
  const errorLabel =
    language === 'go' && !isSuccess && /main\.go:\d+/.test(result.stderr)
      ? 'コンパイル時エラーメッセージ'
      : 'エラーメッセージ';
  const normalizedStdout = normalizeOutput(result.stdout);
  const normalizedExpected = normalizeOutput(expected);
  const comparable = normalizedExpected !== '';
  const matches = comparable && normalizedStdout === normalizedExpected;
  const diff = comparable && !matches ? firstDiffLine(normalizedStdout, normalizedExpected) : null;

  return (
    <div className="rounded-lg border border-surface-3 bg-surface-1 overflow-hidden">
      <table className="w-full text-xs">
        <tbody>
          <tr className="border-b border-surface-3">
            <th className="text-left px-4 py-2 font-medium text-[var(--color-text-muted)] bg-surface-2 w-1/3">
              実行結果ステータス
            </th>
            <td className="px-4 py-2">
              {!isSuccess ? (
                <span className="inline-flex items-center gap-1 font-semibold text-red-400">
                  <XCircleIcon className="w-4 h-4" />
                  実行エラー（exit {result.exitCode}）
                </span>
              ) : !comparable ? (
                <span className="inline-flex items-center gap-1 font-semibold text-green-400">
                  <CheckCircleIcon className="w-4 h-4" />
                  実行成功（エラーなし）
                </span>
              ) : matches ? (
                <span className="inline-flex items-center gap-1 font-semibold text-green-400">
                  <CheckCircleIcon className="w-4 h-4" />
                  実行成功・期待する出力と一致
                </span>
              ) : (
                <>
                  <span className="inline-flex items-center gap-1 font-semibold text-amber-500">
                    <ExclamationTriangleIcon className="w-4 h-4" />
                    実行成功（エラーなし）・期待する出力と不一致
                  </span>
                  {diff && (
                    <p className="mt-1.5 text-amber-600">
                      {diff.line} 行目が異なります — あなたの出力:{' '}
                      {diff.actual !== undefined ? (
                        <code className="font-mono">「{diff.actual}」</code>
                      ) : (
                        '(この行がありません)'
                      )}{' '}
                      / 期待:{' '}
                      {diff.expected !== undefined ? (
                        <code className="font-mono">「{diff.expected}」</code>
                      ) : (
                        '(この行がありません)'
                      )}
                    </p>
                  )}
                </>
              )}
            </td>
          </tr>
          {/* stderr は stdout と混ぜず専用の行に分ける(FRESTYLE-117)。
              エラーの本文を探す場所が一定になり、エディタの行マーカーと突き合わせやすい。 */}
          {result.stderr && (
            <tr className="border-b border-surface-3">
              <th className="text-left px-4 py-2 font-medium text-[var(--color-text-muted)] bg-surface-2 align-top">
                {errorLabel}
              </th>
              <td className="px-4 py-2">
                <pre className="whitespace-pre-wrap break-words font-mono text-red-500 bg-red-500/5 border border-red-500/20 rounded-md p-3">
                  {result.stderr}
                </pre>
              </td>
            </tr>
          )}
          <tr className="border-b border-surface-3">
            <th className="text-left px-4 py-2 font-medium text-[var(--color-text-muted)] bg-surface-2 align-top">
              提出コードのアウトプット
            </th>
            <td className="px-4 py-2">
              <pre className="whitespace-pre-wrap break-words font-mono text-[var(--color-text-primary)]">
                {result.stdout || '(なし)'}
              </pre>
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
      {/* 一致判定はこの画面の期待出力 1 件とのプレビュー比較。正誤の確定はサーバ側採点。 */}
      <div className="px-4 py-2 text-xs text-[var(--color-text-muted)] border-t border-surface-3">
        「一致」はこの画面に表示中の期待出力とのプレビュー比較です。正誤は「提出する」で複数のテストケースによりサーバ側採点されます。
      </div>
    </div>
  );
}
