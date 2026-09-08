import { diffLines } from 'diff';

/**
 * ProseMirror doc から差分用の平文を作る。
 *
 * トップレベルの各ブロック（paragraph・heading・listItem 等）ごとに、そのブロック配下の
 * text ノードを連結した 1 行を作り、ブロックを改行で繋ぐ。backend の extractPageBodyText
 * （Go 実装）と同じ発想の抜き出しだが、あちらは流用できないためフロント側に別実装として持つ。
 *
 * ブロックが object でない・type を持たない等パースできない形は `[変更あり]` という
 * プレースホルダ行にする（何が起きたか分かる程度の情報は残しつつ、例外は投げない）。
 */
export function extractPlainText(doc: unknown): string {
  if (!isPlainObject(doc) || !Array.isArray(doc.content)) return '';
  return doc.content.map(extractBlockLine).join('\n');
}

function extractBlockLine(block: unknown): string {
  if (!isPlainObject(block) || typeof block.type !== 'string') return '[変更あり]';
  return collectText(block);
}

function collectText(node: unknown): string {
  if (!isPlainObject(node)) return '';
  let text = node.type === 'text' && typeof node.text === 'string' ? node.text : '';
  if (Array.isArray(node.content)) {
    for (const child of node.content) text += collectText(child);
  }
  return text;
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

/** 差分の 1 行。unchanged は前後を確かめるための地の文で、色は付けない。 */
export interface SuggestionDiffLine {
  type: 'added' | 'removed' | 'unchanged';
  text: string;
}

/**
 * computeSuggestionDiff は baseDoc（提案時点の本文全体）と doc（提案後の本文全体）を
 * 行単位で突き合わせる。baseDoc が無ければ空文字列として扱い、追加行だけの差分になる
 * （baseSeq を持たない = まだ版が 1 つも無いページへの提案）。
 */
export function computeSuggestionDiff(baseDoc: unknown, doc: unknown): SuggestionDiffLine[] {
  const baseText = baseDoc === undefined ? '' : extractPlainText(baseDoc);
  const newText = extractPlainText(doc);
  const lines: SuggestionDiffLine[] = [];
  for (const part of diffLines(baseText, newText)) {
    const type: SuggestionDiffLine['type'] = part.added ? 'added' : part.removed ? 'removed' : 'unchanged';
    // diffLines は同種の行をまとめて1つの value（内部に \n を含む）に詰めてくる。
    // 行ごとに色分けして表示するため、ここで1行ずつへ割り戻す。
    // 末尾の \n が作る最後の空文字列要素だけは実在する行ではないため落とす。
    const segments = part.value.split('\n');
    if (segments[segments.length - 1] === '') segments.pop();
    for (const text of segments) lines.push({ type, text });
  }
  return lines;
}
