import { describe, it, expect, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router-dom';
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

const PAGE_UUID = '01a045ef-35de-7e9d-b637-84a5eb6fad77';

/** ノートのページを指すリンクを含む本文（教材から参照するのは普通にある）。 */
const docWithNoteLink: RichDocContent = {
  type: 'doc',
  content: [
    {
      type: 'paragraph',
      content: [
        {
          type: 'text',
          text: '補足のページ',
          marks: [{ type: 'link', attrs: { href: `/p/${PAGE_UUID}` } }],
        },
      ],
    },
  ],
};

/** いまの URL を出すだけの部品（遷移したかを見る）。 */
function LocationProbe() {
  const location = useLocation();
  return <span data-testid="path">{location.pathname}</span>;
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
    <MemoryRouter initialEntries={['/courses/5']}>
      <ToastProvider>
        <LocationProbe />
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

describe('教材本文のノートへのリンク', () => {
  it('クリックすると全画面リロードではなくアプリ内で遷移する', async () => {
    renderDetail(material({ doc: docWithNoteLink }));
    const link = await screen.findByRole('link', { name: '補足のページ' });
    expect(screen.getByTestId('path')).toHaveTextContent('/courses/5');

    fireEvent.click(link);

    await waitFor(() =>
      expect(screen.getByTestId('path')).toHaveTextContent(`/p/${PAGE_UUID}`),
    );
  });

});
