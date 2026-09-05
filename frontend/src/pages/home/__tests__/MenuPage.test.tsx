import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { MemoryRouter } from 'react-router-dom';
import MenuPage from '../ui/MenuPage';
import authReducer from '@/entities/user/model/authSlice';
import { createMockStorage } from '@/test/mockStorage';

function renderMenu() {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: {
      auth: { isAuthenticated: true, loading: false },
    },
  });
  return render(
    <Provider store={store}>
      <MemoryRouter>
        <MenuPage />
      </MemoryRouter>
    </Provider>,
  );
}

describe('MenuPage', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createMockStorage());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it('学習・ツールのメニューカードを表示する', () => {
    renderMenu();

    expect(screen.getByRole('heading', { name: 'FreStyle へようこそ', level: 1 })).toBeInTheDocument();
    expect(screen.getByText('コード演習')).toBeInTheDocument();
    expect(screen.getByText('ノート')).toBeInTheDocument();
  });

  it('演習カードに学べる技術ロゴ(Devicon)が出る (FRESTYLE-179)', () => {
    const { container } = renderMenu();
    // LanguageIcon は /lang/<key>.svg を img で描画する。演習(go 等)のロゴが出る。
    expect(container.querySelector('img[src="/lang/go.svg"]')).not.toBeNull();
    expect(container.querySelector('img[src="/lang/typescript.svg"]')).not.toBeNull();
  });
});
