import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

const mockState = { auth: { isAdmin: true, loading: false, role: 'super_admin' } };
vi.mock('react-redux', () => ({
  useSelector: (sel: (s: typeof mockState) => unknown) => sel(mockState),
  useDispatch: () => vi.fn(),
}));

function makeHookReturn() {
  return {
    summary: {
      companyTotal: 3,
      companyActive: 2,
      companyInactive: 1,
    },
    loading: false,
    error: null as string | null,
  };
}
let hookReturn = makeHookReturn();
vi.mock('../model/useAdminDashboard', () => ({
  useAdminDashboard: () => hookReturn,
}));

import AdminDashboardPage from '../ui/AdminDashboardPage';

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminDashboardPage />
    </MemoryRouter>,
  );
}

describe('AdminDashboardPage（運営ダッシュボード）', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockState.auth = { isAdmin: true, loading: false, role: 'super_admin' };
    hookReturn = makeHookReturn();
  });

  it('会社数のカードを表示する', () => {
    renderPage();
    expect(screen.getByText('会社数')).toBeInTheDocument();
    expect(screen.getByText('有効 2 / 無効 1')).toBeInTheDocument();
  });

  // 通過条件（誰が入れて誰が /dashboard へ戻されるか）はルート側の RequireRole が持つ。
  // その表は src/app/__tests__/adminRouteAuthorization.test.tsx で固定している。
});
