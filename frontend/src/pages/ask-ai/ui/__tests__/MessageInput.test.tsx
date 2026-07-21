import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import MessageInput from '../MessageInput';

vi.mock('@/entities/ai-chat/api/aiChatRepository', () => ({
  default: {
    issueAttachmentUploadUrl: vi.fn(),
  },
}));

import aiChatRepository from '@/entities/ai-chat/api/aiChatRepository';

describe('MessageInput', () => {
  const mockOnSend = vi.fn();

  beforeEach(() => {
    mockOnSend.mockClear();
    vi.mocked(aiChatRepository.issueAttachmentUploadUrl).mockReset();
    if (!('createObjectURL' in URL)) {
      // happy-dom fallback: 一部バージョンで未定義
      (URL as unknown as { createObjectURL?: (b: Blob) => string }).createObjectURL = vi.fn(
        () => 'blob:mock'
      );
      (URL as unknown as { revokeObjectURL?: (u: string) => void }).revokeObjectURL = vi.fn();
    }
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

    expect(mockOnSend).toHaveBeenCalledWith('テスト送信', []);
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

    expect(mockOnSend).toHaveBeenCalledWith('テスト', []);
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

    expect(mockOnSend).toHaveBeenCalledWith('ボタン送信', []);
  });

  it('画像を選択すると presigned URL を取得して S3 へ PUT する', async () => {
    vi.mocked(aiChatRepository.issueAttachmentUploadUrl).mockResolvedValue({
      uploadUrl: 'https://s3.example.com/put',
      key: 'ai-chat/7/abc.png',
      expiresIn: 600,
    });
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(null, { status: 200 })
    );

    render(<MessageInput onSend={mockOnSend} />);

    const file = new File(['fake-bytes'], 'cat.png', { type: 'image/png' });
    const fileInput = screen.getByLabelText('添付ファイルを選択') as HTMLInputElement;
    fireEvent.change(fileInput, { target: { files: [file] } });

    await waitFor(() => {
      expect(aiChatRepository.issueAttachmentUploadUrl).toHaveBeenCalledWith({
        filename: 'cat.png',
        contentType: 'image/png',
        sizeBytes: file.size,
      });
    });
    await waitFor(() => {
      expect(fetchSpy).toHaveBeenCalledWith(
        'https://s3.example.com/put',
        expect.objectContaining({ method: 'PUT' })
      );
    });

    // 送信ボタンをクリックすると、presigned URL の key で onSend が呼ばれる
    fireEvent.click(screen.getByLabelText('送信'));
    await waitFor(() => {
      expect(mockOnSend).toHaveBeenCalledWith(
        '',
        expect.arrayContaining([
          expect.objectContaining({ key: 'ai-chat/7/abc.png', filename: 'cat.png' }),
        ])
      );
    });

    fetchSpy.mockRestore();
  });

  it('未対応の MIME はバリデーションエラーで弾かれ presigned URL を呼ばない', async () => {
    render(<MessageInput onSend={mockOnSend} />);

    const file = new File(['x'], 'a.exe', { type: 'application/x-msdownload' });
    const fileInput = screen.getByLabelText('添付ファイルを選択') as HTMLInputElement;
    fireEvent.change(fileInput, { target: { files: [file] } });

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
    expect(aiChatRepository.issueAttachmentUploadUrl).not.toHaveBeenCalled();
  });
});
