import type { TicketCommentReaction } from '@/entities/ticket';

export interface ReactionSummary {
  emoji: string;
  count: number;
  /** 自分（currentUserId）がこの絵文字を押しているか。 */
  mine: boolean;
}

/**
 * 発言 1 件の反応（1 人 1 絵文字ずつのフラットな列）を、絵文字ごとの件数へ畳む。
 *
 * 並びは初出順（backend の応答順のまま）。押した人の名前は userId しか来ず
 * 引けないため、ここでは持たない（件数と自分が押しているかだけを返す）。
 *
 * currentUserId が null（自分の userId がまだ引けていない）のときは、
 * 何を押しても「自分が押しているか」を判断できないので全部 false にする。
 */
export function summarizeReactions(
  reactions: TicketCommentReaction[],
  currentUserId: number | null,
): ReactionSummary[] {
  const order: string[] = [];
  const counts = new Map<string, number>();
  const mineSet = new Set<string>();

  for (const reaction of reactions) {
    if (!counts.has(reaction.emoji)) {
      counts.set(reaction.emoji, 0);
      order.push(reaction.emoji);
    }
    counts.set(reaction.emoji, (counts.get(reaction.emoji) ?? 0) + 1);
    if (currentUserId !== null && reaction.userId === currentUserId) {
      mineSet.add(reaction.emoji);
    }
  }

  return order.map((emoji) => ({ emoji, count: counts.get(emoji) as number, mine: mineSet.has(emoji) }));
}
