import type { JSONContent } from '@tiptap/react';
import { Extension } from '@tiptap/react';
import { Plugin, PluginKey } from '@tiptap/pm/state';
import type { Transaction } from '@tiptap/pm/state';
import type { Node as ProseMirrorNode } from '@tiptap/pm/model';

/**
 * BLOCK_NODE_TYPES は blocks テーブルの1行になるノード名の一覧。
 * backend/internal/domain/block.go の domain.ValidBlockTypes と1対1（増減したら両方直す）。
 * pageRef（インラインの atom）や text・マーク（bold/italic/strike/underline/code/link）は
 * blocks テーブルの行にならないので対象外。
 */
const BLOCK_NODE_TYPES = new Set([
  'paragraph',
  'heading',
  'codeBlock',
  'blockquote',
  'bulletList',
  'orderedList',
  'listItem',
  'taskList',
  'taskItem',
  'table',
  'tableRow',
  'tableHeader',
  'tableCell',
  'image',
  'horizontalRule',
]);

/** hasStableId は attrs.id に空でない文字列が入っているかを返す（ProseMirror Node / JSON 共通）。 */
function hasStableId(id: unknown): boolean {
  return typeof id === 'string' && id.length > 0;
}

/**
 * fillMissingBlockIds は doc 内の対象ノードのうち id を持たないものへ、それぞれ別々の
 * crypto.randomUUID() を振った transaction を組み立てて返す。埋める必要が無ければ null。
 *
 * id の正当性（UUID かどうか）はサーバー（parseBlockNode）が保存のたびに検証し、
 * 通らなければ新規採番へ回す最終防衛線を持つ。ここでの役目は「保存の瞬間には
 * 各ブロックが id を持っている」ことだけを保証すること。
 */
function fillMissingBlockIds(doc: ProseMirrorNode, tr: Transaction): Transaction | null {
  let changed = false;
  doc.descendants((node, pos) => {
    if (!BLOCK_NODE_TYPES.has(node.type.name) || hasStableId(node.attrs.id)) return;
    tr.setNodeMarkup(pos, undefined, { ...node.attrs, id: crypto.randomUUID() });
    changed = true;
  });
  return changed ? tr : null;
}

/**
 * fillMissingBlockIdsInDoc は doc(JSON) を直接（transaction を経由せず）走査し、id を
 * 持たない対象ノードへ crypto.randomUUID() を振った**新しい** doc を返す（変更が無ければ
 * 同じ参照をそのまま返す — sanitizeDocLinks と同じ「構造共有」の流儀）。
 *
 * RichTextEditor が useEditor の content オプション・setContent に渡す**前**の doc を
 * ここで整えるために使う。ProseMirror の transaction を経由しないので "transaction" /
 * "update" イベントを一切発生させない（＝呼び出し元の onChange は絶対に鳴らない）。
 *
 * なぜ appendTransaction 任せにしないか: 初回ロード（useEditor の content オプション）は
 * tiptap が transaction を経由せず直接 doc を組み立てるため、appendTransaction はそもそも
 * 発火しない。エディタ生成直後に別途 transaction を dispatch して埋める案も検討したが、
 * その dispatch は React の act() の外（tiptap 内部の setTimeout 経由）で起き、
 * useEditorState が購読する 'transaction' イベント（"update" と違い preventUpdate で
 * 止められない）が act() の外で再レンダーを誘発し、テスト環境で
 * 「マウント直後の描画がまだ済んでいないタイミングと衝突する」実測の不具合を起こした。
 * doc を渡す前に埋めてしまえば transaction は 1 つも発生せず、この経路の不具合が構造的に無くなる。
 */
export function fillMissingBlockIdsInDoc<T extends JSONContent>(node: T): T {
  const nextContent = fillContentIds(node.content);
  const needsId = typeof node.type === 'string' && BLOCK_NODE_TYPES.has(node.type) && !hasStableId(node.attrs?.id);
  if (!needsId && nextContent === node.content) return node;

  const next: JSONContent = { ...node };
  if (nextContent !== undefined) next.content = nextContent;
  if (needsId) next.attrs = { ...(node.attrs ?? {}), id: crypto.randomUUID() };
  return next as T;
}

function fillContentIds(content: JSONContent[] | undefined): JSONContent[] | undefined {
  if (!Array.isArray(content)) return content;
  let changed = false;
  const next = content.map((child) => {
    const filled = fillMissingBlockIdsInDoc(child);
    if (filled !== child) changed = true;
    return filled;
  });
  return changed ? next : content;
}

/**
 * StableBlockId は編集中に各ブロックノードへ安定した id attribute を保証する ProseMirror
 * プラグイン。初回ロードの穴埋めは fillMissingBlockIdsInDoc（doc(JSON) を editor へ渡す前に
 * 整える）が担うので、ここは編集経路（appendTransaction）だけを持つ。
 *
 * サーバー（backend/internal/usecase/kb/page_usecase.go の parseBlockNode）は保存のたびに
 * attrs.id を読み、有効な UUID ならそのまま使い、無ければ新規採番する。同じ id を送り続ける
 * 限り DB 上の行（と将来のコメントの紐付け）が保たれるので、フロント側は「保存される瞬間には
 * 各ブロックが id を持っている」ことだけ保証すればよい（id の中身自体を厳密に管理する必要は無い）。
 *
 * doc が変わるトランザクションでだけ動く（tr.docChanged のチェックで無駄な処理を避ける。
 * selection-only の変更で doc が変わらずスキップしても、id の無いノードは次に doc が
 * 変わるトランザクションで拾われるので放置される心配は無い）。
 */
export const StableBlockId = Extension.create({
  name: 'stableBlockId',

  addProseMirrorPlugins() {
    return [
      new Plugin({
        key: new PluginKey('stableBlockId'),
        appendTransaction: (transactions, _oldState, newState) => {
          if (!transactions.some((transaction) => transaction.docChanged)) return null;
          return fillMissingBlockIds(newState.doc, newState.tr);
        },
      }),
    ];
  },
});
