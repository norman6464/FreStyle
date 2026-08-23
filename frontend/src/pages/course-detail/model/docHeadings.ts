import GithubSlugger from 'github-slugger';
import type { JSONContent } from '@tiptap/react';
import type { RichDocContent } from '@/shared/ui/RichTextEditor';

export interface DocTocItem {
  id: string;
  text: string;
  level: number;
}

/**
 * extractDocHeadings は doc（tiptap JSON）から h1〜h3 を文書順に抽出し、
 * Markdown 版の目次（rehype-slug / github-slugger）と同じアルゴリズムで anchor id を付ける。
 * 表示 DOM の見出しへ同じ順序で id を振れば（applyHeadingIds）、目次リンクと対応が取れる。
 */
export function extractDocHeadings(doc: RichDocContent): DocTocItem[] {
  const slugger = new GithubSlugger();
  const items: DocTocItem[] = [];
  const walk = (node: JSONContent) => {
    if (node.type === 'heading') {
      const level = Number(node.attrs?.level ?? 1);
      if (level >= 1 && level <= 3) {
        const text = collectText(node).trim();
        if (text) items.push({ level, text, id: slugger.slug(text) });
      }
      return; // 見出しの中に見出しは無い
    }
    node.content?.forEach(walk);
  };
  doc.content?.forEach(walk);
  return items;
}

/**
 * applyHeadingIds はコンテナ内の h1〜h3 に items の id を文書順で付与する。
 * tiptap の描画は見出しへ id を振らないため、目次リンク（#anchor）の着地点をここで作る。
 */
export function applyHeadingIds(container: HTMLElement, items: DocTocItem[]): void {
  const headings = container.querySelectorAll('h1, h2, h3');
  headings.forEach((heading, index) => {
    const item = items[index];
    if (item) heading.id = item.id;
  });
}

function collectText(node: JSONContent): string {
  if (node.type === 'text') return node.text ?? '';
  return (node.content ?? []).map(collectText).join('');
}
