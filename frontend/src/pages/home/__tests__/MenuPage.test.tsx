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

  it('ツールのメニューカードを表示する', () => {
    renderMenu();

    expect(screen.getByRole('heading', { name: 'FreStyle へようこそ', level: 1 })).toBeInTheDocument();
    expect(screen.getByText('ツール')).toBeInTheDocument();
    expect(screen.getByText('ナレッジ')).toBeInTheDocument();
  });
});
