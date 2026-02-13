import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import MessageBubble from '../MessageBubble';

describe('MessageBubble', () => {
  it('テキストメッセージが表示される', () => {
    render(<MessageBubble isSender={false} content="こんにちは" id={1} />);

    expect(screen.getByText('こんにちは')).toBeInTheDocument();
  });

  it('送信者名が表示される（受信側）', () => {
    render(<MessageBubble isSender={false} content="テスト" id={1} senderName="田中太郎" />);

    expect(screen.getByText('田中太郎')).toBeInTheDocument();
  });

  it('送信側では送信者名が表示されない', () => {
    render(<MessageBubble isSender={true} content="テスト" id={1} senderName="田中太郎" />);

    expect(screen.queryByText('田中太郎')).not.toBeInTheDocument();
  });

  it('削除済みメッセージが表示される', () => {
    render(<MessageBubble isSender={true} content="テスト" id={1} isDeleted={true} />);

    expect(screen.getByText('メッセージを削除しました')).toBeInTheDocument();
  });

  it('言い換えボタンが送信者メッセージに表示される', () => {
    const mockRephrase = vi.fn();
    render(<MessageBubble isSender={true} content="テスト" id={1} onRephrase={mockRephrase} />);

    expect(screen.getByText('言い換え')).toBeInTheDocument();
  });

  it('言い換えボタンクリックでonRephraseが呼ばれる', () => {
    const mockRephrase = vi.fn();
    render(<MessageBubble isSender={true} content="テストメッセージ" id={1} onRephrase={mockRephrase} />);

    fireEvent.click(screen.getByText('言い換え'));
    expect(mockRephrase).toHaveBeenCalledWith('テストメッセージ');
  });
});
