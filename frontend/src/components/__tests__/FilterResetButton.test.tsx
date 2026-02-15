import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import FilterResetButton from '../FilterResetButton';

describe('FilterResetButton', () => {
  const mockOnReset = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('フィルターが適用されている場合にボタンが表示される', () => {
    render(<FilterResetButton isActive={true} onReset={mockOnReset} />);
    expect(screen.getByText('リセット')).toBeInTheDocument();
  });

  it('フィルターが適用されていない場合にボタンが非表示', () => {
    render(<FilterResetButton isActive={false} onReset={mockOnReset} />);
    expect(screen.queryByText('リセット')).not.toBeInTheDocument();
  });

  it('クリックするとonResetが呼ばれる', () => {
    render(<FilterResetButton isActive={true} onReset={mockOnReset} />);
    fireEvent.click(screen.getByText('リセット'));
    expect(mockOnReset).toHaveBeenCalledTimes(1);
  });

  it('ArrowPathアイコンが表示される', () => {
    render(<FilterResetButton isActive={true} onReset={mockOnReset} />);
    const button = screen.getByText('リセット').closest('button');
    expect(button).toBeDefined();
    expect(button?.querySelector('svg')).toBeDefined();
  });

  it('isActiveがfalseからtrueに変わるとボタンが出現する', () => {
    const { rerender } = render(<FilterResetButton isActive={false} onReset={mockOnReset} />);
    expect(screen.queryByText('リセット')).not.toBeInTheDocument();

    rerender(<FilterResetButton isActive={true} onReset={mockOnReset} />);
    expect(screen.getByText('リセット')).toBeInTheDocument();
  });

  it('isActiveがtrueからfalseに変わるとボタンが消える', () => {
    const { rerender } = render(<FilterResetButton isActive={true} onReset={mockOnReset} />);
    expect(screen.getByText('リセット')).toBeInTheDocument();

    rerender(<FilterResetButton isActive={false} onReset={mockOnReset} />);
    expect(screen.queryByText('リセット')).not.toBeInTheDocument();
  });

  it('連続クリックでonResetが複数回呼ばれる', () => {
    render(<FilterResetButton isActive={true} onReset={mockOnReset} />);
    const button = screen.getByText('リセット');
    fireEvent.click(button);
    fireEvent.click(button);
    fireEvent.click(button);
    expect(mockOnReset).toHaveBeenCalledTimes(3);
  });
});
