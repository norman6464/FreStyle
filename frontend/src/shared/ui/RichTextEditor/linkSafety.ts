import type { JSONContent } from '@tiptap/core';

/**
 * linkSafety は「リンクの href として安全か」を判定・修復する純関数だけを置くモジュール。
 *
 * ここに DOM / React / tiptap のランタイム依存を持ち込まないこと。理由は 2 つある。
 * 1. スキーマの単一ソース（schemaExtensions.ts）から import するため。あちらは jsdom の無い
 *    素の Node（教材 Markdown 変換器 scripts/md2doc.mjs）からも読み込まれる。
 * 2. 判定そのものを「エディタが動いていること」に依存させないため。保存前・読み込み後といった
 *    エディタの外側でも、同じ関数で同じ結論を出せるようにしておく。
 */

/**
 * href に許可するスキーム（プロトコル）の明示リスト。
 *
 * なぜ tiptap の既定に頼らず自分で列挙するのか:
 * - tiptap の既定 isAllowedUri は http/https/ftp/ftps/mailto/tel/callto/sms/cid/xmpp に加えて
 *   「スキームが無い文字列」も通す。つまり「何を許すか」がライブラリの都合で決まっていて、
 *   版が上がったときに黙って広がりうる。防御の広さがこちらの意思と無関係に動くのは危険。
 * - 逆に「危ないものを拒否する」書き方（javascript: を弾く等）は、新しい危険スキームや
 *   表記ゆれ（大文字・制御文字混じり）が出るたびに漏れる。許可リストなら、知らないスキームは
 *   自動的に不許可側へ倒れる。
 *
 * ここに無いスキーム（javascript: / data: / vbscript: など）は例外なく不許可。
 * 相対パスは原則不許可で、唯一の例外がページ間リンク（下の INTERNAL_PAGE_LINK_PATTERN）。
 */
export const ALLOWED_LINK_PROTOCOLS: readonly string[] = ['http', 'https', 'mailto', 'tel'];

/**
 * ページ間リンク（/kb/{ページID}）の形。ノートのページを指す内部リンクで、
 * 相対パスの中では**この形だけ**を許す。任意の相対パスを開けると、利用者入力から
 * 任意の画面へ踏ませる経路（ログアウトの踏み台や、将来できるかもしれない
 * 副作用つき URL への誘導）になるため、ID は UUID の字面に固定し、
 * クエリ・フラグメント・大文字も認めない。広げる理由が出るまで最小で持つ。
 */
export const INTERNAL_PAGE_LINK_PATTERN =
  /^\/kb\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

/** isInternalPageLinkHref は「同じアプリ内のページを指すリンクか」を判定する述語。 */
export function isInternalPageLinkHref(href: unknown): boolean {
  return typeof href === 'string' && INTERNAL_PAGE_LINK_PATTERN.test(foldAsUrlParserWould(href));
}

/** doc JSON 上のリンクマーク名（ProseMirror のマーク名。tiptap の Link 拡張と一致させる）。 */
export const LINK_MARK_NAME = 'link';

/**
 * URL パーサ（WHATWG）は href を読むとき、タブ・改行・復帰を「無かったもの」として捨て、
 * 前後の C0 制御文字と空白も落とす。攻撃者はここを突いて `java<TAB>script:alert(1)` のように
 * スキーム名を割り、素朴な文字列一致の検査を素通りさせようとする。
 * 判定前に同じ規則で畳んでおけば、「ブラウザが最終的に見る文字列」で判断できる。
 */
function foldAsUrlParserWould(href: string): string {
  // 正規表現で制御文字の範囲を書くと、リテラルが不可視になり lint（no-control-regex）にも触れる。
  // 文字コードで明示的に判定する。0x09/0x0A/0x0D = タブ・改行・復帰、0x20 以下 = C0 制御文字と空白。
  let stripped = '';
  for (let i = 0; i < href.length; i += 1) {
    const code = href.charCodeAt(i);
    if (code === 0x09 || code === 0x0a || code === 0x0d) continue;
    stripped += href[i];
  }
  let start = 0;
  let end = stripped.length;
  while (start < end && stripped.charCodeAt(start) <= 0x20) start += 1;
  while (end > start && stripped.charCodeAt(end - 1) <= 0x20) end -= 1;
  return stripped.slice(start, end);
}

/**
 * sanitizeLinkHref は href を「保存してよい形」に正規化する。許可できないものは null を返す。
 *
 * 戻り値は正規化後の文字列（＝タブ・改行を抜いたもの）で、元の表記はできるだけ残す。
 * URL パーサの出力（parsed.href）で置き換えないのは、`https://example.com` が
 * `https://example.com/` に化けるなど、利用者が書いた文字列を必要以上に書き換えないため。
 */
