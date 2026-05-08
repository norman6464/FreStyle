import { useEffect, useState, useCallback, useRef } from 'react';

/**
 * バックエンド (`/api/v2/health`) の死活ヘルスチェック。
 *
 * 用途: 毎日 22:00〜翌 07:00 JST の scheduled-stop 窓や突発障害で ECS / ALB が
 * 落ちているときに、フロント側で「メンテナンス中ページ」へ切り替えるための
 * シグナルを返す。
 *
 * 仕様:
 *   - 初期値は `'unknown'`。最初の 1 回のチェックが完了するまでアプリ本体は
 *     表示せず、軽い loading で待機する設計を上位コンポーネントに任せる。
 *   - 連続 2 回失敗で `'unhealthy'` → メンテナンス表示
 *   - 1 回でも成功すれば `'healthy'` → 通常表示
 *   - poll 間隔: 通常 60 秒、unhealthy の間は 15 秒に短縮して早く復旧検知
 *   - timeout は 5 秒（ALB が無い時の DNS / connection 失敗を待ちすぎないため）
 *   - 単体の transient エラーで誤って「メンテナンス」表示にしないように
 *     2 回しきい値を採用
 */
export type BackendHealth = 'unknown' | 'healthy' | 'unhealthy';

interface UseBackendHealthResult {
  status: BackendHealth;
  /** 手動で再チェックしたいときに呼ぶ（メンテナンスページの「再試行」ボタン用） */
  recheck: () => void;
}

const HEALTH_PATH = '/api/v2/health';
const TIMEOUT_MS = 5_000;
const POLL_INTERVAL_HEALTHY_MS = 60_000;
const POLL_INTERVAL_UNHEALTHY_MS = 15_000;
const FAILURE_THRESHOLD = 2;

export function useBackendHealth(): UseBackendHealthResult {
  const [status, setStatus] = useState<BackendHealth>('unknown');
  const failureCountRef = useRef(0);
  const cancelledRef = useRef(false);
  const triggerRecheckRef = useRef<() => void>(() => {});
  // schedule() がレンダーをまたいで最新の status を見るための ref。これがないと
  // status を useEffect の依存配列に入れる必要があり、状態遷移のたびに余分なヘルスチェックが発生する。
  const statusRef = useRef<BackendHealth>('unknown');
  statusRef.current = status;
  // インフライトの fetch をアンマウントや再チェック時に即座に中断する用。
  const fetchControllerRef = useRef<AbortController | null>(null);

  const apiBase = (import.meta.env.VITE_API_BASE_URL ?? '') as string;
  const url = apiBase + HEALTH_PATH;

  const check = useCallback(async () => {
    fetchControllerRef.current?.abort();
    const controller = new AbortController();
    fetchControllerRef.current = controller;
    try {
      const res = await fetch(url, {
        method: 'GET',
        cache: 'no-store',
        credentials: 'omit',
        signal: AbortSignal.any([controller.signal, AbortSignal.timeout(TIMEOUT_MS)]),
      });
      if (!res.ok) throw new Error(`health http ${res.status}`);
      if (cancelledRef.current) return;
      failureCountRef.current = 0;
      setStatus('healthy');
    } catch {
      if (cancelledRef.current) return;
      failureCountRef.current += 1;
      if (failureCountRef.current >= FAILURE_THRESHOLD) {
        setStatus('unhealthy');
      }
    }
  }, [url]);

  useEffect(() => {
    cancelledRef.current = false;
    let timer: ReturnType<typeof setTimeout> | null = null;

    const schedule = () => {
      if (cancelledRef.current) return;
      const interval = statusRef.current === 'unhealthy' ? POLL_INTERVAL_UNHEALTHY_MS : POLL_INTERVAL_HEALTHY_MS;
      timer = setTimeout(async () => {
        await check();
        schedule();
      }, interval);
    };

    // 初回チェック → その後 poll を始める
    void (async () => {
      await check();
      schedule();
    })();

    triggerRecheckRef.current = () => {
      if (timer) clearTimeout(timer);
      void (async () => {
        await check();
        schedule();
      })();
    };

    return () => {
      cancelledRef.current = true;
      fetchControllerRef.current?.abort();
      if (timer) clearTimeout(timer);
    };
  }, [check]);

  const recheck = useCallback(() => {
    triggerRecheckRef.current();
  }, []);

  return { status, recheck };
}
