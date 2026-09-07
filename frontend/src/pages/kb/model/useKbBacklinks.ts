import { useEffect, useRef, useState } from 'react';
import { KbRepository, type KbPage } from '@/entities/kb';

export interface KbBacklinksState {
  /** このページを参照しているページの一覧。 */
  pages: KbPage[];
  loading: boolean;
  /** 失敗の理由。null なら失敗していない。 */
  error: string | null;
}

const EMPTY: KbBacklinksState = { pages: [], loading: false, error: null };

const LOAD_FAILED = '参照しているページを読み込めませんでした。';

/**
 * useKbBacklinks はページ 1 枚の逆リンク（このページを参照しているページ）一覧を取得する。
 *
 * **折りたたみ（KbBacklinksSection）の開閉状態には依存せず、ページを開いたら常に取得する**
 * （useKbComments がバッジ表示のため常時取得するのと同じ考え方 — 件数バッジ相当の表示
 * （セクションの見出しに件数を出す）に使うには、閉じている間も取れていないといけない）。
 *
 * 読み取り専用で書き込みが無いデータなので、useKbComments / useKbPageVersions のような
 * seq・writeCount を使った書き込み競合ガードは持たない。**唯一守るのは宛先チェック**
 * （応答が返る前に別ページへ移っていたら、古い応答で新しいページの状態を上書きしない）—
 * useKbPageDoc と同じ、単純な世代番号（generation ref）だけで足りる。
 */
export function useKbBacklinks(
  workspaceSlug: string | undefined,
  pageId: string | undefined,
): KbBacklinksState {
  const [state, setState] = useState<KbBacklinksState>(EMPTY);
  // 速く行き来したときに、古い応答が新しいページの状態を上書きしないための世代番号。
  const generation = useRef(0);

  useEffect(() => {
    if (!workspaceSlug || !pageId) {
      // ページが決まっていない。連番を進めて、飛んでいる応答を無効にする。
      generation.current += 1;
      setState(EMPTY);
      return;
    }
    const token = ++generation.current;
    setState({ pages: [], loading: true, error: null });
    KbRepository.listBacklinks(workspaceSlug, pageId)
      .then((pages) => {
        if (token !== generation.current) return;
        setState({ pages, loading: false, error: null });
      })
      .catch(() => {
        if (token !== generation.current) return;
        setState({ pages: [], loading: false, error: LOAD_FAILED });
      });
  }, [workspaceSlug, pageId]);

  return state;
}
