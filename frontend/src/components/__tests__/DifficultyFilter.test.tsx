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

  it('非選択のボタンはmutedスタイルが適用される', () => {
    render(<DifficultyFilter selected="初級" onChange={mockOnChange} />);

    const unselectedButton = screen.getByText('上級');
    expect(unselectedButton.className).toContain('text-[var(--color-text-muted)]');
    expect(unselectedButton.className).not.toContain('text-rose-400');
  });

  it('各難易度で正しいカラーが適用される', () => {
    const { rerender } = render(<DifficultyFilter selected="初級" onChange={mockOnChange} />);
    expect(screen.getByText('初級').className).toContain('text-emerald-400');

    rerender(<DifficultyFilter selected="上級" onChange={mockOnChange} />);
    expect(screen.getByText('上級').className).toContain('text-rose-400');
  });

  it('DIFFICULTY_OPTIONSの件数とラベルが定数と一致する', () => {
    render(<DifficultyFilter selected={null} onChange={mockOnChange} />);

    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(4);
    expect(buttons.map((b) => b.textContent)).toEqual(['全レベル', '初級', '中級', '上級']);
  });

  it('選択中のボタンにaria-pressed=trueが設定される', () => {
    render(<DifficultyFilter selected="初級" onChange={mockOnChange} />);

    expect(screen.getByText('初級')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByText('中級')).toHaveAttribute('aria-pressed', 'false');
    expect(screen.getByText('上級')).toHaveAttribute('aria-pressed', 'false');
  });

  it('フィルターグループにrole=groupとaria-labelが設定される', () => {
    render(<DifficultyFilter selected={null} onChange={mockOnChange} />);

    const group = screen.getByRole('group', { name: '難易度フィルター' });
    expect(group).toBeInTheDocument();
  });
});
