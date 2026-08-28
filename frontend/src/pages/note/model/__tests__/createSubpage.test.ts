import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createSubpage, type SubpageEditor } from '../createSubpage';

const hoisted = vi.hoisted(() => ({
  createPage: vi.fn(),
  emit: vi.fn(),
}));

vi.mock('@/entities/note', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/entities/note')>();
  return {
    ...actual,
    NoteRepository: { createPage: hoisted.createPage },
    emitNoteTreeEvent: hoisted.emit,
  };
});

const resolved = {
  workspaceSlug: 'w-3f2a9c',
  page: {
    id: 'parent-1',
    spaceId: 'space-1',
    title: '親ページ',
    createdByUserId: 1,
    createdAt: '2026-08-01T00:00:00Z',
    updatedAt: '2026-08-01T00:00:00Z',
  },
  doc: { type: 'doc', content: [] },
  canEdit: true,
};

const child = {
  id: 'child-1',
  spaceId: 'space-1',
  parentId: 'parent-1',
  title: '無題',
  createdByUserId: 1,
  createdAt: '2026-08-28T00:00:00Z',
  updatedAt: '2026-08-28T00:00:00Z',
};

function fakeEditor() {
  const run = vi.fn();
  const insertContent = vi.fn(() => ({ run }));
  const editor: SubpageEditor = { chain: () => ({ focus: () => ({ insertContent }) }) };
  return { editor, insertContent, run };
}

beforeEach(() => {
  vi.clearAllMocks();
  hoisted.createPage.mockResolvedValue(child);
});

describe('createSubpage', () => {
  it('現在のページの子として作り、本文にページ参照を挿し、開く先の URL を返す', async () => {
    const { editor, insertContent, run } = fakeEditor();

    const path = await createSubpage(editor, resolved);

    expect(hoisted.createPage).toHaveBeenCalledWith('w-3f2a9c', 'space-1', {
      title: '無題',
      parentId: 'parent-1',
    });
    // 参照は pageId を正として持ち、title は初回表示のための写し（正本はサーバーが解決）。
    expect(insertContent).toHaveBeenCalledWith({
      type: 'pageRef',
      attrs: { pageId: 'child-1', title: '無題' },
    });
    expect(run).toHaveBeenCalled();
    expect(path).toBe('/p/child-1');
  });

  it('木へ確定後のページで知らせる（サイドバーが追従できる）', async () => {
    const { editor } = fakeEditor();

    await createSubpage(editor, resolved);

    expect(hoisted.emit).toHaveBeenCalledWith({ type: 'page-created', page: child });
  });

  it('作成に失敗したら参照を挿さず、木にも知らせず、失敗を投げる', async () => {
    hoisted.createPage.mockRejectedValue(new Error('403'));
    const { editor, insertContent } = fakeEditor();

    await expect(createSubpage(editor, resolved)).rejects.toThrow();
    expect(insertContent).not.toHaveBeenCalled();
    expect(hoisted.emit).not.toHaveBeenCalled();
  });
});
