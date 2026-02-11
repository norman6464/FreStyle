import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import authReducer from '../../../store/authSlice';
import AppShell from '../AppShell';

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
        <Routes>
          <Route element={<AppShell />}>
            <Route path="/" element={<div>テストコンテンツ</div>} />
          </Route>
        </Routes>
      </MemoryRouter>
    </Provider>
  );
}

describe('AppShell', () => {
  it('サイドバーを表示する', () => {
    renderAppShell();
    expect(screen.getByText('ホーム')).toBeDefined();
    expect(screen.getByText('チャット')).toBeDefined();
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
});
