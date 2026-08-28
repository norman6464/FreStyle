import { INTERNAL_PAGE_LINK_PATTERN } from './linkSafety';

/**
 * クリックからリンクを開く（本文のリンクは編集中でもクリックで飛ぶ）。
 *
 * ProseMirror の handleClick に置かないのは 2 つの理由から:
 *   - あちらは座標からの位置計算（posAtCoords）を経由するため、レイアウトを持たない
 *     テスト環境（jsdom）では発火せず、検証できない振る舞いになる
 *   - 読み取り専用では ProseMirror がクリックを引き受けず、素の <a> の既定遷移
 *     （全画面リロード）に落ちる。ここで一元化すればどちらのモードも同じ動きになる
 */
export function openClickedLink(
  event: MouseEvent,
  navigateToPage?: (path: string) => void,
): boolean {
  if (!(event.target instanceof Element)) return false;
  const anchor = event.target.closest('a[href]');
  if (!(anchor instanceof HTMLAnchorElement)) return false;
  const href = anchor.getAttribute('href');
  if (!href) return false;

  // 修飾キー（Cmd/Ctrl/Shift）は「新しいタブで」の意思。内部・外部を問わず新しいタブへ。
  const wantsNewTab = event.metaKey || event.ctrlKey || event.shiftKey;
  const pagePath = internalPagePath(href);
  if (pagePath !== null && !wantsNewTab) {
    // アプリ内のページはアプリ内遷移（呼び出し側が router を持つ）。
    // 渡されていなければ素の遷移に落とす（shared は router を知らない）。
    if (navigateToPage) navigateToPage(pagePath);
    else window.location.assign(pagePath);
    return true;
  }
  // anchor.href は相対 href も絶対 URL へ解決済み。
  window.open(anchor.href, '_blank', 'noopener,noreferrer');
  return true;
}

/**
 * 内部ページリンクなら /p/{id} のパスを返す。相対の /p/… に加え、
 * 共有 URL の貼り付けで入る「同一オリジンの絶対 URL」も同じ扱いにする
 * （毎回別タブが開くと、アプリ内の行き来なのに窓が増え続ける）。
 */
export function internalPagePath(href: string): string | null {
  if (INTERNAL_PAGE_LINK_PATTERN.test(href)) return href;
  try {
    const url = new URL(href, window.location.origin);
    if (url.origin === window.location.origin && INTERNAL_PAGE_LINK_PATTERN.test(url.pathname)) {
      return url.pathname;
    }
  } catch {
    // URL として読めない値は外部リンク扱いへ落とす（開く側が絶対 URL で開く）。
  }
  return null;
}
