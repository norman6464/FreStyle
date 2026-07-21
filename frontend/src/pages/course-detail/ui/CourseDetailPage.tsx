import { useEffect, useState, useMemo, useRef } from 'react';
import { useSelector } from 'react-redux';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  PlusIcon,
  Bars3Icon,
  TrashIcon,
  ArrowLeftIcon,
  ArrowRightIcon,
  ListBulletIcon,
} from '@heroicons/react/24/outline';
import { CheckCircleIcon as CheckCircleSolidIcon } from '@heroicons/react/24/solid';
import { SecondaryPanel } from '@/widgets/secondary-panel';
import { CourseProgressBar } from '@/entities/course';
import EmptyState from '@/shared/ui/EmptyState';
import FaviconIcon from '@/shared/ui/icons/FaviconIcon';
import ConfirmModal from '@/shared/ui/ConfirmModal';
import Loading from '@/shared/ui/Loading';
import NoteMarkdownEditor from '@/components/NoteMarkdownEditor';
import MarkdownTableOfContents from '@/shared/ui/MarkdownTableOfContents';
import { useTeachingMaterials } from '@/hooks/useTeachingMaterials';
import { useTeachingMaterialEditor } from '@/hooks/useTeachingMaterialEditor';
import { useChapterResume } from '@/hooks/useChapterResume';
import { useNextCourse } from '@/hooks/useNextCourse';
import { useLessonProgress } from '@/hooks/useLessonProgress';
import { useMobilePanelState } from '@/shared/lib/hooks/useMobilePanelState';
import { useLocalStorage } from '@/shared/lib/hooks/useLocalStorage';
import { useToast } from '@/shared/lib/hooks/useToast';
import { CourseRepository } from '@/entities/course';
import { DashboardRepository } from '@/entities/user';
import { ImageUploadRepository } from '@/entities/user';
import type { RootState } from '@/store';
import type { Course, CourseWithProgress, TeachingMaterial } from '@/entities/course';

/**
 * CourseDetailPage — `/courses/:id` 配下の教材一覧 + 編集ページ。
 *
 * - company_admin / super_admin: コース内教材を作成 / 編集 / 削除 / 公開状態切替
 * - trainee: published コース + published 教材のみ閲覧
 *
 * 左パネルに教材リスト、 右側に詳細（NoteMarkdownEditor 流用 = Edit/Preview タブ）。
 */
