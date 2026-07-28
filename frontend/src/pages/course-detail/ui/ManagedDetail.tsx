import { NoteMarkdownEditor } from '@/entities/note';
import { ImageUploadRepository } from '@/entities/user';
import type { useTeachingMaterialEditor } from '../model/useTeachingMaterialEditor';

/**
 * 管理ロール(company_admin / super_admin)向けの教材編集ビュー。
 * NoteMarkdownEditor(Edit/Preview タブ)を流用し、上部に「trainee に公開」トグルを置く。
 */
export default function ManagedDetail({
  editor,
}: {
  editor: ReturnType<typeof useTeachingMaterialEditor>;
}) {
  return (
    <div className="flex-1 flex flex-col min-h-0">
      <div className="px-6 pt-4 pb-2 flex items-center justify-end gap-2 border-b border-surface-3">
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
        <NoteMarkdownEditor
          title={editor.editTitle}
          content={editor.editContent}
          saveStatus={editor.saveStatus}
          onTitleChange={editor.handleTitleChange}
          onContentChange={editor.handleContentChange}
          onImageUpload={(file) => ImageUploadRepository.upload(file)}
        />
      </div>
    </div>
  );
}
