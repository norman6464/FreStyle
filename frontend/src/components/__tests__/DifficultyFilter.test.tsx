import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import DifficultyFilter from '../DifficultyFilter';

describe('DifficultyFilter', () => {
  const mockOnChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('全ての難易度ボタンが表示される', () => {
    render(<DifficultyFilter selected={null} onChange={mockOnChange} />);

    expect(screen.getByText('全レベル')).toBeInTheDocument();
    expect(screen.getByText('初級')).toBeInTheDocument();
    expect(screen.getByText('中級')).toBeInTheDocument();
    expect(screen.getByText('上級')).toBeInTheDocument();
  });

  it('難易度ボタンをクリックするとonChangeが呼ばれる', () => {
    render(<DifficultyFilter selected={null} onChange={mockOnChange} />);

    fireEvent.click(screen.getByText('初級'));
    expect(mockOnChange).toHaveBeenCalledWith('初級');
  });

  it('全レベルをクリックするとnullが渡される', () => {
    render(<DifficultyFilter selected="初級" onChange={mockOnChange} />);

    fireEvent.click(screen.getByText('全レベル'));
    expect(mockOnChange).toHaveBeenCalledWith(null);
  });

  it('選択中の難易度がハイライトされる', () => {
    render(<DifficultyFilter selected="中級" onChange={mockOnChange} />);

    const selectedButton = screen.getByText('中級');
    expect(selectedButton.className).toContain('text-amber-400');
  });
});
