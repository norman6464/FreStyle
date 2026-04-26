import '@testing-library/jest-dom';

// jsdom は Element.prototype.scrollTo を実装していないため、
// scrollTo を呼び出すコンポーネント（EmojiPicker など）のテストで unhandled error を起こす。
// テスト実行時のみ no-op スタブを差し込む。
if (typeof Element !== 'undefined' && !Element.prototype.scrollTo) {
  Element.prototype.scrollTo = () => {};
}
