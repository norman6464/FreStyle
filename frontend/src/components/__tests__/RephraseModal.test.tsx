import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import RephraseModal from '../RephraseModal';

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
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} />);

    expect(screen.getByText('フォーマル')).toBeInTheDocument();
    expect(screen.getByText('ソフト')).toBeInTheDocument();
    expect(screen.getByText('簡潔')).toBeInTheDocument();
    expect(screen.getByText('フォーマルな言い換え文です。')).toBeInTheDocument();
    expect(screen.getByText('ソフトな言い換え文です。')).toBeInTheDocument();
    expect(screen.getByText('簡潔な言い換え文です。')).toBeInTheDocument();
  });

  it('閉じるボタンをクリックするとonCloseが呼ばれる', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} />);

    fireEvent.click(screen.getByText('閉じる'));
    expect(mockOnClose).toHaveBeenCalled();
  });

  it('各パターンにコピーボタンが表示される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} />);

    const copyButtons = screen.getAllByText('コピー');
    expect(copyButtons).toHaveLength(5);
  });

  it('ローディング中はスピナーが表示される', () => {
    render(<RephraseModal result={null} onClose={mockOnClose} />);

    expect(screen.getByText('言い換え中...')).toBeInTheDocument();
  });

  it('質問型パターンが表示される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} />);

    expect(screen.getByText('質問型')).toBeInTheDocument();
    expect(screen.getByText('質問型の言い換え文です。')).toBeInTheDocument();
  });

  it('提案型パターンが表示される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} />);

    expect(screen.getByText('提案型')).toBeInTheDocument();
    expect(screen.getByText('提案型の言い換え文です。')).toBeInTheDocument();
  });

  it('5パターン全てにコピーボタンが表示される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} />);

    const copyButtons = screen.getAllByText('コピー');
    expect(copyButtons).toHaveLength(5);
  });
});
