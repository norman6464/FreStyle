import { useEffect, useMemo, useRef, useState } from 'react';
import { ArrowRightIcon, ListBulletIcon } from '@heroicons/react/24/outline';
import { useLocalStorage } from '@/shared/lib/hooks/useLocalStorage';
import MarkdownTableOfContents from '@/shared/ui/MarkdownTableOfContents';
import { RichTextEditor } from '@/shared/ui/RichTextEditor';
import type { Course, CourseWithProgress, TeachingMaterial } from '@/entities/course';
import CompleteToggleButton from './CompleteToggleButton';
import ChapterNav from './ChapterNav';
import DocTableOfContents from './DocTableOfContents';
import ImageLightbox from './ImageLightbox';
import ReadOnlyMarkdown from './ReadOnlyMarkdown';
import { formatDate } from '../lib/formatDate';
import { stripLeadingTitle } from '../lib/stripLeadingTitle';
import { stripLeadingDocTitle } from '../lib/stripLeadingDocTitle';

/**
 * trainee 向けの教材閲覧ビュー。
 *
 * レイアウトはノート / AI チャットと同じデザイン言語（FRESTYLE-340）:
 * - 本文はカードに入れないフラットな文書（ノートと同じ max-w-3xl 中央カラム + 内部スクロール）
 * - 右の目次・章一覧は SecondaryPanel と同じサイドバー（nav 色 + border 区切り、カードなし）
 * 本文末尾に「完了にする」と「次の章へ / 次のコースへ」を並べ、読み終えた位置から先へ進める。
 */