export default function CourseDetailPage() {
  const { id } = useParams<{ id: string }>();
  const courseId = id ? Number(id) : null;
  const navigate = useNavigate();

  const role = useSelector((s: RootState) => s.auth.role);
  const canManage = role === 'company_admin' || role === 'super_admin';

  const { showToast } = useToast();
  const { isOpen: mobilePanelOpen, open: openMobilePanel, close: closeMobilePanel } = useMobilePanelState();
  // デスクトップの章一覧パネルの開閉。 教材を切り替えても継続するよう localStorage に保持（既定は表示）。
  const [panelOpen, setPanelOpen] = useLocalStorage('course-panel-open', true);

  const [course, setCourse] = useState<Course | null>(null);
  const [courseLoading, setCourseLoading] = useState(true);
  const [courseError, setCourseError] = useState<string | null>(null);

  const {
    materials,
    selectedId,
    selected,
    loading,
    error,
    selectMaterial,
    create,
    update,
    remove,
  } = useTeachingMaterials(courseId);

  const editor = useTeachingMaterialEditor({ selectedId, selected, update });

  // 受講者がコースを開いたら「最後に閲覧した章(無ければ先頭)」を自動表示する(FRESTYLE-99)。
  useChapterResume({ enabled: !canManage, courseId, materials, loading, selectedId, selectMaterial });

  // 最終章の末尾から一覧に戻らず次のコースへ直行できるようにする(FRESTYLE-102)。
  const { nextCourse } = useNextCourse(courseId, !canManage);

  // 章を表示したら閲覧を記録する(受講者のみ・ベストエフォート)。
  // レジュームとダッシュボード「続きから」の基盤データになる。
  useEffect(() => {
    if (canManage || selectedId == null) return;
    DashboardRepository.recordChapterView(selectedId);
  }, [canManage, selectedId]);

  // 進捗トラッキングは学習者（trainee）向け。 教材を管理するロールでは API を叩かない。
  const progress = useLessonProgress(!canManage);
  const completedCount = useMemo(
    () => materials.filter((m) => progress.completedIds.has(m.id)).length,
    [materials, progress.completedIds],
  );

  // 表示順で「次の章」を求める（読み終えたら順に進める導線用）。
  const nextMaterial = useMemo(() => {
    if (selectedId == null) return null;
    const idx = materials.findIndex((m) => m.id === selectedId);
    return idx >= 0 && idx < materials.length - 1 ? materials[idx + 1] : null;
  }, [materials, selectedId]);

  const handleToggleComplete = async (materialId: number, done: boolean) => {
    const ok = await progress.toggle(materialId, done);
    if (!ok) {
      showToast('error', '進捗の更新に失敗しました');
    } else if (done) {
      showToast('success', '完了にしました');
    }
  };

  const [deleteTargetId, setDeleteTargetId] = useState<number | null>(null);

  useEffect(() => {
    if (!courseId) return;
    setCourseLoading(true);
    CourseRepository.get(courseId)
      .then((c) => setCourse(c))
      .catch(() => setCourseError('コースの取得に失敗しました'))
      .finally(() => setCourseLoading(false));
  }, [courseId]);

  // 次の order_in_course を計算（既存の最大値 + 10）。
  const nextOrder = useMemo(() => {
    if (materials.length === 0) return 100;
    return Math.max(...materials.map((m) => m.orderInCourse)) + 10;
  }, [materials]);

  const handleCreate = async () => {
    const created = await create({
      title: '無題の教材',
      content: '',
      orderInCourse: nextOrder,
      isPublished: false,
    });
    if (created) {
      showToast('success', '教材を作成しました');
    } else {
      showToast('error', '教材の作成に失敗しました');
    }
    closeMobilePanel();
  };

  const handleSelect = (mid: number) => {
    selectMaterial(mid);
    closeMobilePanel();
  };

  const handleConfirmDelete = async () => {
    if (deleteTargetId == null) return;
    await remove(deleteTargetId);
    setDeleteTargetId(null);
    showToast('success', '教材を削除しました');
  };

  if (!courseId) {
    return (
      <EmptyState
        icon={FaviconIcon}
        title="コースが指定されていません"
        description="コース一覧から選択してください。"
        action={{ label: 'コース一覧へ', onClick: () => navigate('/courses') }}
      />
    );
  }

  if (courseLoading) {
    return <Loading className="h-full" />;
  }

  if (courseError || !course) {
    return (
      <EmptyState
        icon={FaviconIcon}
        title="コースが見つかりませんでした"
        description={courseError ?? '権限がないか、 コースが削除された可能性があります。'}
        action={{ label: 'コース一覧へ', onClick: () => navigate('/courses') }}
      />
    );
  }

  return (
    <div className="flex h-full">
      {/* 受講者は章一覧を右サイドバーに出すため、デスクトップの左パネルは出さない(FRESTYLE-118)。
          モバイル(md 未満)は右サイドバーが無いので、従来のドロワーを章切替の導線として残す。
          canManage は編集導線(作成/削除)があるため従来の左パネルのまま。 */}
      <div className={canManage ? 'contents' : 'md:hidden'}>
      <SecondaryPanel
        title={course.title || '無題のコース'}
        badge={`${materials.length}件`}
        mobileOpen={mobilePanelOpen}
        onMobileClose={closeMobilePanel}
        collapsible
        collapsed={!panelOpen}
        onToggleCollapsed={() => setPanelOpen((v) => !v)}
        headerContent={
          <div className="space-y-2">
            <Link
              to="/courses"
              className="inline-flex items-center gap-1 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
            >
              <ArrowLeftIcon className="w-3.5 h-3.5" />
              コース一覧
            </Link>
            {canManage && (
              <button
                onClick={handleCreate}
                className="w-full bg-brand-500 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-brand-600 transition-colors flex items-center justify-center gap-2"
              >
                <PlusIcon className="w-4 h-4" />
                新しい教材
              </button>
            )}
            {!canManage && materials.length > 0 && (
              <CourseProgressBar completed={completedCount} total={materials.length} />
            )}
          </div>
        }
      >
        <div className="py-2">
          {loading && materials.length === 0 ? (
            <Loading className="py-8" />
          ) : materials.length === 0 ? (
            <div className="py-12">
              <EmptyState
                icon={FaviconIcon}
                title="教材がありません"
                description={
                  canManage
                    ? '新しい教材を作成しましょう'
                    : '管理者が教材を公開すると、 ここに表示されます'
                }
                action={canManage ? { label: '新しい教材', onClick: handleCreate } : undefined}
              />
            </div>
          ) : (
            materials.map((m) => (
              <MaterialListItem
                key={m.id}
                material={m}
                isActive={selectedId === m.id}
                onSelect={handleSelect}
                onDelete={canManage ? (mid) => setDeleteTargetId(mid) : undefined}
                completed={progress.completedIds.has(m.id)}
              />
            ))
          )}
        </div>
      </SecondaryPanel>
      </div>

      <div className="flex-1 flex flex-col min-w-0">
        {/* モバイルヘッダー */}
        <div className="md:hidden bg-surface-1 border-b border-surface-3 px-4 py-2 flex items-center">
          <button
            onClick={openMobilePanel}
            className="p-1.5 hover:bg-surface-2 rounded transition-colors"
            aria-label="教材一覧を開く"
          >
            <Bars3Icon className="w-5 h-5 text-[var(--color-text-muted)]" />
          </button>
          <span className="ml-2 text-xs text-[var(--color-text-muted)]">{course.title}</span>
        </div>

        {error && <p className="px-6 py-3 text-sm text-red-500">{error}</p>}

        {selectedId != null && !selected && !error ? (
          // 章を選択した瞬間から本文取得が終わるまでローディングを出す。
          // detailLoading は effect 後に立つため、それを待つと一瞬「未選択」表示が
          // ちらつく。 selectedId があるのに本文が無い = 取得中、として即座に出す。
          // 読み手にはレイアウトを保つスケルトン、編集者にはスピナーを表示する。
          canManage ? <Loading className="h-full" /> : <MaterialSkeleton />
        ) : selected ? (
          canManage ? (
            <ManagedDetail editor={editor} />
          ) : (
            <ReadOnlyDetail
              material={selected}
              completed={progress.completedIds.has(selected.id)}
              onToggleComplete={(done) => handleToggleComplete(selected.id, done)}
              nextMaterial={nextMaterial}
              onGoNext={nextMaterial ? () => selectMaterial(nextMaterial.id) : undefined}
              nextCourse={nextCourse}
              onGoNextCourse={nextCourse ? () => navigate(`/courses/${nextCourse.id}`) : undefined}
              course={course}
              materials={materials}
              completedIds={progress.completedIds}
              onSelectMaterial={selectMaterial}
              completedCount={completedCount}
            />
          )
        ) : (
          <EmptyState
            icon={FaviconIcon}
            title="教材を選択してください"
            description={
              canManage
                ? '左のリストから教材を選択するか、 新しい教材を作成しましょう。'
                : '章の一覧から教材を選択してください。'
            }
            action={canManage ? { label: '新しい教材', onClick: handleCreate } : undefined}
          />
        )}
      </div>

      <ConfirmModal
        isOpen={deleteTargetId !== null}
        message="この教材を削除しますか？ 削除後は trainee からも見えなくなります。"
        onConfirm={handleConfirmDelete}
        onCancel={() => setDeleteTargetId(null)}
      />
    </div>
  );
}

