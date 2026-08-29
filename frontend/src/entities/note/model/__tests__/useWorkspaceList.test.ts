import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useWorkspaceList } from '../useWorkspaceList';

const hoisted = vi.hoisted(() => ({
  fetchWorkspaces: vi.fn(),
  createWorkspace: vi.fn(),
  deleteWorkspace: vi.fn(),
}));

vi.mock('../../api/noteRepository', () => ({
  default: {
    fetchWorkspaces: hoisted.fetchWorkspaces,
    createWorkspace: hoisted.createWorkspace,
    deleteWorkspace: hoisted.deleteWorkspace,
  },
}));

const WS_A = { slug: 'a', name: 'A', createdAt: '' };
const WS_B = { slug: 'b', name: 'B', createdAt: '' };

describe('useWorkspaceList', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    hoisted.fetchWorkspaces.mockResolvedValue([WS_A]);
  });

  it('マウント時に一覧を読み込む', async () => {
    const { result } = renderHook(() => useWorkspaceList());
    expect(result.current.loading).toBe(true);
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.workspaces).toEqual([WS_A]);
    expect(result.current.error).toBeNull();
  });

  it('読み込みに失敗するとエラーを返す', async () => {
    hoisted.fetchWorkspaces.mockRejectedValueOnce(new Error('boom'));
    const { result } = renderHook(() => useWorkspaceList());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe('ワークスペースを読み込めませんでした');
  });

  it('作ったワークスペースを一覧へ足す', async () => {
    hoisted.createWorkspace.mockResolvedValue(WS_B);
    const { result } = renderHook(() => useWorkspaceList());
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.createWorkspace({ name: 'B' });
    });

    expect(result.current.workspaces).toEqual([WS_A, WS_B]);
  });

  it('消したワークスペースを一覧から外す', async () => {
    hoisted.deleteWorkspace.mockResolvedValue(undefined);
    const { result } = renderHook(() => useWorkspaceList());
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.deleteWorkspace('a');
    });

    expect(result.current.workspaces).toEqual([]);
  });

  it('retry で読み直せる', async () => {
    const { result } = renderHook(() => useWorkspaceList());
    await waitFor(() => expect(result.current.loading).toBe(false));

    hoisted.fetchWorkspaces.mockResolvedValueOnce([WS_A, WS_B]);
    act(() => result.current.retry());
    await waitFor(() => expect(result.current.workspaces).toEqual([WS_A, WS_B]));
  });
});
