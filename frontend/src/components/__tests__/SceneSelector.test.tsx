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

  it('コードレビューシーンが表示される', () => {
    render(<SceneSelector onSelect={mockOnSelect} onCancel={mockOnCancel} />);

    expect(screen.getByText('コードレビュー')).toBeInTheDocument();
    fireEvent.click(screen.getByText('コードレビュー'));
    expect(mockOnSelect).toHaveBeenCalledWith('code_review');
  });

  it('障害対応シーンが表示される', () => {
    render(<SceneSelector onSelect={mockOnSelect} onCancel={mockOnCancel} />);

    expect(screen.getByText('障害対応')).toBeInTheDocument();
    fireEvent.click(screen.getByText('障害対応'));
    expect(mockOnSelect).toHaveBeenCalledWith('incident');
  });

  it('日報シーンが表示される', () => {
    render(<SceneSelector onSelect={mockOnSelect} onCancel={mockOnCancel} />);

    expect(screen.getByText('日報・週報')).toBeInTheDocument();
    fireEvent.click(screen.getByText('日報・週報'));
    expect(mockOnSelect).toHaveBeenCalledWith('daily_report');
  });

  it('全8シーンが表示される', () => {
    render(<SceneSelector onSelect={mockOnSelect} onCancel={mockOnCancel} />);

    const buttons = screen.getAllByRole('button');
    // 8 scene buttons + 1 skip button = 9
    expect(buttons).toHaveLength(9);
  });

  it('おすすめシーンに「おすすめ」ラベルが表示される', () => {
    render(<SceneSelector onSelect={mockOnSelect} onCancel={mockOnCancel} />);

    // 新卒におすすめの会議・コードレビュー・日報にラベルがつく
    const badges = screen.getAllByText('おすすめ');
    expect(badges.length).toBeGreaterThanOrEqual(1);
  });

  it('具体的な利用例が表示される', () => {
    render(<SceneSelector onSelect={mockOnSelect} onCancel={mockOnCancel} />);

    expect(screen.getByText(/朝会やスプリントレビューで/)).toBeInTheDocument();
    expect(screen.getByText(/PRへのコメントで/)).toBeInTheDocument();
  });

  it('カテゴリヘッダーが表示される', () => {
    render(<SceneSelector onSelect={mockOnSelect} onCancel={mockOnCancel} />);

    expect(screen.getByText('日常業務')).toBeInTheDocument();
    expect(screen.getByText('対面コミュニケーション')).toBeInTheDocument();
  });
});
