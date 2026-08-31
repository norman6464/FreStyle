/**
 * dist/app.html を作る。
 *
 * CloudFront は存在しないパス（/notes/... など SPA 側のルート）へのリクエストに対し、
 * エラー応答としてこのファイルを返す。index.html ではなく別ファイルにしているのは、
 * かつて index.html に公開トップの HTML を焼き込んでいたため。中身入りの index.html を
 * フォールバックに使うと、どの画面を開いても一瞬トップが映ってから目的の画面に切り替わる
 * ちらつきが出ていた（FRESTYLE-230）。
 *
 * 公開トップとその焼き込み（prerender）は廃止したので index.html も空シェルだが、
 * app.html を返す CloudFront の設定はそのまま残っている。生成をやめるとフォールバックが
 * 参照するファイルが消えて本番が壊れるため、build の一部として作り続ける。
 */
import { copyFile, readFile, access } from 'node:fs/promises';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const HERE = dirname(fileURLToPath(import.meta.url));
const DIST = join(HERE, '..', 'dist');
const INDEX = join(DIST, 'index.html');
const SHELL = join(DIST, 'app.html');

async function main() {
  try {
    await access(INDEX);
  } catch {
    throw new Error('dist/index.html が無い。先に `vite build` を実行してください。');
  }

  const html = await readFile(INDEX, 'utf-8');
  // 中身が焼き込まれた HTML をフォールバックに使うと上記のちらつきに戻るので、
  // 空シェルであることを確かめてから複製する。
  if (!/<div id="root">\s*<\/div>/.test(html)) {
    throw new Error('dist/index.html が空シェルでない。app.html はビルド直後の index.html から作る。');
  }

  await copyFile(INDEX, SHELL);
  console.log('wrote: dist/app.html（SPA フォールバック用・空シェル）');
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
