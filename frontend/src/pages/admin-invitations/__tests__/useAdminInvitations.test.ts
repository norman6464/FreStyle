import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';

// Repository は Public API ごと差し替える（DOM も axios も経由しない）。
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
vi.mock('@/shared/lib/logger', () => ({ logger: { error: vi.fn(), warn: vi.fn(), info: vi.fn() } }));

import { useAdminInvitations } from '../model/useAdminInvitations';

const pendingInvitation = {
  id: 10,
  email: 'member@example.com',
  role: 'trainee' as const,
  createdAt: '2026-08-05T00:00:00Z',
  expiresAt: '2026-08-12T00:00:00Z',
};

/** company_admin として hook をマウントし、初期ロードの完了まで待つ。 */
async function mountAsCompanyAdmin(invitations = [pendingInvitation]) {
  getCurrentUser.mockResolvedValue({ id: 2, role: 'company_admin', companyId: 7 });
  listInvitations.mockResolvedValue(invitations);
  const view = renderHook(() => useAdminInvitations());
  await waitFor(() => expect(view.result.current.loading).toBe(false));
  return view;
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('useAdminInvitations の初期ロード', () => {
  it('company_admin では会社一覧を取らず、会社を自社に・役職を受講者に固定する', async () => {
    const { result } = await mountAsCompanyAdmin();

    expect(getCurrentUser).toHaveBeenCalledTimes(1);
    expect(listInvitations).toHaveBeenCalledTimes(1);
    expect(listCompanies).not.toHaveBeenCalled();

    expect(result.current.isCompanyAdmin).toBe(true);
    expect(result.current.isSuperAdmin).toBe(false);
    expect(result.current.invitations).toEqual([pendingInvitation]);
    expect(result.current.companies).toEqual([]);
    expect(result.current.form).toEqual({
      companyId: 7,
      email: '',
      role: 'trainee',
      displayName: '',
    });
    expect(result.current.error).toBeNull();
  });

  it('super_admin では会社一覧を 1 回取得し、役職を会社管理者・会社を先頭社に既定する', async () => {
    getCurrentUser.mockResolvedValue({ id: 1, role: 'super_admin', companyId: null });
    listInvitations.mockResolvedValue([]);
    listCompanies.mockResolvedValue([{ id: 3, name: 'アクメ社' }, { id: 4, name: 'ベータ社' }]);

    const { result } = renderHook(() => useAdminInvitations());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(getCurrentUser).toHaveBeenCalledTimes(1);
    expect(listInvitations).toHaveBeenCalledTimes(1);
    expect(listCompanies).toHaveBeenCalledTimes(1);
    expect(result.current.isSuperAdmin).toBe(true);
    expect(result.current.form.role).toBe('company_admin');
    expect(result.current.form.companyId).toBe(3);
  });

  it('super_admin で会社が 1 社も無いときは会社未選択のままにする', async () => {
    getCurrentUser.mockResolvedValue({ id: 1, role: 'super_admin', companyId: null });
    listInvitations.mockResolvedValue([]);
    listCompanies.mockResolvedValue([]);

    const { result } = renderHook(() => useAdminInvitations());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.form.companyId).toBe(0);
  });

  it('取得に失敗したときはエラーを立てて loading を落とす', async () => {
    getCurrentUser.mockRejectedValue(new Error('network error'));

    const { result } = renderHook(() => useAdminInvitations());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBe('データの取得に失敗しました');
    expect(listInvitations).not.toHaveBeenCalled();
  });
});

describe('useAdminInvitations の招待作成', () => {
  it('招待リンク方式では create を 1 回呼び、成功文言を出してフォームを会社だけ残して空にする', async () => {
    createInvitation.mockResolvedValue({ ...pendingInvitation, email: 'new@example.com' });
    const { result } = await mountAsCompanyAdmin([]);

    act(() => result.current.setForm({ ...result.current.form, email: 'new@example.com', displayName: '山田' }));
    await act(async () => {
      await result.current.submit();
    });

    expect(createInvitation).toHaveBeenCalledTimes(1);
    expect(createInvitation).toHaveBeenCalledWith({
      companyId: 7,
      email: 'new@example.com',
      role: 'trainee',
      displayName: '山田',
    });
    expect(result.current.success).toContain('new@example.com 宛に招待メールを送信しました。');
    expect(result.current.form).toEqual({ companyId: 7, email: '', role: 'trainee', displayName: '' });
    // 作成後の再取得は 1 度だけ（初回 + 再取得 = 2）。
    expect(getCurrentUser).toHaveBeenCalledTimes(2);
    expect(listInvitations).toHaveBeenCalledTimes(2);
  });

  it('初期パスワード方式では一時パスワードを保持し、成功文言は出さない', async () => {
    createTempPassword.mockResolvedValue({
      invitation: { ...pendingInvitation, email: 'new@example.com' },
      temporaryPassword: 'Temp-Pass-9!',
    });
    const { result } = await mountAsCompanyAdmin([]);

    act(() => result.current.setMethod('temporary_password'));
    act(() => result.current.setForm({ ...result.current.form, email: 'new@example.com' }));
    await act(async () => {
      await result.current.submit();
    });

    expect(createTempPassword).toHaveBeenCalledTimes(1);
    expect(createInvitation).not.toHaveBeenCalled();
    expect(result.current.issuedPassword).toEqual({ email: 'new@example.com', password: 'Temp-Pass-9!' });
    expect(result.current.success).toBeNull();
    expect(result.current.copied).toBe(false);

    act(() => result.current.dismissIssuedPassword());
    expect(result.current.issuedPassword).toBeNull();
  });

  it('会社が未選択のときは通信せずエラーだけを出す', async () => {
    getCurrentUser.mockResolvedValue({ id: 1, role: 'super_admin', companyId: null });
    listInvitations.mockResolvedValue([]);
    listCompanies.mockResolvedValue([]);
    const { result } = renderHook(() => useAdminInvitations());
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.submit();
    });

    expect(result.current.error).toBe('会社を選択してください');
    expect(createInvitation).not.toHaveBeenCalled();
    expect(createTempPassword).not.toHaveBeenCalled();
  });

  it('backend の message / error を日本語化して error に入れ、再取得しない', async () => {
    createInvitation.mockRejectedValue({
      response: { data: { error: 'UsernameExistsException' } },
    });
    const { result } = await mountAsCompanyAdmin([]);

    act(() => result.current.setForm({ ...result.current.form, email: 'dup@example.com' }));
    await act(async () => {
      await result.current.submit();
    });

    expect(result.current.error).toBe('このメールアドレスはすでに登録済みです。再招待は不要です。');
    expect(listInvitations).toHaveBeenCalledTimes(1);
    expect(result.current.submitting).toBe(false);
  });

  it('レスポンス本文が無いときは既定の失敗文言を出す', async () => {
    createInvitation.mockRejectedValue(new Error('boom'));
    const { result } = await mountAsCompanyAdmin([]);

    await act(async () => {
      await result.current.submit();
    });

    expect(result.current.error).toBe('招待の作成に失敗しました');
  });
});

