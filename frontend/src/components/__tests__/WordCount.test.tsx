import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import WordCount from '../WordCount';

describe('WordCount', () => {
  it('文字数が表示される', () => {
    render(<WordCount charCount={42} />);
    expect(screen.getByText('42文字')).toBeInTheDocument();
  });

  it('0文字の場合も表示される', () => {
    render(<WordCount charCount={0} />);
    expect(screen.getByText('0文字')).toBeInTheDocument();
  });

  it('大きな数値も表示される', () => {
    render(<WordCount charCount={1500} />);
    expect(screen.getByText('1500文字')).toBeInTheDocument();
  });
});
