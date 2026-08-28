import { Node } from '@tiptap/core';
import type { Extensions } from '@tiptap/core';
import Code from '@tiptap/extension-code';
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight';
import Heading from '@tiptap/extension-heading';
import Image from '@tiptap/extension-image';
import { Link } from '@tiptap/extension-link';
import { TaskItem, TaskList } from '@tiptap/extension-list';
import { TableKit } from '@tiptap/extension-table';
import StarterKit from '@tiptap/starter-kit';
import { common, createLowlight } from 'lowlight';
import { isAllowedLinkHref, isInternalPageLinkHref, sanitizeLinkHref } from './linkSafety';

/**
 * lowlight のインスタンス（highlight.js の common 言語 37 種を登録）。
 * codeBlock スキーマを configure する本モジュールが所有する。エディタ側では
 * トークンの配色に app 全体へ import 済みの Atom One Light テーマ（.hljs-*）が当たる。
 */
export const lowlight = createLowlight(common);

/**
 * CombinableCode は他のマークと共存できるインラインコード。
 *
 * 既定の Code は `excludes: '_'`（＝他の全マークを排他）で、コードを掛けると
 * 太字・斜体・下線・打ち消しがすべて外れてしまう。ノートでは「コード＋太字」等を
 * 重ねたい場面があるため、排他指定を解いて併用できるようにする。マーク名は 'code' のまま。
 */
const CombinableCode = Code.extend({ excludes: '' });

/**
 * リンクを描画するときに必ず付ける固定属性。
 *
 * - `target="_blank"`: 書きかけの本文があるタブを潰さないよう、別タブで開く。
 * - `rel="noopener"`: 開いた先から `window.opener` 越しに元のタブを別ページへ差し替えられる
 *   （reverse tabnabbing ＝ 偽のログイン画面へのすり替え）のを防ぐ。
 * - `rel="noreferrer"`: 遷移先に Referer（＝社内ページの URL）を渡さない。
 * - `rel="nofollow"`: 利用者が自由に書ける外部リンクへ検索評価を渡さない（スパム対策）。
 *
 * 要点は、これらを **doc の attrs に持たせず、描画のたびに固定で付ける**こと。
 * tiptap の既定は target / rel を「マークの属性」として doc に保存するため、
 * `<a href="…" rel="" target="_self">` を貼り付けるだけでその値が保存され、
 * たった 1 行の細工で上の防御が外れてしまう。描画時に固定すれば、保存内容が何であれ必ず付く。
 */
const LINK_RENDER_ATTRIBUTES: Record<string, string> = {
  target: '_blank',
  rel: 'noopener noreferrer nofollow',
};

/**
 * SafeLink はリンクマーク（`link`）。href に許可スキームを明示した形で固めてある。
 *
 * リンクの href は利用者が自由に書けるので、そのまま通すと `javascript:alert(1)` のような
 * 「押すとスクリプトが走る URL」を仕込めてしまう（XSS）。塞ぐべき経路は 3 つあり、
 * 1 つでも空いていれば残りを固めても意味がない。
 *
 *   1. 入力（打ち込み・autolink）    → `isAllowedUri` を差し替えて Link 拡張の全判定を自前にする
 *   2. 貼り付け（HTML の取り込み）   → `href` の `parseHTML` でも同じ関数を通す
 *   3. 保存・再読込の往復             → doc JSON 側を `sanitizeDocLinks`（linkSafety.ts）で洗う。
 *                                       これは RichTextEditor と md2doc が担当する
 *
 * ここで押さえるのは 1 と 2、そして「表示の最後の砦」としての `renderHTML`。
 * 3 は doc JSON が API 経由で丸ごと差し込めるため、エディタの入力経路だけを見ても塞げない。
 */
