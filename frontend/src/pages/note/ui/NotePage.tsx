import { useParams } from 'react-router-dom';
import { NoteSidebar } from '@/widgets/note-sidebar';
import { RichTextEditor, emptyRichDoc, isRichDoc } from '@/shared/ui/RichTextEditor';
import Loading from '@/shared/ui/Loading';
import EmptyState from '@/shared/ui/EmptyState';
import { DocumentTextIcon } from '@heroicons/react/24/outline';
import { useNotePageDoc } from '../model/useNotePageDoc';

/**
 * NotePage はノートの画面（左にサイドバー、右に本文）。
 *
 * URL は /notes（ページ未選択）と /p/{pageId} の 2 つ。ページの URL はページ ID
 * だけを持ち、所属ワークスペースはサーバーの解決 API が返す（テナントを URL に
 * 出さない）。編集可否も同じ応答で来て、編集できる人にはそのまま書ける本文を出す。
 */
export default function NotePage() {
  const { pageId } = useParams<{ pageId: string }>();
  const { data, loading, error, saveStatus, onDocChange } = useNotePageDoc(pageId);

  return (
    <div className="flex h-full">
      <aside className="w-64 shrink-0 border-r border-surface-3 bg-surface-1">
        <NoteSidebar workspaceSlug={data?.workspaceSlug} activePageId={pageId} />
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
                editable={data.canEdit}
                onChange={onDocChange}
                saveStatus={data.canEdit ? saveStatus : 'idle'}
                ariaLabel={`${data.page.title} の本文`}
              />
            </article>
          )}
        </div>
      </main>
    </div>
  );
}