export default function ReadOnlyDetail({
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
  // サイドバー(目次 + 章一覧)の表示状態は localStorage に保持し、 教材を切り替えても選択が続く
  // ようにする（既定は表示）。横幅が狭いときに本文幅を稼げるよう trainee が出し入れできる。
  const [tocOpen, setTocOpen] = useLocalStorage('course-toc-open', true);

  // 章を切り替えたら本文の先頭までスクロールを戻す（末尾の「次の章へ」から進んでも頭から読める）。
  const scrollRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: 0 });
  }, [material.id]);

  // 本文先頭の h1(= タイトル)は、上のタイトル見出しで material.title を大きく出すため取り除く。
  // 残すとタイトルが二重に見える(FRESTYLE-131)。
  const bodyContent = useMemo(() => stripLeadingTitle(material.content), [material.content]);
  // リッチ本文（tiptap JSON）。未移行の章は null で、従来の Markdown 表示へフォールバックする
  // （content 列の撤去はフェーズ E。それまで両対応を保つ）。
  const bodyDoc = useMemo(
    () => (material.doc ? stripLeadingDocTitle(material.doc) : null),
    [material.doc],
  );
  // doc 表示のコンテナ。目次の anchor id 付与（DocTableOfContents）と画像クリック委譲で使う。
  const articleRef = useRef<HTMLDivElement>(null);

  // 本文内の画像クリックでモーダル拡大表示する(FRESTYLE-191)。
  const [lightboxImage, setLightboxImage] = useState<{ src: string; alt: string } | null>(null);
  useEffect(() => setLightboxImage(null), [material.id]);

  return (
    <div className="flex flex-1 min-h-0 bg-[var(--color-surface)]">
      {/* 本文: ノートと同じ「枠のないインライン文書」。中央カラム + 内部スクロール。 */}
      <div ref={scrollRef} className="flex-1 min-h-0 overflow-y-auto">
        <div className="mx-auto w-full max-w-3xl px-6 py-10">
          <div className="mb-3 flex items-start justify-between gap-3">
            <h1 className="min-w-0 flex-1 text-3xl font-bold text-[var(--color-text-primary)] md:text-4xl">
              {material.title || '無題の教材'}
            </h1>
            <div className="shrink-0 pt-2">
              <CompleteToggleButton completed={completed} onToggle={onToggleComplete} />
            </div>
          </div>
          {/* メタ行(最終更新 / サイドバートグル)。サイドバーは lg 以上でのみ表示されるため、
              トグルも lg 未満では隠す。 */}
          <div className="mb-8 flex flex-wrap items-center gap-3">
            <p className="text-xs text-[var(--color-text-muted)]">
              最終更新: {formatDate(material.updatedAt)}
            </p>
            <button
              type="button"
              onClick={() => setTocOpen((isOpen) => !isOpen)}
              aria-pressed={tocOpen}
              title={tocOpen ? '目次を隠す' : '目次を表示'}
              className={`hidden lg:inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-medium border transition-colors ${
                tocOpen
                  ? 'border-taupe-500 text-taupe-400'
                  : 'border-surface-3 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]'
              }`}
            >
              <ListBulletIcon className="w-4 h-4" />
              目次
            </button>
          </div>

          {bodyDoc ? (
            // 画像クリックの拡大表示は、tiptap の描画 DOM へのクリック委譲で拾う
            // （読み取り専用なのでエディタ側のハンドラと競合しない）。
            <div
              ref={articleRef}
              onClick={(e) => {
                const target = e.target;
                if (target instanceof HTMLImageElement && target.src) {
                  setLightboxImage({ src: target.src, alt: target.alt });
                }
              }}
            >
              <RichTextEditor value={bodyDoc} editable={false} ariaLabel="教材本文" />
            </div>
          ) : (
            <div className="prose prose-sm max-w-none course-prose">
              <ReadOnlyMarkdown content={bodyContent} onImageClick={setLightboxImage} />
            </div>
          )}

          {/* 末尾に「完了にする」と「次の章へ」を並べ、 読み終えた位置から次へ進めるようにする。
              最終章では代わりに「次のコースへ」を出す(FRESTYLE-102)。 */}
          <div className="mt-10 pt-6 border-t border-surface-3 flex flex-col sm:flex-row items-center justify-center gap-3">
            <CompleteToggleButton completed={completed} onToggle={onToggleComplete} large />
            {nextMaterial && onGoNext ? (
              <button
                type="button"
                onClick={onGoNext}
                title={`次の章へ: ${nextMaterial.title || '無題の教材'}`}
                className="inline-flex min-w-0 max-w-full sm:max-w-[55%] items-center justify-center gap-2 px-5 py-2.5 rounded-lg text-sm font-medium bg-surface-2 border border-surface-3 text-[var(--color-text-primary)] hover:bg-surface-3 transition-colors"
              >
                <span className="truncate">次の章へ: {nextMaterial.title || '無題の教材'}</span>
                <ArrowRightIcon className="w-4 h-4 flex-shrink-0" />
              </button>
            ) : nextCourse && onGoNextCourse ? (
              <button
                type="button"
                onClick={onGoNextCourse}
                title={`次のコースへ: ${nextCourse.title || '無題のコース'}`}
                className="inline-flex min-w-0 max-w-full sm:max-w-[55%] items-center justify-center gap-2 px-5 py-2.5 rounded-lg text-sm font-medium bg-surface-2 border border-surface-3 text-[var(--color-text-primary)] hover:bg-surface-3 transition-colors"
              >
                <span className="truncate">次のコースへ: {nextCourse.title || '無題のコース'}</span>
                <ArrowRightIcon className="w-4 h-4 flex-shrink-0" />
              </button>
            ) : null}
          </div>
        </div>

        {lightboxImage && (
          <ImageLightbox
            src={lightboxImage.src}
            alt={lightboxImage.alt}
            onClose={() => setLightboxImage(null)}
          />
        )}
      </div>

      {/* 右サイドバー: SecondaryPanel（ノート / AI チャットの左パネル）と同じデザイン言語。
          nav 色の全高パネル + border 区切りで、カード(rounded/shadow)にはしない(FRESTYLE-340)。 */}
      {tocOpen && (
        <aside className="hidden lg:flex w-72 flex-shrink-0 flex-col min-h-0 border-l border-surface-3 bg-[var(--color-nav)]">
          {bodyDoc ? (
            <>
              <div className="px-4 py-3 border-b border-surface-3">
                <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">目次</h2>
              </div>
              <div className="max-h-[45%] min-h-0 overflow-y-auto px-4 py-3 border-b border-surface-3">
                <DocTableOfContents doc={bodyDoc} articleRef={articleRef} />
              </div>
            </>
          ) : (
            // フォールバック(Markdown)時は目次コンポーネントが自前の見出しを持つため、
            // パネルヘッダーは重ねずにそのまま入れる。
            <div className="max-h-[45%] min-h-0 overflow-y-auto px-4 py-3 border-b border-surface-3">
              <MarkdownTableOfContents content={bodyContent} />
            </div>
          )}
          <div className="flex-1 min-h-0 overflow-y-auto px-4 py-3">
            <ChapterNav
              course={course}
              materials={materials}
              selectedId={material.id}
              completedIds={completedIds}
              completedCount={completedCount}
              onSelect={onSelectMaterial}
            />
          </div>
        </aside>
      )}
    </div>
  );
}
