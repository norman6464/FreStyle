import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import CommunicationTipCard from '../CommunicationTipCard';

describe('CommunicationTipCard', () => {
  it('Tipsテキストが表示される', () => {
    render(<CommunicationTipCard />);

    expect(screen.getByTestId('tip-text')).toBeInTheDocument();
  });

  it('カテゴリラベルが表示される', () => {
    render(<CommunicationTipCard />);

    expect(screen.getByTestId('tip-category')).toBeInTheDocument();
  });

  it('タイトルが表示される', () => {
    render(<CommunicationTipCard />);

    expect(screen.getByText('今日のTips')).toBeInTheDocument();
  });

  it('日付によって異なるTipsが表示される', () => {
    vi.useFakeTimers();

    vi.setSystemTime(new Date('2025-01-01'));
    const { unmount } = render(<CommunicationTipCard />);
    const tip1 = screen.getByTestId('tip-text').textContent;
    unmount();

    vi.setSystemTime(new Date('2025-01-02'));
    render(<CommunicationTipCard />);
    const tip2 = screen.getByTestId('tip-text').textContent;

    expect(tip1).not.toBe(tip2);

    vi.useRealTimers();
  });
});
