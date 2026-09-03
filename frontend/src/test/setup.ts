import '@testing-library/jest-dom';
import axios from 'axios';
import { vi } from 'vitest';

// 認可の設定の既定を**明示する**。
//
// vitest は Vite の仕組みで `frontend/.env` を暗黙に読む。手元には .env があり
// CI には無いので、明示しないと同じテストが手元と CI で違う設定で走る
// （実際、手元では一部だけ入っていて「設定なし」と判定されていた）。
//
// 既定を「揃っている」側に置くのは、ほとんどのテストが通常の画面を見たいため。
// 設定が欠けた状態を見たいテストは、そのテストの中で stubEnv して外す。
vi.stubEnv('VITE_OIDC_AUTHORIZE_URI', 'https://issuer.test/oauth/v2/authorize');
vi.stubEnv('VITE_OIDC_CLIENT_ID', 'test-client-id');
vi.stubEnv('VITE_OIDC_REDIRECT_URI', 'http://localhost:3000/login/callback');
vi.stubEnv('VITE_OIDC_SCOPE', 'openid profile email offline_access');

// jsdom は Element.prototype.scrollTo を実装していないため、
// scrollTo を呼び出すコンポーネント（EmojiPicker など）のテストで unhandled error を起こす。
// テスト実行時のみ no-op スタブを差し込む。
if (typeof Element !== 'undefined' && !Element.prototype.scrollTo) {
  Element.prototype.scrollTo = () => {};
}

// jsdom は Range/Element の getClientRects（と Range.getBoundingClientRect）を実装せず、
// ProseMirror（tiptap）の scrollToSelection → coordsAtPos が非同期の unhandled error を起こす
// （テスト自体は全部 pass するのに Vitest が exit 1 になる flaky の原因）。
// 位置はテストで検証しないため、ゼロ矩形で埋める。
const zeroRect = {
  x: 0, y: 0, top: 0, left: 0, right: 0, bottom: 0, width: 0, height: 0,
  toJSON() { return this; },
} as DOMRect;
if (typeof Range !== 'undefined' && !Range.prototype.getClientRects) {
  Range.prototype.getClientRects = () => [zeroRect] as unknown as DOMRectList;
  Range.prototype.getBoundingClientRect = () => zeroRect;
}
if (typeof Element !== 'undefined' && !Element.prototype.getClientRects) {
  Element.prototype.getClientRects = () => [zeroRect] as unknown as DOMRectList;
}

// CI (Node 20+) では axios のテスト未モック分が Node の http/undici に到達し、
// 「InvalidArgumentError: invalid onError method」を unhandled error として吐いて
// すべての Vitest run を exit 1 で落としていた。
// axios のデフォルト adapter を no-op スタブに差し替え、未モックのリクエストが実 HTTP に
// 到達しないようにする。各テストでは vi.mock('../../repositories/...') 等で従来通り
// リポジトリ側をモックすればよく、本スタブは「mock 漏れの保険」として働く。
const stubAdapter: typeof axios.defaults.adapter = (config) =>
  Promise.resolve({
    data: {},
    status: 200,
    statusText: 'OK',
    headers: {},
    config: config as never,
    request: {},
  });

axios.defaults.adapter = stubAdapter;
