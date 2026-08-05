import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Redux: 管理者としてページを表示できる状態にする。
const mockState = { auth: { isAdmin: true, loading: false } };
vi.mock('react-redux', () => ({
  useSelector: (sel: (s: typeof mockState) => unknown) => sel(mockState),
  useDispatch: () => vi.fn(),
}));

const getCurrentUser = vi.fn();
vi.mock('@/entities/user', () => ({
  AuthRepository: { getCurrentUser: () => getCurrentUser() },
}));

const listInvitations = vi.fn();
vi.mock('@/entities/invitation', () => ({
  AdminInvitationRepository: {
    list: () => listInvitations(),
    create: vi.fn(),
    cancel: vi.fn(),
  },
}));

const listCompanies = vi.fn();
vi.mock('@/entities/company', () => ({
  CompanyRepository: { list: () => listCompanies() },
}));

import AdminInvitationsPage from '../ui/AdminInvitationsPage';

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminInvitationsPage />
    </MemoryRouter>,
  );
}

const pendingInvitation = {
  id: 10,
  email: 'member@example.com',
  role: 'trainee',
  createdAt: '2026-08-05T00:00:00Z',
  expiresAt: '2026-08-12T00:00:00Z',
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe('AdminInvitationsPage のロール別データ取得', () => {
  it('company_admin では会社一覧 API を呼ばず、画面がエラーにならない', async () => {
    getCurrentUser.mockResolvedValue({ id: 2, role: 'company_admin', companyId: 1 });
    listInvitations.mockResolvedValue([pendingInvitation]);

    renderPage();

    // 会社欄は自社固定の表示になり、招待一覧も表示される。
    await waitFor(() => {
      expect(screen.getByDisplayValue('所属会社（自社に固定）')).toBeInTheDocument();
    });
    expect(screen.getByText('member@example.com')).toBeInTheDocument();

    // super_admin 専用の会社一覧 API は呼ばない（403 で画面全体が落ちる原因だった）。
    expect(listCompanies).not.toHaveBeenCalled();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();

    // 役職は受講者に固定される。
    expect(screen.getByDisplayValue('受講者（自社のメンバー）')).toBeInTheDocument();
  });

  it('super_admin では会社一覧を取得して選択肢に表示する', async () => {
    getCurrentUser.mockResolvedValue({ id: 1, role: 'super_admin', companyId: null });
    listInvitations.mockResolvedValue([]);
    listCompanies.mockResolvedValue([{ id: 1, name: '株式会社FreStyle' }]);

    renderPage();

    await waitFor(() => {
      expect(screen.getByRole('option', { name: '株式会社FreStyle' })).toBeInTheDocument();
    });
    expect(listCompanies).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();

    // 役職は会社管理者に固定される。
    expect(screen.getByDisplayValue('会社管理者（招待先の会社の管理者）')).toBeInTheDocument();
  });

  it('招待一覧の取得に失敗したときはエラーを表示する', async () => {
    getCurrentUser.mockResolvedValue({ id: 2, role: 'company_admin', companyId: 1 });
    listInvitations.mockRejectedValue(new Error('network error'));

    renderPage();

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('データの取得に失敗しました');
    });
  });
});
