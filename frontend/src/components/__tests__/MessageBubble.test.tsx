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

  it('コピーボタンが非削除メッセージに表示される', () => {
    const mockCopy = vi.fn();
    render(<MessageBubble isSender={false} content="テスト" id={1} onCopy={mockCopy} />);

    expect(screen.getByTitle('コピー')).toBeInTheDocument();
  });

  it('コピーボタンクリックでonCopyが呼ばれる', () => {
    const mockCopy = vi.fn();
    render(<MessageBubble isSender={false} content="コピー対象" id={5} onCopy={mockCopy} />);

    fireEvent.click(screen.getByTitle('コピー'));
    expect(mockCopy).toHaveBeenCalledWith(5, 'コピー対象');
  });

  it('削除済みメッセージにコピーボタンが表示されない', () => {
    const mockCopy = vi.fn();
    render(<MessageBubble isSender={true} content="テスト" id={1} isDeleted={true} onCopy={mockCopy} />);

    expect(screen.queryByTitle('コピー')).not.toBeInTheDocument();
  });

  it('コピー済み状態でチェックアイコンが表示される', () => {
    const mockCopy = vi.fn();
    render(<MessageBubble isSender={false} content="テスト" id={1} onCopy={mockCopy} isCopied={true} />);

    expect(screen.getByTitle('コピー済み')).toBeInTheDocument();
  });

  it('送信者メッセージでもコピーボタンが表示される', () => {
    const mockCopy = vi.fn();
    render(<MessageBubble isSender={true} content="テスト" id={1} onCopy={mockCopy} />);

    expect(screen.getByTitle('コピー')).toBeInTheDocument();
  });

  it('onCopyがnullの場合コピーボタンが表示されない', () => {
    render(<MessageBubble isSender={false} content="テスト" id={1} onCopy={null} />);

    expect(screen.queryByTitle('コピー')).not.toBeInTheDocument();
  });

  it('onCopy未指定の場合コピーボタンが表示されない', () => {
    render(<MessageBubble isSender={false} content="テスト" id={1} />);

    expect(screen.queryByTitle('コピー')).not.toBeInTheDocument();
  });

  it('画像メッセージが表示される', () => {
    render(<MessageBubble isSender={false} content="https://example.com/img.jpg" id={1} type="image" />);

    const img = screen.getByAltText('画像');
    expect(img).toBeInTheDocument();
  });

  it('時刻が正しく表示される', () => {
    render(<MessageBubble isSender={false} content="テスト" id={1} createdAt="2026-02-14T10:30:00" />);

    expect(screen.getByText('10:30')).toBeInTheDocument();
  });

  it('メッセージにarticleロールとaria-labelがある', () => {
    render(<MessageBubble isSender={false} content="こんにちは" id={1} senderName="田中太郎" />);

    const article = screen.getByRole('article');
    expect(article).toBeInTheDocument();
    expect(article).toHaveAttribute('aria-label', '田中太郎のメッセージ');
  });

  it('送信者メッセージにaria-labelがある', () => {
    render(<MessageBubble isSender={true} content="テスト" id={1} />);

    expect(screen.getByRole('article')).toHaveAttribute('aria-label', '自分のメッセージ');
  });

  it('削除ボタンにaria-labelがある', () => {
    const mockDelete = vi.fn();
    render(<MessageBubble isSender={true} content="テスト" id={1} onDelete={mockDelete} />);

    // ホバーで削除ボタン表示
    fireEvent.mouseEnter(screen.getByRole('article').parentElement!);
    expect(screen.getByLabelText('メッセージを削除')).toBeInTheDocument();
  });

  it('言い換えボタンにaria-labelがある', () => {
    render(<MessageBubble isSender={true} content="テスト" id={1} onRephrase={vi.fn()} />);

    expect(screen.getByLabelText('メッセージを言い換え')).toBeInTheDocument();
  });
});
