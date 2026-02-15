interface TiptapNode {
  type: string;
  attrs?: Record<string, unknown>;
  content?: TiptapNode[];
  text?: string;
  marks?: { type: string }[];
}

interface Block {
  type: 'heading' | 'ul' | 'ol' | 'paragraph';
  level?: number;
  items?: string[];
  text?: string;
}

function parseBlocks(content: string): Block[] {
  const normalizedContent = content.replace(/\r\n?/g, '\n');
  const lines = normalizedContent.split('\n');
  const blocks: Block[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    const headingMatch = line.match(/^(#{1,3})\s+(.+)$/);
    if (headingMatch) {
      blocks.push({ type: 'heading', level: headingMatch[1].length, text: headingMatch[2] });
      i++;
      continue;
    }

    if (/^[-・]\s*/.test(line)) {
      const items: string[] = [];
      while (i < lines.length && /^[-・]\s*/.test(lines[i])) {
        items.push(lines[i].replace(/^[-・]\s*/, ''));
        i++;
      }
      blocks.push({ type: 'ul', items });
      continue;
    }

    if (/^\d+\.\s+/.test(line)) {
      const items: string[] = [];
      while (i < lines.length && /^\d+\.\s+/.test(lines[i])) {
        items.push(lines[i].replace(/^\d+\.\s+/, ''));
        i++;
      }
      blocks.push({ type: 'ol', items });
      continue;
    }

    if (line.trim()) {
      blocks.push({ type: 'paragraph', text: line });
    }
    i++;
  }

  return blocks;
}

function parseInlineToTiptap(text: string): TiptapNode[] {
  const result: TiptapNode[] = [];
  let remaining = text;

  while (remaining.length > 0) {
    const boldMatch = remaining.match(/\*\*(.+?)\*\*/);
    const italicMatch = remaining.match(/(?<!\*)\*(?!\*)(.+?)(?<!\*)\*(?!\*)/);

    if (!boldMatch && !italicMatch) {
      if (remaining) result.push({ type: 'text', text: remaining });
      break;
    }

    const boldIndex = boldMatch?.index ?? Infinity;
    const italicIndex = italicMatch?.index ?? Infinity;

    if (boldIndex <= italicIndex && boldMatch) {
      if (boldIndex > 0) {
        result.push({ type: 'text', text: remaining.slice(0, boldIndex) });
      }
      result.push({ type: 'text', text: boldMatch[1], marks: [{ type: 'bold' }] });
      remaining = remaining.slice(boldIndex + boldMatch[0].length);
    } else if (italicMatch) {
      if (italicIndex > 0) {
        result.push({ type: 'text', text: remaining.slice(0, italicIndex) });
      }
      result.push({ type: 'text', text: italicMatch[1], marks: [{ type: 'italic' }] });
      remaining = remaining.slice(italicIndex + italicMatch[0].length);
    }
  }

  return result;
}

function textToTiptap(text: string): TiptapNode[] {
  const nodes = parseInlineToTiptap(text);
  return nodes.length > 0 ? nodes : [{ type: 'text', text }];
}

export function markdownToTiptap(markdown: string): TiptapNode {
  if (!markdown || markdown.trim() === '') {
    return { type: 'doc', content: [] };
  }

  const blocks = parseBlocks(markdown);
  const content: TiptapNode[] = [];

  for (const block of blocks) {
    if (block.type === 'heading') {
      content.push({
        type: 'heading',
        attrs: { level: block.level },
        content: textToTiptap(block.text!),
      });
    } else if (block.type === 'ul') {
      content.push({
        type: 'bulletList',
        content: block.items!.map(item => ({
          type: 'listItem',
          content: [{ type: 'paragraph', content: textToTiptap(item) }],
        })),
      });
    } else if (block.type === 'ol') {
      content.push({
        type: 'orderedList',
        content: block.items!.map(item => ({
          type: 'listItem',
          content: [{ type: 'paragraph', content: textToTiptap(item) }],
        })),
      });
    } else {
      content.push({
        type: 'paragraph',
        content: textToTiptap(block.text!),
      });
    }
  }

  return { type: 'doc', content };
}
