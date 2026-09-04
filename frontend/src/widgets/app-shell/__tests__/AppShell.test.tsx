import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import authReducer from '@/entities/user/model/authSlice';
import AppShell from '../ui/AppShell';
import { ToastProvider } from '@/app/providers/ToastProvider';

function createTestStore() {
  return configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { isAuthenticated: true, loading: false } },
  });
}

function renderAppShell({ initialEntry = '/' } = {}) {
  return render(
    <Provider store={createTestStore()}>
      <MemoryRouter initialEntries={[initialEntry]}>
        <ToastProvider>
          <Routes>
            <Route element={<AppShell />}>
              <Route path="/" element={<div>テストコンテンツ</div>} />
              <Route path="/courses/:id" element={<div>コース詳細コンテンツ</div>} />
            </Route>
          </Routes>
        </ToastProvider>
      </MemoryRouter>
    </Provider>
  );
}

describe('AppShell', () => {
  it('ヘッダーのナビを表示する', () => {
    renderAppShell();
    expect(screen.getAllByText('ホーム').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('ノート').length).toBeGreaterThanOrEqual(1);
  });

  it('子コンテンツを表示する', () => {
    renderAppShell();
    expect(screen.getByText('テストコンテンツ')).toBeDefined();
  });

  it('トップバーを表示する', () => {
    renderAppShell();
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

  // FRESTYLE-122: 教材閲覧はヘッダーごとスクロールするドキュメントスクロール構造になる。
  // コース管理はロールで区別しなくなったため、/courses/:id は誰が開いても同じレイアウト。
  describe('ドキュメントスクロール（FRESTYLE-122）', () => {
    it('/courses/:id ではヘッダーと main がスクロールコンテナに入る', () => {
      const { container } = renderAppShell({ initialEntry: '/courses/2' });
      const scroller = container.querySelector('[data-app-scroll]');
      expect(scroller).not.toBeNull();
      // ヘッダー(banner)と main が同じスクロールコンテナの中にある = 一緒に流れる
      expect(scroller!.querySelector('header, [role="banner"]')).not.toBeNull();
      expect(scroller!.querySelector('#main-content')).not.toBeNull();
      expect(screen.getByText('コース詳細コンテンツ')).toBeInTheDocument();
    });

    it('教材閲覧以外のルートは従来レイアウト（コンテナなし）', () => {
      const { container } = renderAppShell({ initialEntry: '/' });
      expect(container.querySelector('[data-app-scroll]')).toBeNull();
    });
  });
});
