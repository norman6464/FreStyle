import { useEffect, useMemo, useRef, useState } from 'react';
import { ArrowRightIcon, ListBulletIcon } from '@heroicons/react/24/outline';
import { useLocalStorage } from '@/shared/lib/hooks/useLocalStorage';
import MarkdownTableOfContents from '@/shared/ui/MarkdownTableOfContents';
import type { Course, CourseWithProgress, TeachingMaterial } from '@/entities/course';
import CompleteToggleButton from './CompleteToggleButton';
import ChapterNav from './ChapterNav';
import ImageLightbox from './ImageLightbox';
import ReadOnlyMarkdown from './ReadOnlyMarkdown';
import { formatDate } from '../lib/formatDate';
import { stripLeadingTitle } from '../lib/stripLeadingTitle';

/**
 * trainee 向けの教材閲覧ビュー。記事風の本文カード + 右サイドバー(目次 + 章一覧)。
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

  // 本文内の画像クリックでモーダル拡大表示する(FRESTYLE-191)。ページを離れずに図の細部を
  // 確認できる(別タブで開くリンクは学習が中断されるため FRESTYLE-125 で除去済み)。
  const [lightboxImage, setLightboxImage] = useState<{ src: string; alt: string } | null>(null);
  useEffect(() => setLightboxImage(null), [material.id]);

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
        {/* 本文カラム。 サイドバーを隠したときは本文が全幅に伸びて読みにくいため、 読みやすい幅(860px)に
            収めて中央寄せする。 サイドバー表示時は 1fr カラムが既に同程度の幅になる。 */}
        <div className={`min-w-0 ${!tocOpen ? 'mx-auto w-full max-w-[860px]' : ''}`}>
          <article className="bg-white border border-surface-3 rounded-xl shadow-sm px-6 sm:px-10 py-8 sm:py-10">
            {/* 記事サイト風ヘッダー: タイトル + メタを本文カードの先頭に入れる。
                以前はカード外の別ヘッダーに置いていたが、カードをタイトル位置まで上げて
                タイトルもカード内に収めた(FRESTYLE-178)。目次サイドバーは同じグリッド行に来るので
                カード上端と揃う(FRESTYLE-150)。本文先頭の重複 h1 は stripLeadingTitle で除去済み(FRESTYLE-131)。 */}
            <header className="mb-6 pb-6 border-b border-surface-3">
              <h1 className="text-2xl sm:text-3xl font-bold text-[var(--color-text-primary)] leading-snug">
                {material.title || '無題の教材'}
              </h1>
              {/* メタ(最終更新 / 目次トグル / 完了トグル)。 sticky にはしない(FRESTYLE-119)。
                  スクロール途中の完了操作は本文末尾の大きい完了ボタン(FRESTYLE-100)で行える。 */}
              <div className="mt-3 flex flex-wrap items-center gap-3">
                <p className="text-xs text-[var(--color-text-muted)]">
                  最終更新: {formatDate(material.updatedAt)}
                </p>
                {/* 目次は lg 以上でのみ表示されるため、 トグルも lg 未満では隠す。 */}
                <button
                  type="button"
                  onClick={() => setTocOpen((isOpen) => !isOpen)}
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
            <div className="prose prose-sm max-w-none course-prose">
              <ReadOnlyMarkdown content={bodyContent} onImageClick={setLightboxImage} />
            </div>

            {/* 末尾に「完了にする」と「次の章へ」を並べ、 読み終えた位置から次へ進めるようにする。
                崩れ対策(FRESTYLE-189): 完了ボタンは shrink-0 / whitespace-nowrap で常に 1 行を保ち、
                次へボタンは min-w-0 + truncate で長い章タイトルを省略しつつ、 幅を取りすぎないよう
                上限(sm:max-w-[55%])を設けて 2 つのボタンのバランスを取る。
                最終章では代わりに「次のコースへ」を出し、 一覧に戻らず次のコースへ直行できるようにする
                (FRESTYLE-102。 遷移先はレジュームにより 1 章目が自動表示される)。 */}
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
          </article>

          {lightboxImage && (
            <ImageLightbox
              src={lightboxImage.src}
              alt={lightboxImage.alt}
              onClose={() => setLightboxImage(null)}
            />
          )}
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
