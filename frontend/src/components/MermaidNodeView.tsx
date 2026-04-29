import { useEffect, useRef, useState } from 'react';
import { NodeViewWrapper, NodeViewContent, type ReactNodeViewProps } from '@tiptap/react';
import { logger } from '../lib/logger';

/**
 * Mermaid ダイアグラム NodeView。
 *
 * 仕様:
 * - mermaid は約 600KB のため dynamic import で初回利用時のみロードする
 *   (Note 編集ページに mermaid を含まないユーザーへの bundle インパクト回避)
 * - parse 失敗時は raw コードブロックをそのまま表示する
 * - Zenn 制約 (公式 docs より) を踏襲:
 *     - 1 ブロック 2000 文字以内
 *     - クリックイベントは無効
 *
 * NodeView は Tiptap codeBlock の language=mermaid を上書きするが、
 * codeBlock 全体は CodeBlockExtension に任せたいので、こちらは別 atom node として
 * `mermaid` 専用ブロックを切り出す形にしている (markdown ↔ Tiptap でロスレス)。
 */
const MAX_LENGTH = 2000;

type Props = ReactNodeViewProps<HTMLElement> & {
  node: ReactNodeViewProps<HTMLElement>['node'] & { attrs: { code: string } };
};

export default function MermaidNodeView({ node }: Props) {
  const code = (node.attrs.code as string) || '';
  const containerRef = useRef<HTMLDivElement>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    if (!code.trim()) {
      setError(null);
      return;
    }
    if (code.length > MAX_LENGTH) {
      setError(`Mermaid ブロックは ${MAX_LENGTH} 文字以内にしてください (現在 ${code.length} 文字)`);
      return;
    }

    (async () => {
      try {
        const mermaid = (await import('mermaid')).default;
        mermaid.initialize({
          startOnLoad: false,
          // Zenn の制約に合わせて click event を発火させない
          securityLevel: 'strict',
        });
        const id = `mermaid-${Math.random().toString(36).slice(2, 9)}`;
        const { svg } = await mermaid.render(id, code);
        if (cancelled || !containerRef.current) return;
        containerRef.current.innerHTML = svg;
        setError(null);
      } catch (e) {
        if (cancelled) return;
        logger.error('mermaid render failed:', e);
        setError(e instanceof Error ? e.message : 'Mermaid のレンダリングに失敗しました');
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [code]);

  return (
    <NodeViewWrapper className="mermaid-block my-4 rounded-lg border border-surface-3 bg-surface-1 p-4">
      {error ? (
        <pre className="text-xs text-red-400 whitespace-pre-wrap" aria-live="polite">{error}</pre>
      ) : (
        <div ref={containerRef} className="mermaid-rendered overflow-x-auto" aria-label="Mermaid ダイアグラム" />
      )}
      {/* NodeViewContent は atom node では不要だが、Tiptap が選択フォーカスを処理するために span だけ置く */}
      <NodeViewContent style={{ display: 'none' }} />
    </NodeViewWrapper>
  );
}
