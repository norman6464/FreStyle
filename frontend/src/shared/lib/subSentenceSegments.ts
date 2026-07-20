/**
 * subSentenceSegments — Gemini 実物と同じ「句読点区切りチャンク(sub-sentence)」分割。
 *
 * Gemini web の配信バンドルから確認した実装(FRESTYLE-146 調査)では、ストリーミング本文を
 * 正規表現 /[,:\.!\?、。，۔]+/gi で分割し、各チャンクを不可視 span として先行挿入 →
 * 一定リズムで 1 チャンクずつフェードインさせている。本モジュールはその分割部分を
 * 表示側(rehypeFadeSegments)とペーシング側(useSmoothReveal)で共用する。
 *
 * - 区切り文字はチャンクの末尾に含める(「こんにちは、」で 1 チャンク)。
 * - 改行も境界に含める(Gemini は markdown ノード単位で自然に切れるが、こちらは
 *   生テキストの prefix を伸ばす方式のため、コードブロック等の句読点が無い行でも
 *   行単位で放出できるようにする)。
 */

// Gemini 実物の sub-sentence 区切り(, : . ! ? 、 。 ， ۔) + 改行。
const BOUNDARY_RE = /[,:.!?、。，۔\n]+/g;

/** text を句読点チャンクに分割する(区切り文字は前のチャンク末尾に付く。全文が保存される)。 */
export function splitSubSentences(text: string): string[] {
  if (!text) return [];
  const out: string[] = [];
  let last = 0;
  // matchAll でグローバル regex の lastIndex 共有状態を避ける(呼び出しが増えても安全)。
  for (const m of text.matchAll(BOUNDARY_RE)) {
    const end = (m.index ?? 0) + m[0].length;
    out.push(text.slice(last, end));
    last = end;
  }
  if (last < text.length) out.push(text.slice(last));
  return out;
}

/** チャンクが句読点(または改行)で完結しているか。未完チャンクはストリーミング中に保留する。 */
export function endsAtBoundary(chunk: string): boolean {
  return /[,:.!?、。，۔\n]$/.test(chunk);
}
