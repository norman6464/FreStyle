import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import { KbSidebar } from '@/widgets/kb-sidebar';
import { SecondaryPanel } from '@/widgets/secondary-panel';
import { RichTextEditor, emptyRichDoc, isRichDoc, type EditorCommand } from '@/shared/ui/RichTextEditor';
import Loading from '@/shared/ui/Loading';
import EmptyState from '@/shared/ui/EmptyState';
import { useToast } from '@/shared/lib/hooks/useToast';
import { useMobilePanelState } from '@/shared/lib/hooks/useMobilePanelState';
import { DocumentTextIcon, Bars3Icon } from '@heroicons/react/24/outline';
import { useKbPageDoc } from '../model/useKbPageDoc';
import { createSubpage } from '../model/createSubpage';
import { resolveEntryPageId } from '../model/resolveEntryPage';
import { subscribeKbTreeEvents, type KbIcon } from '@/entities/kb';
import KbPageTitle from './KbPageTitle';
import KbPageIconButton from './KbPageIconButton';
import KbPageMeta from './KbPageMeta';
import { SharePanel } from '@/features/permission-sharing';
import { useKbShare } from '../model/useKbShare';

/**
 * KbPage はナレッジの画面（左にサイドバー、右に本文）。
 *
 * URL は素の /kb（ページ未選択）と /kb/{pageId} の 2 つ。ページの URL はページ ID
 * だけを持ち、所属ワークスペースはサーバーの解決 API が返す（テナントを URL に
 * 出さない）。編集可否も同じ応答で来て、編集できる人には題名も本文もその場で書ける。
 *
 * ページ未選択（素の /kb）では、続きを resolveEntryPageId に決めさせて
 * /kb/{pageId} へ即座に移る（見せるための画面ではなく、素通りする入口）。
 */
