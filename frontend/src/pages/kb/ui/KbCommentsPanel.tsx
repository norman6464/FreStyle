import type { KbCommentThread } from '@/entities/kb';
import KbCommentComposer from './KbCommentComposer';
import KbCommentThreadCard from './KbCommentThreadCard';

export interface KbCommentsPanelProps {
  threads: KbCommentThread[];
  loading: boolean;
  error: string | null;
  /** コメント権限が無ければ、作成フォーム・返信欄・解決/再開ボタンを出さない。読むことは誰でもできる。 */
  canComment: boolean;
  onCreateThread: (body: unknown[]) => Promise<void>;
  onReply: (threadId: string, body: unknown[]) => Promise<void>;
  onResolve: (threadId: string) => Promise<void>;
  onReopen: (threadId: string) => Promise<void>;
}

/**
 * KbCommentsPanel はコメントパネルの中身。未解決を先に、解決済みを後に並べる。
 *
 * 状態は受け取るだけで、自分では取りに行かない（取得は useKbComments が持つ）。
 * SharePanel と同じ流儀 — こうしておくと、読み込み中・空・失敗の見た目を story に
 * そのまま並べられる。
 */
export default function KbCommentsPanel({
  threads,
  loading,
  error,
  canComment,
  onCreateThread,
  onReply,
  onResolve,
  onReopen,
}: KbCommentsPanelProps) {
  const unresolved = threads.filter((thread) => thread.resolvedAt === null);
  const resolved = threads.filter((thread) => thread.resolvedAt !== null);

  return (
    <div className="flex flex-col gap-4 p-3">
      {/* 一覧の取得中は作成フォームを出さない（useKbComments 側で取得と書き込みの競合は
          解消済みだが、それでも「まだ読めていない一覧の上に新規作成を重ねさせない」という
          最低限の防御として残す — CodeRabbit 指摘）。 */}
      {canComment && !loading && (
        <div className="rounded-lg border border-surface-3 bg-surface-1 p-3">
          <h3 className="mb-1.5 text-[0.6875rem] font-bold tracking-wide text-[var(--color-text-muted)]">
            新しいスレッドを作成
          </h3>
          <KbCommentComposer placeholder="コメントを書く…" onSubmit={onCreateThread} />
        </div>
      )}

      {loading && (
        // 件数は分からないので 2 行に固定する（実際の件数に寄せると、読み込みのたびに高さが跳ねる）。
        <div className="flex flex-col gap-1.5" role="status" aria-label="コメントを読み込み中">
          <div className="h-16 animate-skeleton rounded bg-surface-2" />
          <div className="h-16 animate-skeleton rounded bg-surface-2" />
        </div>
      )}

      {!loading && error && (
        <p role="alert" className="text-xs leading-relaxed text-red-600">
          {error}
        </p>
      )}

      {!loading && !error && threads.length === 0 && (
        <p className="text-xs leading-relaxed text-[var(--color-text-muted)]">
          まだコメントはありません。
        </p>
      )}

      {!loading && !error && unresolved.length > 0 && (
        <section aria-label="未解決のスレッド">
          <h3 className="mb-1.5 text-[0.6875rem] font-bold tracking-wide text-[var(--color-text-muted)]">
            未解決（{unresolved.length}）
          </h3>
          <ul className="flex flex-col gap-2">
            {unresolved.map((thread) => (
              <li key={thread.id}>
                <KbCommentThreadCard
                  thread={thread}
                  canComment={canComment}
                  onReply={onReply}
                  onResolve={onResolve}
                  onReopen={onReopen}
                />
              </li>
            ))}
          </ul>
        </section>
      )}

      {!loading && !error && resolved.length > 0 && (
        <section aria-label="解決済みのスレッド">
          <h3 className="mb-1.5 text-[0.6875rem] font-bold tracking-wide text-[var(--color-text-muted)]">
            解決済み（{resolved.length}）
          </h3>
          <ul className="flex flex-col gap-2">
            {resolved.map((thread) => (
              <li key={thread.id}>
                <KbCommentThreadCard
                  thread={thread}
                  canComment={canComment}
                  onReply={onReply}
                  onResolve={onResolve}
                  onReopen={onReopen}
                />
              </li>
            ))}
          </ul>
        </section>
      )}
    </div>
  );
}
