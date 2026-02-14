import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import MotivationQuoteCard from '../MotivationQuoteCard';

describe('MotivationQuoteCard', () => {
  it('名言テキストが表示される', () => {
    render(<MotivationQuoteCard />);

    const quoteElements = screen.getAllByTestId('quote-text');
    expect(quoteElements.length).toBeGreaterThanOrEqual(1);
  });

  it('著者名が表示される', () => {
    render(<MotivationQuoteCard />);

    const authorElements = screen.getAllByTestId('quote-author');
    expect(authorElements.length).toBeGreaterThanOrEqual(1);
  });

  it('タイトルが表示される', () => {
    render(<MotivationQuoteCard />);

    expect(screen.getByText("今日の一言")).toBeInTheDocument();
  });

  it('同じ日には同じ名言が表示される', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-03-01'));

    const { unmount } = render(<MotivationQuoteCard />);
    const first = screen.getByTestId('quote-text').textContent;
    unmount();

    render(<MotivationQuoteCard />);
    const second = screen.getByTestId('quote-text').textContent;

    expect(first).toBe(second);
    vi.useRealTimers();
  });
});
