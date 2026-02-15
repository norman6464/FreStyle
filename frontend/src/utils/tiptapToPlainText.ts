import { isLegacyMarkdown } from './isLegacyMarkdown';

interface TiptapNode {
  type: string;
  text?: string;
  content?: TiptapNode[];
}

function extractText(node: TiptapNode): string {
  if (node.type === 'text' && node.text) {
    return node.text;
  }
  if (!node.content) return '';
  return node.content.map(extractText).filter(Boolean).join(' ');
}

export function tiptapToPlainText(content: string): string {
  if (!content) return '';
  if (isLegacyMarkdown(content)) return content;
  try {
    const doc: TiptapNode = JSON.parse(content);
    return extractText(doc).trim();
  } catch {
    return content;
  }
}
