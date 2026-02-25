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

  it('全オプションがボタン要素で4件描画される', () => {
    render(<SortSelector selected="default" onChange={mockOnChange} />);

    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(4);
  });

  it('連続クリックで各ソート値が正しく渡される', () => {
    render(<SortSelector selected="default" onChange={mockOnChange} />);

    fireEvent.click(screen.getByText('難易度↓'));
    fireEvent.click(screen.getByText('名前順'));
    fireEvent.click(screen.getByText('デフォルト'));

    expect(mockOnChange).toHaveBeenNthCalledWith(1, 'difficulty-desc');
    expect(mockOnChange).toHaveBeenNthCalledWith(2, 'name');
    expect(mockOnChange).toHaveBeenNthCalledWith(3, 'default');
  });

  it('rerenderで選択状態が正しく切り替わる', () => {
    const { rerender } = render(<SortSelector selected="default" onChange={mockOnChange} />);
    expect(screen.getByText('デフォルト').className).toContain('text-[var(--color-text-secondary)]');

    rerender(<SortSelector selected="difficulty-desc" onChange={mockOnChange} />);
    expect(screen.getByText('難易度↓').className).toContain('text-[var(--color-text-secondary)]');
    expect(screen.getByText('デフォルト').className).toContain('text-[var(--color-text-muted)]');
  });

  it('選択中のボタンにaria-pressed=trueが設定される', () => {
    render(<SortSelector selected="name" onChange={mockOnChange} />);

    expect(screen.getByText('名前順')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByText('デフォルト')).toHaveAttribute('aria-pressed', 'false');
  });

  it('ソートグループにrole=groupとaria-labelが設定される', () => {
    render(<SortSelector selected="default" onChange={mockOnChange} />);

    const group = screen.getByRole('group', { name: '並び替え' });
    expect(group).toBeInTheDocument();
  });
});
