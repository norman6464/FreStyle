import { render, screen, fireEvent } from '@testing-library/react';
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

function renderSidebar(onNavigate?: () => void) {
  return render(
    <Provider store={createTestStore()}>
      <MemoryRouter initialEntries={['/']}>
        <Sidebar onNavigate={onNavigate} />
      </MemoryRouter>
    </Provider>
  );
}

describe('Sidebar モバイル動作', () => {
  it('ナビ項目クリック時にonNavigateが呼ばれる', () => {
    const onNavigate = vi.fn();
    renderSidebar(onNavigate);
    fireEvent.click(screen.getByText('チャット'));
    expect(onNavigate).toHaveBeenCalledTimes(1);
  });

  it('ログアウトクリック時にonNavigateが呼ばれる', () => {
    const onNavigate = vi.fn();
    global.fetch = vi.fn().mockResolvedValue({ ok: true });
    renderSidebar(onNavigate);
    fireEvent.click(screen.getByText('ログアウト'));
    expect(onNavigate).toHaveBeenCalledTimes(1);
  });

  it('onNavigateが未指定でもエラーにならない', () => {
    renderSidebar();
    expect(() => fireEvent.click(screen.getByText('チャット'))).not.toThrow();
  });
});