export default function KbPage() {
  const { pageId } = useParams<{ pageId: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const { showToast } = useToast();
  const { data, loading, error, saveStatus, onDocChange, renameTitle, changeIcon } = useKbPageDoc(pageId);
  // ヘッダー/サイドバーのワークスペース切替から来たときだけ渡ってくる。
  // ページを開いているときは data.workspaceSlug が正なのでそちらを優先する。
  const navigationWorkspaceSlug = (location.state as { workspaceSlug?: string } | null)?.workspaceSlug;
  const { isOpen: mobilePanelOpen, open: openMobilePanel, close: closeMobilePanel } = useMobilePanelState();

  // ページ未選択(素の /kb)のときだけ動く。続きのページが決まり次第そこへ移るので、
  // 「まだページがありません」を出すのは resolveEntryPageId が null を返したときだけ。
  const [entryResolving, setEntryResolving] = useState(false);
  useEffect(() => {
    if (pageId) return undefined;
    let cancelled = false;
    setEntryResolving(true);
    resolveEntryPageId(navigationWorkspaceSlug)
      .then((id) => {
        if (cancelled) return;
        if (id) {
          navigate(`/kb/${id}`, { replace: true });
          return;
        }
        setEntryResolving(false);
      })
      .catch(() => {
        if (!cancelled) setEntryResolving(false);
      });
    return () => {
      cancelled = true;
    };
  }, [pageId, navigationWorkspaceSlug, navigate]);

  const handleRename = useCallback(
    async (title: string) => {
      try {
        await renameTitle(title);
      } catch (cause) {
        showToast('error', '題名を変更できませんでした');
        // 入力を保たせるため、失敗は握り潰さず投げ直す（KbPageTitle 側の約束）。
        throw cause;
      }
    },
    [renameTitle, showToast],
  );

  const handleChangeIcon = useCallback(
    async (icon: KbIcon | null) => {
      try {
        await changeIcon(icon);
      } catch (cause) {
        showToast('error', icon ? 'アイコンを変更できませんでした' : 'アイコンを外せませんでした');
        // ピッカーを開いたままにするため、握り潰さず投げ直す（KbPageIconPicker 側の約束）。
        throw cause;
      }
    },
    [changeIcon, showToast],
  );

  // 自分か祖先が物理削除されたら一覧へ戻る（消えた場所に立ち続けない）。
  // 祖先はサーバー応答（ancestors — アーカイブ済みも含む）で知っているので、
  // サイドバーの現役の木に載っていないページを開いていても正しく判定できる。
  // ワークスペースごと削除されたとき（配下は FK CASCADE で全消去）も同じ理由で戻る。
  useEffect(() => {
    if (!pageId || !data) return undefined;
    return subscribeKbTreeEvents((event) => {
      if (event.type === 'page-deleted') {
        const hit =
          event.pageId === pageId || (data.ancestors ?? []).some((ancestor) => ancestor.id === event.pageId);
        if (hit) navigate('/kb');
        return;
      }
      if (event.type === 'workspace-deleted' && event.workspaceSlug === data.workspaceSlug) {
        navigate('/kb');
      }
    });
  }, [pageId, data, navigate]);

  // '/page': 子ページを作って本文にリンクを挿し、作ったページを開く。
  //
  // '/' メニューの項目はエディタ生成時に固定される（RichTextEditor の契約）ので、
  // run の closure には ref を握らせ、実行時点の最新の data を読ませる。
  // 「ページ」という業務の語彙はこの画面が持ち、エディタは項目を並べるだけ。
  const subpageContext = useRef({ data, navigate, showToast });
  subpageContext.current = { data, navigate, showToast };
  // 題名で Enter → 本文の先頭へ（見出しから書き出しへ流れるように移る）。
  const [bodyFocusSignal, setBodyFocusSignal] = useState(0);
  // 共有パネルの開閉。ページを移ったら必ず閉じる（別のページの設定を開いたまま
  // 題名だけ変わると、どのページを共有しているのか読めなくなる）。
  const [shareOpen, setShareOpen] = useState(false);
  useEffect(() => {
    setShareOpen(false);
  }, [pageId]);
  // 閉じている間は取りに行かない（開いていないパネルのために毎ページ 2 本引かない）。
  const share = useKbShare(
    shareOpen ? data?.workspaceSlug : undefined,
    shareOpen ? data?.page.id : undefined,
  );

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
      {/* サイドバーはコースの章一覧と同じ機構で出し入れする（« で隠す / 左端ホバーで
          一時表示 / ⌘\ で切替）。画面ごとに別の作りを持たない — 覚えることを増やさない。 */}
      <SecondaryPanel
        title="ナレッジ"
        peekable
        storageKey="frestyle.panel.note"
        mobileOpen={mobilePanelOpen}
        onMobileClose={closeMobilePanel}
      >
        <KbSidebar workspaceSlug={data?.workspaceSlug ?? navigationWorkspaceSlug} activePageId={pageId} />
      </SecondaryPanel>

      <main className="min-w-0 flex-1 overflow-y-auto">
        {/* モバイルヘッダー: md 以上は SecondaryPanel 自身の一時表示機構（左端ホバー / ☰）が
            効くのでここには出さない。md 未満はこのボタンだけがサイドバーを開く唯一の手段。 */}
        <div className="md:hidden bg-surface-1 border-b border-surface-3 px-4 py-2 flex items-center">
          <button
            onClick={openMobilePanel}
            className="p-1.5 hover:bg-surface-2 rounded transition-colors"
            aria-label="ナレッジ一覧を開く"
          >
            <Bars3Icon className="w-5 h-5 text-[var(--color-text-muted)]" />
          </button>
        </div>
        <div className="mx-auto w-full max-w-3xl px-6 py-10">
          {/* pageId 無し(素の /kb)は resolveEntryPageId が続きを決めている間だけ通る道で、
              ほとんどの場合は決まり次第 /kb/{id} へ移ってしまう。ここに残るのは、
              1 枚もページが見つからなかった(ワークスペースが空)ときだけ。 */}
          {!pageId && entryResolving && <Loading className="py-16" />}
          {!pageId && !entryResolving && (
            <EmptyState
              icon={DocumentTextIcon}
              title="まだページがありません"
              description="左のツリーからページを作成してください。"
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
              <div className="mb-2 flex items-start justify-between gap-3">
              <nav aria-label="ページの場所" className="flex min-w-0 flex-wrap items-center gap-1 text-xs text-[var(--color-text-muted)]">
                <span className="truncate">{data.workspaceName ?? data.workspaceSlug}</span>
                {/* ?? [] はデプロイ順の防御 — 旧バックエンドの応答（ancestors なし）でも落とさない */}
                {(data.ancestors ?? []).map((ancestor) => (
                  <span key={ancestor.id} className="flex min-w-0 items-center gap-1">
                    <span aria-hidden="true">/</span>
                    <Link
                      to={`/kb/${ancestor.id}`}
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
              {/*
                共有は canManage のときだけ出す。権限が無い相手に押せるボタンを出しても、
                返るのは 404 だけで「権限が無い」ことすら伝わらない。
              */}
              {data.canManage && (
                <div className="relative shrink-0">
                  <button
                    type="button"
                    onClick={() => setShareOpen((open) => !open)}
                    aria-expanded={shareOpen}
                    className="rounded border border-surface-3 px-2 py-1 text-xs text-[var(--color-text-secondary)] transition-colors hover:bg-surface-2"
                  >
                    共有
                  </button>
                  {shareOpen && (
                    <div className="absolute right-0 top-full z-20 mt-1">
                      <SharePanel
                        targetTitle={data.page.title}
                        inheritedNote="上の段（ワークスペース・スペース・親ページ）から届いている人はここには出ません。"
                        emptyNote="このページではまだ誰にも権限を足していません。上の段から届いている人は、ここが空でもこのページを見られます。"
                        rows={share.rows}
                        candidates={share.candidates}
                        loading={share.loading}
                        error={share.error}
                        saving={share.saving}
                        onGrant={share.grant}
                        onRevoke={share.revoke}
                        onClose={() => setShareOpen(false)}
                      />
                    </div>
                  )}
                </div>
              )}
              </div>
              {/*
                アイコン → 題名の順（Notion 等と同じ、上に乗るものから読む並び）。
                group はアイコン追加ボタンのホバー表示に使う（KbPageIconButton 側の約束）。
                ページごとに作り直す（別ページへ移った瞬間、打ちかけの下書きを持ち越さない）。
              */}
              <div className="group" key={data.page.id}>
                <KbPageIconButton
                  icon={data.page.icon}
                  canEdit={data.canEdit}
                  onChange={handleChangeIcon}
                />
                <KbPageTitle
                  title={data.page.title}
                  canEdit={data.canEdit}
                  onRename={handleRename}
                  onEnter={() => setBodyFocusSignal((prev) => prev + 1)}
                />
              </div>
              <KbPageMeta lastEditedBy={data.lastEditedBy} lastEditedAt={data.lastEditedAt} />
              <RichTextEditor
                // doc は API から来る任意の JSON。形が違えば空の本文として扱い、画面を落とさない。
                value={isRichDoc(data.doc) ? data.doc : emptyRichDoc()}
                editable={data.canEdit}
                onChange={onDocChange}
                saveStatus={data.canEdit ? saveStatus : 'idle'}
                ariaLabel={`${data.page.title} の本文`}
                extraSlashCommands={data.canEdit ? extraSlashCommands : undefined}
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
