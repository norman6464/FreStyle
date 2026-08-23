import { ImageUploadRepository } from '@/entities/user';
import { RichTextEditor } from '@/shared/ui/RichTextEditor';
import type { useTeachingMaterialEditor } from '../model/useTeachingMaterialEditor';

/**
 * 管理ロール(company_admin / super_admin)向けの教材編集ビュー。
 * リッチ本文（doc）を tiptap エディタで編集する（doc が null の新規章は空 doc から始まる）。
 * 上部に「trainee に公開」トグルを置く。
 */
export default function ManagedDetail({
  editor,
}: {
  editor: ReturnType<typeof useTeachingMaterialEditor>;
}) {
  return (
    <div className="flex-1 flex flex-col min-h-0">
      <div className="px-6 pt-4 pb-2 flex items-center justify-end gap-2">
        <label className="flex items-center gap-2 text-xs text-[var(--color-text-secondary)] cursor-pointer">
          <input
            type="checkbox"
            checked={editor.editIsPublished}
            onChange={(event) => editor.handleIsPublishedChange(event.target.checked)}
            className="rounded border-surface-3"
          />
          trainee に公開
        </label>
      </div>
      <div className="flex-1 min-h-0">
        <div className="h-full overflow-y-auto">
          <div className="mx-auto w-full max-w-3xl px-6 py-6">
            <input
              type="text"
              value={editor.editTitle}
              onChange={(e) => editor.handleTitleChange(e.target.value)}
              placeholder="無題の教材"
              aria-label="教材のタイトル"
              className="text-2xl font-bold text-[var(--color-text-primary)] bg-transparent border-none outline-none w-full mb-4 placeholder:text-[var(--color-text-faint)]"
            />
            <RichTextEditor
              className="course-doc"
              value={editor.editDoc}
              onChange={editor.handleDocChange}
              saveStatus={editor.saveStatus}
              ariaLabel="教材本文"
              placeholder="本文を入力…（'/' でコマンド）"
              onImageUpload={(file) => ImageUploadRepository.upload(file)}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
