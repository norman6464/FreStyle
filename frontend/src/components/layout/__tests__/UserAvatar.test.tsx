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
});
