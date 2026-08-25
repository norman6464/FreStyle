import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

// Redux: 管理者としてページを表示できる状態にする。
const mockState = { auth: { isAdmin: true, loading: false } };
vi.mock('react-redux', () => ({
  useSelector: (sel: (s: typeof mockState) => unknown) => sel(mockState),
  useDispatch: () => vi.fn(),
}));

const { getCurrentUser, listInvitations, createInvitation, createTempPassword, cancelInvitation, listCompanies } =
  vi.hoisted(() => ({
    getCurrentUser: vi.fn(),
    listInvitations: vi.fn(),
    createInvitation: vi.fn(),
    createTempPassword: vi.fn(),
    cancelInvitation: vi.fn(),
    listCompanies: vi.fn(),
  }));

vi.mock('@/entities/user', () => ({
  AuthRepository: { getCurrentUser: () => getCurrentUser() },
}));

vi.mock('@/entities/invitation', () => ({
  AdminInvitationRepository: {
    list: () => listInvitations(),
    create: (form: unknown) => createInvitation(form),
    createWithTemporaryPassword: (form: unknown) => createTempPassword(form),
    cancel: (id: number) => cancelInvitation(id),
  },
}));

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

/** company_admin としてページを開き、初期ロードが終わるまで待つ。 */
async function renderAsCompanyAdmin(invitations = [pendingInvitation]) {
  getCurrentUser.mockResolvedValue({ id: 2, role: 'company_admin', companyId: 1 });
  listInvitations.mockResolvedValue(invitations);
  const view = renderPage();
  await waitFor(() => {
    expect(screen.getByDisplayValue('所属会社（自社に固定）')).toBeInTheDocument();
  });
  return view;
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('AdminInvitationsPage のロール別データ取得', () => {
  it('company_admin では会社一覧 API を呼ばず、画面がエラーにならない', async () => {
    await renderAsCompanyAdmin();

    // 会社欄は自社固定の表示になり、招待一覧も表示される。
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

  it('初期パスワード方式で発行された一時パスワードを 1 度だけ表示する', async () => {
    createTempPassword.mockResolvedValue({
      invitation: { ...pendingInvitation, email: 'new@example.com' },
      temporaryPassword: 'Temp-Pass-9!',
    });
    await renderAsCompanyAdmin([]);

    // 方式を初期パスワードに切り替え、email を入れて送信。
    fireEvent.change(screen.getByPlaceholderText('newmember@example.com'), {
      target: { value: 'new@example.com' },
    });
    fireEvent.click(screen.getByLabelText(/初期パスワードを発行/));
    fireEvent.click(screen.getByRole('button', { name: '初期パスワードを発行' }));

    await waitFor(() => {
      expect(createTempPassword).toHaveBeenCalledTimes(1);
    });
    // 一時パスワードが表示される。
    expect(await screen.findByText('Temp-Pass-9!')).toBeInTheDocument();
    expect(screen.getByText(/二度と表示できません/)).toBeInTheDocument();

    // 「閉じる」で表示が消える。
    fireEvent.click(screen.getByRole('button', { name: /閉じる/ }));
    expect(screen.queryByText('Temp-Pass-9!')).not.toBeInTheDocument();
  });
});

/*
 * 構造変更（model/ 層の切り出し）で通信の回数・順序が変わっていないことを固定するテスト。
 * 「1 回だけ呼ぶ」を明示的に数えるのは、hook 化に伴う useEffect の依存配列の取り違えで
 * 初期ロードが 2 回走る事故が起きやすいため。
 */
describe('AdminInvitationsPage の API 呼び出し回数', () => {
  it('company_admin の初回表示では getCurrentUser と招待一覧を 1 回ずつ呼ぶ', async () => {
    await renderAsCompanyAdmin();

    expect(getCurrentUser).toHaveBeenCalledTimes(1);
    expect(listInvitations).toHaveBeenCalledTimes(1);
    expect(listCompanies).toHaveBeenCalledTimes(0);
  });

  it('super_admin の初回表示では会社一覧も含めて 1 回ずつ呼ぶ', async () => {
    getCurrentUser.mockResolvedValue({ id: 1, role: 'super_admin', companyId: null });
    listInvitations.mockResolvedValue([]);
    listCompanies.mockResolvedValue([{ id: 1, name: '株式会社FreStyle' }]);

    renderPage();
    await waitFor(() => {
      expect(screen.getByRole('option', { name: '株式会社FreStyle' })).toBeInTheDocument();
    });

    expect(getCurrentUser).toHaveBeenCalledTimes(1);
    expect(listInvitations).toHaveBeenCalledTimes(1);
    expect(listCompanies).toHaveBeenCalledTimes(1);
  });

  it('招待メールを送信すると create を 1 回呼び、その後に一覧を 1 回だけ再取得する', async () => {
    createInvitation.mockResolvedValue({ ...pendingInvitation, email: 'new@example.com' });
    await renderAsCompanyAdmin([]);

    fireEvent.change(screen.getByPlaceholderText('newmember@example.com'), {
      target: { value: 'new@example.com' },
    });
    fireEvent.change(screen.getByPlaceholderText('例: 山田太郎'), { target: { value: '山田太郎' } });
    fireEvent.click(screen.getByRole('button', { name: '招待メールを送信' }));

    await waitFor(() => {
      expect(screen.getByRole('status')).toHaveTextContent(
        'new@example.com 宛に招待メールを送信しました。',
      );
    });

    // 送信フォームの中身はそのまま backend へ渡る（会社は自社固定 / 役職は SoD 固定）。
    expect(createInvitation).toHaveBeenCalledTimes(1);
    expect(createInvitation).toHaveBeenCalledWith({
      companyId: 1,
      email: 'new@example.com',
      role: 'trainee',
      displayName: '山田太郎',
    });
    // 作成後の再取得は 1 度だけ（初回 + 再取得 = 2）。
    expect(getCurrentUser).toHaveBeenCalledTimes(2);
    expect(listInvitations).toHaveBeenCalledTimes(2);
    expect(listCompanies).toHaveBeenCalledTimes(0);

    // メールアドレスと表示名は送信後に空へ戻る（会社は保持）。
    expect(screen.getByPlaceholderText('newmember@example.com')).toHaveValue('');
    expect(screen.getByPlaceholderText('例: 山田太郎')).toHaveValue('');
  });

  it('招待の取り消しは確認モーダルの確定で cancel を 1 回呼び、一覧を再取得する', async () => {
    cancelInvitation.mockResolvedValue(undefined);
    await renderAsCompanyAdmin();

    fireEvent.click(screen.getByRole('button', { name: '取り消し' }));

    // 確認モーダルには対象のメールアドレスが出る。
    const dialog = await screen.findByRole('dialog');
    expect(dialog).toHaveTextContent('member@example.com 宛の招待を取り消します。');

    fireEvent.click(screen.getByRole('button', { name: '取り消す' }));

    await waitFor(() => {
      expect(cancelInvitation).toHaveBeenCalledTimes(1);
    });
    expect(cancelInvitation).toHaveBeenCalledWith(10);
    expect(getCurrentUser).toHaveBeenCalledTimes(2);
    expect(listInvitations).toHaveBeenCalledTimes(2);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('確認モーダルを「戻る」で閉じると cancel を呼ばない', async () => {
    await renderAsCompanyAdmin();

    fireEvent.click(screen.getByRole('button', { name: '取り消し' }));
    await screen.findByRole('dialog');
    fireEvent.click(screen.getByRole('button', { name: '戻る' }));

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(cancelInvitation).not.toHaveBeenCalled();
    expect(listInvitations).toHaveBeenCalledTimes(1);
  });

  it('招待の取り消しに失敗したときはエラーを表示し、モーダルを開いたままにする', async () => {
    cancelInvitation.mockRejectedValue(new Error('boom'));
    await renderAsCompanyAdmin();

    fireEvent.click(screen.getByRole('button', { name: '取り消し' }));
    await screen.findByRole('dialog');
    fireEvent.click(screen.getByRole('button', { name: '取り消す' }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('招待のキャンセルに失敗しました');
    });
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    // 失敗時は再取得しない。
    expect(listInvitations).toHaveBeenCalledTimes(1);
  });

  it('招待の作成に失敗したときはエラーコードを日本語にして表示し、再取得しない', async () => {
    createInvitation.mockRejectedValue({
      response: { data: { error: 'company_admin_can_only_invite_trainee' } },
    });
    await renderAsCompanyAdmin([]);

    fireEvent.change(screen.getByPlaceholderText('newmember@example.com'), {
      target: { value: 'new@example.com' },
    });
    fireEvent.click(screen.getByRole('button', { name: '招待メールを送信' }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('会社管理者が招待できるのは受講者のみです。');
    });
    expect(listInvitations).toHaveBeenCalledTimes(1);
  });
});
