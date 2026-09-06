const STORAGE_KEY = 'frestyle.kb.lastVisitedPageId';

/**
 * 直近に開いた(閲覧・編集した)ページの ID を覚えておくためだけの localStorage 越しの薄い層。
 *
 * ヘッダーの「ナレッジ」ボタンや素の /kb から、前回の続きへ自動で戻るために使う
 * (pages/kb/model/resolveEntryPage.ts)。消えても機能が壊れるわけではない
 * (単に最初に見つかったページへ戻るだけ)ので、書き込み・読み出しとも例外は握り潰す
 * (プライベートブラウズ等で localStorage が使えない環境でも画面を落とさない)。
 */
export function rememberVisitedPage(pageId: string): void {
  try {
    localStorage.setItem(STORAGE_KEY, pageId);
  } catch {
    // 書けなくても致命的ではない。
  }
}

export function getLastVisitedPageId(): string | null {
  try {
    return localStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
}

/**
 * 覚えていたページが自分と同じ ID のときだけ忘れる。
 *
 * ページを開けなかった(削除済み等)ときに呼ぶ。無条件に消すと、それとは無関係な
 * ページを直リンクで開いて失敗しただけの場合にも「前回の続き」を失ってしまう。
 */
export function forgetVisitedPageIfMatches(pageId: string): void {
  if (getLastVisitedPageId() === pageId) {
    try {
      localStorage.removeItem(STORAGE_KEY);
    } catch {
      // 消せなくても致命的ではない。
    }
  }
}
