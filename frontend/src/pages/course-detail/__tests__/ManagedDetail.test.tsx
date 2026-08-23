import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { ToastProvider } from '@/app/providers/ToastProvider';
import type { TeachingMaterial } from '@/entities/course';
import type { RichDocContent } from '@/shared/ui/RichTextEditor';
import ManagedDetail from '../ui/ManagedDetail';
import { useTeachingMaterialEditor } from '../model/useTeachingMaterialEditor';

vi.mock('@/entities/course/api/teachingMaterialRepository', () => ({
  default: { get: vi.fn(), updateDoc: vi.fn() },
}));

const richDoc: RichDocContent = {
  type: 'doc',
  content: [{ type: 'paragraph', content: [{ type: 'text', text: 'リッチ本文' }] }],
};

function material(overrides: Partial<TeachingMaterial> = {}): TeachingMaterial {
  return {
    id: 1,
    companyId: 10,
    courseId: 5,
    createdByUserId: 1,
    title: '章タイトル',
    content: '',
    doc: richDoc,
    revision: 1,
    orderInCourse: 1,
    isPublished: false,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-08T00:00:00Z',
    ...overrides,
  };
}

function Harness({ selected }: { selected: TeachingMaterial }) {
  const editor = useTeachingMaterialEditor({
    selectedId: selected.id,
    selected,
    update: vi.fn().mockResolvedValue(undefined),
  });
  return <ManagedDetail editor={editor} />;
}

function renderDetail(selected: TeachingMaterial) {
  return render(
    <ToastProvider>
      <Harness selected={selected} />
    </ToastProvider>,
  );
}

describe('ManagedDetail のエディタ分岐 (FRESTYLE-339)', () => {
  it('doc がある章は tiptap エディタ（編集可能）とタイトル入力になる', async () => {
    renderDetail(material());
    expect(screen.getByLabelText('教材のタイトル')).toHaveValue('章タイトル');
    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: '教材本文' })).toBeInTheDocument();
    });
    expect(screen.getByText('リッチ本文')).toBeInTheDocument();
    // 編集可能（contenteditable=true）。
    expect(screen.getByRole('textbox', { name: '教材本文' })).toHaveAttribute('contenteditable', 'true');
    // 旧 Markdown エディタの Edit/Preview タブは出ない。
    expect(screen.queryByRole('tab', { name: /edit/i })).not.toBeInTheDocument();
  });

  it('doc 未移行で Markdown content を持つ章は従来の Markdown エディタになる', () => {
    renderDetail(material({ doc: null, content: '# 既存の Markdown 本文' }));
    expect(screen.queryByRole('textbox', { name: '教材本文' })).not.toBeInTheDocument();
    // NoteMarkdownEditor のタイトル入力（aria-label はノート用のまま流用）。
    expect(screen.getByLabelText('ノートのタイトル')).toHaveValue('章タイトル');
  });

  it('「trainee に公開」トグルはどちらのモードでも表示される', () => {
    renderDetail(material());
    expect(screen.getByLabelText('trainee に公開')).not.toBeChecked();
  });
});
