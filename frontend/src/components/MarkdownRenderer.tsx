interface MarkdownRendererProps {
  content: string;
}

function parseInline(text: string): (string | JSX.Element)[] {
  const result: (string | JSX.Element)[] = [];
  let remaining = text;
  let key = 0;

  while (remaining.length > 0) {
    // **bold**
    const boldMatch = remaining.match(/\*\*(.+?)\*\*/);
    if (boldMatch && boldMatch.index !== undefined) {
      if (boldMatch.index > 0) {
        result.push(remaining.slice(0, boldMatch.index));
      }
      result.push(<strong key={key++}>{boldMatch[1]}</strong>);
      remaining = remaining.slice(boldMatch.index + boldMatch[0].length);
      continue;
    }

    // *italic*
    const italicMatch = remaining.match(/\*(.+?)\*/);
    if (italicMatch && italicMatch.index !== undefined) {
      if (italicMatch.index > 0) {
        result.push(remaining.slice(0, italicMatch.index));
      }
      result.push(<em key={key++}>{italicMatch[1]}</em>);
      remaining = remaining.slice(italicMatch.index + italicMatch[0].length);
      continue;
    }

    result.push(remaining);
    break;
  }

  return result;
}

interface Block {
  type: 'heading' | 'ul' | 'ol' | 'paragraph';
  level?: number;
  items?: string[];
  text?: string;
}

function parseBlocks(content: string): Block[] {
  const lines = content.split('\n');
  const blocks: Block[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // Heading
    const headingMatch = line.match(/^(#{1,3})\s+(.+)$/);
    if (headingMatch) {
      blocks.push({ type: 'heading', level: headingMatch[1].length, text: headingMatch[2] });
      i++;
      continue;
    }

    // Unordered list (- or ・)
    if (/^[-・]\s*/.test(line)) {
      const items: string[] = [];
      while (i < lines.length && /^[-・]\s*/.test(lines[i])) {
        items.push(lines[i].replace(/^[-・]\s*/, ''));
        i++;
      }
      blocks.push({ type: 'ul', items });
      continue;
    }

    // Ordered list (1. 2. etc.)
    if (/^\d+\.\s+/.test(line)) {
      const items: string[] = [];
      while (i < lines.length && /^\d+\.\s+/.test(lines[i])) {
        items.push(lines[i].replace(/^\d+\.\s+/, ''));
        i++;
      }
      blocks.push({ type: 'ol', items });
      continue;
    }

    // Paragraph (skip empty lines)
    if (line.trim()) {
      blocks.push({ type: 'paragraph', text: line });
    }
    i++;
  }

  return blocks;
}

export default function MarkdownRenderer({ content }: MarkdownRendererProps) {
  if (!content) return null;

  const blocks = parseBlocks(content);

  return (
    <div className="prose-note space-y-2">
      {blocks.map((block, i) => {
        if (block.type === 'heading') {
          const Tag = `h${block.level}` as 'h1' | 'h2' | 'h3';
          const sizeClass = block.level === 1
            ? 'text-lg font-bold'
            : block.level === 2
              ? 'text-base font-semibold'
              : 'text-sm font-semibold';
          return (
            <Tag key={i} className={`${sizeClass} text-[var(--color-text-primary)]`}>
              {parseInline(block.text!)}
            </Tag>
          );
        }

        if (block.type === 'ul') {
          return (
            <ul key={i} className="list-disc list-inside space-y-0.5 text-sm text-[var(--color-text-primary)]">
              {block.items!.map((item, j) => (
                <li key={j}>{parseInline(item)}</li>
              ))}
            </ul>
          );
        }

        if (block.type === 'ol') {
          return (
            <ol key={i} className="list-decimal list-inside space-y-0.5 text-sm text-[var(--color-text-primary)]">
              {block.items!.map((item, j) => (
                <li key={j}>{parseInline(item)}</li>
              ))}
            </ol>
          );
        }

        return (
          <p key={i} className="text-sm text-[var(--color-text-primary)] leading-relaxed">
            {parseInline(block.text!)}
          </p>
        );
      })}
    </div>
  );
}
