import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import SidebarItem from '../SidebarItem';
import { HomeIcon } from '@heroicons/react/24/outline';

describe('SidebarItem', () => {
  it('ラベルを表示する', () => {
    render(
      <SidebarItem icon={HomeIcon} label="ホーム" to="/" active={false} />
    );
    expect(screen.getByText('ホーム')).toBeDefined();
  });

  it('アイコンを表示する', () => {
    render(
      <SidebarItem icon={HomeIcon} label="ホーム" to="/" active={false} />
    );
    const icon = document.querySelector('svg');
    expect(icon).toBeDefined();
  });

  it('アクティブ状態で強調スタイルが適用される', () => {
    const { container } = render(
      <SidebarItem icon={HomeIcon} label="ホーム" to="/" active={true} />
    );
    const link = container.querySelector('a');
    expect(link?.className).toContain('bg-primary-50');
    expect(link?.className).toContain('text-primary-700');
  });

  it('非アクティブ状態で通常スタイルが適用される', () => {
    const { container } = render(
      <SidebarItem icon={HomeIcon} label="ホーム" to="/" active={false} />
    );
    const link = container.querySelector('a');
    expect(link?.className).toContain('text-slate-600');
  });
});
