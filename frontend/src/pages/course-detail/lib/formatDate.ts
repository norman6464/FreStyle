/** ISO 8601 の日時文字列を `YYYY/MM/DD` 形式に整形する（空文字は空文字のまま返す）。 */
export function formatDate(iso: string): string {
  if (!iso) return '';
  const date = new Date(iso);
  return `${date.getFullYear()}/${padTwoDigits(date.getMonth() + 1)}/${padTwoDigits(date.getDate())}`;
}

/** 1 桁の数値を 2 桁ゼロ埋めの文字列にする（例: 3 → "03"）。 */
function padTwoDigits(value: number): string {
  return String(value).padStart(2, '0');
}