export function sanitizeLinkHref(href: unknown): string | null {
  if (typeof href !== 'string') return null;
  const folded = foldAsUrlParserWould(href);
  if (folded === '') return null;

  // ページ間リンクだけは相対パスのまま許す（形は正規表現で固定）。
  if (INTERNAL_PAGE_LINK_PATTERN.test(folded)) return folded;

  let parsed: URL;
  try {
    parsed = new URL(folded);
  } catch {
    // それ以外の絶対 URL として読めないもの（相対パス・スキーム無し）は許可しない。
    // 「スキームが無いから安全」ではなく「安全と判断できる形になっていない」ので落とす。
    return null;
  }
  // parsed.protocol は URL パーサが小文字化した `https:` の形。末尾のコロンを外して比べる。
  const protocol = parsed.protocol.slice(0, -1);
  if (!ALLOWED_LINK_PROTOCOLS.includes(protocol)) return null;
  return folded;
}

/** isAllowedLinkHref は sanitizeLinkHref が値を返すか（＝許可できるか）だけを見る述語。 */
export function isAllowedLinkHref(href: unknown): boolean {
  return sanitizeLinkHref(href) !== null;
}

/**
 * normalizeLinkInput は「人がリンク入力欄に打った文字列」を href へ変換する。許可できなければ null。
 *
 * `example.com/a` のようにスキームを省いた入力には https:// を補う（毎回 https:// と打たせない）。
 * ただし補うのは **コロンを 1 つも含まない入力に限る**。`javascript:alert(1)` の頭に https:// を
 * 足して `https://javascript:alert(1)` として通してしまう、という補完由来の抜け道を作らないため。
 * コロンを含む入力は「スキームを自分で書いた」とみなし、許可リストに無ければそのまま不許可にする。
 */
export function normalizeLinkInput(rawInput: string): string | null {
  const trimmed = rawInput.trim();
  if (trimmed === '') return null;
  // ページ間リンクは https:// を補う前に判定する。補ってしまうと
  // `https:///kb/{ID}` がホスト名 `kb` の外部 URL として読まれ、壊れた絶対 URL が保存される。
  if (isInternalPageLinkHref(trimmed)) return sanitizeLinkHref(trimmed);
  if (trimmed.includes(':')) return sanitizeLinkHref(trimmed);
  return sanitizeLinkHref(`https://${trimmed}`);
}

type DocMark = NonNullable<JSONContent['marks']>[number];

/**
 * sanitizeDocLinks は doc JSON を歩いて、許可できない href のリンクマークを取り除く
 * （マークだけを外して文字は残す。読み手から本文が消えないようにするため）。
 * 許可できる href は正規化した値へ書き直す。
 *
 * なぜ「描画時に無害化する」だけでは足りないのか:
 * 入力・貼り付けの経路をエディタ側でいくら塞いでも、doc JSON は API から丸ごと差し込める。
 * 既に汚染された doc が DB にあれば、それを読み込んで保存し直すたびに危険な値が生き延びる。
 * 表示だけを直すやり方は「見えないところに攻撃文字列が残り続ける」状態を許すので、
 * 読み込み時と保存時の両方でこの関数を通し、doc そのものを綺麗にしておく。
 *
 * 変更が無ければ入力と同じ参照を返す（無用なコピーを避ける）。
 */
export function sanitizeDocLinks<T extends JSONContent>(node: T): T {
  const nextMarks = sanitizeMarks(node.marks);
  const nextContent = sanitizeContent(node.content);
  if (nextMarks === node.marks && nextContent === node.content) return node;

  const next: JSONContent = { ...node };
  if (nextMarks === undefined) delete next.marks;
  else next.marks = nextMarks;
  if (nextContent === undefined) delete next.content;
  else next.content = nextContent;
  return next as T;
}

function sanitizeContent(content: JSONContent[] | undefined): JSONContent[] | undefined {
  if (!Array.isArray(content)) return content;
  let changed = false;
  const next = content.map((child) => {
    const sanitized = sanitizeDocLinks(child);
    if (sanitized !== child) changed = true;
    return sanitized;
  });
  return changed ? next : content;
}

function sanitizeMarks(marks: DocMark[] | undefined): DocMark[] | undefined {
  if (!Array.isArray(marks)) return marks;
  let changed = false;
  const next: DocMark[] = [];
  for (const mark of marks) {
    if (mark.type !== LINK_MARK_NAME) {
      next.push(mark);
      continue;
    }
    const href = sanitizeLinkHref(mark.attrs?.href);
    if (href === null) {
      // 許可できないリンクはマークごと落とす（テキストは content 側に残る）。
      changed = true;
      continue;
    }
    if (href === mark.attrs?.href) {
      next.push(mark);
      continue;
    }
    changed = true;
    next.push({ ...mark, attrs: { ...mark.attrs, href } });
  }
  if (!changed) return marks;
  // マークが 1 つも残らなかったら marks 自体を落とす。tiptap の getJSON も空の marks は書かないので、
  // 空配列を残すと「同じ内容なのに JSON が違う」状態になり、保存の差分検出や往復比較が濁る。
  return next.length > 0 ? next : undefined;
}
