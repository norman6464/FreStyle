import { renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useKbPageTemplates } from '../useKbPageTemplates';

const hoisted = vi.hoisted(() => ({
  listPageTemplates: vi.fn(),
  deletePageTemplate: vi.fn(),
  createPageFromTemplate: vi.fn(),
}));

vi.mock('@/entities/kb', () => ({
  KbRepository: {
    listPageTemplates: hoisted.listPageTemplates,
    deletePageTemplate: hoisted.deletePageTemplate,
    createPageFromTemplate: hoisted.createPageFromTemplate,
  },
}));

const SLUG = 'w-3f2a9c';
const SPACE = 's-1';

const template = (id: string, name: string) => ({
  id,
  name,
  createdAt: '2026-09-01T00:00:00Z',
});

beforeEach(() => {
  vi.clearAllMocks();
  hoisted.listPageTemplates.mockResolvedValue([template('t-2', '週次レポート'), template('t-1', '議事録')]);
  hoisted.deletePageTemplate.mockResolvedValue(undefined);
  hoisted.createPageFromTemplate.mockResolvedValue({ id: 'p-new', spaceId: SPACE, title: '議事録' });
});

describe('useKbPageTemplates の一覧取得（open ゲート）', () => {
  it('open=false の間は取りに行かない', () => {
    renderHook(() => useKbPageTemplates(SLUG, SPACE, false));

    expect(hoisted.listPageTemplates).not.toHaveBeenCalled();
  });

  it('workspaceSlug/spaceId のどちらかが欠けていれば open=true でも取りに行かない', () => {
    renderHook(() => useKbPageTemplates(undefined, SPACE, true));
    renderHook(() => useKbPageTemplates(SLUG, undefined, true));

    expect(hoisted.listPageTemplates).not.toHaveBeenCalled();
  });

  it('open=true になると取得する', async () => {
    const { result } = renderHook(() => useKbPageTemplates(SLUG, SPACE, true));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(hoisted.listPageTemplates).toHaveBeenCalledWith(SLUG, SPACE);
    expect(result.current.templates.map((t) => t.id)).toEqual(['t-2', 't-1']);
  });

  it('閉じると state を畳む（次に開いたとき前のスペースのテンプレートを出さない）', async () => {
    const { result, rerender } = renderHook(
      ({ open }: { open: boolean }) => useKbPageTemplates(SLUG, SPACE, open),
      { initialProps: { open: true } },
    );
    await waitFor(() => expect(result.current.templates).toHaveLength(2));

    rerender({ open: false });

    expect(result.current.templates).toHaveLength(0);
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it('読み込みに失敗したら理由を出し、古い一覧を残さない', async () => {
    hoisted.listPageTemplates.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useKbPageTemplates(SLUG, SPACE, true));

    await waitFor(() => expect(result.current.error).not.toBeNull());
    expect(result.current.templates).toHaveLength(0);
  });

  it('スペースを切り替えると新しい宛先を取得し直す', async () => {
    const { result, rerender } = renderHook(
      ({ spaceId }: { spaceId: string }) => useKbPageTemplates(SLUG, spaceId, true),
      { initialProps: { spaceId: SPACE } },
    );
    await waitFor(() => expect(result.current.templates).toHaveLength(2));

    hoisted.listPageTemplates.mockResolvedValue([template('t-9', '別スペース用')]);
    rerender({ spaceId: 's-2' });

    await waitFor(() => expect(result.current.templates.map((t) => t.id)).toEqual(['t-9']));
    expect(hoisted.listPageTemplates).toHaveBeenLastCalledWith(SLUG, 's-2');
  });
});

describe('useKbPageTemplates.deleteTemplate', () => {
  it('成功したら一覧から取り除く', async () => {
    const { result } = renderHook(() => useKbPageTemplates(SLUG, SPACE, true));
    await waitFor(() => expect(result.current.templates).toHaveLength(2));

    await result.current.deleteTemplate('t-1');

    expect(hoisted.deletePageTemplate).toHaveBeenCalledWith(SLUG, 't-1');
    await waitFor(() => expect(result.current.templates.map((t) => t.id)).toEqual(['t-2']));
  });

  it('失敗は握り潰さず投げる（一覧は変わらない）', async () => {
    hoisted.deletePageTemplate.mockRejectedValue(new Error('forbidden'));
    const { result } = renderHook(() => useKbPageTemplates(SLUG, SPACE, true));
    await waitFor(() => expect(result.current.templates).toHaveLength(2));

    await expect(result.current.deleteTemplate('t-1')).rejects.toThrow();
    expect(result.current.templates).toHaveLength(2);
  });
});

describe('useKbPageTemplates.createPageFromTemplate', () => {
  it('workspaceSlug・spaceId を添えてリポジトリを呼ぶ（open に依存しない）', async () => {
    const { result } = renderHook(() => useKbPageTemplates(SLUG, SPACE, false));

    const page = await result.current.createPageFromTemplate({ templateId: 't-1', title: '議事録' });

    expect(hoisted.createPageFromTemplate).toHaveBeenCalledWith(SLUG, SPACE, {
      templateId: 't-1',
      title: '議事録',
    });
    expect(page).toEqual({ id: 'p-new', spaceId: SPACE, title: '議事録' });
  });

  it('parentId を渡すとそのまま送る', async () => {
    const { result } = renderHook(() => useKbPageTemplates(SLUG, SPACE, false));

    await result.current.createPageFromTemplate({ templateId: 't-1', parentId: 'p-1', title: '議事録' });

    expect(hoisted.createPageFromTemplate).toHaveBeenCalledWith(SLUG, SPACE, {
      templateId: 't-1',
      parentId: 'p-1',
      title: '議事録',
    });
  });

  it('失敗は握り潰さず投げる', async () => {
    hoisted.createPageFromTemplate.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useKbPageTemplates(SLUG, SPACE, false));

    await expect(
      result.current.createPageFromTemplate({ templateId: 't-1', title: '議事録' }),
    ).rejects.toThrow();
  });

  it('workspaceSlug/spaceId が未確定なら投げる', async () => {
    const { result } = renderHook(() => useKbPageTemplates(undefined, undefined, false));

    await expect(
      result.current.createPageFromTemplate({ templateId: 't-1', title: 'x' }),
    ).rejects.toThrow();
  });
});
