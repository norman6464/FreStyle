import { useParams } from 'react-router-dom';
import { KbSidebar } from '@/widgets/kb-sidebar';
import { RichTextEditor, emptyRichDoc, isRichDoc } from '@/shared/ui/RichTextEditor';
import Loading from '@/shared/ui/Loading';
import EmptyState from '@/shared/ui/EmptyState';
import { DocumentTextIcon } from '@heroicons/react/24/outline';
import { useKbPageDoc } from '../model/useKbPageDoc';

/**
 * KnowledgeBasePage はナレッジ基盤の画面（左にサイドバー、右に本文）。
 *
 * いまは**読むだけ**。作る・名前を変える・動かすは次の段で足す。
 * 既存の /notes（リッチ文書）とは別系統で、あちらには手を触れない。2 系統が並ぶのは
 * 移行の判断を先送りするためだが、先送りした分だけ移行は重くなる。期限を決めること。
 */
export default function KnowledgeBasePage() {
  const { workspaceSlug, pageId } = useParams<{ workspaceSlug: string; pageId: string }>();
  const { data, loading, error } = useKbPageDoc(workspaceSlug, pageId);

  return (
    <div className="flex h-full">
      <aside className="w-64 shrink-0 border-r border-surface-3 bg-surface-1">
        <KbSidebar workspaceSlug={workspaceSlug} activePageId={pageId} />
      </aside>

      <main className="min-w-0 flex-1 overflow-y-auto">
        <div className="mx-auto w-full max-w-3xl px-6 py-10">
          {!pageId && (
            <EmptyState
              icon={DocumentTextIcon}
              title="ページを選んでください"
              description="左のツリーからページを開きます。"
            />
          )}

          {pageId && loading && <Loading className="py-16" />}

          {/*
            404 は「無い」と「見えない」の両方。どちらかを名指しすると、
            ID を総当たりするだけで隠したページの実在が分かってしまう。
          */}
          {pageId && !loading && error && (
            <EmptyState
              icon={DocumentTextIcon}
              title="ページを開けません"
              description={error}
            />
          )}

          {pageId && !loading && !error && data && (
            <article>
              <h1 className="mb-4 text-3xl font-bold text-[var(--color-text-primary)] md:text-4xl">
                {data.page.title}
              </h1>
              <RichTextEditor
                // doc は API から来る任意の JSON。形が違えば空の本文として扱い、画面を落とさない。
                value={isRichDoc(data.doc) ? data.doc : emptyRichDoc()}
                editable={false}
                ariaLabel={`${data.page.title} の本文`}
              />
            </article>
          )}
        </div>
      </main>
    </div>
  );
}