function MaterialListItem({
  material,
  isActive,
  onSelect,
  onDelete,
  completed = false,
}: {
  material: TeachingMaterial;
  isActive: boolean;
  onSelect: (id: number) => void;
  onDelete?: (id: number) => void;
  completed?: boolean;
}) {
  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => onSelect(material.id)}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onSelect(material.id);
        }
      }}
      className={`group flex items-start gap-2.5 pl-4 pr-2 py-2.5 cursor-pointer transition-colors border-l-2 ${
        isActive
          ? 'border-brand-500 bg-[var(--color-nav-active)]'
          : 'border-transparent hover:bg-[var(--color-nav-hover)]'
      }`}
    >
      <p className={`flex-1 min-w-0 text-[13px] leading-snug line-clamp-2 ${
        isActive
          ? 'font-medium text-[var(--color-text-primary)]'
          : 'text-[var(--color-text-secondary)]'
      }`}>
        {material.title || '無題の教材'}
      </p>

      {onDelete && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete(material.id);
          }}
          className="opacity-0 group-hover:opacity-100 p-1 hover:bg-red-900/30 rounded transition-opacity flex-shrink-0"
          aria-label="教材を削除"
        >
          <TrashIcon className="w-3.5 h-3.5 text-[var(--color-text-muted)] hover:text-red-400" />
        </button>
      )}
    </div>
  );
}

