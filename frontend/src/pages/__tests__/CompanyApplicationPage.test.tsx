import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import CompanyApplicationPage from '../CompanyApplicationPage';

const mockApply = vi.fn();

vi.mock('../../repositories/CompanyApplicationRepository', () => ({
  CompanyApplicationRepository: {
    apply: (form: unknown) => mockApply(form),
  },
}));

function renderPage() {
  return render(
    <MemoryRouter>
      <CompanyApplicationPage />
    </MemoryRouter>,
  );
}

function fillRequired() {
  fireEvent.change(screen.getByLabelText('会社名'), {
    target: { value: 'Example Corp' },
  });
  fireEvent.change(screen.getByLabelText('お名前（ご担当者）'), {
    target: { value: '山田太郎' },
  });
  fireEvent.change(screen.getByLabelText('メールアドレス'), {
    target: { value: 'yamada@example.com' },
  });
}

describe('CompanyApplicationPage', () => {
  beforeEach(() => {
    mockApply.mockReset();
  });

  it('フォーム項目が表示される', () => {
    renderPage();
    expect(screen.getByLabelText('会社名')).toBeInTheDocument();
    expect(screen.getByLabelText('お名前（ご担当者）')).toBeInTheDocument();
    expect(screen.getByLabelText('メールアドレス')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '申請する' })).toBeInTheDocument();
  });

  it('必須項目を入力して送信すると repository.apply が呼ばれ、成功表示になる', async () => {
    mockApply.mockResolvedValueOnce(undefined);
    renderPage();
    fillRequired();
    fireEvent.click(screen.getByRole('button', { name: '申請する' }));

    await waitFor(() => {
      expect(mockApply).toHaveBeenCalledWith({
        companyName: 'Example Corp',
        applicantName: '山田太郎',
        email: 'yamada@example.com',
        message: '',
      });
    });
    expect(
      await screen.findByText(/申請を受け付けました/),
    ).toBeInTheDocument();
  });

  it('必須が空のまま送信するとバリデーションエラーを出し、送信しない', async () => {
    renderPage();
    fireEvent.click(screen.getByRole('button', { name: '申請する' }));
    expect(
      await screen.findByText(/会社名・お名前・メールアドレスは必須です/),
    ).toBeInTheDocument();
    expect(mockApply).not.toHaveBeenCalled();
  });

  it('送信が失敗するとエラーメッセージを表示する', async () => {
    mockApply.mockRejectedValueOnce(new Error('network'));
    renderPage();
    fillRequired();
    fireEvent.click(screen.getByRole('button', { name: '申請する' }));
    expect(await screen.findByText(/送信に失敗しました/)).toBeInTheDocument();
  });
});
