import { useEffect, useMemo, useState, type RefObject } from 'react';
import type { RichDocContent } from '@/shared/ui/RichTextEditor';
import { extractDocHeadings, applyHeadingIds } from '../model/docHeadings';

/**
 * DocTableOfContents — doc（tiptap JSON）ベースの目次サイドバー。
 *
 * Markdown 版（shared/ui/MarkdownTableOfContents）と同じ見た目・同じ id 規則（github-slugger）で、
 * ソースだけ doc JSON に置き換えたもの。tiptap の表示 DOM は見出しに id を持たないため、
 * articleRef のコンテナへ id を振る役目もここで担う（描画完了が非同期なので MutationObserver で追従）。
 */
export default function DocTableOfContents({
  doc,
  articleRef,
}: {
  doc: RichDocContent;
  articleRef: RefObject<HTMLElement | null>;
}) {
  const items = useMemo(() => extractDocHeadings(doc), [doc]);
  const [activeId, setActiveId] = useState<string | null>(null);

  useEffect(() => {
    if (items.length === 0) return;
    const container = articleRef.current;
    if (!container) return;

    // ビューポート上端 30% に入った最初の heading を active 扱い（Markdown 版と同じ規則）。
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((e) => e.isIntersecting)
          .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);
        if (visible.length > 0) {
          setActiveId(visible[0].target.id);
        }
      },
      { rootMargin: '0px 0px -70% 0px', threshold: 0 },
    );

    // tiptap の描画は editor 初期化後に非同期で現れるため、DOM 変化のたびに
    // id を振り直してから observe し直す（章切替で見出しが総入れ替えになる）。
    const wire = () => {
      applyHeadingIds(container, items);
      observer.disconnect();
      items.forEach((item) => {
        const el = document.getElementById(item.id);
        if (el) observer.observe(el);
      });
    };
    wire();
    const mutationObserver = new MutationObserver(wire);
    mutationObserver.observe(container, { childList: true, subtree: true });

    return () => {
      mutationObserver.disconnect();
      observer.disconnect();
    };
  }, [items, articleRef]);

  if (items.length === 0) return null;

  return (
    // 親カードが高さを制限したときに見出し(目次)は固定し、リストだけを内側でスクロールさせる。
    <nav aria-label="目次" className="flex min-h-0 flex-col text-sm">
      <h2 className="mb-3 flex-shrink-0 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-secondary)]">
        目次
      </h2>
      <ul className="min-h-0 flex-1 space-y-1.5 overflow-y-auto border-l border-surface-3 pr-1">
        {items.map((item) => {
          const isActive = activeId === item.id;
          return (
            <li key={item.id} style={{ paddingLeft: `${(item.level - 1) * 12}px` }}>
              <a
                href={`#${item.id}`}
                className={`block -ml-px pl-3 py-0.5 border-l-2 transition-colors ${
                  isActive
                    ? 'border-taupe-500 text-[var(--color-text-primary)] font-medium'
                    : 'border-transparent text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]'
                }`}
              >
                {item.text}
              </a>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