function ManagedDetail({
  editor,
}: {
  editor: ReturnType<typeof useTeachingMaterialEditor>;
}) {
  return (
    <div className="flex-1 flex flex-col min-h-0">
      <div className="px-6 pt-4 pb-2 flex items-center justify-end gap-2 border-b border-surface-3">
        <label className="flex items-center gap-2 text-xs text-[var(--color-text-secondary)] cursor-pointer">
          <input
            type="checkbox"
            checked={editor.editIsPublished}
            onChange={(e) => editor.handleIsPublishedChange(e.target.checked)}
            className="rounded border-surface-3"
          />
          trainee に公開
        </label>
      </div>
      <div className="flex-1 min-h-0">
        <NoteMarkdownEditor
          title={editor.editTitle}
          content={editor.editContent}
          saveStatus={editor.saveStatus}
          onTitleChange={editor.handleTitleChange}
          onContentChange={editor.handleContentChange}
          onImageUpload={(file) => ImageUploadRepository.upload(file)}
        />
      </div>
    </div>
  );
}

function ReadOnlyDetail({
  material,
  completed,
  onToggleComplete,
  nextMaterial,
  onGoNext,
  nextCourse,
  onGoNextCourse,
  course,
  materials,
  completedIds,
  onSelectMaterial,
  completedCount,
}: {
  material: TeachingMaterial;
  completed: boolean;
  onToggleComplete: (done: boolean) => void;
  nextMaterial?: TeachingMaterial | null;
  onGoNext?: () => void;
  nextCourse?: CourseWithProgress | null;
  onGoNextCourse?: () => void;
  course: Course;
  materials: TeachingMaterial[];
  completedIds: Set<number>;
  onSelectMaterial: (id: number) => void;
  completedCount: number;
}) {
  // 目次の表示状態は localStorage に保持し、 教材を切り替えても選択が続くようにする（既定は表示）。
  // 横幅が狭いときに本文幅を稼げるよう trainee が出し入れできる。
  const [tocOpen, setTocOpen] = useLocalStorage('course-toc-open', true);

  // 章を切り替えたら本文の先頭までスクロールを戻す（末尾の「次の章へ」から進んでも頭から読める）。
  // スクロールは AppShell のドキュメントスクロールコンテナ(FRESTYLE-122)側で起きるため、
  // そちらへ遡って scrollTo する（テスト等で見つからない場合は自要素へフォールバック）。
  const scrollRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const scroller = scrollRef.current?.closest('[data-app-scroll]') ?? scrollRef.current;
    scroller?.scrollTo({ top: 0 });
  }, [material.id]);

  // 本文先頭の h1(= タイトル)は、下のヘッダーで material.title を大きく出すため取り除く。
  // 残すとカードの外(ヘッダー)とカードの中(本文)でタイトルが二重に見える(FRESTYLE-131)。
  const bodyContent = useMemo(() => stripLeadingTitle(material.content), [material.content]);

  return (
    // 背景は読み物用の灰青(--color-reading-surface)、本文は白カード。背景と内容のコントラストで
    // 読み物として視線が本文に集まるようにする(FRESTYLE-118)。body は白に戻したが、教材閲覧だけ
    // 灰青を維持する(FRESTYLE-147)。内部スクロールは持たない(FRESTYLE-122 でページ全体スクロール)。
    <div ref={scrollRef} className="flex-1 bg-[var(--color-reading-surface)]">
      {/* 読み物ページなので外側の余白は広め(FRESTYLE-115)。中央寄せなので左右は自然に余白になる。 */}
      <div
        className={`mx-auto w-full max-w-6xl px-6 sm:px-10 py-8 sm:py-10 grid grid-cols-1 gap-8 ${
          tocOpen ? 'lg:grid-cols-[minmax(0,1fr)_280px]' : ''
        }`}
      >
        {/* 記事サイト風ヘッダー(Zenn 風): タイトル + メタをグリッド直下の子にして col-span-full で
            行全体に広げる。 本文カード(左カラム)と右サイドバー(目次/章一覧)は次の行に並ぶので、
            右の目次カードが本文カードと同じ高さから始まる(FRESTYLE-150 / 131)。
            行間は grid の gap-8 が担うので header に mb は付けない。 */}
        <header className="col-span-full text-center">
          <h1 className="text-3xl sm:text-4xl font-bold text-[var(--color-text-primary)] leading-snug">
            {material.title || '無題の教材'}
          </h1>
          {/* メタ(最終更新 / 目次トグル / 完了トグル)。 sticky にはしない(FRESTYLE-119)。
              スクロール途中の完了操作は本文末尾の大きい完了ボタン(FRESTYLE-100)で行える。 */}
          <div className="mt-4 flex flex-wrap items-center justify-center gap-3">
            <p className="text-xs text-[var(--color-text-muted)]">
              最終更新: {formatDate(material.updatedAt)}
            </p>
            {/* 目次は lg 以上でのみ表示されるため、 トグルも lg 未満では隠す。 */}
            <button
              type="button"
              onClick={() => setTocOpen((v) => !v)}
              aria-pressed={tocOpen}
              title={tocOpen ? '目次を隠す' : '目次を表示'}
              className={`hidden lg:inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-medium border transition-colors ${
                tocOpen
                  ? 'border-taupe-500 text-taupe-400'
                  : 'border-surface-3 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]'
              }`}
            >
              <ListBulletIcon className="w-4 h-4" />
              目次
            </button>
            <CompleteToggleButton completed={completed} onToggle={onToggleComplete} />
          </div>
        </header>

        {/* 本文カラム。 サイドバーを隠したときは本文が全幅に伸びて読みにくいため、 読みやすい幅(860px)に
            収めて中央寄せする。 サイドバー表示時は 1fr カラムが既に同程度の幅になる。 */}
        <div className={`min-w-0 ${!tocOpen ? 'mx-auto w-full max-w-[860px]' : ''}`}>
          <article className="bg-white border border-surface-3 rounded-xl shadow-sm px-6 sm:px-10 py-8 sm:py-10">
            <div className="prose prose-sm max-w-none course-prose">
              <ReadOnlyMarkdown content={bodyContent} />
            </div>

            {/* 末尾に「完了にする」と「次の章へ」を並べ、 読み終えた位置から次へ進めるようにする。
                最終章では代わりに「次のコースへ」を出し、 一覧に戻らず次のコースへ直行できるようにする
                (FRESTYLE-102。 遷移先はレジュームにより 1 章目が自動表示される)。 */}
            <div className="mt-10 pt-6 border-t border-surface-3 flex flex-col sm:flex-row items-center justify-center gap-3">
              <CompleteToggleButton completed={completed} onToggle={onToggleComplete} large />
              {nextMaterial && onGoNext ? (
                <button
                  type="button"
                  onClick={onGoNext}
                  title={`次の章へ: ${nextMaterial.title || '無題の教材'}`}
                  className="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-lg text-sm font-medium bg-surface-2 border border-surface-3 text-[var(--color-text-primary)] hover:bg-surface-3 transition-colors max-w-full"
                >
                  <span className="truncate">次の章へ: {nextMaterial.title || '無題の教材'}</span>
                  <ArrowRightIcon className="w-4 h-4 flex-shrink-0" />
                </button>
              ) : nextCourse && onGoNextCourse ? (
                <button
                  type="button"
                  onClick={onGoNextCourse}
                  title={`次のコースへ: ${nextCourse.title || '無題のコース'}`}
                  className="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-lg text-sm font-medium bg-surface-2 border border-surface-3 text-[var(--color-text-primary)] hover:bg-surface-3 transition-colors max-w-full"
                >
                  <span className="truncate">次のコースへ: {nextCourse.title || '無題のコース'}</span>
                  <ArrowRightIcon className="w-4 h-4 flex-shrink-0" />
                </button>
              ) : null}
            </div>
          </article>
        </div>

        {tocOpen && (
          <aside className="hidden lg:block">
            {/* サイドバー全体をビューポート高さに収める(FRESTYLE-144)。目次カードと章一覧カードは
                内容ぶんの高さを取りつつ、合計が収まらないときは flex で縮み各カードが内部スクロールする。
                これで章が長くてもサイドバーが間延びせず、各カードに独立したスクロールバーが出る。 */}
            <div className="sticky top-6 flex max-h-[calc(100vh-3rem)] flex-col gap-4">
              <div className="flex min-h-0 flex-col rounded-xl border border-surface-3 bg-white p-4 shadow-sm">
                {/* タイトルはヘッダーで大きく出すため、 目次からも先頭 h1 を除いた本文を渡す。 */}
                <MarkdownTableOfContents content={bodyContent} />
              </div>
              {/* 章一覧は左パネルから右サイドバー(目次の下)へ移動(FRESTYLE-118)。 */}
              <div className="flex min-h-0 flex-col rounded-xl border border-surface-3 bg-white p-4 shadow-sm">
                <ChapterNav
                  course={course}
                  materials={materials}
                  selectedId={material.id}
                  completedIds={completedIds}
                  completedCount={completedCount}
                  onSelect={onSelectMaterial}
                />
              </div>
            </div>
          </aside>
        )}
      </div>
    </div>
  );
}

