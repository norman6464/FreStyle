import { lazy, type ComponentType } from 'react';

// React の lazy はジェネリックを ComponentType<any> 相当で受ける必要がある
// （props を持つコンポーネントもラップ可能にするため）。型パラメータ T の
// 制約のみ any を許容し、ファイル全体の disable は副作用が広いので避ける。

/**
 * lazy() のラッパ。dynamic import が失敗したとき（典型的にはデプロイ後に
 * フロントの bundle hash が変わって古い chunk が S3 から消えているケース）に
 * 1 回だけ自動でフルリロードして、新しい index.html を取り直す。
 *
 * 症状:
 *   - 旧 index.html を SPA としてロードしているクライアントが、ナビゲーション時に
 *     `import('./pages/NotesPage')` 等を解決しようとして
 *     `Failed to fetch dynamically imported module: .../assets/NotesPage-XXXX.js` を出す
 *   - S3/CloudFront は missing object に対して index.html を返す設定なので、
 *     ブラウザは「JS module を期待したのに text/html が来た」とエラーする
 *
 * 対策:
 *   - sessionStorage の `lazy-reload:<key>` フラグで「直前 1 回 reload した」かを判定し、
 *     未 reload なら window.location.reload() で復旧
 *   - 既に 1 回 reload 済（フラグあり）なら catch せず元のエラーを投げる
 *     （= 真のバグなら無限 reload に陥らない）
 *
 * key を渡すのは複数の lazyWithReload が同時にエラーを起こしたときに、
 * sessionStorage の key を分けて重複 reload を避けるため。
 */
export function lazyWithReload<T extends ComponentType<any>>(
  factory: () => Promise<{ default: T }>,
  key: string
): ReturnType<typeof lazy<T>> {
  return lazy<T>(async () => {
    try {
      return await factory();
    } catch (err) {
      const isChunkLoadError =
        err instanceof Error &&
        (/Failed to fetch dynamically imported module/i.test(err.message) ||
          /Loading chunk \d+ failed/i.test(err.message) ||
          /Importing a module script failed/i.test(err.message));

      if (!isChunkLoadError) {
        throw err;
      }

      const storageKey = `lazy-reload:${key}`;
      const alreadyReloaded =
        typeof window !== 'undefined' && window.sessionStorage?.getItem(storageKey) === '1';

      if (alreadyReloaded) {
        // 既に reload 済 → 真のバグ。フォールスルーして ErrorBoundary に表示させる。
        throw err;
      }

      try {
        window.sessionStorage?.setItem(storageKey, '1');
      } catch {
        // sessionStorage が使えない環境（Safari private mode 等）は無視。
      }
      window.location.reload();
      // reload は async に走るので、ここで永遠に解決しない promise を返して
      // 上位 (Suspense) を一時的にハングさせる（その間に reload が走る）。
      return new Promise<never>(() => {});
    }
  });
}

/**
 * ナビゲーション成功（成功した URL に到達）時に呼んで、reload フラグを掃除する。
 * 失敗した chunk が解決された証拠なので、次回また失敗したら再度 reload を許可する。
 */
export function clearLazyReloadFlags(): void {
  if (typeof window === 'undefined' || !window.sessionStorage) return;
  for (let i = 0; i < window.sessionStorage.length; i++) {
    const k = window.sessionStorage.key(i);
    if (k && k.startsWith('lazy-reload:')) {
      window.sessionStorage.removeItem(k);
      i--;
    }
  }
}
