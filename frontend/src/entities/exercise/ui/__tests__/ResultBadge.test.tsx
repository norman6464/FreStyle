import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ResultBadge from '../ResultBadge';

describe('ResultBadge', () => {
  it('合格時は「全テストケース合格」を表示する', () => {
    render(<ResultBadge isCorrect />);
    expect(screen.getByText('全テストケース合格')).toBeInTheDocument();
    expect(screen.queryByText('不合格')).not.toBeInTheDocument();
  });

  it('不合格時は「不合格」を表示する', () => {
    render(<ResultBadge isCorrect={false} />);
    expect(screen.getByText('不合格')).toBeInTheDocument();
    expect(screen.queryByText('全テストケース合格')).not.toBeInTheDocument();
  });
});
