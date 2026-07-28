import { useEffect, useState, useMemo } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  PlusIcon,
  Bars3Icon,
  ArrowLeftIcon,
} from '@heroicons/react/24/outline';
import { useAppSelector } from '@/shared/lib/store';
import { SecondaryPanel } from '@/widgets/secondary-panel';
import { CourseProgressBar, CourseRepository } from '@/entities/course';
import EmptyState from '@/shared/ui/EmptyState';
import FaviconIcon from '@/shared/ui/icons/FaviconIcon';
import ConfirmModal from '@/shared/ui/ConfirmModal';
import Loading from '@/shared/ui/Loading';
import { DashboardRepository } from '@/entities/user';
import { useMobilePanelState } from '@/shared/lib/hooks/useMobilePanelState';
import { useLocalStorage } from '@/shared/lib/hooks/useLocalStorage';
import { useToast } from '@/shared/lib/hooks/useToast';
import { useTeachingMaterials } from '../model/useTeachingMaterials';
import { useTeachingMaterialEditor } from '../model/useTeachingMaterialEditor';
import { useChapterResume } from '../model/useChapterResume';
import { useNextCourse } from '../model/useNextCourse';
import { useLessonProgress } from '../model/useLessonProgress';
import MaterialListItem from './MaterialListItem';
import MaterialSkeleton from './MaterialSkeleton';
import ManagedDetail from './ManagedDetail';
import ReadOnlyDetail from './ReadOnlyDetail';
import type { Course } from '@/entities/course';

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

  const role = useAppSelector((state) => state.auth.role);
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
    () => materials.filter((material) => progress.completedIds.has(material.id)).length,
    [materials, progress.completedIds],
  );

  // 表示順で「次の章」を求める（読み終えたら順に進める導線用）。
  const nextMaterial = useMemo(() => {
    if (selectedId == null) return null;
    const selectedIndex = materials.findIndex((material) => material.id === selectedId);
    return selectedIndex >= 0 && selectedIndex < materials.length - 1
      ? materials[selectedIndex + 1]
      : null;
  }, [materials, selectedId]);

  const handleToggleComplete = async (materialId: number, done: boolean) => {
    const succeeded = await progress.toggle(materialId, done);
    if (!succeeded) {
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
      .then((fetchedCourse) => setCourse(fetchedCourse))
      .catch(() => setCourseError('コースの取得に失敗しました'))
      .finally(() => setCourseLoading(false));
  }, [courseId]);

  // 次の order_in_course を計算（既存の最大値 + 10）。
  const nextOrder = useMemo(() => {
    if (materials.length === 0) return 100;
    return Math.max(...materials.map((material) => material.orderInCourse)) + 10;
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

  const handleSelect = (materialId: number) => {
    selectMaterial(materialId);
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
        onToggleCollapsed={() => setPanelOpen((isOpen) => !isOpen)}
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
            materials.map((material) => (
              <MaterialListItem
                key={material.id}
                material={material}
                isActive={selectedId === material.id}
                onSelect={handleSelect}
                onDelete={canManage ? (materialId) => setDeleteTargetId(materialId) : undefined}
                completed={progress.completedIds.has(material.id)}
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
