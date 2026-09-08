import { describe, it, expect, vi } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useSidebar } from '../useSidebar';
import { useAuth } from '@/features/auth';

// ログアウトの実体は useAuth に一本化した。ここでは配線（useAuth の logout/loading を
// handleLogout/loggingOut としてそのまま渡していること）だけを確かめる。
// logout 自体の挙動（Redux 更新・遷移・失敗時の扱い）は useAuth.test.ts が検証する。
vi.mock('@/features/auth', () => ({
  useAuth: vi.fn(),
}));

describe('useSidebar', () => {
  it('useAuth の logout を handleLogout として返す', () => {
    const mockLogout = vi.fn();
    vi.mocked(useAuth).mockReturnValue({ logout: mockLogout, loading: false } as never);

    const { result } = renderHook(() => useSidebar());
    result.current.handleLogout();

    expect(mockLogout).toHaveBeenCalledOnce();
  });

  it('useAuth の loading を loggingOut として返す', () => {
    vi.mocked(useAuth).mockReturnValue({ logout: vi.fn(), loading: true } as never);

    const { result } = renderHook(() => useSidebar());

    expect(result.current.loggingOut).toBe(true);
  });
});
