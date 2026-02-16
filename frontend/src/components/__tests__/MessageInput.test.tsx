import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import MessageInput from '../MessageInput';

describe('MessageInput', () => {
  const mockOnSend = vi.fn();

  beforeEach(() => {
    mockOnSend.mockClear();
  });

  it('テキスト入力と送信ボタンが表示される', () => {
    render(<MessageInput onSend={mockOnSend} />);

    expect(screen.getByPlaceholderText('メッセージを入力...')).toBeInTheDocument();
    expect(screen.getByLabelText('送信')).toBeInTheDocument();
  });

  it('空テキストでは送信ボタンが無効', () => {
    render(<MessageInput onSend={mockOnSend} />);

    expect(screen.getByLabelText('送信')).toBeDisabled();
  });

  it('テキスト入力後に送信ボタンが有効になる', () => {
    render(<MessageInput onSend={mockOnSend} />);

    fireEvent.change(screen.getByPlaceholderText('メッセージを入力...'), {
      target: { value: 'テスト' },
    });
    expect(screen.getByLabelText('送信')).not.toBeDisabled();
  });

  it('送信中はインジケーターが表示される', () => {
    render(<MessageInput onSend={mockOnSend} isSending={true} />);

    expect(screen.getByText('送信中...')).toBeInTheDocument();
  });

  it('送信中は入力欄が無効化される', () => {
    render(<MessageInput onSend={mockOnSend} isSending={true} />);

    expect(screen.getByPlaceholderText('メッセージを入力...')).toBeDisabled();
  });

  it('送信中は送信ボタンが無効化される', () => {
    render(<MessageInput onSend={mockOnSend} isSending={true} />);

    expect(screen.getByLabelText('送信')).toBeDisabled();
  });

  it('送信完了後に入力欄にフォーカスが戻る', () => {
    const { rerender } = render(<MessageInput onSend={mockOnSend} isSending={true} />);

    rerender(<MessageInput onSend={mockOnSend} isSending={false} />);

    expect(screen.getByPlaceholderText('メッセージを入力...')).toHaveFocus();
  });

  it('Enterキーでメッセージが送信される', () => {
    render(<MessageInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText('メッセージを入力...');
    fireEvent.change(textarea, { target: { value: 'テスト送信' } });
    fireEvent.keyDown(textarea, { key: 'Enter', shiftKey: false });

    expect(mockOnSend).toHaveBeenCalledWith('テスト送信');
  });

  it('Shift+Enterでは送信されない', () => {
    render(<MessageInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText('メッセージを入力...');
    fireEvent.change(textarea, { target: { value: 'テスト' } });
    fireEvent.keyDown(textarea, { key: 'Enter', shiftKey: true });

    expect(mockOnSend).not.toHaveBeenCalled();
  });

  it('IME入力中はEnterで送信されない', () => {
    render(<MessageInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText('メッセージを入力...');
    fireEvent.change(textarea, { target: { value: 'テスト' } });
    fireEvent.compositionStart(textarea);
    fireEvent.keyDown(textarea, { key: 'Enter', shiftKey: false });

    expect(mockOnSend).not.toHaveBeenCalled();
  });

  it('IME確定後はEnterで送信される', () => {
    render(<MessageInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText('メッセージを入力...');
    fireEvent.change(textarea, { target: { value: 'テスト' } });
    fireEvent.compositionStart(textarea);
    fireEvent.compositionEnd(textarea);
    fireEvent.keyDown(textarea, { key: 'Enter', shiftKey: false });

    expect(mockOnSend).toHaveBeenCalledWith('テスト');
  });

  it('空テキストではEnterで送信されない', () => {
    render(<MessageInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText('メッセージを入力...');
    fireEvent.keyDown(textarea, { key: 'Enter', shiftKey: false });

    expect(mockOnSend).not.toHaveBeenCalled();
  });

  it('送信後にテキストがクリアされる', () => {
    render(<MessageInput onSend={mockOnSend} />);

    const textarea = screen.getByPlaceholderText('メッセージを入力...');
    fireEvent.change(textarea, { target: { value: '送信テスト' } });
    fireEvent.keyDown(textarea, { key: 'Enter', shiftKey: false });

    expect(textarea).toHaveValue('');
  });

  it('送信ボタンクリックでメッセージが送信される', () => {
    render(<MessageInput onSend={mockOnSend} />);

    fireEvent.change(screen.getByPlaceholderText('メッセージを入力...'), {
      target: { value: 'ボタン送信' },
    });
    fireEvent.click(screen.getByLabelText('送信'));

    expect(mockOnSend).toHaveBeenCalledWith('ボタン送信');
  });
});
