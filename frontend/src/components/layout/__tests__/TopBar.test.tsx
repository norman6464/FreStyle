import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import TopBar from '../TopBar';

function renderTopBar(title = 'ホーム', onMenuToggle = () => {}) {
  return render(
    <MemoryRouter>
      <TopBar title={title} onMenuToggle={onMenuToggle} />
    </MemoryRouter>
  );
}

describe('TopBar', () => {
  it('ページタイトルを表示する', () => {
    renderTopBar('チャット');
    expect(screen.getByText('チャット')).toBeDefined();
  });

  it('異なるタイトルを表示できる', () => {
    renderTopBar('AI');
    expect(screen.getByText('AI')).toBeDefined();
  });

  it('モバイルメニューボタンを表示する', () => {
    renderTopBar();
    const menuButton = screen.getByRole('button', { name: /メニュー/i });
    expect(menuButton).toBeDefined();
  });

  it('メニューボタンクリックでonMenuToggleが呼ばれる', () => {
    const onMenuToggle = vi.fn();
    renderTopBar('テスト', onMenuToggle);

    fireEvent.click(screen.getByRole('button', { name: /メニュー/i }));
    expect(onMenuToggle).toHaveBeenCalledOnce();
  });

  it('headerタグでレンダリングされる', () => {
    renderTopBar();
    expect(document.querySelector('header')).toBeTruthy();
  });
});
