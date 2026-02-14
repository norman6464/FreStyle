import { render, screen } from '@testing-library/react';
import SectionHeader from '../SectionHeader';
import { UserCircleIcon, LightBulbIcon, ChatBubbleLeftRightIcon } from '@heroicons/react/24/solid';

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

  it('異なるアイコンでも正しくレンダリングされる', () => {
    const { container } = render(<SectionHeader icon={LightBulbIcon} title="設定" />);
    expect(container.querySelector('svg')).toBeInTheDocument();
    expect(screen.getByText('設定')).toBeInTheDocument();
  });

  it('長いタイトルが正しく表示される', () => {
    const longTitle = 'コミュニケーションスタイルの設定について';
    render(<SectionHeader icon={ChatBubbleLeftRightIcon} title={longTitle} />);
    expect(screen.getByText(longTitle)).toBeInTheDocument();
  });

  it('アイコンとタイトルの間にgapがある', () => {
    const { container } = render(<SectionHeader icon={UserCircleIcon} title="テスト" />);
    expect(container.firstChild).toHaveClass('gap-2');
  });

  it('下マージンが設定されている', () => {
    const { container } = render(<SectionHeader icon={UserCircleIcon} title="テスト" />);
    expect(container.firstChild).toHaveClass('mb-3');
  });
});
