import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import WeakAxisAdviceCard from '../WeakAxisAdviceCard';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

describe('WeakAxisAdviceCard', () => {
  beforeEach(() => {
    mockNavigate.mockClear();
  });

  it('おすすめ練習のタイトルが表示される', () => {
    render(<WeakAxisAdviceCard axis="論理的構成力" />);
    expect(screen.getByText('おすすめ練習')).toBeInTheDocument();
  });

  it('既知の軸に対応するアドバイスが表示される', () => {
    render(<WeakAxisAdviceCard axis="配慮表現" />);
    expect(screen.getByText('配慮表現を伸ばすシナリオで練習しましょう')).toBeInTheDocument();
  });

  it('未知の軸に対してもアドバイスが表示される', () => {
    render(<WeakAxisAdviceCard axis="カスタム軸" />);
    expect(screen.getByText('カスタム軸を伸ばすシナリオで練習しましょう')).toBeInTheDocument();
  });

  it('練習一覧ボタンが表示される', () => {
    render(<WeakAxisAdviceCard axis="論理的構成力" />);
    expect(screen.getByText('練習一覧を見る')).toBeInTheDocument();
  });

  it('ボタンクリックで練習ページに遷移する', () => {
    render(<WeakAxisAdviceCard axis="論理的構成力" />);
    fireEvent.click(screen.getByText('練習一覧を見る'));
    expect(mockNavigate).toHaveBeenCalledWith('/practice');
  });

  it('要約力のアドバイスが正しく表示される', () => {
    render(<WeakAxisAdviceCard axis="要約力" />);
    expect(screen.getByText('要約力を伸ばすシナリオで練習しましょう')).toBeInTheDocument();
  });
});
