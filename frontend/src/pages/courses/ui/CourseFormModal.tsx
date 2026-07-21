import { useState } from 'react';
import { COURSE_CATEGORIES, COURSE_LANGUAGES } from '@/entities/course';
import type { Course } from '@/entities/course';

export interface CourseFormPayload {
  title: string;
  description: string;
  category: string;
  language: string;
  sortOrder: number;
  isPublished: boolean;
}

/**
 * CourseFormModal — コースの作成 / 編集フォーム（管理者用）。
 *
 * 新規作成時は `defaultCategory` を初期選択にする（領域一覧ページから作ると、
 * その領域が初期値になる。FRESTYLE-177）。
 */
export default function CourseFormModal({
  initial,
  defaultCategory = '',
  onClose,
  onSubmit,
}: {
  initial: Course | null;
  defaultCategory?: string;
  onClose: () => void;
  onSubmit: (payload: CourseFormPayload) => Promise<void>;
}) {
  const [title, setTitle] = useState(initial?.title ?? '');
  const [description, setDescription] = useState(initial?.description ?? '');
  const [category, setCategory] = useState(initial?.category ?? defaultCategory);
  const [language, setLanguage] = useState(initial?.language ?? '');
  const [sortOrder, setSortOrder] = useState(initial?.sortOrder ?? 100);
  const [isPublished, setIsPublished] = useState(initial?.isPublished ?? false);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    setSubmitting(true);
    try {
      await onSubmit({
        title: title.trim(),
        description: description.trim(),
        category,
        language,
        sortOrder,
        isPublished,
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <div
        className="bg-surface-1 rounded-lg shadow-xl border border-surface-3 max-w-lg w-full p-6 space-y-4"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-bold text-[var(--color-text-primary)]">
          {initial ? 'コースを編集' : '新しいコース'}
        </h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <label className="block">
            <span className="block text-sm text-[var(--color-text-secondary)] mb-1">タイトル *</span>
            <input
              required
              maxLength={200}
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full px-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm focus:outline-none focus:border-brand-400"
            />
          </label>
          <label className="block">
            <span className="block text-sm text-[var(--color-text-secondary)] mb-1">説明</span>
            <textarea
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="このコースで学べる内容の概要"
              className="w-full px-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm focus:outline-none focus:border-brand-400 resize-y"
            />
          </label>
          <label className="block">
            <span className="block text-sm text-[var(--color-text-secondary)] mb-1">カテゴリ（学習領域）</span>
            <select
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              className="w-full px-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm focus:outline-none focus:border-brand-400"
            >
              <option value="">未分類</option>
              {COURSE_CATEGORIES.map((c) => (
                <option key={c.key} value={c.key}>
                  {c.label}
                </option>
              ))}
            </select>
            <span className="block text-xs text-[var(--color-text-muted)] mt-1">
              一覧カードの色分け・領域選択に使われます（色＝学習領域）
            </span>
          </label>
          <label className="block">
            <span className="block text-sm text-[var(--color-text-secondary)] mb-1">主に扱う言語・技術</span>
            <select
              value={language}
              onChange={(e) => setLanguage(e.target.value)}
              className="w-full px-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm focus:outline-none focus:border-brand-400"
            >
              <option value="">未設定（言語が主題でないコース）</option>
              {COURSE_LANGUAGES.map((l) => (
                <option key={l.key} value={l.key}>
                  {l.label}
                </option>
              ))}
            </select>
            <span className="block text-xs text-[var(--color-text-muted)] mt-1">
              一覧カードに言語バッジとして表示されます
            </span>
          </label>
          <label className="block">
            <span className="block text-sm text-[var(--color-text-secondary)] mb-1">並び順 (昇順)</span>
            <input
              type="number"
              value={sortOrder}
              onChange={(e) => setSortOrder(Number(e.target.value) || 0)}
              className="w-32 px-3 py-1.5 bg-surface-2 border border-surface-3 rounded-lg text-sm focus:outline-none focus:border-brand-400"
            />
            <span className="block text-xs text-[var(--color-text-muted)] mt-1">
              小さい値が上に来ます。 既定: 100
            </span>
          </label>
          <label className="flex items-center gap-2 text-sm text-[var(--color-text-secondary)]">
            <input
              type="checkbox"
              checked={isPublished}
              onChange={(e) => setIsPublished(e.target.checked)}
            />
            trainee に公開
          </label>
          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="px-3 py-1.5 text-sm text-[var(--color-text-muted)] hover:bg-surface-2 rounded"
            >
              キャンセル
            </button>
            <button
              type="submit"
              disabled={!title.trim() || submitting}
              className="bg-brand-500 text-white px-4 py-1.5 rounded text-sm font-medium hover:bg-brand-600 transition-colors disabled:opacity-50"
            >
              {submitting ? '保存中...' : initial ? '更新' : '作成'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
