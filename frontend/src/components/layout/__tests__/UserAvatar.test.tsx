import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import UserAvatar from '../UserAvatar';

describe('UserAvatar', () => {
  it('名前の頭文字が表示される', () => {
    render(<UserAvatar name="テストユーザー" />);

    expect(screen.getByText('テ')).toBeInTheDocument();
  });

  it('英字名の大文字頭文字が表示される', () => {
    render(<UserAvatar name="alice" />);

    expect(screen.getByText('A')).toBeInTheDocument();
  });

  it('名前が空の場合?が表示される', () => {
    render(<UserAvatar name="" />);

    expect(screen.getByText('?')).toBeInTheDocument();
  });

  it('size="sm"でsmクラスが適用される', () => {
    const { container } = render(<UserAvatar name="太郎" size="sm" />);

    const avatar = container.firstElementChild!;
    expect(avatar.className).toContain('w-7');
    expect(avatar.className).toContain('h-7');
  });

  it('デフォルトサイズでmdクラスが適用される', () => {
    const { container } = render(<UserAvatar name="太郎" />);

    const avatar = container.firstElementChild!;
    expect(avatar.className).toContain('w-8');
    expect(avatar.className).toContain('h-8');
  });
});
