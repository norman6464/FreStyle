import { useEffect, useState, useMemo } from 'react';
import { useSelector } from 'react-redux';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  PlusIcon,
  MagnifyingGlassIcon,
  Bars3Icon,
  TrashIcon,
  CheckCircleIcon,
  ClockIcon,
  ArrowLeftIcon,
  ListBulletIcon,
} from '@heroicons/react/24/outline';
import { CheckCircleIcon as CheckCircleSolidIcon } from '@heroicons/react/24/solid';
import SecondaryPanel from '../components/layout/SecondaryPanel';
import EmptyState from '../components/EmptyState';
import FaviconIcon from '../components/icons/FaviconIcon';
import ConfirmModal from '../components/ConfirmModal';
import Loading from '../components/Loading';
import NoteMarkdownEditor from '../components/NoteMarkdownEditor';
import MarkdownTableOfContents from '../components/MarkdownTableOfContents';
import { useTeachingMaterials } from '../hooks/useTeachingMaterials';
import { useTeachingMaterialEditor } from '../hooks/useTeachingMaterialEditor';
import { useLessonProgress } from '../hooks/useLessonProgress';
import { useMobilePanelState } from '../hooks/useMobilePanelState';
import { useLocalStorage } from '../hooks/useLocalStorage';
import { useToast } from '../hooks/useToast';
import CourseRepository from '../repositories/CourseRepository';
import type { RootState } from '../store';
import type { Course, TeachingMaterial } from '../types';

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

  const [course, setCourse] = useState<Course | null>(null);
  const [courseLoading, setCourseLoading] = useState(true);
  const [courseError, setCourseError] = useState<string | null>(null);

  const {
    materials,
    filtered,
    selectedId,
    selected,
    loading,
    error,
    searchQuery,
    setSearchQuery,
    selectMaterial,
    create,
    update,
    remove,
  } = useTeachingMaterials(courseId);

  const editor = useTeachingMaterialEditor({ selectedId, selected, update });

  // 進捗トラッキングは学習者（trainee）向け。 教材を管理するロールでは API を叩かない。
  const progress = useLessonProgress(!canManage);
  const completedCount = useMemo(
    () => materials.filter((m) => progress.completedIds.has(m.id)).length,
    [materials, progress.completedIds],
  );

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
      <SecondaryPanel
        title={course.title || '無題のコース'}
        badge={`${materials.length}件`}
        mobileOpen={mobilePanelOpen}
        onMobileClose={closeMobilePanel}
        headerContent={
          <div className="space-y-2">
            <Link
              to="/courses"
              className="inline-flex items-center gap-1 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
            >
              <ArrowLeftIcon className="w-3.5 h-3.5" />
              コース一覧
            </Link>
            <div className="relative">
              <MagnifyingGlassIcon className="w-4 h-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" />
              <input
                type="text"
                placeholder="教材を検索..."
                aria-label="教材を検索"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-8 pr-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:outline-none focus:border-primary-500 transition-colors"
              />
            </div>
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
        <div className="p-2 space-y-0.5">
          {loading && materials.length === 0 ? (
            <Loading className="py-8" />
          ) : filtered.length === 0 ? (
            <div className="py-12">
              <EmptyState
                icon={FaviconIcon}
                title={searchQuery ? '該当する教材がありません' : '教材がありません'}
                description={
                  searchQuery
                    ? '検索条件を変更してみてください'
                    : canManage
                      ? '新しい教材を作成しましょう'
                      : '管理者が教材を公開すると、 ここに表示されます'
                }
                action={canManage && !searchQuery ? { label: '新しい教材', onClick: handleCreate } : undefined}
              />
            </div>
          ) : (
            filtered.map((m) => (
              <MaterialListItem
                key={m.id}
                material={m}
                isActive={selectedId === m.id}
                onSelect={handleSelect}
                onDelete={canManage ? (mid) => setDeleteTargetId(mid) : undefined}
                showCompletion={!canManage}
                completed={progress.completedIds.has(m.id)}
              />
            ))
          )}
        </div>
      </SecondaryPanel>

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

        {selected ? (
          canManage ? (
            <ManagedDetail editor={editor} />
          ) : (
            <ReadOnlyDetail
              material={selected}
              completed={progress.completedIds.has(selected.id)}
              onToggleComplete={(done) => handleToggleComplete(selected.id, done)}
            />
          )
        ) : (
          <EmptyState
            icon={FaviconIcon}
            title="教材を選択してください"
            description={
              canManage
                ? '左のリストから教材を選択するか、 新しい教材を作成しましょう。'
                : '左のリストから教材を選択してください。'
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
  showCompletion = false,
  completed = false,
}: {
  material: TeachingMaterial;
  isActive: boolean;
  onSelect: (id: number) => void;
  onDelete?: (id: number) => void;
  // showCompletion=true（trainee）のときは公開状態の代わりに完了状態を表示する。
  showCompletion?: boolean;
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
      className={`group flex items-start gap-2 px-3 py-2 rounded-md cursor-pointer transition-colors ${
        isActive
          ? 'bg-[var(--color-nav-active)]'
          : 'hover:bg-[var(--color-nav-hover)]'
      }`}
    >
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5">
          {showCompletion ? (
            completed ? (
              <CheckCircleSolidIcon className="w-4 h-4 text-green-500 flex-shrink-0" />
            ) : (
              <span className="w-3.5 h-3.5 rounded-full border-2 border-[var(--color-text-muted)] flex-shrink-0" />
            )
          ) : material.isPublished ? (
            <CheckCircleIcon className="w-3 h-3 text-green-400 flex-shrink-0" />
          ) : (
            <ClockIcon className="w-3 h-3 text-amber-400 flex-shrink-0" />
          )}
          <p className="text-sm text-[var(--color-text-primary)] truncate font-medium">
            {material.title || '無題の教材'}
          </p>
        </div>
        <p className="text-xs text-[var(--color-text-muted)] mt-0.5 truncate">
          {showCompletion
            ? completed
              ? '完了済み'
              : '未完了'
            : material.isPublished
              ? '公開中'
              : '下書き'}{' '}
          ・ {formatDate(material.updatedAt)}
        </p>
      </div>
      {onDelete && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete(material.id);
          }}
          className="opacity-0 group-hover:opacity-100 p-1 hover:bg-red-900/30 rounded transition-opacity"
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
        />
      </div>
    </div>
  );
}

function ReadOnlyDetail({
  material,
  completed,
  onToggleComplete,
}: {
  material: TeachingMaterial;
  completed: boolean;
  onToggleComplete: (done: boolean) => void;
}) {
  // 目次の表示状態は localStorage に保持し、 教材を切り替えても選択が続くようにする（既定は表示）。
  // 横幅が狭いときに本文幅を稼げるよう trainee が出し入れできる。
  const [tocOpen, setTocOpen] = useLocalStorage('course-toc-open', true);

  return (
    <div className="flex-1 overflow-y-auto">
      <div
        className={`mx-auto w-full max-w-6xl px-6 py-6 grid grid-cols-1 gap-8 ${
          tocOpen ? 'lg:grid-cols-[minmax(0,1fr)_240px]' : ''
        }`}
      >
        {/* 目次を隠したときは本文が全幅に伸びて読みにくいため、 読みやすい幅(~896px)に
            収めて中央寄せする。 目次表示時は 1fr カラムが既に同程度の幅になる。 */}
        <article className={`min-w-0 ${!tocOpen ? 'mx-auto w-full max-w-4xl' : ''}`}>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)] mb-2">
            {material.title || '無題の教材'}
          </h1>
          <div className="flex items-center justify-between gap-3 mb-6">
            <p className="text-xs text-[var(--color-text-muted)]">
              最終更新: {formatDate(material.updatedAt)}
            </p>
            <div className="flex items-center gap-2">
              {/* 目次は lg 以上でのみ表示されるため、 トグルも lg 未満では隠す。 */}
              <button
                type="button"
                onClick={() => setTocOpen((v) => !v)}
                aria-pressed={tocOpen}
                title={tocOpen ? '目次を隠す' : '目次を表示'}
                className={`hidden lg:inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-medium border transition-colors ${
                  tocOpen
                    ? 'border-primary-500 text-primary-400'
                    : 'border-surface-3 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]'
                }`}
              >
                <ListBulletIcon className="w-4 h-4" />
                目次
              </button>
              <CompleteToggleButton completed={completed} onToggle={onToggleComplete} />
            </div>
          </div>
          <div className="prose prose-sm max-w-none">
            <ReadOnlyMarkdown content={material.content} />
          </div>

          {/* 末尾にも完了ボタンを置き、 読み終えた位置から戻らず完了にできるようにする。 */}
          <div className="mt-10 pt-6 border-t border-surface-3 flex justify-center">
            <CompleteToggleButton completed={completed} onToggle={onToggleComplete} large />
          </div>
        </article>

        {tocOpen && (
          <aside className="hidden lg:block">
            <div className="sticky top-6">
              <MarkdownTableOfContents content={material.content} />
            </div>
          </aside>
        )}
      </div>
    </div>
  );
}

/** CourseProgressBar はコース内の完了割合を示す進捗バー。 trainee の左パネルに表示する。 */
function CourseProgressBar({ completed, total }: { completed: number; total: number }) {
  const pct = total > 0 ? Math.round((completed / total) * 100) : 0;
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-xs text-[var(--color-text-muted)]">
        <span>学習の進捗</span>
        <span>
          {completed}/{total}（{pct}%）
        </span>
      </div>
      <div className="h-1.5 w-full rounded-full bg-surface-3 overflow-hidden">
        <div
          className="h-full rounded-full bg-green-500 transition-all"
          style={{ width: `${pct}%` }}
          role="progressbar"
          aria-label="学習の進捗"
          aria-valuenow={pct}
          aria-valuemin={0}
          aria-valuemax={100}
        />
      </div>
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
        large ? 'px-5 py-2.5 text-sm' : 'px-3 py-1.5 text-sm'
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
            className="text-primary-400 underline-offset-2 hover:underline"
          >
            {children as ReactNode}
          </a>
        ),
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
