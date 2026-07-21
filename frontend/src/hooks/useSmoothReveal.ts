import { useEffect, useRef, useState } from 'react';
import { splitSubSentences, endsAtBoundary } from '@/shared/lib/subSentenceSegments';

/**
 * useSmoothReveal — ストリーミング本文を Gemini 実物と同じリズムで放出するペーシングフック
 * (FRESTYLE-146)。
 *
 * Gemini web の配信バンドルから確認したアルゴリズムの再現:
 * - 本文を句読点チャンク(sub-sentence)に分割し、1 チャンクずつ表示に追加する
 * - 放出間隔 fb = (チャンク到着間隔の累積平均 avg + 600) / (残チャンク数 + 1)
 *   (avg は各サンプルを 1/n 重みで足す累積算術平均。 Gemini 実装と同型。 サンプルは 1000ms で
 *    キャップ、初期値 3000ms。到着が速いほど・残りが多いほど速く出す)
 * - 遅延が fb×10 を超えていたら floor(遅延/fb) 個をまとめて表示(catch-up)
 * - 応答完了(active true→false)で avg×0.8 に加速し、残りをリズム付きで流し切る(fast-drain)
 *
 * 設計上のルール(既存挙動とテストを壊さないための急所):
 * - mount 時は content 全文を即時表示する(履歴ロード・確定メッセージ・既存テストは無影響)
 * - content が表示済み prefix の延長でないとき(SSE error の全文置換等)は即スナップ
 * - streaming 中、句読点で終わっていない末尾の未完チャンクは保留する
 *   (出してしまうと次の delta で span テキストが書き換わり、フェード無しで文字が増える)。
 *   ただしコードブロック等で境界が来ない場合に備え、一定文字数を超えたら放出する
 */

const AVG_INITIAL_MS = 3000;
const AVG_SAMPLE_CAP_MS = 1000;
const FB_BASE_MS = 600;
const COMPLETE_ACCEL = 0.8;
const CATCHUP_FACTOR = 10;
// 句読点境界が来ない未完チャンクを保留する上限。超えたら放出する(長い無句読点行の停滞防止)。
const MAX_TAIL_HOLD_CHARS = 120;

export interface SmoothRevealResult {
  /** 表示してよい prefix。MarkdownView にはこれを渡す。 */
  text: string;
  /** backlog を出し切った(完了マーカー等の表示に使う)。 */
  settled: boolean;
}

interface RevealState {
  visible: string;
  content: string;
  active: boolean;
  everActive: boolean;
  avg: number;
  samples: number;
  lastArrival: number;
  lastReveal: number;
  timer: ReturnType<typeof setTimeout> | null;
}

export function useSmoothReveal(content: string, active: boolean): SmoothRevealResult {
  // mount 時スナップショットは即時全文表示(履歴・remount・既存テストをすべて無変更で満たす)。
  const [visible, setVisible] = useState(content);

  const stateRef = useRef<RevealState | null>(null);
  if (stateRef.current === null) {
    stateRef.current = {
      visible: content,
      content,
      active,
      everActive: active,
      avg: AVG_INITIAL_MS,
      samples: 0,
      lastArrival: 0,
      lastReveal: Date.now(),
      timer: null,
    };
  }

  useEffect(() => {
    const s = stateRef.current as RevealState;

    const revealTo = (next: string) => {
      s.visible = next;
      s.lastReveal = Date.now();
      setVisible(next);
    };

    const stopTimer = () => {
      if (s.timer !== null) {
        clearTimeout(s.timer);
        s.timer = null;
      }
    };

    // 未放出 backlog を句読点チャンクに分割する。streaming 中は未完の末尾チャンクを保留。
    const pendingChunks = (): string[] => {
      const backlog = s.content.slice(s.visible.length);
      if (!backlog) return [];
      const chunks = splitSubSentences(backlog);
      if (s.active && chunks.length > 0) {
        const tail = chunks[chunks.length - 1];
        if (!endsAtBoundary(tail) && tail.length <= MAX_TAIL_HOLD_CHARS) chunks.pop();
      }
      return chunks;
    };

    // Gemini の適応放出間隔: fb = (累積平均 + 600) / (残チャンク数 + 1)。
    const fbMs = (pending: number) => (s.avg + FB_BASE_MS) / (pending + 1);

    const tick = () => {
      s.timer = null;
      const chunks = pendingChunks();
      if (chunks.length === 0) {
        // 保留中の未完 tail だけが残った状態で完了した場合はここで出し切る。
        if (!s.active && s.visible !== s.content) revealTo(s.content);
        return;
      }
      const fb = fbMs(chunks.length);
      const elapsed = Date.now() - s.lastReveal;
      // catch-up: 大きく遅れていたら floor(遅延/fb) 個まとめて表示(Gemini と同じ)。
      const take =
        elapsed > fb * CATCHUP_FACTOR
          ? Math.min(chunks.length, Math.max(1, Math.floor(elapsed / fb)))
          : 1;
      revealTo(s.visible + chunks.slice(0, take).join(''));
      schedule();
    };

    const schedule = () => {
      if (s.timer !== null) return;
      const chunks = pendingChunks();
      if (chunks.length === 0) {
        if (!s.active && s.visible !== s.content) {
          // 未完 tail のみ残して完了 → 一拍おいて出し切る。
          s.timer = setTimeout(tick, fbMs(1));
        }
        return;
      }
      s.timer = setTimeout(tick, fbMs(chunks.length));
    };

    // チャンク到着間隔の累積平均(Gemini と同型。サンプルは 1000ms キャップ)。
    if (active && content.length > s.content.length) {
      const now = Date.now();
      if (s.lastArrival > 0) {
        const sample = Math.min(AVG_SAMPLE_CAP_MS, now - s.lastArrival);
        s.samples += 1;
        s.avg = ((s.samples - 1) / s.samples) * s.avg + sample / s.samples;
      }
      s.lastArrival = now;
    }
    s.content = content;

    // 完了(active true→false)で放出を加速(Gemini は COMPLETE で ×0.8)。
    if (s.active && !active) s.avg *= COMPLETE_ACCEL;
    s.active = active;
    if (active) s.everActive = true;

    // 表示済み prefix の延長でない content(error の全文置換等)は即スナップ。
    if (!content.startsWith(s.visible)) {
      stopTimer();
      revealTo(content);
      return;
    }
    // 一度もストリーミングしていないメッセージ(履歴・確定・編集)は常に全文即時。
    if (!active && !s.everActive) {
      if (s.visible !== content) revealTo(content);
      return;
    }
    // 「考え中」からの最初のチャンクは fb を待たず即時に出す。
    if (s.visible === '' && content !== '' && s.timer === null) {
      tick();
      return;
    }
    schedule();
  }, [content, active]);

  // unmount で進行中のタイマーを破棄する。
  useEffect(() => {
    return () => {
      const s = stateRef.current;
      if (s?.timer != null) clearTimeout(s.timer);
    };
  }, []);

  return { text: visible, settled: !active && visible === content };
}
