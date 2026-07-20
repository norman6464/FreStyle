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
    events: [
      {
        id: 1,
        actorId: 9,
        actorEmail: 'ops@frestyle.dev',
        actorRole: 'super_admin',
        action: 'DELETE /api/v2/admin/members/:userId',
        targetId: 42,
        createdAt: '2026-06-15T00:00:00Z',
      },
      {
        id: 2,
        actorId: 9,
        actorEmail: 'ops@frestyle.dev',
        actorRole: 'super_admin',
        action: 'PATCH /api/v2/admin/companies/:id/active',
        targetId: 3,
        createdAt: '2026-06-14T00:00:00Z',
      },
    ],
    loading: false,
    error: null as string | null,
  };
}
let hookReturn = makeHookReturn();
vi.mock('@/hooks/useAuditLog', () => ({
  useAuditLog: () => hookReturn,
}));

import AdminAuditLogPage from '../ui/AdminAuditLogPage';

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminAuditLogPage />
    </MemoryRouter>,
  );
}

describe('AdminAuditLogPage（監査ログ）', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockState.auth = { isAdmin: true, loading: false, role: 'super_admin' };
    hookReturn = makeHookReturn();
  });

  it('操作を日本語化し、実行者・対象IDを表示する', () => {
    renderPage();
    expect(screen.getByText('従業員を削除')).toBeInTheDocument();
    expect(screen.getByText('会社の有効/無効を変更')).toBeInTheDocument();
    expect(screen.getByText(/対象 ID: 42/)).toBeInTheDocument();
  });

  it('super_admin 以外はリダイレクトする', () => {
    mockState.auth = { isAdmin: true, loading: false, role: 'company_admin' };
    renderPage();
    expect(screen.queryByText('従業員を削除')).not.toBeInTheDocument();
  });
});
