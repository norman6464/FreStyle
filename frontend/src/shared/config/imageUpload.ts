/**
 * 画像アップロードで受け入れる MIME タイプの単一ソース。
 *
 * プロフィール画像・AI チャット・リッチテキストエディタなどで共通に使う。
 * `<input accept>` と JS 側の絞り込みで同じ定義を参照し、ずれ（バックエンド非互換 MIME の受理）を防ぐ。
 */
export const ACCEPTED_IMAGE_MIME_TYPES = [
  'image/png',
  'image/jpeg',
  'image/gif',
  'image/webp',
] as const;

/** `<input type="file" accept>` に渡す文字列（カンマ区切り）。 */
export const ACCEPTED_IMAGE_ACCEPT_ATTR = ACCEPTED_IMAGE_MIME_TYPES.join(',');

/** file.type が受け入れ対象の画像 MIME かを返す。 */
export function isAcceptedImageMimeType(type: string): boolean {
  return (ACCEPTED_IMAGE_MIME_TYPES as readonly string[]).includes(type);
}

/**
 * 画像アップロードの上限バイト数（10 MiB）。backend の上限と一致させる。
 *
 * サーバー側の判定を待たず、選んだ直後にクライアント側でも早期に弾く（無駄な
 * アップロード開始と、待たされた末の失敗表示を避けるため）。
 */
export const MAX_IMAGE_UPLOAD_BYTES = 10 * 1024 * 1024;
