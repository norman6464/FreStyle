import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
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
  // 本文のリンクをアプリ内遷移で開くため router を使う（本番も Router 配下）。
  return render(
    <MemoryRouter>
      <ToastProvider>
        <Harness selected={selected} />
      </ToastProvider>
    </MemoryRouter>,
  );
}

describe('ManagedDetail の tiptap 編集 (FRESTYLE-347)', () => {
  it('doc がある章は tiptap エディタ（編集可能）とタイトル入力になる', async () => {
    renderDetail(material());
    expect(screen.getByLabelText('教材のタイトル')).toHaveValue('章タイトル');
    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: '教材本文' })).toBeInTheDocument();
    });
    expect(screen.getByText('リッチ本文')).toBeInTheDocument();
    // 編集可能（contenteditable=true）。
    expect(screen.getByRole('textbox', { name: '教材本文' })).toHaveAttribute('contenteditable', 'true');
  });

  it('doc が null の章（本文未保存の新規章）も空 doc の tiptap エディタになる', async () => {
    renderDetail(material({ doc: null }));
    expect(screen.getByLabelText('教材のタイトル')).toHaveValue('章タイトル');
    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: '教材本文' })).toBeInTheDocument();
    });
    expect(screen.getByRole('textbox', { name: '教材本文' })).toHaveAttribute('contenteditable', 'true');
  });

  it('「trainee に公開」トグルが表示される', () => {
    renderDetail(material());
    expect(screen.getByLabelText('trainee に公開')).not.toBeChecked();
  });
});