/**
 * ChapterNav — 閲覧ビューの右サイドバーに出す章一覧(FRESTYLE-118)。
 * コース一覧への戻り + コース名 + 進捗バー + 章リスト(完了はチェック・現在章はハイライト)。
 */
function ChapterNav({
  course,
  materials,
  selectedId,
  completedIds,
  completedCount,
  onSelect,
}: {
  course: Course;
  materials: TeachingMaterial[];
  selectedId: number;
  completedIds: Set<number>;
  completedCount: number;
  onSelect: (id: number) => void;
}) {
  return (
    // 親カードが高さを制限したとき、見出し(コース名・進捗)は固定して章リストだけを内側でスクロールさせる(FRESTYLE-144)。
    <div className="flex min-h-0 flex-col">
      <div className="flex-shrink-0">
        <Link
          to="/courses"
          className="inline-flex items-center gap-1 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
        >
          <ArrowLeftIcon className="w-3.5 h-3.5" />
          コース一覧
        </Link>
        <p className="mt-2 text-sm font-semibold text-[var(--color-text-primary)]">
          {course.title || '無題のコース'}
          <span className="ml-2 text-xs font-normal text-[var(--color-text-muted)]">
            {materials.length} 章
          </span>
        </p>
        {materials.length > 0 && (
          <div className="mt-2">
            <CourseProgressBar completed={completedCount} total={materials.length} />
          </div>
        )}
      </div>
      <nav aria-label="章一覧" className="mt-3 min-h-0 flex-1 space-y-0.5 overflow-y-auto pr-1">
        {materials.map((m, i) => {
          const active = m.id === selectedId;
          return (
            <button
              key={m.id}
              type="button"
              onClick={() => onSelect(m.id)}
              aria-current={active ? 'page' : undefined}
              className={`w-full flex items-start gap-2 rounded-md px-2 py-1.5 text-left text-[13px] leading-snug transition-colors ${
                active
                  ? 'bg-brand-500/10 font-medium text-[var(--color-text-primary)]'
                  : 'text-[var(--color-text-secondary)] hover:bg-surface-2'
              }`}
            >
              {completedIds.has(m.id) ? (
                <CheckCircleSolidIcon className="w-4 h-4 mt-0.5 text-emerald-500 flex-shrink-0" aria-label="完了" />
              ) : (
                <span className="w-4 h-4 mt-0.5 flex items-center justify-center text-[10px] rounded-full border border-surface-3 text-[var(--color-text-muted)] flex-shrink-0">
                  {i + 1}
                </span>
              )}
              <span className="line-clamp-2">{m.title || '無題の教材'}</span>
            </button>
          );
        })}
      </nav>
    </div>
  );
}

