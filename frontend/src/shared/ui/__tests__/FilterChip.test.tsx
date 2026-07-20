import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import FilterChip from '../FilterChip';

describe('FilterChip', () => {
  it('label を表示し aria-pressed で選択状態を公開する', () => {
    render(<FilterChip label="PHP" active={true} onClick={() => {}} />);
    const chip = screen.getByRole('button', { name: 'PHP' });
    expect(chip).toHaveAttribute('aria-pressed', 'true');
  });

  it('非 active のときは aria-pressed=false', () => {
    render(<FilterChip label="Go" active={false} onClick={() => {}} />);
    expect(screen.getByRole('button', { name: 'Go' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('クリックで onClick を呼ぶ', () => {
    const onClick = vi.fn();
    render(<FilterChip label="Docker" active={false} onClick={onClick} />);
    fireEvent.click(screen.getByRole('button', { name: 'Docker' }));
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('active 時に activeClass を適用する(カテゴリ色)', () => {
    render(
      <FilterChip label="データベース" active={true} activeClass="bg-emerald-500/15" onClick={() => {}} />,
    );
    expect(screen.getByRole('button', { name: 'データベース' })).toHaveClass('bg-emerald-500/15');
  });
});
