import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import SessionTimeCard from '../SessionTimeCard';

describe('SessionTimeCard', () => {
  it('タイトルが表示される', () => {
    render(<SessionTimeCard dates={['2026-01-15T09:00:00']} />);
    expect(screen.getByText('練習時間帯')).toBeInTheDocument();
  });

  it('4つの時間帯ラベルが表示される', () => {
    render(<SessionTimeCard dates={['2026-01-15T09:00:00']} />);
    expect(screen.getByText('朝')).toBeInTheDocument();
    expect(screen.getByText('昼')).toBeInTheDocument();
    expect(screen.getByText('夕方')).toBeInTheDocument();
    expect(screen.getByText('夜')).toBeInTheDocument();
  });

  it('各時間帯の件数が表示される', () => {
    const dates = [
      '2026-01-15T08:00:00',
      '2026-01-15T10:00:00',
      '2026-01-15T14:00:00',
      '2026-01-15T19:00:00',
      '2026-01-15T23:00:00',
    ];
    render(<SessionTimeCard dates={dates} />);
    const counts = screen.getAllByTestId('time-count');
    expect(counts[0]).toHaveTextContent('2'); // 朝 (8, 10)
    expect(counts[1]).toHaveTextContent('1'); // 昼 (14)
    expect(counts[2]).toHaveTextContent('1'); // 夕方 (19)
    expect(counts[3]).toHaveTextContent('1'); // 夜 (23)
  });

  it('最頻時間帯のメッセージが表示される', () => {
    const dates = [
      '2026-01-15T08:00:00',
      '2026-01-15T09:00:00',
      '2026-01-15T10:00:00',
    ];
    render(<SessionTimeCard dates={dates} />);
    expect(screen.getByText(/朝型/)).toBeInTheDocument();
  });

  it('夜型のメッセージが表示される', () => {
    const dates = [
      '2026-01-15T22:00:00',
      '2026-01-15T23:00:00',
      '2026-01-15T01:00:00',
    ];
    render(<SessionTimeCard dates={dates} />);
    expect(screen.getByText(/夜型/)).toBeInTheDocument();
  });

  it('空配列の場合は何も表示しない', () => {
    const { container } = render(<SessionTimeCard dates={[]} />);
    expect(container.firstChild).toBeNull();
  });
});