/**
 * MaterialSkeleton は章本文の取得中に出す骨組み。
 *
 * バーのスピナーではなく記事レイアウト（タイトル / メタ / 本文行 / 図ブロック）を
 * 模した pulse プレースホルダにすることで、 章切り替え時のちらつきを抑え体感速度を上げる。
 */
function MaterialSkeleton() {
  return (
    // 実表示(灰青背景 + 白カード)と同じ配色にして、取得完了時の切り替わりで背景が変わらないようにする。
    <div className="flex-1 bg-[var(--color-reading-surface)]">
      <div
        className="mx-auto w-full max-w-[860px] px-6 sm:px-10 py-8 sm:py-10 animate-pulse"
        aria-hidden="true"
      >
        {/* タイトル + メタはカードの外・上(FRESTYLE-131 の実表示に合わせて跳ねを防ぐ)。 */}
        <div className="mb-6 flex flex-col items-center gap-3">
          <div className="h-8 w-3/4 rounded bg-surface-3" />
          <div className="h-3 w-40 rounded bg-surface-2" />
        </div>
        <div className="bg-white border border-surface-3 rounded-xl shadow-sm px-6 sm:px-10 py-8 sm:py-10">
          <div className="space-y-3">
            <div className="h-4 w-full rounded bg-surface-2" />
            <div className="h-4 w-11/12 rounded bg-surface-2" />
            <div className="h-4 w-4/5 rounded bg-surface-2" />
          </div>
          <div className="mt-6 h-40 w-full rounded-lg bg-surface-2" />
          <div className="mt-6 space-y-3">
            <div className="h-4 w-full rounded bg-surface-2" />
            <div className="h-4 w-10/12 rounded bg-surface-2" />
            <div className="h-4 w-3/4 rounded bg-surface-2" />
          </div>
        </div>
      </div>
      <span className="sr-only" role="status">
        読み込み中
      </span>
    </div>
  );
}

