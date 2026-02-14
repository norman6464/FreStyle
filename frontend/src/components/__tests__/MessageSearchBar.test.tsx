import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import MessageSearchBar from '../MessageSearchBar';

describe('MessageSearchBar', () => {
  const mockOnSearch = vi.fn();
  const mockOnClear = vi.fn();

  beforeEach(() => {
    mockOnSearch.mockClear();
    mockOnClear.mockClear();
  });

  it('検索入力フィールドが表示される', () => {
    render(<MessageSearchBar onSearch={mockOnSearch} onClear={mockOnClear} matchCount={0} />);

    expect(screen.getByPlaceholderText('メッセージを検索...')).toBeInTheDocument();
  });

  it('テキスト入力でonSearchが呼ばれる', () => {
    render(<MessageSearchBar onSearch={mockOnSearch} onClear={mockOnClear} matchCount={0} />);

    fireEvent.change(screen.getByPlaceholderText('メッセージを検索...'), {
      target: { value: 'テスト' },
    });
    expect(mockOnSearch).toHaveBeenCalledWith('テスト');
  });

  it('マッチ件数が表示される', () => {
    render(<MessageSearchBar onSearch={mockOnSearch} onClear={mockOnClear} matchCount={3} />);

    expect(screen.getByText('3件')).toBeInTheDocument();
  });

  it('クリアボタンクリックでonClearが呼ばれる', () => {
    render(<MessageSearchBar onSearch={mockOnSearch} onClear={mockOnClear} matchCount={0} />);

    fireEvent.click(screen.getByLabelText('検索をクリア'));
    expect(mockOnClear).toHaveBeenCalled();
  });

  it('マッチ件数が0のときは件数を表示しない', () => {
    render(<MessageSearchBar onSearch={mockOnSearch} onClear={mockOnClear} matchCount={0} />);

    expect(screen.queryByText('0件')).not.toBeInTheDocument();
  });

  it('role="search"がコンテナに適用される', () => {
    render(<MessageSearchBar onSearch={mockOnSearch} onClear={mockOnClear} matchCount={0} />);
    expect(screen.getByRole('search')).toBeInTheDocument();
  });

  it('aria-labelがinputに適用される', () => {
    render(<MessageSearchBar onSearch={mockOnSearch} onClear={mockOnClear} matchCount={0} />);
    const input = screen.getByLabelText('メッセージを検索');
    expect(input.tagName).toBe('INPUT');
  });

  it('マッチ件数がaria-liveで通知される', () => {
    render(<MessageSearchBar onSearch={mockOnSearch} onClear={mockOnClear} matchCount={5} />);
    const liveRegion = screen.getByText('5件');
    expect(liveRegion).toHaveAttribute('aria-live', 'polite');
  });

  it('入力をクリアするとonSearchが空文字で呼ばれる', () => {
    render(<MessageSearchBar onSearch={mockOnSearch} onClear={mockOnClear} matchCount={0} />);
    const input = screen.getByPlaceholderText('メッセージを検索...');

    fireEvent.change(input, { target: { value: 'テスト' } });
    mockOnSearch.mockClear();
    fireEvent.change(input, { target: { value: '' } });
    expect(mockOnSearch).toHaveBeenCalledWith('');
  });

  it('クリアボタンがsearchロール内にある', () => {
    render(<MessageSearchBar onSearch={mockOnSearch} onClear={mockOnClear} matchCount={0} />);
    const searchContainer = screen.getByRole('search');
    const clearButton = screen.getByLabelText('検索をクリア');
    expect(searchContainer).toContainElement(clearButton);
  });

  it('大きなマッチ件数が正しく表示される', () => {
    render(<MessageSearchBar onSearch={mockOnSearch} onClear={mockOnClear} matchCount={999} />);
    expect(screen.getByText('999件')).toBeInTheDocument();
  });
});
