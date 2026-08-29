import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { NoteSidebar } from '@/widgets/note-sidebar';
import { RichTextEditor, emptyRichDoc, isRichDoc, type EditorCommand } from '@/shared/ui/RichTextEditor';
import Loading from '@/shared/ui/Loading';
import EmptyState from '@/shared/ui/EmptyState';
import { useToast } from '@/shared/lib/hooks/useToast';
import { DocumentTextIcon } from '@heroicons/react/24/outline';
import { useNotePageDoc } from '../model/useNotePageDoc';
import { createSubpage } from '../model/createSubpage';
import { subscribeNoteTreeEvents } from '@/entities/note';
import NotePageTitle from './NotePageTitle';

/**
 * NotePage はノートの画面（左にサイドバー、右に本文）。
 *
 * URL は /notes（ページ未選択）と /p/{pageId} の 2 つ。ページの URL はページ ID
 * だけを持ち、所属ワークスペースはサーバーの解決 API が返す（テナントを URL に
 * 出さない）。編集可否も同じ応答で来て、編集できる人には題名も本文もその場で書ける。
 */
export default function NotePage() {
  const { pageId } = useParams<{ pageId: string }>();
  const navigate = useNavigate();
  const { showToast } = useToast();
  const { data, loading, error, saveStatus, onDocChange, renameTitle } = useNotePageDoc(pageId);

  const handleRename = useCallback(
    async (title: string) => {
      try {
        await renameTitle(title);
      } catch (cause) {
        showToast('error', '題名を変更できませんでした');
        // 入力を保たせるため、失敗は握り潰さず投げ直す（NotePageTitle 側の約束）。
        throw cause;
      }
    },
    [renameTitle, showToast],
  );

  // 自分か祖先が物理削除されたら一覧へ戻る（消えた場所に立ち続けない）。
  // 祖先はサーバー応答（ancestors — アーカイブ済みも含む）で知っているので、
  // サイドバーの現役の木に載っていないページを開いていても正しく判定できる。
  useEffect(() => {
    if (!pageId || !data) return undefined;
    return subscribeNoteTreeEvents((event) => {
      if (event.type !== 'page-deleted') return;
      const hit =
        event.pageId === pageId || (data.ancestors ?? []).some((ancestor) => ancestor.id === event.pageId);
      if (hit) navigate('/notes');
    });
  }, [pageId, data, navigate]);

  // '/page': 子ページを作って本文にリンクを挿し、作ったページを開く。
  //
  // '/' メニューの項目はエディタ生成時に固定される（RichTextEditor の契約）ので、
  // run の closure には ref を握らせ、実行時点の最新の data を読ませる。
  // 「ページ」という業務の語彙はこの画面が持ち、エディタは項目を並べるだけ。
  const subpageContext = useRef({ data, navigate, showToast });
  subpageContext.current = { data, navigate, showToast };
  // ツールバーの描画先（ヘッダー直下の sticky バー）。callback ref で持つのは、
  // 要素が現れた瞬間に再レンダーへつなげてポータルを張り直すため（useRef だと張られない）。
  const [toolbarHost, setToolbarHost] = useState<HTMLDivElement | null>(null);
  // 題名で Enter → 本文の先頭へ（見出しから書き出しへ流れるように移る）。
  const [bodyFocusSignal, setBodyFocusSignal] = useState(0);

  const extraSlashCommands = useMemo<EditorCommand[]>(
    () => [
      {
        id: 'page',
        label: 'ページ',
        group: 'insert',
        glyph: '📄',
        keywords: ['page', 'subpage', 'child'],
        run: (editor) => {
          const ctx = subpageContext.current;
          if (!ctx.data) return;
          void createSubpage(editor, ctx.data)
            .then((path) => ctx.navigate(path))
            .catch(() => ctx.showToast('error', '子ページを作成できませんでした'));
        },
      },
    ],
    [],
  );

  return (
    <div className="flex h-full">
      <aside className="w-64 shrink-0 border-r border-surface-3 bg-surface-1">
        <NoteSidebar workspaceSlug={data?.workspaceSlug} activePageId={pageId} />
      </aside>

      <main className="min-w-0 flex-1 overflow-y-auto">
        {/*
          書式ツールバーはヘッダー直下に固定する（題名より上）。本文の途中で書式を
          変えるときにスクロールで手が届かなくならないよう、sticky で留める。
          中身はエディタがポータルで描き込む（editor はエディタの中に閉じたまま）。
        */}
        {pageId && !loading && !error && data?.canEdit && (
          // z-30: 本文側の浮遊 UI（コードブロックの操作バー z-10・その言語メニュー z-30）
          // より上に出す。選択中に出るバブル（z-50）はこの上のまま — あちらは
          // 選んだ場所に出るので、覆われると書式を変えられない。
          <div className="sticky top-0 z-30 border-b border-surface-3 bg-surface">
            <div ref={setToolbarHost} className="mx-auto w-full max-w-3xl px-6 py-1.5" />
          </div>
        )}
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
              {/*
                パンくず（場所の表示）。ワークスペース名 → 閲覧できる祖先 → 現在のページ。
                見えない祖先は応答に含まれず、穴があいたまま出す（木と同じ見え方。
                フロントで埋めると、サーバーが伏せた実在を推測で喋ることになる）。
              */}
              <nav aria-label="ページの場所" className="mb-2 flex min-w-0 flex-wrap items-center gap-1 text-xs text-[var(--color-text-muted)]">
                <span className="truncate">{data.workspaceName ?? data.workspaceSlug}</span>
                {/* ?? [] はデプロイ順の防御 — 旧バックエンドの応答（ancestors なし）でも落とさない */}
                {(data.ancestors ?? []).map((ancestor) => (
                  <span key={ancestor.id} className="flex min-w-0 items-center gap-1">
                    <span aria-hidden="true">/</span>
                    <Link
                      to={`/p/${ancestor.id}`}
                      className="max-w-40 truncate hover:text-[var(--color-text-primary)] hover:underline"
                    >
                      {ancestor.title}
                    </Link>
                  </span>
                ))}
                {/* 区切りは題名と組にして折り返す（独立させると「/」だけが行末に残る） */}
                <span className="flex min-w-0 items-center gap-1">
                  <span aria-hidden="true">/</span>
                  <span aria-current="page" className="max-w-40 truncate text-[var(--color-text-secondary)]">
                    {data.page.title}
                  </span>
                </span>
              </nav>
              {/* ページごとに作り直す（別ページへ移った瞬間、打ちかけの下書きを持ち越さない） */}
              <NotePageTitle
                key={data.page.id}
                title={data.page.title}
                canEdit={data.canEdit}
                onRename={handleRename}
                onEnter={() => setBodyFocusSignal((prev) => prev + 1)}
              />
              <RichTextEditor
                // doc は API から来る任意の JSON。形が違えば空の本文として扱い、画面を落とさない。
                value={isRichDoc(data.doc) ? data.doc : emptyRichDoc()}
                editable={data.canEdit}
                onChange={onDocChange}
                saveStatus={data.canEdit ? saveStatus : 'idle'}
                ariaLabel={`${data.page.title} の本文`}
                extraSlashCommands={data.canEdit ? extraSlashCommands : undefined}
                // host が現れてから出す。最初の描画で host はまだ null なので、
                // 常に true にすると一瞬だけ本文の直上に出てから跳ぶ（ちらつく）。
                toolbar={toolbarHost !== null}
                toolbarContainer={toolbarHost}
                onNavigateToPage={(path) => navigate(path)}
                focusSignal={bodyFocusSignal}

              />
            </article>
          )}
        </div>
      </main>
    </div>
  );
}
