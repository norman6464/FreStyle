/**
 * YouTube / Twitter (X) / GitHub permalink などのサービス専用埋め込み NodeView。
 *
 * Zenn 仕様:
 *   - YouTube: 動画 URL 単独行 → iframe
 *   - Twitter (X): ポスト URL 単独行 → blockquote (twitter widget)
 *   - GitHub: ファイル/permalink URL 単独行 → コードブロック (#L行範囲指定 もサポート)
 *
 * 本 PR では iframe / blockquote ベースの軽量描画のみ。GitHub コード本体取得は
 * Phase 2 (raw.githubusercontent.com 経由) で扱う。
 */
import { useMemo } from 'react';
import { NodeViewWrapper, type ReactNodeViewProps } from '@tiptap/react';

type EmbedKind = 'youtube' | 'twitter' | 'github';

type Props = ReactNodeViewProps<HTMLElement> & {
  node: ReactNodeViewProps<HTMLElement>['node'] & {
    attrs: { kind: EmbedKind; url: string };
  };
};

function youtubeId(url: string): string | null {
  // 対応: https://www.youtube.com/watch?v=ID / https://youtu.be/ID
  try {
    const u = new URL(url);
    if (u.hostname === 'youtu.be') return u.pathname.slice(1) || null;
    if (u.hostname.endsWith('youtube.com')) {
      return u.searchParams.get('v');
    }
    return null;
  } catch {
    return null;
  }
}

function twitterStatusUrl(url: string): string | null {
  try {
    const u = new URL(url);
    if (u.hostname === 'twitter.com' || u.hostname === 'x.com') {
      // /<user>/status/<id>
      if (/^\/[^/]+\/status\/\d+/.test(u.pathname)) return url;
    }
    return null;
  } catch {
    return null;
  }
}

export default function ServiceEmbedNodeView({ node }: Props) {
  const kind = (node.attrs.kind as EmbedKind) || 'youtube';
  const url = (node.attrs.url as string) || '';

  const body = useMemo(() => {
    if (kind === 'youtube') {
      const id = youtubeId(url);
      if (!id) return null;
      return (
        <iframe
          src={`https://www.youtube.com/embed/${id}`}
          title="YouTube video player"
          loading="lazy"
          frameBorder={0}
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
          allowFullScreen
          className="w-full aspect-video rounded-lg"
        />
      );
    }
    if (kind === 'twitter') {
      const tweet = twitterStatusUrl(url);
      if (!tweet) return null;
      return (
        // Twitter のオフィシャル widget JS は外部スクリプト読み込みが必要なので
        // Phase 1 では blockquote プレースホルダだけ出す。実際の rich render は
        // フロントの読者ページ側で widgets.js を有効化するか別 PR で扱う。
        <blockquote className="twitter-tweet">
          <a href={tweet} target="_blank" rel="noreferrer noopener">
            {tweet}
          </a>
        </blockquote>
      );
    }
    if (kind === 'github') {
      // GitHub permalink は Phase 1 ではコードブロック取得を行わず、
      // タイトル付きリンクカードに fallback する。
      return (
        <a
          href={url}
          target="_blank"
          rel="noreferrer noopener"
          className="block rounded-lg border border-surface-3 bg-surface-1 p-3 hover:bg-[var(--color-surface-2)] no-underline"
        >
          <div className="text-xs text-[var(--color-text-muted)]">GitHub</div>
          <div className="text-sm font-semibold text-[var(--color-text-primary)] truncate">
            {url}
          </div>
        </a>
      );
    }
    return null;
  }, [kind, url]);

  return (
    <NodeViewWrapper className="service-embed my-3">
      {body ?? (
        <a href={url} target="_blank" rel="noreferrer noopener" className="text-sky-400 underline">
          {url}
        </a>
      )}
    </NodeViewWrapper>
  );
}
