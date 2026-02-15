export function isLegacyMarkdown(content: string): boolean {
  if (!content || content.trim() === '') return false;
  try {
    const parsed = JSON.parse(content);
    return !(parsed && parsed.type === 'doc' && Array.isArray(parsed.content));
  } catch {
    return true;
  }
}
