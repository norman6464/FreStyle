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

  it('空の値ではクリアボタンが表示されない', () => {
    render(<InputField label="メール" name="email" value="" onChange={mockOnChange} />);
    expect(screen.queryByRole('button')).toBeNull();
  });

  it('デフォルトのtypeがtextである', () => {
    render(<InputField label="名前" name="name" value="" onChange={mockOnChange} />);
    expect(screen.getByLabelText('名前')).toHaveAttribute('type', 'text');
  });

  it('エラーメッセージが表示される', () => {
    render(<InputField label="メール" name="email" value="" onChange={mockOnChange} error="必須項目です" />);
    expect(screen.getByText('必須項目です')).toBeInTheDocument();
  });

  it('エラー時にaria-invalidがtrueになる', () => {
    render(<InputField label="メール" name="email" value="" onChange={mockOnChange} error="エラー" />);
    expect(screen.getByLabelText('メール')).toHaveAttribute('aria-invalid', 'true');
  });

  it('エラー時にaria-describedbyが設定される', () => {
    render(<InputField label="メール" name="email" value="" onChange={mockOnChange} error="エラー" />);
    expect(screen.getByLabelText('メール')).toHaveAttribute('aria-describedby', 'email-error');
    expect(screen.getByText('エラー')).toHaveAttribute('id', 'email-error');
  });

  it('エラーがない場合はエラーメッセージが表示されない', () => {
    render(<InputField label="メール" name="email" value="" onChange={mockOnChange} />);
    expect(screen.queryByText('必須項目です')).not.toBeInTheDocument();
  });

  it('エラー時にボーダーが赤色になる', () => {
    render(<InputField label="メール" name="email" value="" onChange={mockOnChange} error="エラー" />);
    const input = screen.getByLabelText('メール');
    expect(input.className).toContain('border-rose-500');
  });
});
