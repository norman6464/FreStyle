/**
 * rehypeFadeSegments — ストリーミング中の Markdown 本文を「語(word)単位でフェードイン」させる rehype プラグイン。
 *
 * react-markdown が生成した hast のテキストノードを Intl.Segmenter で語に分割し、可視な語を
 * `<span class="fade-seg">` で包む。実際のフェードは CSS 側(.fade-seg)で mount 時に一度だけ
 * opacity(+ reduced-motion でない時のみ blur) を再生する。
 *
 * 既表示の語は再パースで位置が変わらなければ React が DOM ノードを再利用するため、アニメが
 * 再生し直されない(＝新しく mount した span だけがフェードする)。char-offset で「既出」を潰す
 * 方式を採らないのは、フェード途中の語を potenzialに opacity:1 へ瞬間スナップさせてしまう(pop)ため。
 *
 * - pre / code 配下は分割しない(コードのハイライト・コピーを壊さない / 1 文字ずつ光らせない)。
 * - 空白のみのセグメントは span で包まず素のテキストで残す(折り返しとレイアウトを保つ)。
 * - 非ストリーミング時(完了・履歴・テスト)はこのプラグインを付けず、素の Markdown に戻す。
 */

interface HastText {
  type: 'text';
  value: string;
}

interface HastElement {
  type: 'element';
  tagName: string;
  properties?: Record<string, unknown>;
  children: HastChild[];
}

type HastChild = HastText | HastElement | { type: string; [key: string]: unknown };

interface HastRoot {
  type: 'root';
  children: HastChild[];
}

// Intl.Segmenter は ES2020 の lib 型に含まれないため、必要な部分だけローカルに型定義する。
interface WordSegmenter {
  segment(input: string): Iterable<{ segment: string }>;
}
interface WordSegmenterCtor {
  new (
    locale?: string,
    options?: { granularity?: 'grapheme' | 'word' | 'sentence' },
  ): WordSegmenter;
}

// pre / code 配下のテキストは分割対象外。
const SKIP_TAGS = new Set(['pre', 'code']);

// 語分割器。Intl.Segmenter があれば日本語などの非空白区切り言語も語単位に分けられる。
// 無い実行環境向けに空白区切りのフォールバックを持つ。
const SegmenterImpl = (Intl as unknown as { Segmenter?: WordSegmenterCtor }).Segmenter;
const segmenter: WordSegmenter | null = SegmenterImpl
  ? new SegmenterImpl('ja', { granularity: 'word' })
  : null;

function segments(text: string): string[] {
  if (segmenter) {
    return Array.from(segmenter.segment(text), (s) => s.segment);
  }
  // フォールバック: 空白を保持したまま語と空白に分ける。
  return text.split(/(\s+)/).filter((s) => s.length > 0);
}

function isText(node: HastChild): node is HastText {
  return node.type === 'text';
}

function isElement(node: HastChild): node is HastElement {
  return node.type === 'element';
}

function wrapText(value: string): HastChild[] {
  const out: HastChild[] = [];
  for (const seg of segments(value)) {
    if (!/\S/.test(seg)) {
      // 空白のみのセグメントは折り返しのためそのまま残す。
      out.push({ type: 'text', value: seg });
    } else {
      out.push({
        type: 'element',
        tagName: 'span',
        properties: { className: ['fade-seg'] },
        children: [{ type: 'text', value: seg }],
      });
    }
  }
  return out;
}

function transform(node: HastElement | HastRoot): void {
  const next: HastChild[] = [];
  for (const child of node.children) {
    if (isText(child)) {
      next.push(...wrapText(child.value));
    } else if (isElement(child)) {
      if (!SKIP_TAGS.has(child.tagName)) {
        transform(child);
      }
      next.push(child);
    } else {
      next.push(child);
    }
  }
  node.children = next;
}

/** ストリーミング中の本文をフェードイン用の語 span に分割する rehype プラグイン。 */
export default function rehypeFadeSegments() {
  return (tree: HastRoot): void => {
    transform(tree);
  };
}
