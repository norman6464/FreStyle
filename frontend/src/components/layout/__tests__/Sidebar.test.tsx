import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import authReducer from '../../../store/authSlice';
import Sidebar from '../Sidebar';

function createTestStore() {
  return configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { isAuthenticated: true, loading: false } },
  });
}

function renderSidebar(currentPath = '/') {
  return render(
    <Provider store={createTestStore()}>
      <MemoryRouter initialEntries={[currentPath]}>
        <Sidebar />
      </MemoryRouter>
    </Provider>
  );
}

describe('Sidebar', () => {
  it('全ナビゲーション項目を表示する', () => {
    renderSidebar();
    expect(screen.getByText('ホーム')).toBeDefined();
    expect(screen.getByText('チャット')).toBeDefined();
    expect(screen.getByText('AI')).toBeDefined();
    expect(screen.getByText('練習')).toBeDefined();
    expect(screen.getByText('ユーザー検索')).toBeDefined();
    expect(screen.getByText('プロフィール')).toBeDefined();
    expect(screen.getByText('パーソナリティ')).toBeDefined();
  });

  it('ログアウトボタンを表示する', () => {
    renderSidebar();
    expect(screen.getByText('ログアウト')).toBeDefined();
  });

  it('FreStyleロゴ画像を表示する', () => {
    renderSidebar();
    const logo = screen.getByAltText('FreStyle');
    expect(logo).toBeDefined();
    expect(logo.getAttribute('src')).toBe('/image.png');
  });

  it('ホームルートでホームがアクティブになる', () => {
    renderSidebar('/');
    const homeLink = screen.getByText('ホーム').closest('a');
    expect(homeLink?.className).toContain('bg-primary-50');
  });

  it('チャットルートでチャットがアクティブになる', () => {
    renderSidebar('/chat');
    const chatLink = screen.getByText('チャット').closest('a');
    expect(chatLink?.className).toContain('bg-primary-50');
  });
});
