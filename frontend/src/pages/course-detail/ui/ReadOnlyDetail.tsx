import { useEffect, useMemo, useState, type RefObject } from 'react';
import { ArrowRightIcon } from '@heroicons/react/24/outline';
import { emptyRichDoc, RichTextEditor, type RichDocContent } from '@/shared/ui/RichTextEditor';
import type { CourseWithProgress, TeachingMaterial } from '@/entities/course';
import CompleteToggleButton from './CompleteToggleButton';
import ImageLightbox from './ImageLightbox';
import { formatDate } from '../lib/formatDate';

/**
 * trainee 向けの教材閲覧ビュー。
 *
 * レイアウトはノートと同じ「枠のないインライン文書」（max-w-3xl 中央カラム + 内部スクロール。
 * FRESTYLE-340）。目次・章一覧は左の SecondaryPanel（CourseDetailPage 側）に集約した
 * （FRESTYLE-341。右サイドバーは廃止）。
 * 本文末尾に「完了にする」と「次の章へ / 次のコースへ」を並べ、読み終えた位置から先へ進める。
 */
export default function ReadOnlyDetail({
  material,
  bodyDoc,
  articleRef,
  scrollContainerRef,
  completed,
  onToggleComplete,
  nextMaterial,
  onGoNext,
  nextCourse,
  onGoNextCourse,
}: {
  material: TeachingMaterial;
  /** 先頭 h1 除去済みのリッチ本文。null（まだ本文を保存していない新規章）は空 doc として描画する。 */
  bodyDoc: RichDocContent | null;
  /** 本文コンテナの ref。左パネルの目次(DocTableOfContents)が anchor id を振るために参照する。 */
  articleRef: RefObject<HTMLDivElement | null>;
  /** 本文スクロールコンテナを親（ヘッダー自動隠しの監視）へ渡す callback ref。 */
  scrollContainerRef?: (node: HTMLElement | null) => void;
  completed: boolean;
  onToggleComplete: (done: boolean) => void;
  nextMaterial?: TeachingMaterial | null;
  onGoNext?: () => void;
  nextCourse?: CourseWithProgress | null;
  onGoNextCourse?: () => void;
}) {
  // 章を切り替えたら本文の先頭までスクロールを戻す（末尾の「次の章へ」から進んでも頭から読める）。
  useEffect(() => {
    articleRef.current?.closest('[data-course-scroll]')?.scrollTo({ top: 0 });
  }, [material.id, articleRef]);

  // doc が null の章（理論上は本文未保存の新規章のみ）は空 doc として描画する。
  const displayDoc = useMemo(() => bodyDoc ?? emptyRichDoc(), [bodyDoc]);

  // 本文内の画像クリックでモーダル拡大表示する(FRESTYLE-191)。
  const [lightboxImage, setLightboxImage] = useState<{ src: string; alt: string } | null>(null);
  useEffect(() => setLightboxImage(null), [material.id]);

  return (
    <div className="flex flex-1 min-h-0 bg-[var(--color-surface)]">
      {/* 本文: ノートと同じ「枠のないインライン文書」。中央カラム + 内部スクロール。 */}
      <div ref={scrollContainerRef} data-course-scroll className="flex-1 min-h-0 overflow-y-auto">
        <div className="mx-auto w-full max-w-3xl px-6 py-10">
          <div className="mb-3 flex items-start justify-between gap-3">
            <h1 className="min-w-0 flex-1 text-3xl font-bold text-[var(--color-text-primary)] md:text-4xl">
              {material.title || '無題の教材'}
            </h1>
            <div className="shrink-0 pt-2">
              <CompleteToggleButton completed={completed} onToggle={onToggleComplete} />
            </div>
          </div>
          {/* メタ行。目次・章一覧の出し入れは左パネル(SecondaryPanel の «/☰)に統合済み。 */}
          <div className="mb-8 flex flex-wrap items-center gap-3">
            <p className="text-xs text-[var(--color-text-muted)]">
              最終更新: {formatDate(material.updatedAt)}
            </p>
          </div>

          {/* 画像クリックの拡大表示は、tiptap の描画 DOM へのクリック委譲で拾う
              （読み取り専用なのでエディタ側のハンドラと競合しない）。 */}
          <div
            ref={articleRef}
            onClick={(e) => {
              const target = e.target;
              if (target instanceof HTMLImageElement && target.src) {
                setLightboxImage({ src: target.src, alt: target.alt });
              }
            }}
          >
            <RichTextEditor value={displayDoc} editable={false} ariaLabel="教材本文" className="course-doc" />
          </div>

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

    </div>
  );
}
