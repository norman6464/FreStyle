import type { EditorState } from '@tiptap/pm/state';
import type { ResolvedPos } from '@tiptap/pm/model';

/**
 * COMMENT_ANCHOR_BLOCK_TYPES は「錨付きコメント」の対象になりうるブロックノード名の一覧。
 * backend/internal/domain/block.go の domain.ValidBlockTypes と1対1で、
 * stableBlockId.ts の BLOCK_NODE_TYPES とも同じ内容（＝blocks テーブルの1行になるノード）。
 * 2 箇所に同じ集合を持たせているのは、schemaExtensions.ts（教材変換器と共有）へ
 * 依存を増やしたくないため — 増減したら両方（このファイルと stableBlockId.ts）を直す、
 * という stableBlockId.ts 側のコメントに書かれた既存の慣習にそのまま倣う。
 */
const COMMENT_ANCHOR_BLOCK_TYPES = new Set([
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

/** CommentAnchor は「本文のどこを指しているか」のスナップショット。 */
export interface CommentAnchor {
  /** 錨を張ったブロックノードの安定 id（stableBlockId.ts が保証する attrs.id）。 */
  blockId: string;
  /** ブロックの内容開始位置からの相対オフセット（開始側）。下のコメント参照。 */
  anchorFrom: number;
  /** ブロックの内容開始位置からの相対オフセット（終了側）。 */
  anchorTo: number;
  /** 選択していた文字列（前後の空白を trim 済み）。人が読める手がかり。 */
  quote: string;
}

/** hasStableId は attrs.id に空でない文字列が入っているかを返す（stableBlockId.ts と同じ判定）。 */
function hasStableId(id: unknown): id is string {
  return typeof id === 'string' && id.length > 0;
}

/**
 * findAnchorBlock は resolved position から depth を「選択位置の深さ→1」の順に浅くしながら、
 * COMMENT_ANCHOR_BLOCK_TYPES に含まれる型で、かつ id を持つノードを探す。
 *
 * 深い方から探すことで、ネストしたノード（listItem の中の paragraph 等）では
 * **最も内側の id 付きノード**（paragraph）が選ばれる。外側（listItem）まで
 * 遡ってしまうと、リスト全体にコメントが付いたかのような粒度になり、選んだ文と
 * 対応しなくなるため。
 */
function findAnchorBlock(pos: ResolvedPos): { depth: number; blockId: string } | null {
  for (let depth = pos.depth; depth >= 1; depth -= 1) {
    const node = pos.node(depth);
    if (COMMENT_ANCHOR_BLOCK_TYPES.has(node.type.name) && hasStableId(node.attrs.id)) {
      return { depth, blockId: node.attrs.id as string };
    }
  }
  return null;
}

/**
 * resolveCommentAnchor は現在の選択範囲から「錨付きコメント」の元になる情報を計算する。
 *
 * 計算できない（＝コメントボタンを無効化すべき）場合は null を返す:
 *   - 選択が空（カーソルのみ）
 *   - 選択が複数ブロックにまたがる
 *   - id を持たないブロック（例: 保存前で StableBlockId が未採番）の中にいる
 *   - 選択文字列が空白だけで、trim すると空になる
 *
 * anchorFrom / anchorTo は **doc 全体の position ではなく、そのブロックノードの
 * 「コンテンツ開始位置」（$from.start(depth)）からの相対位置**にする。理由:
 * doc 全体の position はドキュメント内の他の場所（前のブロックの追加・削除等）が
 * 変わるだけでずれる。保存後にブロックの外側で編集があっても、**このブロック自身の
 * 中身が変わらない限り**相対オフセットはずれない。ブロック単位で錨を持たせる設計
 * （backend の block_id + ブロック内オフセット）と直接対応する。
 *
 * 本文編集への追従（re-anchoring）はスコープ外 — 錨はスナップショットで、編集後に
 * ズレても quote が人間可読な手がかりとして残るだけでよい（将来の改善候補）。
 */
export function resolveCommentAnchor(state: EditorState): CommentAnchor | null {
  const { selection } = state;
  if (selection.empty) return null;

  const { $from, $to } = selection;

  const from = findAnchorBlock($from);
  if (!from) return null;
  const blockStart = $from.start(from.depth);

  const to = findAnchorBlock($to);
  // to 側が見つからない、または to 側のブロック開始位置が from 側と一致しない場合は
  // 選択が複数ブロックにまたがっている（同じブロックの中に収まっていない）。
  if (!to || $to.start(to.depth) !== blockStart) return null;

  const quote = state.doc.textBetween($from.pos, $to.pos, ' ').trim();
  if (quote === '') return null;

  return {
    blockId: from.blockId,
    anchorFrom: $from.pos - blockStart,
    anchorTo: $to.pos - blockStart,
    quote,
  };
}
