import type { ReactNode } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import rehypeHighlight from 'rehype-highlight';
import rehypeSlug from 'rehype-slug';

/** trainee の閲覧用に Markdown を render するだけのコンポーネント（画像クリックで拡大表示できる）。 */
export default function ReadOnlyMarkdown({
  content,
  onImageClick,
}: {
  content: string;
  onImageClick?: (image: { src: string; alt: string }) => void;
}) {
  if (!content.trim()) {
    return <p className="text-[var(--color-text-muted)]">この教材にはまだ本文がありません。</p>;
  }
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm, remarkBreaks]}
      rehypePlugins={[rehypeSlug, rehypeHighlight]}
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
        // 図（draw.io から書き出した PNG/SVG 等）を本文に埋め込めるようにする。
        // 中央寄せ + 枠 + 白背景（透過 SVG が見えるよう）。リンクにはしない
        // （クリックで画像 URL が別タブに開き学習が中断される、というユーザー要望で FRESTYLE-125 にて除去）。
        // 代わりにクリックでモーダル拡大表示する(FRESTYLE-191。ページを離れないので中断されない)。
        img: ({ src, alt }) => {
          const url = typeof src === 'string' ? src : undefined;
          const image = (
            <img
              src={url}
              alt={alt ?? ''}
              loading="lazy"
              className="mx-auto max-w-[90%] h-auto rounded-lg border border-surface-3 bg-white"
            />
          );
          return (
            <figure className="my-5">
              {onImageClick && url ? (
                <button
                  type="button"
                  onClick={() => onImageClick({ src: url, alt: alt ?? '' })}
                  aria-label={`${alt || '図'}を拡大表示`}
                  className="block w-full cursor-zoom-in"
                >
                  {image}
                </button>
              ) : (
                image
              )}
              {alt && (
                <figcaption className="mt-2 text-center text-xs text-[var(--color-text-muted)]">
                  {alt}
                </figcaption>
              )}
            </figure>
          );
        },
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
        table: ({ children }) => (
          <div className="overflow-x-auto">
            <table className="text-sm border-collapse">{children as ReactNode}</table>
          </div>
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  );
}
