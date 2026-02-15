import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import CardHeading from '../CardHeading';

describe('CardHeading', () => {
  it('テキストが表示される', () => {
    render(<CardHeading>テストタイトル</CardHeading>);
    expect(screen.getByText('テストタイトル')).toBeInTheDocument();
  });

  it('正しいスタイルクラスが適用される', () => {
    render(<CardHeading>タイトル</CardHeading>);
    const heading = screen.getByText('タイトル');
    expect(heading.className).toContain('text-xs');
    expect(heading.className).toContain('font-medium');
  });

  it('p要素としてレンダリングされる', () => {
    render(<CardHeading>タイトル</CardHeading>);
    const heading = screen.getByText('タイトル');
    expect(heading.tagName).toBe('P');
  });
});
