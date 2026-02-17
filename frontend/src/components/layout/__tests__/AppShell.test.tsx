import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import authReducer from '../../../store/authSlice';
import AppShell from '../AppShell';
import { ToastProvider } from '../../../hooks/useToast';

function createTestStore() {
  return configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { isAuthenticated: true, loading: false } },
  });
}

function renderAppShell() {
  return render(
    <Provider store={createTestStore()}>
      <MemoryRouter initialEntries={['/']}>
        <ToastProvider>
          <Routes>
            <Route element={<AppShell />}>
              <Route path="/" element={<div>テストコンテンツ</div>} />
            </Route>
          </Routes>
        </ToastProvider>
      </MemoryRouter>
    </Provider>
  );
}

describe('AppShell', () => {
  it('サイドバーを表示する', () => {
    renderAppShell();
    // デスクトップ+モバイルで2つのSidebarが存在
    expect(screen.getAllByText('ホーム').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('チャット').length).toBeGreaterThanOrEqual(1);
  });

  it('子コンテンツを表示する', () => {
    renderAppShell();
    expect(screen.getByText('テストコンテンツ')).toBeDefined();
  });

  it('トップバーを表示する', () => {
    renderAppShell();
    // TopBarにはモバイルメニューボタンがある
    const menuButton = screen.getByRole('button', { name: /メニュー/i });
    expect(menuButton).toBeDefined();
  });

  it('Cmd+Kでコマンドパレットが開く', () => {
    renderAppShell();
    expect(screen.queryByPlaceholderText('コマンドを検索...')).not.toBeInTheDocument();
    fireEvent.keyDown(document, { key: 'k', metaKey: true });
    expect(screen.getByPlaceholderText('コマンドを検索...')).toBeInTheDocument();
  });

  it('Ctrl+Kでコマンドパレットが開く', () => {
    renderAppShell();
    fireEvent.keyDown(document, { key: 'k', ctrlKey: true });
    expect(screen.getByPlaceholderText('コマンドを検索...')).toBeInTheDocument();
  });
});