/** CompleteToggleButton は教材の完了 / 未完了を切り替えるトグルボタン。 */
function CompleteToggleButton({
  completed,
  onToggle,
  large = false,
}: {
  completed: boolean;
  onToggle: (done: boolean) => void;
  large?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={() => onToggle(!completed)}
      aria-pressed={completed}
      className={`inline-flex items-center justify-center gap-1.5 rounded-lg font-medium transition-colors ${
        large ? 'px-5 pt-[9px] pb-[11px] text-sm' : 'px-3 pt-[5px] pb-[7px] text-sm'
      } ${
        completed
          ? 'bg-green-500/15 text-green-500 hover:bg-green-500/25'
          : 'bg-brand-500 text-white hover:bg-brand-600'
      }`}
    >
      <CheckCircleSolidIcon className="w-4 h-4" />
      {completed ? '完了済み' : '完了にする'}
    </button>
  );
}

function formatDate(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  return `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())}`;
}

function pad(n: number): string {
  return String(n).padStart(2, '0');
}

/**
 * stripLeadingTitle — 本文の先頭にある h1(章タイトル)を取り除く(FRESTYLE-131)。
 *
 * 教材本文は規約により `# タイトル`(= material.title と同じ)で始まる。閲覧ビューでは
 * タイトルをカード外のヘッダーで大きく表示するため、本文側の先頭 h1 を消して二重表示を防ぐ。
 * 「先頭の空行を飛ばした最初の非空行が h1 のときだけ」除去するので、コードブロック内の
 * `# コメント` や 2 個目以降の見出しには一切触れない(先頭の 1 個だけが対象)。
 */
