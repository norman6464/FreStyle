import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

const mockState = { auth: { isAdmin: true, loading: false, role: 'super_admin' } };
vi.mock('react-redux', () => ({
  useSelector: (sel: (s: typeof mockState) => unknown) => sel(mockState),
  useDispatch: () => vi.fn(),
}));

// entities/company の Public API をまるごと差し替える（実 axios は呼ばない）。
const { listStats, updateActive } = vi.hoisted(() => ({
  listStats: vi.fn(),
  updateActive: vi.fn(),
}));
vi.mock('@/entities/company', () => ({
  CompanyRepository: {
    listStats: () => listStats(),
    updateActive: (id: number, active: boolean) => updateActive(id, active),
  },
}));

import AdminCompaniesPage from '../ui/AdminCompaniesPage';

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
    updateActive.mockResolvedValue(undefined);
  });

  it('会社と各社のメンバー集計を表示する', async () => {
    renderPage();
    expect(await screen.findByText('アクメ社')).toBeInTheDocument();
    expect(
      screen.getByText((content, el) => el?.tagName === 'P' && /メンバー\s*5/.test(content)),
    ).toBeInTheDocument();
    // 初期ロードで横断ビューを 1 回だけ取得する（重複呼び出しを防ぐ）。
    expect(listStats).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('会社一覧の取得に失敗したときはエラーを表示する', async () => {
    listStats.mockRejectedValue(new Error('boom'));
    renderPage();

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('会社一覧の取得に失敗しました');
    });
    expect(listStats).toHaveBeenCalledTimes(1);
  });

  it('無効化ボタンで updateActive を 1 回呼び、一覧は再取得せず楽観的に反映する', async () => {
    renderPage();
    fireEvent.click(await screen.findByRole('button', { name: '無効化' }));

    // 楽観的更新: レスポンスを待たずに「無効」バッジとボタン表示が切り替わる。
    expect(await screen.findByRole('button', { name: '有効化' })).toBeInTheDocument();
    expect(screen.getByText('無効')).toBeInTheDocument();

    await waitFor(() => {
      expect(updateActive).toHaveBeenCalledTimes(1);
    });
    expect(updateActive).toHaveBeenCalledWith(1, false);
    // 一覧の再取得はしない（楽観的更新のみ）。
    expect(listStats).toHaveBeenCalledTimes(1);
  });

  it('切り替えに失敗したときは元の状態へ戻してエラーを表示する', async () => {
    updateActive.mockRejectedValue(new Error('boom'));
    renderPage();
    fireEvent.click(await screen.findByRole('button', { name: '無効化' }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('会社状態の更新に失敗しました');
    });
    // ロールバックして「無効化」ボタンに戻る。
    expect(screen.getByRole('button', { name: '無効化' })).toBeInTheDocument();
    expect(screen.queryByText('無効')).not.toBeInTheDocument();
    expect(listStats).toHaveBeenCalledTimes(1);
  });

  // 通過条件（誰が入れて誰が /dashboard へ戻されるか）はルート側の RequireRole が持つ。
  // その表は src/app/__tests__/adminRouteAuthorization.test.tsx で固定している。
});
