import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import MenuNavigationCard from '../MenuNavigationCard';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

describe('MenuNavigationCard', () => {
  beforeEach(() => {
    mockNavigate.mockClear();
  });

  it('メニュー項目が全て表示される', () => {
    render(<MenuNavigationCard latestScore={null} />);
    expect(screen.getByText('AI アシスタント')).toBeInTheDocument();
    expect(screen.getByText('練習モード')).toBeInTheDocument();
    expect(screen.getByText('スコア履歴')).toBeInTheDocument();
  });

  it('説明文が表示される', () => {
    render(<MenuNavigationCard latestScore={null} />);
    expect(screen.getByText('AIにコミュニケーションを分析・フィードバック')).toBeInTheDocument();
    expect(screen.getByText('ビジネスシナリオでロールプレイ練習')).toBeInTheDocument();
    expect(screen.getByText('フィードバック結果の振り返り')).toBeInTheDocument();
  });

  it('最新スコアバッジが表示される', () => {
    render(<MenuNavigationCard latestScore={8.5} />);
    expect(screen.getByText('最新: 8.5')).toBeInTheDocument();
  });

  it('最新スコアがnullの場合バッジが非表示', () => {
    render(<MenuNavigationCard latestScore={null} />);
    expect(screen.queryByText(/最新:/)).toBeNull();
  });

  it('メニュー項目クリックで画面遷移する', () => {
    render(<MenuNavigationCard latestScore={null} />);
    fireEvent.click(screen.getByText('AI アシスタント'));
    expect(mockNavigate).toHaveBeenCalledWith('/chat/ask-ai');
  });

  it('練習モードクリックで画面遷移する', () => {
    render(<MenuNavigationCard latestScore={null} />);
    fireEvent.click(screen.getByText('練習モード'));
    expect(mockNavigate).toHaveBeenCalledWith('/practice');
  });

  it('8つのメニュー項目のボタンが存在する', () => {
    render(<MenuNavigationCard latestScore={null} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(8);
  });

  it('Enterキーで画面遷移する', () => {
    render(<MenuNavigationCard latestScore={null} />);
    const aiButton = screen.getByRole('button', { name: 'AI アシスタント' });
    fireEvent.keyDown(aiButton, { key: 'Enter' });
    expect(mockNavigate).toHaveBeenCalledWith('/chat/ask-ai');
  });

  it('Spaceキーで画面遷移する', () => {
    render(<MenuNavigationCard latestScore={null} />);
    const practiceButton = screen.getByRole('button', { name: '練習モード' });
    fireEvent.keyDown(practiceButton, { key: ' ' });
    expect(mockNavigate).toHaveBeenCalledWith('/practice');
  });

  it('aria-labelが各メニュー項目に設定されている', () => {
    render(<MenuNavigationCard latestScore={null} />);
    expect(screen.getByRole('button', { name: 'AI アシスタント' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '練習モード' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'スコア履歴' })).toBeInTheDocument();
  });
});
