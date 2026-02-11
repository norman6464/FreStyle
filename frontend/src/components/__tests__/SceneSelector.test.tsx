import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import SceneSelector from '../SceneSelector';

describe('SceneSelector', () => {
  const mockOnSelect = vi.fn();
  const mockOnCancel = vi.fn();

  beforeEach(() => {
    mockOnSelect.mockClear();
    mockOnCancel.mockClear();
  });

  it('5つのシーン選択肢が表示される', () => {
    render(<SceneSelector onSelect={mockOnSelect} onCancel={mockOnCancel} />);

    expect(screen.getByText('会議')).toBeInTheDocument();
    expect(screen.getByText('1on1')).toBeInTheDocument();
    expect(screen.getByText('メール')).toBeInTheDocument();
    expect(screen.getByText('プレゼン')).toBeInTheDocument();
    expect(screen.getByText('商談')).toBeInTheDocument();
  });

  it('シーンをクリックするとonSelectが呼ばれる', () => {
    render(<SceneSelector onSelect={mockOnSelect} onCancel={mockOnCancel} />);

    fireEvent.click(screen.getByText('会議'));
    expect(mockOnSelect).toHaveBeenCalledWith('meeting');
  });

  it('1on1をクリックするとone_on_oneが渡される', () => {
    render(<SceneSelector onSelect={mockOnSelect} onCancel={mockOnCancel} />);

    fireEvent.click(screen.getByText('1on1'));
    expect(mockOnSelect).toHaveBeenCalledWith('one_on_one');
  });

  it('キャンセルボタンをクリックするとonCancelが呼ばれる', () => {
    render(<SceneSelector onSelect={mockOnSelect} onCancel={mockOnCancel} />);

    fireEvent.click(screen.getByText('スキップ'));
    expect(mockOnCancel).toHaveBeenCalled();
  });

  it('各シーンに説明文が表示される', () => {
    render(<SceneSelector onSelect={mockOnSelect} onCancel={mockOnCancel} />);

    expect(screen.getByText(/発言のタイミング/)).toBeInTheDocument();
    expect(screen.getByText(/心理的安全性/)).toBeInTheDocument();
    expect(screen.getByText(/件名の明確さ/)).toBeInTheDocument();
    expect(screen.getByText(/ストーリー構成/)).toBeInTheDocument();
    expect(screen.getByText(/ニーズヒアリング/)).toBeInTheDocument();
  });
});
