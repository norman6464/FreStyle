import { isLegacyMarkdown } from './isLegacyMarkdown';

interface TiptapNode {
  type: string;
  text?: string;
  content?: TiptapNode[];
}

const CONTAINER_BLOCK_TYPES = new Set([
  'doc', 'blockquote', 'bulletList', 'orderedList', 'listItem',
  'toggleList', 'toggleContent',
]);

function extractText(node: TiptapNode): string {
  if (node.type === 'text' && node.text) {
    return node.text;
  }
  if (!node.content || node.content.length === 0) return '';
  const childTexts = node.content.map(extractText).filter(Boolean);
  if (childTexts.length === 0) return '';
  const separator = CONTAINER_BLOCK_TYPES.has(node.type) ? ' ' : '';
  return childTexts.join(separator);
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
