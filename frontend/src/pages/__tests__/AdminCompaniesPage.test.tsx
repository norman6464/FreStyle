import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

const mockState = { auth: { isAdmin: true, loading: false, role: 'super_admin' } };
vi.mock('react-redux', () => ({
  useSelector: (sel: (s: typeof mockState) => unknown) => sel(mockState),
  useDispatch: () => vi.fn(),
}));

// CompanyRepository（default export インスタンス）をモック。listStats が横断ビューを返す。
const listStats = vi.fn();
const updateActive = vi.fn();
vi.mock('../../repositories/CompanyRepository', () => ({
  default: {
    listStats: () => listStats(),
    updateActive: (...args: unknown[]) => updateActive(...args),
  },
}));

import AdminCompaniesPage from '../AdminCompaniesPage';

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminCompaniesPage />
    </MemoryRouter>,
  );
}

describe('AdminCompaniesPage（会社横断ビュー）', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockState.auth = { isAdmin: true, loading: false, role: 'super_admin' };
    listStats.mockResolvedValue([
      {
        id: 1,
        name: 'アクメ社',
        isActive: true,
        createdAt: '2026-06-01T00:00:00Z',
        memberTotal: 5,
        activeMembers: 4,
        traineeCount: 3,
      },
    ]);
  });

  it('会社と各社のメンバー集計を表示する', async () => {
    renderPage();
    expect(await screen.findByText('アクメ社')).toBeInTheDocument();
    expect(
      screen.getByText((content, el) => el?.tagName === 'P' && /メンバー\s*5/.test(content)),
    ).toBeInTheDocument();
    expect(listStats).toHaveBeenCalled();
  });
});
