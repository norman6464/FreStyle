import { describe, it, expect } from 'vitest';
import rehypeFadeSegments from '../rehypeFadeSegments';
import { splitSubSentences, endsAtBoundary } from '../subSentenceSegments';

// hast を最小限に自前で組んでプラグインを直接動かす（react-markdown を介さない純粋な変換テスト）。
interface Node {
  type: string;
  tagName?: string;
  value?: string;
  properties?: Record<string, unknown>;
  children?: Node[];
}

function run(tree: Node): Node {
  // プラグインは transformer を返す。tree を破壊的に書き換える。
  rehypeFadeSegments()(tree as never);
  return tree;
}

function collect(node: Node, pred: (n: Node) => boolean, acc: Node[] = []): Node[] {
  if (pred(node)) acc.push(node);
  node.children?.forEach((c) => collect(c, pred, acc));
  return acc;
}

function textOf(node: Node): string {
  if (node.type === 'text') return node.value ?? '';
  return (node.children ?? []).map(textOf).join('');
}

const isFadeSpan = (n: Node) =>
  n.type === 'element' &&
  n.tagName === 'span' &&
  Array.isArray(n.properties?.className) &&
  (n.properties?.className as unknown[]).includes('fade-seg');

describe('splitSubSentences', () => {
  it('句読点(読点含む)で分割し、区切り文字は前のチャンク末尾に付く', () => {
    expect(splitSubSentences('こんにちは、今日は晴れ。明日は雨。')).toEqual([
      'こんにちは、',
      '今日は晴れ。',
      '明日は雨。',
    ]);
  });

  it('句読点が無いテキストは 1 チャンクのまま', () => {
    expect(splitSubSentences('今日は良い天気です')).toEqual(['今日は良い天気です']);
  });

  it('改行も境界になる(コードブロック等の行単位放出)', () => {
    expect(splitSubSentences('line1\nline2')).toEqual(['line1\n', 'line2']);
  });

  it('全チャンクを連結すると元のテキストに戻る(全文保存)', () => {
    const text = 'A, B: C. こんにちは、世界。改行\nもある！最後は未完';
    expect(splitSubSentences(text).join('')).toBe(text);
  });

  it('endsAtBoundary は句読点/改行で終わるチャンクだけ true', () => {
    expect(endsAtBoundary('こんにちは、')).toBe(true);
    expect(endsAtBoundary('晴れ。')).toBe(true);
    expect(endsAtBoundary('line\n')).toBe(true);
    expect(endsAtBoundary('未完のまま')).toBe(false);
  });
});

describe('rehypeFadeSegments', () => {
  it('本文のテキストを句読点チャンクの fade-seg span に分割し、テキストは失われない', () => {
    const tree: Node = {
      type: 'root',
      children: [
        {
          type: 'element',
          tagName: 'p',
          properties: {},
          children: [{ type: 'text', value: 'こんにちは、今日は晴れ。明日は雨。' }],
        },
      ],
    };
    run(tree);
    const spans = collect(tree, isFadeSpan);
    // 読点・句点で 3 チャンクに分かれる。
    expect(spans).toHaveLength(3);
    expect(spans.map(textOf)).toEqual(['こんにちは、', '今日は晴れ。', '明日は雨。']);
    // 変換後も全文が保たれる。
    expect(textOf(tree)).toBe('こんにちは、今日は晴れ。明日は雨。');
  });

  it('句読点の無いテキストは 1 つの span になる', () => {
    const tree: Node = {
      type: 'root',
      children: [
        {
          type: 'element',
          tagName: 'p',
          properties: {},
          children: [{ type: 'text', value: '今日は良い天気です' }],
        },
      ],
    };
    run(tree);
    expect(collect(tree, isFadeSpan)).toHaveLength(1);
    expect(textOf(tree)).toBe('今日は良い天気です');
  });

  it('pre / code 配下のテキストは分割しない（ハイライト/コピーを壊さない）', () => {
    const tree: Node = {
      type: 'root',
      children: [
        {
          type: 'element',
          tagName: 'pre',
          properties: {},
          children: [
            {
              type: 'element',
              tagName: 'code',
              properties: {},
              children: [{ type: 'text', value: 'const a = 1;' }],
            },
          ],
        },
      ],
    };
    run(tree);
    expect(collect(tree, isFadeSpan)).toHaveLength(0);
    expect(textOf(tree)).toBe('const a = 1;');
  });

  it('空白のみのセグメントは span で包まず素のテキストで残す', () => {
    const tree: Node = {
      type: 'root',
      children: [
        {
          type: 'element',
          tagName: 'p',
          properties: {},
          // 先頭が句点 → 「。」で 1 チャンク、続く空白のみのチャンクは text のまま。
          children: [{ type: 'text', value: '。 　' }],
        },
      ],
    };
    run(tree);
    const p = tree.children![0];
    const spans = collect(tree, isFadeSpan);
    expect(spans.map(textOf)).toEqual(['。']);
    const hasWhitespaceText = (p.children ?? []).some(
      (c) => c.type === 'text' && /^\s+$/.test(c.value ?? ''),
    );
    expect(hasWhitespaceText).toBe(true);
    expect(textOf(tree)).toBe('。 　');
  });
});
