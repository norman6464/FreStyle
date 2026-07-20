import { RefObject, useEffect, useLayoutEffect, useRef } from 'react';

/**
 * useFadeOnVisible — `.fade-seg` のフェードを「mount 時」ではなく「初めて画面に入った時」に走らせる
 * （FRESTYLE-153）。
 *
 * `.fade-seg` の CSS アニメーションは DOM に mount した瞬間に再生される。ストリーミング中の自動
 * スクロール追従をやめた（FRESTYLE-149）ため本文は画面外へ伸びていき、画面外で mount した
 * チャンクはそこでフェードを再生し終える。ユーザーが自分でスクロールして到達した頃には
 * 不透明になっていて「ただ文字が順に出るだけ」に見えていた。
 *
 * そこで、新しく現れたチャンクのうち **画面外のものだけ** アニメーションを保留し
 * （`data-fade="defer"`）、IntersectionObserver で画面に入った時点で解除して再生する。
 *
 * 安全側の設計:
 * - `IntersectionObserver` が無い環境（jsdom / 旧ブラウザ）では何も保留しない = 従来どおり
 *   mount 時フェード。テキストが見えなくなることはない。
 * - 保留するのは画面外の要素だけなので、見えている本文が隠れることはない。
 * - 応答完了後はフェードプラグイン自体が外れて `.fade-seg` が DOM から消えるため、
 *   保留されたまま不可視で固定される要素は残らない。
 */

/** 画面外判定に使う余白。少し先読みしておき、到達と同時にフェードが始まるようにする。 */
const ROOT_MARGIN_PX = 64;

export function useFadeOnVisible(
  containerRef: RefObject<HTMLElement | null>,
  /** ストリーミング中など、フェード対象の span が生成されうる間だけ true。 */
  enabled: boolean,
  /** 本文が変わるたびに走査し直すためのキー（表示中テキストなど）。 */
  contentKey: unknown,
) {
  const observerRef = useRef<IntersectionObserver | null>(null);

  // 走査は paint 前に行う。mount 直後にアニメーションが 1 フレームでも見えてしまうと
  // 「画面外で再生済み」を防ぐ意味がなくなるため useLayoutEffect を使う。
  useLayoutEffect(() => {
    const root = containerRef.current;
    if (!root || !enabled) return;
    if (typeof IntersectionObserver === 'undefined') return; // 保留しない（従来動作）

    if (!observerRef.current) {
      observerRef.current = new IntersectionObserver(
        (entries) => {
          for (const entry of entries) {
            if (!entry.isIntersecting) continue;
            const el = entry.target as HTMLElement;
            // defer を外すと .fade-seg の animation が改めて走る。
            el.dataset.fade = 'now';
            observerRef.current?.unobserve(el);
          }
        },
        { rootMargin: `${ROOT_MARGIN_PX}px 0px` },
      );
    }
    const observer = observerRef.current;

    // まだ判定していない span だけを対象にする（既表示分は再生し直さない）。
    const segments = root.querySelectorAll<HTMLElement>('.fade-seg:not([data-fade])');
    for (const el of segments) {
      const rect = el.getBoundingClientRect();
      const visible =
        rect.top < window.innerHeight + ROOT_MARGIN_PX && rect.bottom > -ROOT_MARGIN_PX;
      if (visible) {
        // 見えているので今フェードさせる（CSS の mount アニメがそのまま走る）。
        el.dataset.fade = 'now';
      } else {
        el.dataset.fade = 'defer';
        observer.observe(el);
      }
    }
  }, [containerRef, enabled, contentKey]);

  // unmount / 無効化で監視を止める。
  useEffect(() => {
    return () => {
      observerRef.current?.disconnect();
      observerRef.current = null;
    };
  }, []);
}
