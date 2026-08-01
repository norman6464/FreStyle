/**
 * API 応答を配列として安全に受け取るためのヘルパ（FRESTYLE-77）。
 *
 * バックエンドが 0 件のときに空配列ではなく null を返すと、`map` / `filter` / `for-of` が
 * TypeError で落ちて画面が開けなくなる。データがまだ無い新規ユーザー・新規コース・
 * 未提出演習など、使い始めの動線を直撃する（FRESTYLE-70 で実機観測）。
 *
 * バックエンド側でも空配列を保証しているが、片側だけの対策は「もう一方が壊れたら
 * ユーザーに影響が出る」状態のままなので、受け取る側でも守る。
 */
export function toArray<T>(value: unknown): T[] {
  return Array.isArray(value) ? (value as T[]) : [];
}
