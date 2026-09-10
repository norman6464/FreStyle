import { useCallback, useRef, useState } from 'react';
import { TicketRepository, type TicketCommentEdit } from '@/entities/ticket';

export interface CommentEditsState {
  edits: TicketCommentEdit[];
  loading: boolean;
  error: string | null;
}

const EMPTY: CommentEditsState = { edits: [], loading: false, error: null };

/**
 * useCommentEdits は「（編集済み）」を押した発言 1 件ぶんの編集前の本文を引く。
 *
 * 一覧に混ぜて先読みする口が無いので、押した瞬間に 1 本叩く。同時に開けるのは
 * 呼び出し側（TicketCommentItem の並び）が 1 件に絞る前提で、ここでは宛先の
 * 世代管理はしない（1 発言 1 インスタンスで使う想定の軽い hook）。
 */
export function useCommentEdits(workspaceSlug: string, ticketId: string, commentId: string) {
  const [state, setState] = useState<CommentEditsState>(EMPTY);
  const seq = useRef(0);

  const load = useCallback(async () => {
    const request = ++seq.current;
    setState({ edits: [], loading: true, error: null });
    try {
      const edits = await TicketRepository.fetchTicketCommentEdits(workspaceSlug, ticketId, commentId);
      if (seq.current !== request) return;
      setState({ edits, loading: false, error: null });
    } catch {
      if (seq.current !== request) return;
      setState({ edits: [], loading: false, error: '編集履歴を読み込めませんでした。' });
    }
  }, [workspaceSlug, ticketId, commentId]);

  return { ...state, load };
}
