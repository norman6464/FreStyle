import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import authReducer from '@/entities/user/model/authSlice';
import NavSidebar from '../ui/NavSidebar';
import type { NavSidebarProps } from '../ui/NavSidebar';

function renderSidebar(
  authState: Record<string, unknown>,
  props: Partial<NavSidebarProps> = {},
  initialPath = '/dashboard',
) {
  const store = configureStore({
    reducer: { auth: authReducer },
    preloadedState: { auth: { isAuthenticated: true, loading: false, ...authState } },
  });
  const merged: NavSidebarProps = {
    mode: 'pinned',
    onPin: vi.fn(),
    onCollapse: vi.fn(),
    ...props,
  };
  render(
    <Provider store={store}>
      <MemoryRouter initialEntries={[initialPath]}>
        <NavSidebar {...merged} />
      </MemoryRouter>
    </Provider>,
  );
  return merged;
}

describe('NavSidebar', () => {
  it('主要ナビ項目を縦に表示する（trainee）', () => {
    renderSidebar({ role: 'trainee', isAdmin: false, aiChatEnabledForTrainees: true });
    expect(screen.getByRole('navigation', { name: 'サイドナビゲーション' })).toBeInTheDocument();
    for (const label of ['ホーム', 'AI', '演習', 'コース', 'ノート', 'レポート']) {
      expect(screen.getByRole('link', { name: label })).toBeInTheDocument();
    }
    // trainee には管理セクションを出さない。
    expect(screen.queryByText('管理')).not.toBeInTheDocument();
  });

  it('AI が無効な trainee には AI を出さない', () => {
    renderSidebar({ role: 'trainee', isAdmin: false, aiChatEnabledForTrainees: false });
    expect(screen.queryByRole('link', { name: 'AI' })).not.toBeInTheDocument();
  });

  it('company_admin には管理メニュー（従業員一覧・招待管理）を出す', () => {
    renderSidebar({ role: 'company_admin', isAdmin: true, aiChatEnabledForTrainees: true });
    expect(screen.getByText('管理')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '従業員一覧' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '招待管理' })).toBeInTheDocument();
    // super_admin 限定の項目は出さない。
    expect(screen.queryByRole('link', { name: '監査ログ' })).not.toBeInTheDocument();
  });

  it('super_admin はホームのみ＋管理フルメニュー', () => {
    renderSidebar({ role: 'super_admin', isAdmin: true, aiChatEnabledForTrainees: true });
    expect(screen.getByRole('link', { name: 'ホーム' })).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'コース' })).not.toBeInTheDocument();
    expect(screen.getByRole('link', { name: '監査ログ' })).toBeInTheDocument();
  });

  it('現在ページに aria-current を付ける', () => {
    renderSidebar(
      { role: 'trainee', isAdmin: false, aiChatEnabledForTrainees: true },
      {},
      '/notes',
    );
    expect(screen.getByRole('link', { name: 'ノート' })).toHaveAttribute('aria-current', 'page');
    expect(screen.getByRole('link', { name: 'ホーム' })).not.toHaveAttribute('aria-current');
  });

  it('pinned では «（閉じる）を出し、クリックで onCollapse を呼ぶ', () => {
    const props = renderSidebar(
      { role: 'trainee', isAdmin: false, aiChatEnabledForTrainees: true },
      { mode: 'pinned' },
    );
    const btn = screen.getByRole('button', { name: 'サイドバーを閉じる' });
    fireEvent.click(btn);
    expect(props.onCollapse).toHaveBeenCalled();
    expect(screen.queryByRole('button', { name: 'サイドバーを固定表示する' })).not.toBeInTheDocument();
  });

  it('collapsed では »（固定表示する）を出し、クリックで onPin を呼ぶ', () => {
    const props = renderSidebar(
      { role: 'trainee', isAdmin: false, aiChatEnabledForTrainees: true },
      { mode: 'collapsed' },
    );
    const btn = screen.getByRole('button', { name: 'サイドバーを固定表示する' });
    fireEvent.click(btn);
    expect(props.onPin).toHaveBeenCalled();
  });

  it('マウスの出入りを onPointerEnter / onPointerLeave で親へ伝える', () => {
    const props = renderSidebar(
      { role: 'trainee', isAdmin: false, aiChatEnabledForTrainees: true },
      { mode: 'collapsed', onPointerEnter: vi.fn(), onPointerLeave: vi.fn() },
    );
    const nav = screen.getByRole('navigation', { name: 'サイドナビゲーション' });
    fireEvent.mouseEnter(nav);
    expect(props.onPointerEnter).toHaveBeenCalled();
    fireEvent.mouseLeave(nav);
    expect(props.onPointerLeave).toHaveBeenCalled();
  });

  it('設定リンクを表示する', () => {
    renderSidebar({ role: 'trainee', isAdmin: false, aiChatEnabledForTrainees: true });
    expect(screen.getByRole('link', { name: '設定' })).toBeInTheDocument();
  });
});
