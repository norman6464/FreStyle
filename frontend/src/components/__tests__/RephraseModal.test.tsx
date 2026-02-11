import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import RephraseModal from '../RephraseModal';

describe('RephraseModal', () => {
  const mockOnClose = vi.fn();
  const rephraseResult = {
    formal: 'フォーマルな言い換え文です。',
    soft: 'ソフトな言い換え文です。',
    concise: '簡潔な言い換え文です。',
  };

  beforeEach(() => {
    mockOnClose.mockClear();
  });

  it('3パターンの言い換え結果が表示される', () => {
    render(<RephraseModal result={rephraseResult} onClose={mockOnClose} />);

    expect(screen.getByText('フォーマル版')).toBeInTheDocument();
    expect(screen.getByText('ソフト版')).toBeInTheDocument();
    expect(screen.getByText('簡潔版')).toBeInTheDocument();
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
    expect(copyButtons).toHaveLength(3);
  });

  it('ローディング中はスピナーが表示される', () => {
    render(<RephraseModal result={null} onClose={mockOnClose} />);

    expect(screen.getByText('言い換え中...')).toBeInTheDocument();
  });
});
