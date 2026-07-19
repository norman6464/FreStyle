import { describe, it, expect } from 'vitest';
import rehypeFadeSegments from '../rehypeFadeSegments';

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

describe('rehypeFadeSegments', () => {
  it('本文のテキストを語単位の fade-seg span に分割し、テキストは失われない', () => {
    const tree: Node = {
      type: 'root',
      children: [
        { type: 'element', tagName: 'p', properties: {}, children: [{ type: 'text', value: '今日は良い天気です' }] },
      ],
    };
    run(tree);
    const spans = collect(tree, isFadeSpan);
    // 語分割されるので複数の span になる。
    expect(spans.length).toBeGreaterThan(1);
    // 変換後も全文が保たれる。
    expect(textOf(tree)).toBe('今日は良い天気です');
  });

  it('pre / code 配下のテキストは分割しない（コードを 1 文字ずつ光らせない）', () => {
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
    // code 配下に fade-seg span は生成されない。
    const spans = collect(tree, isFadeSpan);
    expect(spans).toHaveLength(0);
    // code の中身はそのまま。
    expect(textOf(tree)).toBe('const a = 1;');
  });

  it('空白のみのセグメントは span で包まず素のテキストで残す（折り返しを保つ）', () => {
    const tree: Node = {
      type: 'root',
      children: [
        { type: 'element', tagName: 'p', properties: {}, children: [{ type: 'text', value: 'hello world' }] },
      ],
    };
    run(tree);
    const p = tree.children![0];
    // 語(hello / world)は span、間の空白は text ノードで残る。
    const spanTexts = collect(tree, isFadeSpan).map(textOf);
    expect(spanTexts).toContain('hello');
    expect(spanTexts).toContain('world');
    const hasWhitespaceText = (p.children ?? []).some((c) => c.type === 'text' && /^\s+$/.test(c.value ?? ''));
    expect(hasWhitespaceText).toBe(true);
    expect(textOf(tree)).toBe('hello world');
  });
});
