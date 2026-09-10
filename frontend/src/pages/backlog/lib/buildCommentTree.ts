import type { TicketComment } from '@/entities/ticket';

/** 幹 1 件と、その下に平らに並んだ返信。 */
export interface CommentThread {
  root: TicketComment;
  /**
   * 返信。幹への直接の返信でなければ（＝孫以降）宛先（返信先の発言 ID）を持つ。
   * 表示側はこの ID から宛先の発言を引いて「◂ 誰々 へ」を添える。
   */
  replies: Array<{ comment: TicketComment; replyToId: string | null }>;
}

/**
 * backend は `created_at ASC, id ASC` のフラット配列を返し、木は組まない。深さも制限しない。
 * この関数がここで木を組み立てる規則を固定する:
 *
 * - `parentCommentId` が null → 幹
 * - 幹を指す → その幹の返信（宛先なし）
 * - 返信を指す（孫以降） → 親を辿って最上位の祖先（幹）を求め、その幹の返信列へ
 *   **平坦化**する（表示のインデントは 1 段までのため）。直接の返信先が幹自身でなければ
 *   宛先を持たせる
 * - **親を辿る途中で見つからない（親だけ削除された孤児）** → そこで辿るのを止め、
 *   見つからなかった時点のノード自身を幹にする。その下にぶら下がる返信は、
 *   この昇格した幹へ平坦化される（孤児の子孫まで個別に幹へ昇格させない）
 * - 親を辿って辿ってきた経路に戻る（循環。`parentCommentId` は作成時に固定で
 *   既存の発言しか指せないため実データでは起こらないはずだが、防御として置く）
 *   → その場で辿るのを止める（無限ループにしない。ただし循環の当事者どうしが
 *   互いを見て別々に「自分が幹」と判断しうるため、この形だけは重複して現れる
 *   ことがある——実データでは作れない形なので、これ以上は追わない）
 * - 並びは幹も返信列も作成日時の昇順（backend の順のまま。幹の並びは各幹が
 *   ぶら下がる相手として最初に現れた時刻＝実質その幹自身の作成時刻の順になる）
 */
export function buildCommentTree(comments: TicketComment[]): CommentThread[] {
  const byId = new Map(comments.map((c) => [c.id, c]));
  const threads = new Map<string, CommentThread>();
  const order: string[] = [];

  /** comment から親を辿り、実質の幹の ID と、幹への直接の返信でなければ宛先 ID を返す。 */
  function resolveRoot(comment: TicketComment): { rootId: string; replyToId: string | null } {
    const seen = new Set<string>([comment.id]);
    let current = comment;
    while (current.parentCommentId !== null) {
      const parent = byId.get(current.parentCommentId);
      if (!parent || seen.has(parent.id)) break; // 孤児 or 循環。current をここで幹にする。
      seen.add(parent.id);
      current = parent;
    }
    if (current.id === comment.id) return { rootId: comment.id, replyToId: null };
    const replyToId = comment.parentCommentId === current.id ? null : comment.parentCommentId;
    return { rootId: current.id, replyToId };
  }

  function ensureThread(rootId: string): CommentThread {
    let thread = threads.get(rootId);
    if (!thread) {
      thread = { root: byId.get(rootId) as TicketComment, replies: [] };
      threads.set(rootId, thread);
      order.push(rootId);
    }
    return thread;
  }

  for (const comment of comments) {
    const { rootId, replyToId } = resolveRoot(comment);
    if (rootId === comment.id) {
      ensureThread(comment.id).root = comment;
    } else {
      ensureThread(rootId).replies.push({ comment, replyToId });
    }
  }

  return order.map((id) => threads.get(id) as CommentThread);
}
