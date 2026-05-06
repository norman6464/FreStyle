import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import authReducer from '../../store/authSlice';
import authRepository from '../../repositories/AuthRepository';
import WelcomePage from '../WelcomePage';

vi.mock('../../repositories/AuthRepository');

function renderPage(opts?: { onboarded?: boolean }) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: {
        isAuthenticated: true,
        loading: false,
        isAdmin: false,
        onboarded: opts?.onboarded ?? false,
      },
    },
  });
  return render(
    <Provider store={store}>
      <MemoryRouter initialEntries={['/welcome']}>
        <Routes>
          <Route path="/welcome" element={<WelcomePage />} />
          <Route path="/" element={<div>ホーム</div>} />
        </Routes>
      </MemoryRouter>
    </Provider>
  );
}

describe('WelcomePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('役割と表示名で挨拶し、できることを表示する', async () => {
    vi.mocked(authRepository.getCurrentUser).mockResolvedValue({
      id: 1,
      email: 't@example.com',
      role: 'company_admin',
      displayName: '佐藤',
      onboarded: false,
    });

    renderPage();

    await waitFor(() => {
      expect(screen.getByText(/佐藤 さん/)).toBeInTheDocument();
    });
    expect(screen.getByText(/会社管理者/)).toBeInTheDocument();
    expect(
      screen.getByText('管理画面で自社メンバーの招待 / 管理')
    ).toBeInTheDocument();
  });

  it('trainee には管理機能の項目を出さない', async () => {
    vi.mocked(authRepository.getCurrentUser).mockResolvedValue({
      id: 2,
      role: 'trainee',
      displayName: '山田',
      onboarded: false,
    });

    renderPage();

    await waitFor(() => {
      expect(screen.getByText(/山田 さん/)).toBeInTheDocument();
    });
    expect(
      screen.queryByText('管理画面で自社メンバーの招待 / 管理')
    ).not.toBeInTheDocument();
  });

  it('「はじめる」を押すと completeOnboarding を呼んでホームに遷移', async () => {
    vi.mocked(authRepository.getCurrentUser).mockResolvedValue({
      id: 1,
      role: 'trainee',
      displayName: '佐藤',
      onboarded: false,
    });
    vi.mocked(authRepository.completeOnboarding).mockResolvedValue();

    renderPage();

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'はじめる' })).toBeEnabled();
    });
    fireEvent.click(screen.getByRole('button', { name: 'はじめる' }));

    await waitFor(() => {
      expect(authRepository.completeOnboarding).toHaveBeenCalledTimes(1);
      expect(screen.getByText('ホーム')).toBeInTheDocument();
    });
  });

  // onboarded=true の状態で /welcome に来たら即ホームへリダイレクトする。
  it('onboarded=true で /welcome に来た場合は即ホームへリダイレクト', async () => {
    vi.mocked(authRepository.getCurrentUser).mockResolvedValue({
      id: 1,
      role: 'trainee',
      onboarded: true,
    });

    renderPage({ onboarded: true });

    await waitFor(() => {
      expect(screen.getByText('ホーム')).toBeInTheDocument();
    });
  });
});
