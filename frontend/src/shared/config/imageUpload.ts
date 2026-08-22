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
