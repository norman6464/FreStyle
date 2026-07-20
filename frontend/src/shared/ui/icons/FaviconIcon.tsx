/**
 * FaviconIcon — `/favicon.svg` を `EmptyState` の icon prop と同じ型で描画する。
 *
 * `EmptyState` 等は `icon: ComponentType<{ className?: string }>` を要求するため、
 * `<img>` をラップしてアイコンコンポーネント互換にする。
 */
export default function FaviconIcon({ className }: { className?: string }) {
  return <img src="/favicon.svg" alt="" aria-hidden="true" className={className} />;
}
