import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Redux: 会社管理者としてページを表示できる状態にする。
const mockState = { auth: { isAdmin: true, loading: false } };
vi.mock('react-redux', () => ({
  useSelector: (sel: (s: typeof mockState) => unknown) => sel(mockState),
  useDispatch: () => vi.fn(),
}));

// 従業員一覧は固定データを返す（API/axios は呼ばない）。
const members = [
  {
    id: 1,
    displayName: '河野拓真',
    email: 'kawano@example.com',
    role: 'company_admin',
    aiChatEnabled: null,
    isActive: true,
  },
  {
    id: 2,
    displayName: '木村',
    email: 'kimura@example.com',
    role: 'trainee',
    aiChatEnabled: null,
    isActive: true,
  },
];
vi.mock('../../hooks/useAdminMembers', () => ({
  useAdminMembers: () => ({
    members,
    loading: false,
    error: null,
    updatingId: null,
    setAiAccess: vi.fn(),
    setActive: vi.fn(),
    remove: vi.fn(),
  }),
}));

import AdminMembersPage from '../AdminMembersPage';

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminMembersPage />
    </MemoryRouter>,
  );
}

describe('AdminMembersPage の検索', () => {
  it('初期は全従業員が表示される', () => {
    renderPage();
    expect(screen.getByText('河野拓真')).toBeInTheDocument();
    expect(screen.getByText('木村')).toBeInTheDocument();
  });

  it('氏名で部分一致フィルタできる（一致しない行は消える）', () => {
    renderPage();
    fireEvent.change(screen.getByLabelText('従業員を検索'), { target: { value: '木村' } });
    expect(screen.queryByText('河野拓真')).not.toBeInTheDocument();
    expect(screen.getByText('木村')).toBeInTheDocument();
  });

  it('メールアドレスでもフィルタできる', () => {
    renderPage();
    fireEvent.change(screen.getByLabelText('従業員を検索'), { target: { value: 'kawano' } });
    expect(screen.getByText('河野拓真')).toBeInTheDocument();
    expect(screen.queryByText('木村')).not.toBeInTheDocument();
  });

  it('一致しないと no-results メッセージを出す', () => {
    renderPage();
    fireEvent.change(screen.getByLabelText('従業員を検索'), { target: { value: 'zzzznomatch' } });
    expect(screen.getByText(/に一致する従業員がいません/)).toBeInTheDocument();
  });
});