function stripLeadingTitle(content: string): string {
  if (!content) return content;
  const lines = content.split('\n');
  let i = 0;
  while (i < lines.length && lines[i].trim() === '') i++;
  if (i >= lines.length || !/^#\s+\S/.test(lines[i])) return content;
  lines.splice(0, i + 1);
  while (lines.length && lines[0].trim() === '') lines.shift();
  return lines.join('\n');
}

// trainee の閲覧用に Markdown を render するだけのコンポーネント。
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import rehypeHighlight from 'rehype-highlight';
import rehypeSlug from 'rehype-slug';
import type { ReactNode } from 'react';

function ReadOnlyMarkdown({ content }: { content: string }) {
  if (!content.trim()) {
    return <p className="text-[var(--color-text-muted)]">この教材にはまだ本文がありません。</p>;
  }
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm, remarkBreaks]}
      rehypePlugins={[rehypeSlug, rehypeHighlight]}
      components={{
        a: ({ href, children }) => (
          <a
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            className="text-brand-400 underline-offset-2 hover:underline"
          >
            {children as ReactNode}
          </a>
        ),
        // 図（draw.io から書き出した PNG/SVG 等）を本文に埋め込めるようにする。
        // 中央寄せ + 枠 + 白背景（透過 SVG が見えるよう）。リンクにはしない
        // （クリックで画像 URL が別タブに開き学習が中断される、というユーザー要望で FRESTYLE-125 にて除去）。
        img: ({ src, alt }) => {
          const url = typeof src === 'string' ? src : undefined;
          return (
            <figure className="my-5">
              <img
                src={url}
                alt={alt ?? ''}
                loading="lazy"
                className="mx-auto max-w-[90%] h-auto rounded-lg border border-surface-3 bg-white"
              />
              {alt && (
                <figcaption className="mt-2 text-center text-xs text-[var(--color-text-muted)]">
                  {alt}
                </figcaption>
              )}
            </figure>
          );
        },
        code: ({ className, children, ...props }) => {
          const isBlock = className?.includes('language-');
          if (isBlock) {
            return (
              <code className={className} {...props}>
                {children as ReactNode}
              </code>
            );
          }
          return (
            <code className="px-1 py-0.5 rounded bg-[var(--color-surface-3)] text-[0.85em]">
              {children as ReactNode}
            </code>
          );
        },
        table: ({ children }) => (
          <div className="overflow-x-auto">
            <table className="text-sm border-collapse">{children as ReactNode}</table>
          </div>
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  );
}
