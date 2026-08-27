import { useEffect, useRef, useState } from 'react';
import { KnowledgeBaseRepository, type KbPageDoc } from '@/entities/knowledge-base';

export interface KbPageDocState {
  data: KbPageDoc | null;
  loading: boolean;
  /** 失敗の理由。null なら失敗していない。 */
  error: string | null;
}

/**
 * useKbPageDoc は 1 ページ分の本文を取ってくる。
 *
 * 404 は「無い」と「見えない」の両方を意味する。backend が撃ち分けていないので
 * （撃ち分けると ID の総当たりで実在が分かる）、**フロントで「見る権限がありません」と
 * 書いてはいけない。** 書いた瞬間に、フロントが backend の隠していることを喋る。
 */
export function useKbPageDoc(workspaceSlug: string | undefined, pageId: string | undefined) {
  const [state, setState] = useState<KbPageDocState>({ data: null, loading: false, error: null });

  // 速く行き来したときに、古い応答が新しいページを上書きするのを防ぐ。
  const generation = useRef(0);

  useEffect(() => {
    if (!workspaceSlug || !pageId) {
      setState({ data: null, loading: false, error: null });
      return;
    }
    const token = ++generation.current;
    setState((prev) => ({ ...prev, loading: true, error: null }));

    KnowledgeBaseRepository.fetchPage(workspaceSlug, pageId)
      .then((data) => {
        if (token !== generation.current) return;
        setState({ data, loading: false, error: null });
      })
      .catch(() => {
        if (token !== generation.current) return;
        setState({
          data: null,
          loading: false,
          error: 'このページを開けませんでした。移動または削除された可能性があります。',
        });
      });
  }, [workspaceSlug, pageId]);

  return state;
}
