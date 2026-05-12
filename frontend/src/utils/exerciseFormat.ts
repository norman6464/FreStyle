/**
 * ExerciseDetailPage 系で共有する整形ユーティリティ。
 * 1 ファイルに集約することで、 各コンポーネントが個別に同等の関数を持たないようにする。
 */

/**
 * master_exercises.language → Monaco のシンタックスハイライト言語ID へのマッピング。
 * Monaco は `php` / `go` / `shell` を組み込みでサポート。
 * 該当しない言語は plaintext に fallback する。
 */
export function monacoLanguageOf(lang: string): string {
  switch (lang) {
    case 'php':
      return 'php';
    case 'go':
      return 'go';
    case 'bash':
      return 'shell';
    default:
      return 'plaintext';
  }
}

/**
 * バックエンドの normalizeOutput と同等の正規化（提出前確認の「期待値と一致」表示用）。
 * 末尾の改行 / 空白を吸収して厳密一致判定する。
 */
export function normalizeOutput(s: string): string {
  return s.replace(/\r\n/g, '\n').replace(/\r/g, '\n').replace(/[\s]+$/, '');
}

/** 1 桁数値を 0 埋めの 2 桁文字列にする。日時表示用。 */
export function pad(n: number): string {
  return String(n).padStart(2, '0');
}
