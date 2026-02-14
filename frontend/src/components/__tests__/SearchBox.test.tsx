import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import SearchBox from '../SearchBox';

describe('SearchBox', () => {
  it('プレースホルダーが表示される', () => {
    render(<SearchBox value="" onChange={vi.fn()} placeholder="ユーザーを検索" />);

    expect(screen.getByPlaceholderText('ユーザーを検索')).toBeInTheDocument();
  });

  it('デフォルトプレースホルダーが表示される', () => {
    render(<SearchBox value="" onChange={vi.fn()} />);

    expect(screen.getByPlaceholderText('検索')).toBeInTheDocument();
  });

  it('入力時にonChangeが呼ばれる', () => {
    const mockOnChange = vi.fn();
    render(<SearchBox value="" onChange={mockOnChange} />);

    fireEvent.change(screen.getByPlaceholderText('検索'), { target: { value: 'テスト' } });
    expect(mockOnChange).toHaveBeenCalledWith('テスト');
  });

  it('値が反映される', () => {
    render(<SearchBox value="初期値" onChange={vi.fn()} />);
    expect(screen.getByDisplayValue('初期値')).toBeInTheDocument();
  });

  it('テキスト入力フィールドとしてレンダリングされる', () => {
    render(<SearchBox value="" onChange={vi.fn()} />);
    const input = screen.getByPlaceholderText('検索');
    expect(input).toHaveAttribute('type', 'text');
  });

  it('検索アイコンが表示される', () => {
    const { container } = render(<SearchBox value="" onChange={vi.fn()} />);
    const svg = container.querySelector('svg');
    expect(svg).toBeTruthy();
  });

  it('値が入力されている時にクリアボタンが表示される', () => {
    render(<SearchBox value="テスト" onChange={vi.fn()} />);

    expect(screen.getByLabelText('検索をクリア')).toBeInTheDocument();
  });

  it('値が空の時にクリアボタンが表示されない', () => {
    render(<SearchBox value="" onChange={vi.fn()} />);

    expect(screen.queryByLabelText('検索をクリア')).not.toBeInTheDocument();
  });

  it('クリアボタンクリックで空文字列でonChangeが呼ばれる', () => {
    const mockOnChange = vi.fn();
    render(<SearchBox value="テスト" onChange={mockOnChange} />);

    fireEvent.click(screen.getByLabelText('検索をクリア'));
    expect(mockOnChange).toHaveBeenCalledWith('');
  });
});
