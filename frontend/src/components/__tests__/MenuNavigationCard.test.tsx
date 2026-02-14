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
    render(<MenuNavigationCard totalUnread={0} latestScore={null} />);
    expect(screen.getByText('チャット')).toBeInTheDocument();
    expect(screen.getByText('AI アシスタント')).toBeInTheDocument();
    expect(screen.getByText('練習モード')).toBeInTheDocument();
    expect(screen.getByText('スコア履歴')).toBeInTheDocument();
  });

  it('説明文が表示される', () => {
    render(<MenuNavigationCard totalUnread={0} latestScore={null} />);
    expect(screen.getByText('メンバーとメッセージをやり取り')).toBeInTheDocument();
    expect(screen.getByText('AIにコミュニケーションを分析・フィードバック')).toBeInTheDocument();
    expect(screen.getByText('ビジネスシナリオでロールプレイ練習')).toBeInTheDocument();
    expect(screen.getByText('フィードバック結果の振り返り')).toBeInTheDocument();
  });

  it('未読バッジが表示される', () => {
    render(<MenuNavigationCard totalUnread={5} latestScore={null} />);
    expect(screen.getByText('5件の未読')).toBeInTheDocument();
  });

  it('未読0件ではバッジが非表示', () => {
    render(<MenuNavigationCard totalUnread={0} latestScore={null} />);
    expect(screen.queryByText(/件の未読/)).toBeNull();
  });

  it('最新スコアバッジが表示される', () => {
    render(<MenuNavigationCard totalUnread={0} latestScore={8.5} />);
    expect(screen.getByText('最新: 8.5')).toBeInTheDocument();
  });

  it('最新スコアがnullの場合バッジが非表示', () => {
    render(<MenuNavigationCard totalUnread={0} latestScore={null} />);
    expect(screen.queryByText(/最新:/)).toBeNull();
  });

  it('メニュー項目クリックで画面遷移する', () => {
    render(<MenuNavigationCard totalUnread={0} latestScore={null} />);
    fireEvent.click(screen.getByText('チャット'));
    expect(mockNavigate).toHaveBeenCalledWith('/chat');
  });

  it('練習モードクリックで画面遷移する', () => {
    render(<MenuNavigationCard totalUnread={0} latestScore={null} />);
    fireEvent.click(screen.getByText('練習モード'));
    expect(mockNavigate).toHaveBeenCalledWith('/practice');
  });

  it('4つのメニュー項目のボタンが存在する', () => {
    render(<MenuNavigationCard totalUnread={0} latestScore={null} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(4);
  });
});
