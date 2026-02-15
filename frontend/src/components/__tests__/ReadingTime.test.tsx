import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ReadingTime from '../ReadingTime';

describe('ReadingTime', () => {
  it('0文字の場合「約0分」を表示する', () => {
    render(<ReadingTime charCount={0} />);
    expect(screen.getByText('約0分')).toBeInTheDocument();
  });

  it('400文字の場合「約1分」を表示する', () => {
    render(<ReadingTime charCount={400} />);
    expect(screen.getByText('約1分')).toBeInTheDocument();
  });

  it('2000文字の場合「約4分」を表示する', () => {
    render(<ReadingTime charCount={2000} />);
    expect(screen.getByText('約4分')).toBeInTheDocument();
  });

  it('200文字の場合「約1分」（切り上げ）を表示する', () => {
    render(<ReadingTime charCount={200} />);
    expect(screen.getByText('約1分')).toBeInTheDocument();
  });
});
