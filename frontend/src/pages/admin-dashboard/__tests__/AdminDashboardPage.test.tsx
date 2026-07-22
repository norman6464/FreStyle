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
      applicationTotal: 4,
      pendingApplications: 2,
      recentPending: [
        { id: 1, companyName: 'アクメ社', applicantName: '山田', email: 'y@acme.co.jp', message: '', status: 'pending', createdAt: '', updatedAt: '' },
        { id: 2, companyName: 'ベータ社', applicantName: '佐藤', email: 's@beta.co.jp', message: '', status: 'pending', createdAt: '', updatedAt: '' },
      ],
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

  it('会社数と承認待ち件数のカードを表示する', () => {
    renderPage();
    expect(screen.getByText('会社数')).toBeInTheDocument();
    expect(screen.getByText('承認待ちの申請')).toBeInTheDocument();
    expect(screen.getByText('有効 2 / 無効 1')).toBeInTheDocument();
  });

  it('承認待ちの申請を一覧に出す', () => {
    renderPage();
    expect(screen.getByText('アクメ社')).toBeInTheDocument();
    expect(screen.getByText('ベータ社')).toBeInTheDocument();
  });

  it('super_admin 以外はリダイレクトする', () => {
    mockState.auth = { isAdmin: true, loading: false, role: 'company_admin' };
    renderPage();
    expect(screen.queryByText('会社数')).not.toBeInTheDocument();
  });
});
