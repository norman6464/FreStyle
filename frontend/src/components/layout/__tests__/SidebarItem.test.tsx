import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import SidebarItem from '../SidebarItem';
import { HomeIcon } from '@heroicons/react/24/outline';

function renderSidebarItem(active = false) {
  return render(
    <MemoryRouter>
      <SidebarItem icon={HomeIcon} label="ホーム" to="/" active={active} />
    </MemoryRouter>
  );
}

describe('SidebarItem', () => {
  it('ラベルを表示する', () => {
    renderSidebarItem();
    expect(screen.getByText('ホーム')).toBeDefined();
  });

  it('アイコンを表示する', () => {
    renderSidebarItem();
    const icon = document.querySelector('svg');
    expect(icon).toBeDefined();
  });

  it('アクティブ状態で強調スタイルが適用される', () => {
    const { container } = renderSidebarItem(true);
    const link = container.querySelector('a');
    expect(link?.className).toContain('bg-surface-3');
    expect(link?.className).toContain('text-primary-300');
  });

  it('非アクティブ状態で通常スタイルが適用される', () => {
    const { container } = renderSidebarItem(false);
    const link = container.querySelector('a');
    expect(link?.className).toContain('text-[var(--color-text-tertiary)]');
  });

  it('バッジが渡されると未読数を表示する', () => {
    render(
      <MemoryRouter>
        <SidebarItem icon={HomeIcon} label="チャット" to="/chat" active={false} badge={5} />
      </MemoryRouter>
    );
    expect(screen.getByText('5')).toBeInTheDocument();
  });

  it('バッジが0の場合は表示しない', () => {
    render(
      <MemoryRouter>
        <SidebarItem icon={HomeIcon} label="チャット" to="/chat" active={false} badge={0} />
      </MemoryRouter>
    );
    expect(screen.queryByTestId('sidebar-badge')).not.toBeInTheDocument();
  });
});
