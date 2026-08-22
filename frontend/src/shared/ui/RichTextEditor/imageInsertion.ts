import type { Editor } from '@tiptap/react';
import { isAcceptedImageMimeType } from '@/shared/config/imageUpload';

/** ImageUploader は File を受け取り、表示用 URL を返す関数。 */
export type ImageUploader = (file: File) => Promise<string>;

/**
 * acceptedImageFiles は FileList/配列から、受け入れ可能な MIME の画像ファイルだけを取り出す。
 * ドロップ・貼り付け・ファイル選択のいずれの入力も同じ基準で正規化する。
 */
export function acceptedImageFiles(files: FileList | File[] | null | undefined): File[] {
  return Array.from(files ?? []).filter((file) => isAcceptedImageMimeType(file.type));
}

/**
 * insertUploadedImages は画像ファイル群をアップロードして editor へ挿入する。
 *
 * - 選択順を保つため 1 つずつ順次（await）に処理する（並列だと URL 取得の早い順に並んでしまう）
 * - 各挿入の直前に isAlive() と editor.isDestroyed を確認し、別文書へ切り替え済みなら誤挿入しない
 * - alt にはファイル名を既定で入れる（代替テキストの土台。専用 UI は後続）
 *
 * アップロード失敗は握りつぶす（通知は呼び出し側の uploader 内の方針に委ねる）。
 */
export async function insertUploadedImages(
  editor: Editor,
  files: File[],
  upload: ImageUploader,
  isAlive: () => boolean = () => true,
): Promise<void> {
  for (const file of files) {
    try {
      const url = await upload(file);
      if (!isAlive() || editor.isDestroyed) continue;
      editor.chain().focus().setImage({ src: url, alt: file.name }).run();
    } catch {
      /* アップロード失敗は無視（通知は呼び出し側） */
    }
  }
}
