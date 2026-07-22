/**
 * app 層の Public API。
 *
 * app は FSD の最上位層なので、他の層から import されることは無い。
 * ここはエントリ（index.tsx）から参照する組み立て済みのアプリを公開するためだけに置く。
 * ワイルドカード re-export は使わない（FSD の Public API 規約）。
 */
export { default as App } from './App';
