import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Redux: super_admin としてページを表示できる状態。テストごとに role を差し替える。
const mockState = { auth: { isAdmin: true, loading: false, role: 'super_admin' } };
vi.mock('react-redux', () => ({
  useSelector: (sel: (s: typeof mockState) => unknown) => sel(mockState),
  useDispatch: () => vi.fn(),
}));

const showToast = vi.fn();
vi.mock('@/hooks/useToast', () => ({ useToast: () => ({ showToast }) }));

// useCompanyApplications フックは固定データを返す（API/axios は呼ばない）。
const setStatus = vi.fn().mockResolvedValue(true);
const applications = [
  {
    id: 1,
    companyName: 'アクメ社',
    applicantName: '山田',
    email: 'yamada@acme.co.jp',
    message: '新卒研修で使いたい',
    status: 'pending',
    createdAt: '2026-06-01T00:00:00Z',
    updatedAt: '2026-06-01T00:00:00Z',
  },
  {
    id: 2,
    companyName: '承認済み社',
    applicantName: '佐藤',
    email: 'sato@ok.co.jp',
    message: '',
    status: 'approved',
    createdAt: '2026-05-20T00:00:00Z',
    updatedAt: '2026-05-21T00:00:00Z',
  },
];
function makeHookReturn() {
  return {
    applications,
    pendingCount: 1,
    loading: false,
    error: null as string | null,
    updatingId: null as number | null,
    setStatus,
    reload: vi.fn(),
  };
}
// 宣言時に初期化しておく（モック factory がモジュール読み込み時に評価されても undefined にならないように）。
let hookReturn = makeHookReturn();
vi.mock('@/hooks/useCompanyApplications', () => ({
  useCompanyApplications: () => hookReturn,
}));

import AdminCompanyApplicationsPage from '../ui/AdminCompanyApplicationsPage';

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminCompanyApplicationsPage />
    </MemoryRouter>,
  );
}

describe('AdminCompanyApplicationsPage（利用申請の承認）', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setStatus.mockResolvedValue(true);
    mockState.auth = { isAdmin: true, loading: false, role: 'super_admin' };
    hookReturn = makeHookReturn();
  });

  it('申請の一覧を表示し、承認済みには招待への導線を出す', () => {
    renderPage();
    expect(screen.getByText('アクメ社')).toBeInTheDocument();
    expect(screen.getByText('承認済み社')).toBeInTheDocument();
    expect(screen.getByText('招待へ')).toBeInTheDocument();
  });

  it('承認ボタンで status を approved に更新し、トーストを出す', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: '承認' }));
    expect(setStatus).toHaveBeenCalledWith(1, 'approved');
    await waitFor(() => expect(showToast).toHaveBeenCalledWith('success', expect.stringContaining('承認')));
  });

  it('却下は確認ダイアログで OK のとき rejected に更新する', () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: '却下' }));
    expect(setStatus).toHaveBeenCalledWith(1, 'rejected');
  });

  it('却下をキャンセルしたら更新しない', () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false);
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: '却下' }));
    expect(setStatus).not.toHaveBeenCalled();
  });

  it('super_admin 以外はリダイレクトして一覧を表示しない', () => {
    mockState.auth = { isAdmin: true, loading: false, role: 'company_admin' };
    renderPage();
    expect(screen.queryByText('アクメ社')).not.toBeInTheDocument();
  });
});
