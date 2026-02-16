import { useCallback } from 'react';
import { tiptapToMarkdown } from '../utils/tiptapToMarkdown';

export function useMarkdownExport() {
  const exportAsMarkdown = useCallback((title: string, content: string) => {
    const markdown = tiptapToMarkdown(content);
    const fullContent = title ? `# ${title}\n\n${markdown}` : markdown;

    const blob = new Blob([fullContent], { type: 'text/markdown;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${title || '無題'}.md`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, []);

  const copyAsMarkdown = useCallback((title: string, content: string): string => {
    const markdown = tiptapToMarkdown(content);
    return title ? `# ${title}\n\n${markdown}` : markdown;
  }, []);

  return { exportAsMarkdown, copyAsMarkdown };
}
