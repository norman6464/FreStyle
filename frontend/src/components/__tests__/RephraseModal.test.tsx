import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import RephraseModal from '../RephraseModal';

vi.mock('../../hooks/useFavoritePhrase', () => ({
  useFavoritePhrase: () => ({
    saveFavorite: vi.fn(),
    removeFavorite: vi.fn(),
    isFavorite: vi.fn().mockReturnValue(false),
    phrases: [],
  }),
}));

describe('RephraseModal', () => {
  const mockOnClose = vi.fn();
  const rephraseResult = {
    formal: 'フォーマルな言い換え文です。',
    soft: 'ソフトな言い換え文です。',
    concise: '簡潔な言い換え文です。',
    questioning: '質問型の言い換え文です。',
    proposal: '提案型の言い換え文です。',
  };

  beforeEach(() => {
    mockOnClose.mockClear();
  });

  it('3パターンの言い換え結果が表示される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} originalText="元のテキスト" />);

    expect(screen.getByText('フォーマル')).toBeInTheDocument();
    expect(screen.getByText('ソフト')).toBeInTheDocument();
    expect(screen.getByText('簡潔')).toBeInTheDocument();
    expect(screen.getByText('フォーマルな言い換え文です。')).toBeInTheDocument();
    expect(screen.getByText('ソフトな言い換え文です。')).toBeInTheDocument();
    expect(screen.getByText('簡潔な言い換え文です。')).toBeInTheDocument();
  });

  it('閉じるボタンをクリックするとonCloseが呼ばれる', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} originalText="元のテキスト" />);

    fireEvent.click(screen.getByText('閉じる'));
    expect(mockOnClose).toHaveBeenCalled();
  });

  it('各パターンにコピーボタンが表示される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} originalText="元のテキスト" />);

    const copyButtons = screen.getAllByText('コピー');
    expect(copyButtons).toHaveLength(5);
  });

  it('ローディング中はスピナーが表示される', () => {
    render(<RephraseModal result={null} onClose={mockOnClose} originalText="元のテキスト" />);

    expect(screen.getByText('言い換え中...')).toBeInTheDocument();
  });

  it('質問型パターンが表示される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} originalText="元のテキスト" />);

    expect(screen.getByText('質問型')).toBeInTheDocument();
    expect(screen.getByText('質問型の言い換え文です。')).toBeInTheDocument();
  });

  it('提案型パターンが表示される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} originalText="元のテキスト" />);

    expect(screen.getByText('提案型')).toBeInTheDocument();
    expect(screen.getByText('提案型の言い換え文です。')).toBeInTheDocument();
  });

  it('5パターン全てにコピーボタンが表示される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} originalText="元のテキスト" />);

    const copyButtons = screen.getAllByText('コピー');
    expect(copyButtons).toHaveLength(5);
  });

  it('各パターンに利用シーンの説明が表示される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} originalText="元のテキスト" />);

    expect(screen.getByText(/上司や顧客への報告/)).toBeInTheDocument();
    expect(screen.getByText(/指摘やお願いをする時/)).toBeInTheDocument();
    expect(screen.getByText(/チャットやSlackで/)).toBeInTheDocument();
  });

  it('各パターンにお気に入りボタンが表示される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} originalText="元のテキスト" />);

    const favButtons = screen.getAllByLabelText('お気に入りに追加');
    expect(favButtons).toHaveLength(5);
  });

  it('ESCキーでonCloseが呼ばれる', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} originalText="元のテキスト" />);

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(mockOnClose).toHaveBeenCalled();
  });

  it('コピーボタンクリックでフィードバックが表示される', async () => {
    // clipboard mockを設定
    Object.assign(navigator, {
      clipboard: { writeText: vi.fn().mockResolvedValue(undefined) },
    });

    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} originalText="元のテキスト" />);

    const copyButtons = screen.getAllByText('コピー');
    fireEvent.click(copyButtons[0]);

    // コピー成功フィードバック
    expect(await screen.findByText('コピーしました')).toBeInTheDocument();
  });

  it('モーダルにrole=dialogとaria-modal属性が設定される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} originalText="元のテキスト" />);
    const dialog = screen.getByRole('dialog');
    expect(dialog).toHaveAttribute('aria-modal', 'true');
    expect(dialog).toHaveAttribute('aria-labelledby');
  });
});
