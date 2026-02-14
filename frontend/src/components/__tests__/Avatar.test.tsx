import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import Avatar from '../Avatar';

describe('Avatar', () => {
  it('名前のイニシャルが表示される', () => {
    render(<Avatar name="田中太郎" />);
    expect(screen.getByText('田')).toBeInTheDocument();
  });

  it('画像URLが指定された場合imgタグが表示される', () => {
    render(<Avatar name="田中" src="https://example.com/photo.jpg" />);
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'https://example.com/photo.jpg');
    expect(img).toHaveAttribute('alt', '田中');
  });

  it('画像URLがない場合イニシャルフォールバックが表示される', () => {
    render(<Avatar name="Taro" />);
    expect(screen.getByText('T')).toBeInTheDocument();
    expect(screen.queryByRole('img')).not.toBeInTheDocument();
  });

  it('名前が空文字の場合?が表示される', () => {
    render(<Avatar name="" />);
    expect(screen.getByText('?')).toBeInTheDocument();
  });

  it('デフォルトでmdサイズが適用される', () => {
    const { container } = render(<Avatar name="A" />);
    const avatar = container.firstChild as HTMLElement;
    expect(avatar.className).toContain('w-10');
    expect(avatar.className).toContain('h-10');
  });

  it('smサイズが適用される', () => {
    const { container } = render(<Avatar name="A" size="sm" />);
    const avatar = container.firstChild as HTMLElement;
    expect(avatar.className).toContain('w-8');
    expect(avatar.className).toContain('h-8');
  });

  it('lgサイズが適用される', () => {
    const { container } = render(<Avatar name="A" size="lg" />);
    const avatar = container.firstChild as HTMLElement;
    expect(avatar.className).toContain('w-14');
    expect(avatar.className).toContain('h-14');
  });

  it('xlサイズが適用される', () => {
    const { container } = render(<Avatar name="A" size="xl" />);
    const avatar = container.firstChild as HTMLElement;
    expect(avatar.className).toContain('w-16');
    expect(avatar.className).toContain('h-16');
  });

  it('rounded-fullクラスが適用される', () => {
    const { container } = render(<Avatar name="A" />);
    const avatar = container.firstChild as HTMLElement;
    expect(avatar.className).toContain('rounded-full');
  });

  it('画像にobject-coverクラスが適用される', () => {
    render(<Avatar name="A" src="https://example.com/photo.jpg" />);
    const img = screen.getByRole('img');
    expect(img.className).toContain('object-cover');
  });
});