const SafeLink = Link.configure({
  // 入力経路: URL を打って空白などで区切ると自動でリンクになる。
  autolink: true,
  // `[文字](URL)` という Markdown 記法も入力・貼り付けから拾う（href の可否は isAllowedUri が見る）。
  markdownLinks: true,
  // スキームを省いて `example.com` と書かれたときに補うスキーム。既定は 'http' なので https にする。
  defaultProtocol: 'https',
  // protocols は linkify に「これも URL として認識してよい」と教える口であって、安全判定ではない。
  // 判定は下の isAllowedUri（＝ ALLOWED_LINK_PROTOCOLS）へ一本化したいのでここは空のままにする。
  protocols: [],
  // 【安全判定の一本化】tiptap の既定判定を丸ごと差し替える。
  // 既定は ftp/ftps/callto/sms/cid/xmpp なども、さらに「スキームが無い文字列」も通す。
  // つまり許可範囲がライブラリの都合で決まり、版が上がると黙って広がりうる。
  // Link 拡張は入力・貼り付け・setLink/toggleLink・HTML 解析・描画のすべてでこの関数を呼ぶので、
  // ここを自前の許可リストに差し替えれば、経路ごとの取りこぼしが起きにくくなる。
  isAllowedUri: (uri) => isAllowedLinkHref(uri),
  // 編集中にリンクを踏んで画面が飛ぶのを防ぐ（クリックはキャレット移動として扱う）。
  // 読み取り専用（editable=false）では tiptap のクリックハンドラが降りるので、素の <a> として開く。
  openOnClick: false,
}).extend({
  addAttributes() {
    // 既定の Link は href / target / rel / class / title を doc に保存する。
    // target / rel / class は保存せず描画時に固定する（LINK_RENDER_ATTRIBUTES のコメント参照）ので、
    // doc に残すのは href と title だけにする。保存する値が減るほど、細工できる余地も減る。
    return {
      href: {
        default: null,
        // 貼り付けた HTML から href を読む経路。ここでも同じ関数を通し、
        // 通らない値は null にして「href の無いリンク」に落とす（マーク自体は
        // 親の parseHTML ルールが isAllowedUri で弾くので、実際には二重の壁になる）。
        parseHTML: (element) => sanitizeLinkHref(element.getAttribute('href')),
      },
      title: { default: null },
    };
  },

  renderHTML({ HTMLAttributes }) {
    const href = sanitizeLinkHref(HTMLAttributes.href);
    if (href === null) {
      // 表示経路の最後の砦。万一 doc に許可できない href が残っていても <a> にはしない。
      // href="" の <a> にすると「押せるのにどこへも行かない要素」が残るため、span で出す。
      return ['span', {}, 0];
    }
    // ページ間リンク（/p/…）は同じタブで開く。_blank と rel の束は外部サイト向けの防御
    // （tabnabbing・Referer 漏れ・スパム評価）で、同一アプリ内の遷移には当てはまらない。
    const attributes: Record<string, string> = isInternalPageLinkHref(href)
      ? { href }
      : { ...LINK_RENDER_ATTRIBUTES, href };
    if (typeof HTMLAttributes.title === 'string' && HTMLAttributes.title !== '') {
      attributes.title = HTMLAttributes.title;
    }
    return ['a', attributes, 0];
  },
});

/**
 * PageRef は「ページ参照」— ノートのページを指すインラインの 1 要素（atom）。
 *
 * 文字を持たず、表示は attrs.title（**表示のための写し**）。題名の正本はページ側にあり、
 * サーバーが読み出しのたびに「読み手が閲覧できる参照だけ」現在の題名へ差し替える。
 * atom なので文字は編集できない（編集できると「題名に追従する」約束が壊れる）し、
 * Backspace で 1 要素として消える。
 *
 * href は pageId から組み立てる。pageId は貼り付けや API 由来の doc からも入るので、
 * UUID の字面（INTERNAL_PAGE_LINK_PATTERN と同じ形）を検証し、通らなければ
 * リンクにしない（押せるのにどこへも行かない要素を作らないため span で出す）。
 */
const PAGE_REF_UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