describe('useAdminInvitations の招待取り消し', () => {
  it('確定で cancel を 1 回呼び、対象を閉じて一覧を再取得する', async () => {
    cancelInvitation.mockResolvedValue(undefined);
    const { result } = await mountAsCompanyAdmin();

    act(() => result.current.requestCancel(pendingInvitation));
    expect(result.current.cancelTarget).toEqual(pendingInvitation);

    await act(async () => {
      await result.current.confirmCancel();
    });

    expect(cancelInvitation).toHaveBeenCalledTimes(1);
    expect(cancelInvitation).toHaveBeenCalledWith(10);
    expect(result.current.cancelTarget).toBeNull();
    expect(listInvitations).toHaveBeenCalledTimes(2);
  });

  it('対象が無いときは何も呼ばない', async () => {
    const { result } = await mountAsCompanyAdmin();

    await act(async () => {
      await result.current.confirmCancel();
    });

    expect(cancelInvitation).not.toHaveBeenCalled();
  });

  it('失敗したときはエラーを出し、対象を開いたままにする', async () => {
    cancelInvitation.mockRejectedValue(new Error('boom'));
    const { result } = await mountAsCompanyAdmin();

    act(() => result.current.requestCancel(pendingInvitation));
    await act(async () => {
      await result.current.confirmCancel();
    });

    expect(result.current.error).toBe('招待のキャンセルに失敗しました');
    expect(result.current.cancelTarget).toEqual(pendingInvitation);
    expect(listInvitations).toHaveBeenCalledTimes(1);
  });

  it('モーダルを閉じると対象が消える', async () => {
    const { result } = await mountAsCompanyAdmin();

    act(() => result.current.requestCancel(pendingInvitation));
    act(() => result.current.closeCancelModal());

    expect(result.current.cancelTarget).toBeNull();
  });

  it('取り消し処理中はモーダルを閉じられない', async () => {
    let release: () => void = () => {};
    cancelInvitation.mockReturnValue(new Promise<void>((resolve) => { release = resolve; }));
    const { result } = await mountAsCompanyAdmin();

    act(() => result.current.requestCancel(pendingInvitation));
    act(() => { result.current.confirmCancel(); });
    await waitFor(() => expect(result.current.canceling).toBe(true));

    act(() => result.current.closeCancelModal());
    expect(result.current.cancelTarget).toEqual(pendingInvitation);

    await act(async () => { release(); });
  });
});

describe('useAdminInvitations の初期パスワードのコピー', () => {
  async function mountWithIssuedPassword() {
    createTempPassword.mockResolvedValue({
      invitation: { ...pendingInvitation, email: 'new@example.com' },
      temporaryPassword: 'Temp-Pass-9!',
    });
    const view = await mountAsCompanyAdmin([]);
    act(() => view.result.current.setMethod('temporary_password'));
    await act(async () => {
      await view.result.current.submit();
    });
    return view;
  }

  it('クリップボードへ書けたら copied を立てる', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal('navigator', { clipboard: { writeText } });
    const { result } = await mountWithIssuedPassword();

    await act(async () => {
      await result.current.copyIssuedPassword();
    });

    expect(writeText).toHaveBeenCalledWith('Temp-Pass-9!');
    expect(result.current.copied).toBe(true);
    vi.unstubAllGlobals();
  });

  it('クリップボードが使えなくても落ちず copied は立たない', async () => {
    vi.stubGlobal('navigator', { clipboard: { writeText: vi.fn().mockRejectedValue(new Error('denied')) } });
    const { result } = await mountWithIssuedPassword();

    await act(async () => {
      await result.current.copyIssuedPassword();
    });

    expect(result.current.copied).toBe(false);
    vi.unstubAllGlobals();
  });

  it('発行前は何も起きない', async () => {
    const writeText = vi.fn();
    vi.stubGlobal('navigator', { clipboard: { writeText } });
    const { result } = await mountAsCompanyAdmin([]);

    await act(async () => {
      await result.current.copyIssuedPassword();
    });

    expect(writeText).not.toHaveBeenCalled();
    vi.unstubAllGlobals();
  });
});
