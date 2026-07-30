/**
 * ビルド時プリレンダー。
 *
 * vite build 後の dist/ を SPA フォールバック付きの静的サーバで配信し、Playwright(既存の
 * devDependency)の headless Chromium で公開ルートを実際に描画してから DOM をスナップショットし、
 * 「中身入りの HTML」を dist/<route>/index.html に書き戻す。これにより検索ボットの初回取得で
 * 本文がクロールできる状態にする(CSR の空シェル対策)。
 *
 * 対象は Phase 1 では公開トップ "/" のみ(dist/index.html を上書き = CloudFront/S3 の
 * デフォルトルートオブジェクトがそのまま返るためインフラ変更不要)。追加ルートは Phase 2 で
 * CloudFront Function を入れてから対象に加える。
 *
 * 使い方: npm run build && node scripts/prerender.mjs
 */
import { createServer } from 'node:http';
import { readFile, writeFile, mkdir, readdir, stat } from 'node:fs/promises';
import { join, extname, dirname, relative, sep } from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromium } from '@playwright/test';

const ROUTES = ['/']; // Phase 1: 公開トップのみ
const HERE = dirname(fileURLToPath(import.meta.url));
const DIST = join(HERE, '..', 'dist');
const READY_SELECTOR = 'html[data-prerender-ready]';
const NAV_TIMEOUT_MS = 20000;

const CONTENT_TYPES = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.mjs': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.ico': 'image/x-icon',
  '.webp': 'image/webp',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.ttf': 'font/ttf',
  '.txt': 'text/plain; charset=utf-8',
  '.xml': 'application/xml; charset=utf-8',
  '.webmanifest': 'application/manifest+json',
};

async function fileExists(p) {
  try {
    return (await stat(p)).isFile();
  } catch {
    return false;
  }
}

/**
 * dist/ 配下の全ファイルを列挙し「URL パス → 絶対パス」の対応表を作る。
 * リクエスト由来の文字列は fs API に渡さず、この表の値(readdir 由来)だけを使うことで
 * path traversal を構造的に不可能にする(CodeQL js/path-injection 対策)。
 */
async function buildFileMap() {
  const entries = await readdir(DIST, { recursive: true, withFileTypes: true });
  const map = new Map();
  for (const e of entries) {
    if (!e.isFile()) continue;
    const abs = join(e.parentPath ?? e.path, e.name);
    map.set('/' + relative(DIST, abs).split(sep).join('/'), abs);
  }
  return map;
}

/** dist/ を配信する静的サーバ。対応表にないパスは index.html にフォールバック(SPA)。 */
async function startServer() {
  const files = await buildFileMap();
  const indexPath = join(DIST, 'index.html');
  const server = createServer(async (req, res) => {
    try {
      // ヘルスチェックはローカルでは常に OK を返し、メンテナンス表示への誤遷移を防ぐ。
      if (req.url && req.url.startsWith('/api/v2/health')) {
        res.writeHead(200, { 'content-type': 'application/json' });
        res.end('{"status":"ok"}');
        return;
      }
      const urlPath = decodeURIComponent((req.url || '/').split('?')[0]);
      const filePath = (urlPath !== '/' && files.get(urlPath)) || indexPath;
      const body = await readFile(filePath);
      res.writeHead(200, { 'content-type': CONTENT_TYPES[extname(filePath)] || 'application/octet-stream' });
      res.end(body);
    } catch {
      // 例外の内容(スタックトレース等)はレスポンスに含めない(CodeQL stack-trace-exposure 対策)。
      res.writeHead(500, { 'content-type': 'text/plain; charset=utf-8' });
      res.end('internal error');
    }
  });
  return new Promise((resolve) => {
    server.listen(0, '127.0.0.1', () => resolve({ server, port: server.address().port }));
  });
}

async function main() {
  if (!(await fileExists(join(DIST, 'index.html')))) {
    throw new Error('dist/index.html が無い。先に `npm run build` を実行してください。');
  }

  const { server, port } = await startServer();
  const base = `http://127.0.0.1:${port}`;
  const browser = await chromium.launch();
  const snapshots = [];

  try {
    for (const route of ROUTES) {
      const page = await browser.newPage();
      await page.goto(`${base}${route}`, { waitUntil: 'load', timeout: NAV_TIMEOUT_MS });
      // LP がマウント完了で付ける目印を待ってからスナップショット(ローディング途中を撮らない)。
      await page.waitForSelector(READY_SELECTOR, { timeout: NAV_TIMEOUT_MS });
      const html = await page.content();
      await page.close();

      // 回帰ガード: 空シェルや本文欠落なら fail させ、空シェルへの逆戻りを防ぐ。
      if (/<div id="root">\s*<\/div>/.test(html)) {
        throw new Error(`prerender: ${route} が空シェルのまま(#root が空)。`);
      }
      if (!html.includes('新卒ITエンジニア向け研修プラットフォーム')) {
        throw new Error(`prerender: ${route} に想定コンテンツ(ヒーロー文言)が含まれていない。`);
      }
      snapshots.push({ route, html });
      console.log(`prerendered: ${route} (${html.length} bytes)`);
    }
  } finally {
    await browser.close();
    server.close();
  }

  // すべて撮り終えてから書き出す(途中で index.html を上書きしてフォールバックが変わるのを防ぐ)。
  for (const { route, html } of snapshots) {
    const outPath = route === '/' ? join(DIST, 'index.html') : join(DIST, route, 'index.html');
    await mkdir(dirname(outPath), { recursive: true });
    await writeFile(outPath, html, 'utf-8');
    console.log(`wrote: ${outPath.replace(DIST, 'dist')}`);
  }
  console.log(`\nprerender 完了: ${snapshots.length} ルート`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
