import '@testing-library/jest-dom';
import { vi, beforeEach } from 'vitest';

// jsdom は Element.prototype.scrollTo を実装していないため、
// scrollTo を呼び出すコンポーネント（EmojiPicker など）のテストで unhandled error を起こす。
// テスト実行時のみ no-op スタブを差し込む。
if (typeof Element !== 'undefined' && !Element.prototype.scrollTo) {
  Element.prototype.scrollTo = () => {};
}

// CI（Node 20+）では axios のテスト未モック分が Node の fetch (undici) に到達し、
// 「InvalidArgumentError: invalid onError method」を unhandled error として吐いて
// すべての Vitest run を exit 1 で落とす。
// 個別テストの mock 漏れを許容するため、グローバル fetch をデフォルトで no-op に差し替える。
// 各テストが必要に応じて vi.fn().mockResolvedValue(...) などで上書きできる。
beforeEach(() => {
  vi.stubGlobal(
    'fetch',
    vi.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        json: async () => ({}),
        text: async () => '',
        headers: new Headers(),
      } as Response),
    ),
  );
});
