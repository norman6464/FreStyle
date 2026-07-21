import { ReactNode, memo, useMemo } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import CodeBlock from './CodeBlock';
import rehypeFadeSegments from '../lib/rehypeFadeSegments';

/**
 * react-markdown のラッパ。コンポーネントマップで pre / code / a / table などを
 * プロジェクトのトーンに揃える。`prose` クラス（Tailwind Typography）が当たって
 * いる前提なので、要素ごとにクラス上書きは最小限。
 *
 * `isStreaming` が true の間だけ、本文を句読点チャンク単位の `<span class="fade-seg">` に分割する
 * rehype プラグインを差し込み、新しく現れたチャンクを CSS でフェードインさせる(Gemini 実物仕様)。
 * 完了後・履歴・非ストリーミング時はプラグインを外し、素の Markdown(span なし)に戻すことで、
 * DOM を軽く保ち、既存テストの `getByText` 完全一致も維持する。
 *
 * memo 化: ストリーミング中の再パースを「SSE token 毎」ではなく「ペーシングの放出毎」に
 * 抑えるため(props の content は useSmoothReveal の visible prefix が変わったときだけ変わる)。
 */
export default memo(function MarkdownView({
  content,
  isStreaming = false,
}: {
  content: string;
  isStreaming?: boolean;
}) {
  const rehypePlugins = useMemo(
    () => (isStreaming ? [rehypeHighlight, rehypeFadeSegments] : [rehypeHighlight]),
    [isStreaming],
  );

  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      rehypePlugins={rehypePlugins}
      components={{
        a: ({ href, children }) => (
          <a
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            className="text-brand-400 underline-offset-2 hover:underline"
          >
            {children as ReactNode}
          </a>
        ),
        // インラインコード（pre 直下でない code）にだけスタイルを当てる。
        // ブロックコードは rehype-highlight + highlight.js テーマで装飾済。
        code: ({ className, children, ...props }) => {
          const isBlock = className?.includes('language-');
          if (isBlock) {
            return (
              <code className={className} {...props}>
                {children as ReactNode}
              </code>
            );
          }
          return (
            <code className="px-1 py-0.5 rounded bg-[var(--color-surface-3)] text-[0.85em]">
              {children as ReactNode}
            </code>
          );
        },
        pre: ({ children }) => <CodeBlock>{children as ReactNode}</CodeBlock>,
        table: ({ children }) => (
          <div className="overflow-x-auto">
            <table className="text-sm border-collapse">{children as ReactNode}</table>
          </div>
        ),
        th: ({ children }) => (
          <th className="border border-[var(--color-surface-3)] px-2 py-1 bg-[var(--color-surface-2)] text-left">
            {children as ReactNode}
          </th>
        ),
        td: ({ children }) => (
          <td className="border border-[var(--color-surface-3)] px-2 py-1">
            {children as ReactNode}
          </td>
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  );
});
