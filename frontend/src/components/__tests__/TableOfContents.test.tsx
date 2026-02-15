import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import TableOfContents from '../TableOfContents';

describe('TableOfContents', () => {
  const headings = [
    { level: 1, text: 'タイトル', id: 'heading-0' },
    { level: 2, text: 'セクション1', id: 'heading-1' },
    { level: 3, text: 'サブセクション', id: 'heading-2' },
    { level: 2, text: 'セクション2', id: 'heading-3' },
  ];

  it('見出しリストを表示する', () => {
    render(<TableOfContents headings={headings} onHeadingClick={vi.fn()} />);
    expect(screen.getByText('タイトル')).toBeInTheDocument();
    expect(screen.getByText('セクション1')).toBeInTheDocument();
    expect(screen.getByText('サブセクション')).toBeInTheDocument();
    expect(screen.getByText('セクション2')).toBeInTheDocument();
  });

  it('見出しクリックでonHeadingClickが呼ばれる', () => {
    const onHeadingClick = vi.fn();
    render(<TableOfContents headings={headings} onHeadingClick={onHeadingClick} />);

    fireEvent.click(screen.getByText('セクション1'));
    expect(onHeadingClick).toHaveBeenCalledWith('heading-1');
  });

  it('「目次」タイトルが表示される', () => {
    render(<TableOfContents headings={headings} onHeadingClick={vi.fn()} />);
    expect(screen.getByText('目次')).toBeInTheDocument();
  });

  it('見出しがない場合は何も表示しない', () => {
    const { container } = render(<TableOfContents headings={[]} onHeadingClick={vi.fn()} />);
    expect(container.firstChild).toBeNull();
  });

  it('見出しレベルに応じてインデントが異なる', () => {
    render(<TableOfContents headings={headings} onHeadingClick={vi.fn()} />);
    const h1 = screen.getByText('タイトル');
    const h2 = screen.getByText('セクション1');
    const h3 = screen.getByText('サブセクション');

    expect(h1.closest('button')?.className).toContain('pl-0');
    expect(h2.closest('button')?.className).toContain('pl-4');
    expect(h3.closest('button')?.className).toContain('pl-8');
  });
});
