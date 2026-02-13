import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import InputField from '../InputField';

describe('InputField', () => {
  const mockOnChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('ラベルが表示される', () => {
    render(<InputField label="メール" name="email" value="" onChange={mockOnChange} />);

    expect(screen.getByLabelText('メール')).toBeInTheDocument();
  });

  it('入力値が変更されるとonChangeが呼ばれる', () => {
    render(<InputField label="メール" name="email" value="" onChange={mockOnChange} />);

    fireEvent.change(screen.getByLabelText('メール'), { target: { value: 'test@example.com' } });
    expect(mockOnChange).toHaveBeenCalled();
  });

  it('値があるときクリアボタンが表示される', () => {
    render(<InputField label="メール" name="email" value="test" onChange={mockOnChange} />);

    const clearButton = screen.getByRole('button');
    expect(clearButton).toBeInTheDocument();
  });

  it('クリアボタンで入力値がリセットされる', () => {
    render(<InputField label="メール" name="email" value="test" onChange={mockOnChange} />);

    fireEvent.click(screen.getByRole('button'));
    expect(mockOnChange).toHaveBeenCalledWith({ target: { name: 'email', value: '' } });
  });
});
