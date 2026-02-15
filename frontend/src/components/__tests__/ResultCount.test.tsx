import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import ResultCount from '../ResultCount';

describe('ResultCount', () => {
  it('フィルター未適用時に全件数のみ表示する', () => {
    render(<ResultCount filteredCount={5} totalCount={5} isFilterActive={false} />);
    expect(screen.getByText('5件')).toBeInTheDocument();
  });

  it('フィルター適用時にフィルタ後件数と全件数を表示する', () => {
    render(<ResultCount filteredCount={2} totalCount={5} isFilterActive={true} />);
    expect(screen.getByText('2 / 5件')).toBeInTheDocument();
  });

  it('フィルタ後件数が0の場合も表示する', () => {
    render(<ResultCount filteredCount={0} totalCount={5} isFilterActive={true} />);
    expect(screen.getByText('0 / 5件')).toBeInTheDocument();
  });
});
