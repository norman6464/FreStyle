import { createContext, useContext, useMemo, useState } from 'react';

export interface HeaderVisibilityValue {
  /** ヘッダーを隠すべきか。既定 false（表示）。 */
  headerHidden: boolean;
  /** ページ側（スクロール検知など）から表示状態を切り替える。 */
  setHeaderHidden: (hidden: boolean) => void;
}

/**
 * HeaderVisibilityContext はアプリ共通ヘッダーの表示/非表示状態を配る。
 * AppShell が Provider として値を持ち、ページ（例: ノートの本文スクロール）が
 * setHeaderHidden で「下スクロール中は隠す」を実現する。復帰は AppShell 側が行う。
 */
export const HeaderVisibilityContext = createContext<HeaderVisibilityValue>({
  headerHidden: false,
  setHeaderHidden: () => {},
});

/** useHeaderVisibility はヘッダー表示状態の読み書き口。AppShell 配下でのみ意味を持つ。 */
export function useHeaderVisibility(): HeaderVisibilityValue {
  return useContext(HeaderVisibilityContext);
}

/** useHeaderVisibilityState は AppShell が Provider に渡す状態を組み立てる。 */
export function useHeaderVisibilityState(): HeaderVisibilityValue {
  const [headerHidden, setHeaderHidden] = useState(false);
  return useMemo(() => ({ headerHidden, setHeaderHidden }), [headerHidden]);
}
