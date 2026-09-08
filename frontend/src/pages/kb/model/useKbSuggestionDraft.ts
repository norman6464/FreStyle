import { useCallback, useEffect, useState } from 'react';
import { KbRepository } from '@/entities/kb';

export interface KbSuggestionDraftState {
  /** ドラフトモード中か。true の間だけ本文表示エリアが編集可能な下書きに切り替わる。 */
  open: boolean;
  /** 下書き中の doc。open=false の間は null。 */
  draft: unknown;
  /** createSuggestion が飛んでいる間 true。 */
  submitting: boolean;
  /** 失敗の理由。null なら失敗していない。ドラフトは消さずここへ出す。 */
  error: string | null;
}

const CLOSED: KbSuggestionDraftState = { open: false, draft: null, submitting: false, error: null };

/**
 * useKbSuggestionDraft は commenter の「変更を提案する」ドラフトモードを持つ。
 *
 * **既存の useKbPageDoc の自動保存パイプライン（flushSave・デバウンス・onDocChange）とは
 * 完全に別系統。** backend の提案作成 API は作成のみ（既存の open な提案を更新する API が無い）
 * ため、自動保存のようにデバウンスのたびに送ると新しい提案行が量産されてしまう。
 * ここに置くのは「送信」ボタンが押されたときに 1 回だけ createSuggestion を呼ぶ、
 * それだけの状態機械。
 *
 * 失敗は投げない — KbSaveAsTemplateButton と同じ流儀で、エラーはこの state の中に持ち、
 * ドラフトモードのまま・入力を保持したまま呼び出し側（KbPage）が帯に表示する。
 * 成功したかどうかは submit の戻り値（boolean）で呼び出し側へ伝える
 * （成功トーストを出す・失敗トーストは出さない、という出し分けを呼び出し側に委ねるため）。
 */
export function useKbSuggestionDraft(workspaceSlug: string | undefined, pageId: string | undefined) {
  const [state, setState] = useState<KbSuggestionDraftState>(CLOSED);

  // ページを移ったら、書きかけの下書きを持ち越さない（共有・コメント・履歴の各パネルと同じ理由）。
  useEffect(() => {
    setState(CLOSED);
  }, [workspaceSlug, pageId]);

  const start = useCallback((initialDoc: unknown) => {
    setState({ open: true, draft: initialDoc, submitting: false, error: null });
  }, []);

  const cancel = useCallback(() => {
    setState(CLOSED);
  }, []);

  const changeDraft = useCallback((doc: unknown) => {
    setState((prev) => (prev.open ? { ...prev, draft: doc } : prev));
  }, []);

  /** submit は下書きを 1 回だけ提案として送る。成功したら true、失敗したら false を返す。 */
  const submit = useCallback(async (): Promise<boolean> => {
    if (!workspaceSlug || !pageId) return false;
    setState((prev) => ({ ...prev, submitting: true, error: null }));
    try {
      await KbRepository.createSuggestion(workspaceSlug, pageId, state.draft);
      setState(CLOSED);
      return true;
    } catch {
      setState((prev) => ({ ...prev, submitting: false, error: '提案を送信できませんでした。' }));
      return false;
    }
  }, [workspaceSlug, pageId, state.draft]);

  return { ...state, start, cancel, changeDraft, submit };
}
