/**
 * 画像アップロードで許可する MIME の一覧と上限サイズ。
 *
 * backend の許可リスト（usecase の allowedNoteImageContentTypes / AllowedAttachmentContentTypes）と
 * 一致させる契約値。ここがずれるとファイル選択できたのに 415 で弾かれる状態になる。
 *
 * image/svg+xml は意図的に含めない（SVG は script を埋め込めるため保存型 XSS の媒体になる）。
 * image/heic / image/heif も Safari 以外が表示できないため含めない。
 */
export const ACCEPTED_IMAGE_MIME_TYPES = [
  'image/png',
  'image/jpeg',
  'image/gif',
  'image/webp',
] as const;

/** input[type=file] の accept 属性に渡す文字列。 */
export const ACCEPTED_IMAGE_ACCEPT_ATTR = ACCEPTED_IMAGE_MIME_TYPES.join(',');

/** 画像 1 枚あたりの上限（backend の maxNoteImageBytes と同値の 5MB）。 */
export const MAX_IMAGE_UPLOAD_BYTES = 5 * 1024 * 1024;

/**
 * isAcceptedImageMimeType は許可 MIME かどうかを判定する。
 * backend が互換で受け入れる非標準の image/jpg も通す（ブラウザは image/jpeg を送るため通常は発生しない）。
 */
export function isAcceptedImageMimeType(mimeType: string): boolean {
  return mimeType === 'image/jpg' || (ACCEPTED_IMAGE_MIME_TYPES as readonly string[]).includes(mimeType);
}
