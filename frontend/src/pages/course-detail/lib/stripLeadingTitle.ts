/**
 * 本文の先頭にある h1(章タイトル)を取り除く(FRESTYLE-131)。
 *
 * 教材本文は規約により `# タイトル`(= material.title と同じ)で始まる。閲覧ビューでは
 * タイトルをカード外のヘッダーで大きく表示するため、本文側の先頭 h1 を消して二重表示を防ぐ。
 * 「先頭の空行を飛ばした最初の非空行が h1 のときだけ」除去するので、コードブロック内の
 * `# コメント` や 2 個目以降の見出しには一切触れない(先頭の 1 個だけが対象)。
 */
export function stripLeadingTitle(content: string): string {
  if (!content) return content;
  const lines = content.split('\n');
  let firstNonEmptyLine = 0;
  while (firstNonEmptyLine < lines.length && lines[firstNonEmptyLine].trim() === '') firstNonEmptyLine++;
  if (firstNonEmptyLine >= lines.length || !/^#\s+\S/.test(lines[firstNonEmptyLine])) return content;
  lines.splice(0, firstNonEmptyLine + 1);
  while (lines.length && lines[0].trim() === '') lines.shift();
  return lines.join('\n');
}