export const PageRef = Node.create({
  name: 'pageRef',
  group: 'inline',
  inline: true,
  atom: true,
  selectable: true,

  addAttributes() {
    return {
      pageId: { default: null },
      title: { default: null },
    };
  },

  parseHTML() {
    return [
      {
        tag: 'a[data-page-ref]',
        // SafeLink の a[href] 規則より先に効かせる（同じ <a> に両方が一致し、
        // 既定の優先度だとリンクマークが勝って、コピー＆ペーストで参照が
        // ただのリンクに劣化する）。
        priority: 100,
        getAttrs: (element) => {
          const pageId = element.getAttribute('data-page-id');
          // ID の形をここでも検証する。貼り付けは外部の HTML からも来るので、
          // 通らないものは参照として取り込まない（この規則ごと不一致にし、
          // 後続の規則＝リンクや素の文字に落とす）。
          if (pageId === null || !PAGE_REF_UUID_PATTERN.test(pageId)) return false;
          return { pageId, title: element.textContent || null };
        },
      },
    ];
  },

  renderHTML({ node }) {
    const pageId = typeof node.attrs.pageId === 'string' ? node.attrs.pageId : '';
    const title = typeof node.attrs.title === 'string' && node.attrs.title !== ''
      ? node.attrs.title
      : 'ページ';
    if (!PAGE_REF_UUID_PATTERN.test(pageId)) {
      return ['span', { 'data-page-ref': 'true' }, title];
    }
    // 同一アプリ内の遷移なので _blank や rel の束は付けない（SafeLink の内部リンクと同じ扱い）。
    return [
      'a',
      { 'data-page-ref': 'true', 'data-page-id': pageId, href: `/p/${pageId}`, class: 'rte-page-ref' },
      title,
    ];
  },

  // editor.getText() やプレーンテキスト化で参照が消えないよう、題名を文字として出す。
  renderText({ node }) {
    return typeof node.attrs.title === 'string' && node.attrs.title !== ''
      ? node.attrs.title
      : 'ページ';
  },
});

/** createSchemaExtensions の組み立てオプション。 */
export interface CreateSchemaExtensionsOptions {
  /** 画像ノードをスキーマに含めるか（既定 true）。 */
  image?: boolean;
}

/**
 * createSchemaExtensions は「ドキュメントのスキーマ（ノード/マーク名・attrs・content 式）を
 * 決める」拡張だけを組み立てる共有 factory。
 *
 * エディタ（editorExtensions.ts）と、Node で動く教材 Markdown 変換器（scripts/md2doc.mjs）が
 * 同一スキーマで doc(JSON) を扱うための単一ソース。変換器は本ファイルを jsdom なしの Node から
 * 直接 import するため、ここには React・CSS・DOM 依存を（推移的にも）置かないこと。
 * NodeView・input rule・プレースホルダ等の表示/入力の挙動は editorExtensions.ts 側で上掛けする。
 *
 * ノード/マークを足すときに、エディタ側だけへ足してはいけない理由はここにある。
 * 例えばリンクをエディタにだけ足すと、教材の `[文字](URL)` は変換器側でマークを持てず
 * ただの文字列になり、同じ文書がエディタと教材で別物に見える（逆に変換器にだけ足すと、
 * 教材が生成した doc をエディタが開けない）。スキーマは必ずこの factory に 1 か所で置く。
 */
export function createSchemaExtensions(
  options: CreateSchemaExtensionsOptions = {},
): Extensions {
  const { image = true } = options;

  const extensions: Extensions = [
    // StarterKit の code は排他指定、heading は levels 無制限、codeBlock はハイライトなし、
    // link は許可スキームが tiptap 既定任せのため、それぞれ無効化してこちらの拡張へ差し替える。
    StarterKit.configure({ heading: false, code: false, codeBlock: false, link: false }),
    CombinableCode,
    // リンク。href の許可スキームを明示した SafeLink（エディタと教材変換器で同じ判定を使う）。
    SafeLink,
    // 見出しは 1〜3 のみ（エディタ UI・教材の章構造とも 3 段で揃える）。
    Heading.configure({ levels: [1, 2, 3] }),
    // 構文ハイライト付きコードブロック。ノード名は 'codeBlock' のまま既存 doc と互換。
    CodeBlockLowlight.configure({ lowlight, defaultLanguage: 'plaintext' }),
    // 表（GFM テーブル相当）。教材の本文とノートの両方で使う。
    // resizable は列幅ドラッグ UI が必要になるため、まずは固定幅で表現力を優先する。
    TableKit.configure({ table: { resizable: false } }),
    // タスクリスト（チェックボックス）。教材のチェックリスト章とノートの TODO で使う。
    TaskList,
    TaskItem.configure({ nested: true }),
    // ページ参照（インラインの atom）。題名はサーバーが読み出し時に解決する。
    PageRef,
  ];

  if (image) {
    extensions.push(Image.configure({ inline: false, allowBase64: false }));
  }

  return extensions;
}
