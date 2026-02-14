import { render, screen } from '@testing-library/react';
import SectionHeader from '../SectionHeader';
import { UserCircleIcon } from '@heroicons/react/24/solid';

describe('SectionHeader', () => {
  it('タイトルが表示される', () => {
    render(<SectionHeader icon={UserCircleIcon} title="基本情報" />);
    expect(screen.getByText('基本情報')).toBeInTheDocument();
  });

  it('アイコンがレンダリングされる', () => {
    const { container } = render(<SectionHeader icon={UserCircleIcon} title="基本情報" />);
    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('h3要素でタイトルが表示される', () => {
    render(<SectionHeader icon={UserCircleIcon} title="基本情報" />);
    const heading = screen.getByText('基本情報');
    expect(heading.tagName).toBe('H3');
  });

  it('flexレイアウトのコンテナがレンダリングされる', () => {
    const { container } = render(<SectionHeader icon={UserCircleIcon} title="テスト" />);
    expect(container.firstChild).toHaveClass('flex');
    expect(container.firstChild).toHaveClass('items-center');
  });
});
