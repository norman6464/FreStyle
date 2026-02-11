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
    expect(link?.className).toContain('bg-primary-50');
    expect(link?.className).toContain('text-primary-700');
  });

  it('非アクティブ状態で通常スタイルが適用される', () => {
    const { container } = renderSidebarItem(false);
    const link = container.querySelector('a');
    expect(link?.className).toContain('text-slate-600');
  });
});
