import { useEffect, useState } from 'react';

/**
 * 「今開いているスペース」を、ページ本体からヘッダーへ伝える共有値。
 *
 * ヘッダー（widgets/app-shell）は個々のページの解決結果（workspaceSlug・spaceId）を
 * props で受け取れない（AppShell がルーティングの外側で描画するため）。props で結ぶには
 * 両者が遠すぎるので、entities/kb/model/kbTreeEvents.ts と同じ理由で、entities 層の
 * 小さな共有値で結ぶ。
 *
 * 書き手（値を確定させたページ側）: KbPage・useKbSpaceEntry（概要・すべてのページ・
 * お気に入り・メンバーの 4 画面が共有）・useBacklogSpace（バックログ）。
 * 読み手: widgets/app-shell の「スペース ▾」「作成」。
 *
 * 値はページ遷移のたびに書き手が上書きするので、古いスペースを指したまま
 * 残り続けることはない（KB 以外の画面では読み手側が使わないだけで、値自体は
 * 最後に開いていたスペースを指したままになるが、表示に使われないため実害は無い）。
 */
export interface CurrentKbSpace {
  workspaceSlug: string;
  spaceId: string;
}

let current: CurrentKbSpace | null = null;
const listeners = new Set<(value: CurrentKbSpace | null) => void>();

/** setCurrentKbSpace は「今いるスペース」を確定させたページ側が呼ぶ。 */
export function setCurrentKbSpace(value: CurrentKbSpace | null): void {
  current = value;
  for (const listener of [...listeners]) listener(value);
}

/** useCurrentKbSpace は「今いるスペース」を読むだけの側（ヘッダー）が使う。 */
export function useCurrentKbSpace(): CurrentKbSpace | null {
  const [value, setValue] = useState(current);
  useEffect(() => {
    setValue(current);
    listeners.add(setValue);
    return () => {
      listeners.delete(setValue);
    };
  }, []);
  return value;
}
