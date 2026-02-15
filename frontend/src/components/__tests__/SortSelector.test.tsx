import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import SortSelector from '../SortSelector';

describe('SortSelector', () => {
  const mockOnChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('全てのソートオプションが表示される', () => {
    render(<SortSelector selected="default" onChange={mockOnChange} />);

    expect(screen.getByText('デフォルト')).toBeInTheDocument();
    expect(screen.getByText('難易度↑')).toBeInTheDocument();
    expect(screen.getByText('難易度↓')).toBeInTheDocument();
    expect(screen.getByText('名前順')).toBeInTheDocument();
  });

  it('ソートオプションをクリックするとonChangeが呼ばれる', () => {
    render(<SortSelector selected="default" onChange={mockOnChange} />);

    fireEvent.click(screen.getByText('難易度↑'));
    expect(mockOnChange).toHaveBeenCalledWith('difficulty-asc');
  });

  it('選択中のオプションがハイライトされる', () => {
    render(<SortSelector selected="name" onChange={mockOnChange} />);

    const selectedButton = screen.getByText('名前順');
    expect(selectedButton.className).toContain('text-[var(--color-text-secondary)]');
  });

  it('非選択のオプションはmutedスタイルが適用される', () => {
    render(<SortSelector selected="default" onChange={mockOnChange} />);

    const unselectedButton = screen.getByText('名前順');
    expect(unselectedButton.className).toContain('text-[var(--color-text-muted)]');
  });
});
