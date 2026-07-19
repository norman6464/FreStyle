/**
 * rehypeFadeSegments — ストリーミング中の Markdown 本文を「句読点チャンク単位でフェードイン」
 * させる rehype プラグイン(Gemini 実物仕様の再現、FRESTYLE-145 → 146 で語単位から変更)。
 *
 * Gemini web の実装(配信バンドルで確認)は、本文を句読点 regex で sub-sentence チャンクに分割して
 * span 化し、1 チャンクずつ opacity 0→1 (400ms / ease-out / 1 回) でフェードインさせる。
 * 本プラグインは react-markdown が生成した hast のテキストノードを同じ規則
 * (subSentenceSegments.splitSubSentences) で分割し、`<span class="fade-seg">` で包む。
 * 実際のフェードは CSS 側(.fade-seg)が mount 時に一度だけ再生する。
 *
 * 既表示のチャンクは、再パースで位置が変わらなければ React が DOM ノードを再利用するため
 * アニメが再生し直されない(＝新しく mount した span だけがフェードする)。放出タイミングは
 * useSmoothReveal(ペーシング)が句読点境界で visible prefix を伸ばすことで揃うので、
 * プレーンテキスト部分ではチャンク境界がずれて span テキストが書き換わることはない。
 * ただし markdown 構文がチャンク境界をまたいで後から確定する場合(例: `**強調、太字**` が
 * 閉じて段落が strong 要素へ再構成される、連続句読点が delta を跨ぐ)は、その部分の hast が
 * 作り直されて既表示 span が再マウント・再フェードすることがある(FRESTYLE-145 から続く既知の
 * 局所アーティファクト。今まさにストリーミング中の末尾に限られ全文チラつきにはならない)。
 *
 * - pre / code 配下は分割しない(コードのハイライト・コピーを壊さない)。
 * - 空白のみのセグメントは span で包まず素のテキストで残す。
 * - 非ストリーミング時(完了・履歴・テスト)はこのプラグインを付けず、素の Markdown に戻す。
 */
import { splitSubSentences } from './subSentenceSegments';

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

// pre / code 配下のテキストは分割対象外。
const SKIP_TAGS = new Set(['pre', 'code']);

function isText(node: HastChild): node is HastText {
  return node.type === 'text';
}

function isElement(node: HastChild): node is HastElement {
  return node.type === 'element';
}

function wrapText(value: string): HastChild[] {
  const out: HastChild[] = [];
  for (const seg of splitSubSentences(value)) {
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

/** ストリーミング中の本文をフェードイン用の句読点チャンク span に分割する rehype プラグイン。 */
export default function rehypeFadeSegments() {
  return (tree: HastRoot): void => {
    transform(tree);
  };
}
