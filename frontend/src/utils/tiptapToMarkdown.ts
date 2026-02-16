interface TiptapNode {
  type: string;
  attrs?: Record<string, unknown>;
  content?: TiptapNode[];
  text?: string;
  marks?: { type: string; attrs?: Record<string, unknown> }[];
}

function renderInline(node: TiptapNode): string {
  if (node.type === 'text') {
    let text = node.text || '';
    if (node.marks) {
      for (const mark of node.marks) {
        switch (mark.type) {
          case 'bold':
            text = `**${text}**`;
            break;
          case 'italic':
            text = `*${text}*`;
            break;
          case 'code':
            text = `\`${text}\``;
            break;
          case 'strike':
            text = `~~${text}~~`;
            break;
          case 'link':
            text = `[${text}](${mark.attrs?.href || ''})`;
            break;
        }
      }
    }
    return text;
  }
  if (node.type === 'image') {
    const alt = (node.attrs?.alt as string) || '';
    const src = (node.attrs?.src as string) || '';
    return `![${alt}](${src})`;
  }
  if (node.type === 'hardBreak') {
    return '\n';
  }
  return '';
}

function renderInlineContent(nodes: TiptapNode[] | undefined): string {
  if (!nodes) return '';
  return nodes.map(renderInline).join('');
}

function renderBlock(node: TiptapNode, _depth: number = 0): string {
  switch (node.type) {
    case 'paragraph':
      return renderInlineContent(node.content);

    case 'heading': {
      const level = (node.attrs?.level as number) || 1;
      const prefix = '#'.repeat(level);
      return `${prefix} ${renderInlineContent(node.content)}`;
    }

    case 'bulletList':
      return (node.content || [])
        .map(item => {
          const text = renderListItemContent(item);
          return `- ${text}`;
        })
        .join('\n');

    case 'orderedList':
      return (node.content || [])
        .map((item, i) => {
          const text = renderListItemContent(item);
          return `${i + 1}. ${text}`;
        })
        .join('\n');

    case 'taskList':
      return (node.content || [])
        .map(item => {
          const checked = item.attrs?.checked ? 'x' : ' ';
          const text = renderListItemContent(item);
          return `- [${checked}] ${text}`;
        })
        .join('\n');

    case 'blockquote':
      return (node.content || [])
        .map(child => `> ${renderBlock(child)}`)
        .join('\n');

    case 'codeBlock': {
      const lang = (node.attrs?.language as string) || '';
      const code = renderInlineContent(node.content);
      return `\`\`\`${lang}\n${code}\n\`\`\``;
    }

    case 'horizontalRule':
      return '---';

    case 'image': {
      const alt = (node.attrs?.alt as string) || '';
      const src = (node.attrs?.src as string) || '';
      return `![${alt}](${src})`;
    }

    case 'toggleList': {
      // トグルの要約をテキストとして出力
      const summary = node.content?.find(c => c.type === 'toggleSummary');
      const content = node.content?.find(c => c.type === 'toggleContent');
      const parts: string[] = [];
      if (summary) parts.push(renderInlineContent(summary.content));
      if (content?.content) {
        parts.push(...content.content.map(c => renderBlock(c)));
      }
      return parts.join('\n\n');
    }

    case 'callout': {
      const text = (node.content || []).map(c => renderBlock(c)).join('\n');
      return `> ${text}`;
    }

    case 'table': {
      return renderTable(node);
    }

    default:
      if (node.content) {
        return node.content.map(c => renderBlock(c)).join('\n\n');
      }
      return renderInlineContent(node.content);
  }
}

function renderListItemContent(item: TiptapNode): string {
  if (!item.content) return '';
  return item.content.map(child => renderBlock(child)).join('\n');
}

function renderTable(table: TiptapNode): string {
  if (!table.content) return '';
  const rows = table.content.filter(r => r.type === 'tableRow');
  if (rows.length === 0) return '';

  const lines: string[] = [];
  for (let i = 0; i < rows.length; i++) {
    const cells = (rows[i].content || []).map(cell => {
      return renderInlineContent(cell.content?.[0]?.content);
    });
    lines.push(`| ${cells.join(' | ')} |`);
    if (i === 0) {
      lines.push(`| ${cells.map(() => '---').join(' | ')} |`);
    }
  }
  return lines.join('\n');
}

export function tiptapToMarkdown(jsonString: string): string {
  if (!jsonString) return '';
  try {
    const doc: TiptapNode = JSON.parse(jsonString);
    if (!doc.content || doc.content.length === 0) return '';
    return doc.content.map(node => renderBlock(node)).join('\n\n');
  } catch {
    return jsonString;
  }
}
