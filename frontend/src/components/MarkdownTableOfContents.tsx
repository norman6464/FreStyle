import { useEffect, useMemo, useState } from 'react';
import GithubSlugger from 'github-slugger';

interface TocItem {
  id: string;
  text: string;
  level: number;
}

/**
 * MarkdownTableOfContents — 拡張記法風の目次サイドバー。
 *
 * Markdown 本文から heading (#, ##, ###) を抽出し、 rehype-slug が生成する id と
 * 同じアルゴリズム（github-slugger）で anchor を組み立てる。
 * IntersectionObserver で現在表示中の見出しをハイライトする。
 */
export default function MarkdownTableOfContents({ content }: { content: string }) {
  const items = useMemo(() => extractHeadings(content), [content]);
  const [activeId, setActiveId] = useState<string | null>(null);

  useEffect(() => {
    if (items.length === 0) return;

    // 各見出しの可視状態を監視。 ビューポート上端 30% に入った最後の heading を active 扱い。
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

    items.forEach((it) => {
      const el = document.getElementById(it.id);
      if (el) observer.observe(el);
    });

    return () => observer.disconnect();
  }, [items]);

  if (items.length === 0) return null;

  return (
    <nav aria-label="目次" className="text-sm">
      <h2 className="text-xs font-semibold text-[var(--color-text-secondary)] tracking-wider uppercase mb-3">
        目次
      </h2>
      <ul className="space-y-1.5 border-l border-surface-3">
        {items.map((it) => {
          const isActive = activeId === it.id;
          return (
            <li
              key={it.id}
              style={{ paddingLeft: `${(it.level - 1) * 12}px` }}
            >
              <a
                href={`#${it.id}`}
                className={`block -ml-px pl-3 py-0.5 border-l-2 transition-colors ${
                  isActive
                    ? 'border-taupe-500 text-[var(--color-text-primary)] font-medium'
                    : 'border-transparent text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]'
                }`}
              >
                {it.text}
              </a>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}

const HEADING_RE = /^(#{1,3})\s+(.+?)\s*#*\s*$/;
const FENCE_RE = /^```/;

function extractHeadings(content: string): TocItem[] {
  if (!content) return [];
  const slugger = new GithubSlugger();
  const items: TocItem[] = [];
  let inFence = false;

  for (const rawLine of content.split('\n')) {
    if (FENCE_RE.test(rawLine.trim())) {
      inFence = !inFence;
      continue;
    }
    if (inFence) continue;

    const match = HEADING_RE.exec(rawLine);
    if (!match) continue;
    const level = match[1].length;
    const text = stripMarkdownInline(match[2]);
    if (!text) continue;
    items.push({ level, text, id: slugger.slug(text) });
  }

  return items;
}

// インライン記法（**bold** / `code` / [link](url)）を表示用に剥がす。
function stripMarkdownInline(s: string): string {
  return s
    .replace(/`([^`]+)`/g, '$1')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\*([^*]+)\*/g, '$1')
    .replace(/~~([^~]+)~~/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
    .trim();
}
