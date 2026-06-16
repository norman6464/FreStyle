/**
 * ExerciseDetailPage 系で共有する整形ユーティリティ。
 * 1 ファイルに集約することで、 各コンポーネントが個別に同等の関数を持たないようにする。
 */

/**
 * master_exercises.language → Monaco のシンタックスハイライト言語ID へのマッピング。
 * Monaco は php / go / python / javascript / typescript / sql / shell などを組み込みでサポートする。
 * 未対応の言語(docker など)は plaintext に fallback する。
 */
export function monacoLanguageOf(lang: string): string {
  switch (lang) {
    case 'php':
      return 'php';
    case 'go':
      return 'go';
    case 'sql':
      return 'sql';
    case 'python':
      return 'python';
    case 'javascript':
      return 'javascript';
    case 'typescript':
      return 'typescript';
    case 'bash':
    case 'sh':
      return 'shell';
    default:
      return 'plaintext';
  }
}

/**
 * 数値 を 最小 2 桁 の 0 埋め 文字列 に する。 日時 表示 用。
 * 例: `pad(3)` → `"03"`、 `pad(12)` → `"12"`、 `pad(123)` → `"123"` (3 桁 以上 は そのまま)。
 */
export function pad(n: number): string {
  return String(n).padStart(2, '0');
}
