import type { TicketCommentSegment } from '@/entities/ticket';

export interface TicketCommentBodyProps {
  body: TicketCommentSegment[];
  /** userId から表示名を引く。引けなければ null（画面側が「不明なユーザー」に落とす）。 */
  resolveMentionName: (userId: string) => string | null;
}

/**
 * 発言の本文を素の React で描く。
 *
 * 発言 1 件ごとに入力欄（tiptap）を立てる作りは採らない — エディタ 1 台につき
 * 内部の拡張が 92 個立つ（実測）。20 件のスレッドで 1,840 個になる。
 */
export default function TicketCommentBody({ body, resolveMentionName }: TicketCommentBodyProps) {
  return (
    <p className="whitespace-pre-wrap break-words text-sm text-[var(--color-text-secondary)]">
      {body.map((segment, i) => {
        if (segment.kind === 'mention') {
          const name = resolveMentionName(segment.userId);
          return (
            <span key={i} className="rounded bg-brand-50 px-1 text-brand-700">
              @{name ?? '不明なユーザー'}
            </span>
          );
        }
        return <span key={i}>{segment.text}</span>;
      })}
    </p>
  );
}
