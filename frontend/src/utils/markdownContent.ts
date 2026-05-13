import { isValidElement, ReactNode } from 'react';

/**
 * React 要素ツリーから text node だけを連結して返す。
 * rehype-highlight 後の `<span class="hljs-...">code</span>` 構造から
 * 元のコードを復元する用途。
 */
export function extractTextContent(node: ReactNode): string {
  if (typeof node === 'string') return node;
  if (typeof node === 'number') return String(node);
  if (Array.isArray(node)) return node.map(extractTextContent).join('');
  if (isValidElement(node)) {
    const props = node.props as { children?: ReactNode };
    return extractTextContent(props.children);
  }
  return '';
}

/**
 * 子要素の className から `language-xxx` を抽出する。
 * 複数行コードのときだけ言語が付く。
 * 言語不明（インラインコード相当 / 未知の言語）は空文字を返す。
 */
export function extractLanguage(node: ReactNode): string {
  if (!isValidElement(node)) return '';
  const el = node as React.ReactElement<{ className?: string }>;
  const match = el.props.className?.match(/language-(\w+)/);
  return match?.[1] ?? '';
}
