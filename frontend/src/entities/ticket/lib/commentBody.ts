import type { TicketCommentSegment } from '../model/types';

/**
 * 発言の本文（インラインノードの配列）と、画面が扱う区間の列との往復。
 *
 * backend が受け取る本文は段落を持たないインラインノードの配列で、チケット本体の `doc`
 * （`{type:'doc',content:[…]}`）とは別の形をしている。往復をこの 1 か所に閉じ、画面側は
 * `TicketCommentSegment[]` だけを見ればよいようにする。
 */

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

/**
 * 応答の本文を区間の列へ畳む。
 *
 * text と名指し以外のノードは、文字を持っていればその文字だけを拾い、持っていなければ捨てる。
 * backend は本文のノード種別を検証しないので、この画面が書いたもの以外が入っている可能性を
 * 完全には否定できない — 読めない飾りのために発言そのものが表示できなくなる方が困る。
 *
 * **marks（太字・リンク等）は意図的に一切保持しない。** text ノードの `text` だけを拾い、
 * `marks` は読まない。backend は marks を検証しないので、リンクの飛び先が入っていても
 * 描画に使わなければ無害化の必要そのものが無くなる（このアプリのどの入力欄も marks 付きの
 * 本文をチケットの発言に書き込まない・書き込めない）。ノート側のコメントと同じ
 * 「本文はプレーンテキスト中心」の方針。
 */
export function readCommentBody(body: unknown): TicketCommentSegment[] {
  if (!Array.isArray(body)) return [];
  const segments: TicketCommentSegment[] = [];
  for (const node of body) {
    if (!isRecord(node)) continue;
    if (node.type === 'mention') {
      const attrs = isRecord(node.attrs) ? node.attrs : {};
      const userId = attrs.userId;
      if (typeof userId === 'string' && userId !== '') {
        segments.push({ kind: 'mention', userId });
      } else if (typeof userId === 'number' && Number.isFinite(userId)) {
        segments.push({ kind: 'mention', userId: String(userId) });
      }
      continue;
    }
    if (typeof node.text === 'string' && node.text !== '') {
      segments.push({ kind: 'text', text: node.text });
    }
  }
  return segments;
}

/**
 * 区間の列を、送信できる本文へ組み立てる。
 *
 * 空白だけの text は backend が**本文全体ごと**拒む（400）。名指しを 2 つ続けて書いたときの
 * 区切りの空白がまさにこれに当たるので、ここで落とす。落とした空白は表示側の余白で補う。
 * 全部落ちて空配列になったら送信してはいけない（空の本文も同じく 400）— 呼び出し側は
 * 戻り値の長さを見て送信可否を決めること。
 */
export function buildCommentBody(segments: TicketCommentSegment[]): unknown[] {
  const nodes: unknown[] = [];
  for (const segment of segments) {
    if (segment.kind === 'mention') {
      if (segment.userId === '') continue;
      nodes.push({ type: 'mention', attrs: { userId: segment.userId } });
      continue;
    }
    if (segment.text.trim() === '') continue;
    nodes.push({ type: 'text', text: segment.text });
  }
  return nodes;
}
