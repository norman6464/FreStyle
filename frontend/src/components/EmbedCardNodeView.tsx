import { useEffect, useState } from 'react';
import { NodeViewWrapper, type ReactNodeViewProps } from '@tiptap/react';
import { EmbedRepository, type EmbedCardDto } from '../repositories/EmbedRepository';
import { logger } from '../lib/logger';

/**
 * URL 単独行 / `@[card](url)` を render するカード。
 *
 * 表示パターン:
 *   1. fetching:    URL のみ表示するスケルトン
 *   2. resolved:    OGP の画像 + タイトル + サイト名 + URL のリンク
 *   3. error:       fallback として通常リンクを表示
 *
 * Note: URL は Tiptap node の attrs.url から取り、Go backend の /api/v2/embeds/oembed
 * を経由してメタを取得する (SSRF 対策はバックエンド側)。
 */
type Props = ReactNodeViewProps<HTMLElement> & {
  node: ReactNodeViewProps<HTMLElement>['node'] & { attrs: { url: string } };
};

export default function EmbedCardNodeView({ node }: Props) {
  const url = (node.attrs.url as string) || '';
  const [card, setCard] = useState<EmbedCardDto | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    let cancelled = false;
    if (!url) {
      setError(true);
      return;
    }
    EmbedRepository.resolve(url)
      .then((c) => {
        if (!cancelled) setCard(c);
      })
      .catch((e) => {
        if (!cancelled) {
          logger.error('embed resolve failed:', e);
          setError(true);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [url]);

  if (error || !url) {
    return (
      <NodeViewWrapper>
        <a href={url} target="_blank" rel="noreferrer noopener" className="text-sky-400 underline">
          {url || '(URL なし)'}
        </a>
      </NodeViewWrapper>
    );
  }

  if (!card) {
    return (
      <NodeViewWrapper>
        <div className="my-3 rounded-lg border border-surface-3 bg-surface-1 p-3 text-xs text-[var(--color-text-muted)]">
          読み込み中: <span className="break-all">{url}</span>
        </div>
      </NodeViewWrapper>
    );
  }

  return (
    <NodeViewWrapper>
      <a
        href={card.url}
        target="_blank"
        rel="noreferrer noopener"
        className="my-3 flex gap-3 rounded-lg border border-surface-3 bg-surface-1 p-3 hover:bg-[var(--color-surface-2)] transition-colors no-underline"
      >
        {card.imageUrl && (
          <img
            src={card.imageUrl}
            alt=""
            loading="lazy"
            className="w-24 h-24 object-cover rounded-md flex-shrink-0"
          />
        )}
        <div className="flex-1 min-w-0">
          <div className="text-sm font-semibold text-[var(--color-text-primary)] line-clamp-2">
            {card.title || card.url}
          </div>
          {card.description && (
            <div className="text-xs text-[var(--color-text-muted)] line-clamp-2 mt-1">
              {card.description}
            </div>
          )}
          <div className="text-[10px] text-[var(--color-text-faint)] mt-1 truncate">
            {card.siteName ? `${card.siteName} · ` : ''}
            {card.url}
          </div>
        </div>
      </a>
    </NodeViewWrapper>
  );
}
