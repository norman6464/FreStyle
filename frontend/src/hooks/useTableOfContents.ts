import { useMemo, useState, useCallback } from 'react';

export interface TocHeading {
  level: number;
  text: string;
  id: string;
}

interface TiptapNode {
  type: string;
  text?: string;
  attrs?: { level?: number };
  content?: TiptapNode[];
}

function extractText(nodes: TiptapNode[]): string {
  return nodes
    .map((n) => {
      if (n.type === 'text' && n.text) return n.text;
      if (n.content) return extractText(n.content);
      return '';
    })
    .join('');
}

export function useTableOfContents(content: string) {
  const [isOpen, setIsOpen] = useState(false);

  const headings = useMemo<TocHeading[]>(() => {
    if (!content) return [];
    try {
      const doc: TiptapNode = JSON.parse(content);
      if (!doc.content) return [];

      let index = 0;
      return doc.content
        .filter((node) => node.type === 'heading' && node.attrs?.level)
        .map((node) => ({
          level: node.attrs!.level!,
          text: node.content ? extractText(node.content) : '',
        }))
        .filter((h) => h.text.trim() !== '')
        .map((h) => ({ ...h, id: `heading-${index++}` }));
    } catch {
      return [];
    }
  }, [content]);

  const toggle = useCallback(() => {
    setIsOpen((prev) => !prev);
  }, []);

  return { headings, isOpen, toggle };
}
