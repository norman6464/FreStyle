// 起動時に localStorage の theme を反映する。
// FOUC を防ぐため <script src> で head 同期実行する想定。
// inline script を外出ししたのは Content-Security-Policy: script-src 'self' を有効にするため。
try {
  if (localStorage.getItem('theme') === 'light') {
    document.documentElement.classList.add('light');
  }
} catch (e) {
  // localStorage 不可な環境ではデフォルトテーマで起動。
}
