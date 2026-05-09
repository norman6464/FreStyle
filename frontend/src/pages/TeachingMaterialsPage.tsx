import { useEffect } from 'react';
import { useSelector } from 'react-redux';
import {
  AcademicCapIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  Bars3Icon,
  TrashIcon,
  CheckCircleIcon,
  ClockIcon,
} from '@heroicons/react/24/outline';
import SecondaryPanel from '../components/layout/SecondaryPanel';
import EmptyState from '../components/EmptyState';
import ConfirmModal from '../components/ConfirmModal';
import Loading from '../components/Loading';
import NoteMarkdownEditor from '../components/NoteMarkdownEditor';
import { useTeachingMaterials } from '../hooks/useTeachingMaterials';
import { useTeachingMaterialEditor } from '../hooks/useTeachingMaterialEditor';
import { useMobilePanelState } from '../hooks/useMobilePanelState';
import { useToast } from '../hooks/useToast';
import { useState } from 'react';
import type { RootState } from '../store';
import type { TeachingMaterial } from '../types';

/**
 * TeachingMaterialsPage — `/teaching-materials` 教材一覧 + 編集ページ。
 *
 * - company_admin / super_admin: 教材を作成 / 編集 / 削除 / 公開状態切替
 * - trainee: 自社の `isPublished=true` 教材を閲覧のみ
 *
 * 左パネルに教材リスト、 右側に詳細（NoteMarkdownEditor 流用 = Edit/Preview タブ）。
 */
export default function TeachingMaterialsPage() {
  const role = useSelector((s: RootState) => s.auth.role);
  const canManage = role === 'company_admin' || role === 'super_admin';

  const { showToast } = useToast();
  const { isOpen: mobilePanelOpen, open: openMobilePanel, close: closeMobilePanel } = useMobilePanelState();

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
    fetchAll,
  } = useTeachingMaterials();

  const editor = useTeachingMaterialEditor({ selectedId, selected, update });

  const [deleteTargetId, setDeleteTargetId] = useState<number | null>(null);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  const handleCreate = async () => {
    const created = await create({ title: '無題の教材', content: '', isPublished: false });
    if (created) {
      showToast('success', '教材を作成しました');
    } else {
      showToast('error', '教材の作成に失敗しました');
    }
    closeMobilePanel();
  };

  const handleSelect = (id: number) => {
    selectMaterial(id);
    closeMobilePanel();
  };

  const handleConfirmDelete = async () => {
    if (deleteTargetId == null) return;
    await remove(deleteTargetId);
    setDeleteTargetId(null);
    showToast('success', '教材を削除しました');
  };

  return (
    <div className="flex h-full">
      <SecondaryPanel
        title="教材"
        badge={`${materials.length}件`}
        mobileOpen={mobilePanelOpen}
        onMobileClose={closeMobilePanel}
        headerContent={
          <div className="space-y-2">
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
                className="w-full bg-primary-500 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-primary-600 transition-colors flex items-center justify-center gap-2"
              >
                <PlusIcon className="w-4 h-4" />
                新しい教材
              </button>
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
                icon={AcademicCapIcon}
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
                onDelete={canManage ? (id) => setDeleteTargetId(id) : undefined}
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
          <span className="ml-2 text-xs text-[var(--color-text-muted)]">教材一覧</span>
        </div>

        {error && <p className="px-6 py-3 text-sm text-red-500">{error}</p>}

        {selected ? (
          canManage ? (
            <ManagedDetail
              editor={editor}
            />
          ) : (
            <ReadOnlyDetail material={selected} />
          )
        ) : (
          <EmptyState
            icon={AcademicCapIcon}
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
}: {
  material: TeachingMaterial;
  isActive: boolean;
  onSelect: (id: number) => void;
  onDelete?: (id: number) => void;
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
          {material.isPublished
            ? <CheckCircleIcon className="w-3 h-3 text-green-400 flex-shrink-0" />
            : <ClockIcon className="w-3 h-3 text-amber-400 flex-shrink-0" />}
          <p className="text-sm text-[var(--color-text-primary)] truncate font-medium">
            {material.title || '無題の教材'}
          </p>
        </div>
        <p className="text-xs text-[var(--color-text-muted)] mt-0.5 truncate">
          {material.isPublished ? '公開中' : '下書き'} ・ {formatDate(material.updatedAt)}
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

function ReadOnlyDetail({ material }: { material: TeachingMaterial }) {
  // trainee は読み取り専用のため、 NoteMarkdownEditor を Preview 固定状態で渡す代わりに
  // 簡易的な表示モードにする（編集不可なので saveStatus / 入力は不要）。
  return (
    <div className="flex-1 overflow-y-auto p-6 max-w-3xl mx-auto w-full">
      <h1 className="text-2xl font-bold text-[var(--color-text-primary)] mb-2">
        {material.title || '無題の教材'}
      </h1>
      <p className="text-xs text-[var(--color-text-muted)] mb-6">
        最終更新: {formatDate(material.updatedAt)}
      </p>
      <div className="prose prose-invert prose-sm max-w-none">
        {/* trainee は編集できないので NoteMarkdownEditor の onContentChange は no-op で
            渡し、 Preview 表示扱いにする。 ただしカスタマイズが面倒なので、
            最初から <ReactMarkdown> で描画する別経路にする。 */}
        <ReadOnlyMarkdown content={material.content} />
      </div>
    </div>
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
// NoteMarkdownEditor の MarkdownView と同じだが import を分離しないために
// 同じパッケージを直接使う。
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import rehypeHighlight from 'rehype-highlight';
import type { ReactNode } from 'react';

function ReadOnlyMarkdown({ content }: { content: string }) {
  if (!content.trim()) {
    return <p className="text-[var(--color-text-muted)]">この教材にはまだ本文がありません。</p>;
  }
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm, remarkBreaks]}
      rehypePlugins={[rehypeHighlight]}
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
