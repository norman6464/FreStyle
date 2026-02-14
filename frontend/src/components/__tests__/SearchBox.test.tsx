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

  it('スペースのみの値でもクリアボタンが表示される', () => {
    render(<SearchBox value=" " onChange={vi.fn()} />);
    expect(screen.getByLabelText('検索をクリア')).toBeInTheDocument();
  });

  it('長い文字列でも正常に表示される', () => {
    const longText = 'テ'.repeat(200);
    render(<SearchBox value={longText} onChange={vi.fn()} />);
    expect(screen.getByDisplayValue(longText)).toBeInTheDocument();
    expect(screen.getByLabelText('検索をクリア')).toBeInTheDocument();
  });

  it('クリアボタンにbutton要素が使用される', () => {
    render(<SearchBox value="テスト" onChange={vi.fn()} />);
    const clearButton = screen.getByLabelText('検索をクリア');
    expect(clearButton.tagName).toBe('BUTTON');
  });

  it('role="search"が外側のdivに適用される', () => {
    render(<SearchBox value="" onChange={vi.fn()} />);
    expect(screen.getByRole('search')).toBeInTheDocument();
  });

  it('aria-labelがinputに適用される', () => {
    render(<SearchBox value="" onChange={vi.fn()} placeholder="名前で検索" />);
    const input = screen.getByLabelText('名前で検索');
    expect(input).toBeInTheDocument();
    expect(input.tagName).toBe('INPUT');
  });

  it('デフォルトplaceholderがaria-labelに使用される', () => {
    render(<SearchBox value="" onChange={vi.fn()} />);
    const input = screen.getByLabelText('検索');
    expect(input).toBeInTheDocument();
  });

  it('searchロール内にinputが含まれる', () => {
    render(<SearchBox value="" onChange={vi.fn()} />);
    const searchRegion = screen.getByRole('search');
    const input = searchRegion.querySelector('input');
    expect(input).toBeTruthy();
  });

  it('クリアボタンがsearchロール内に含まれる', () => {
    render(<SearchBox value="テスト" onChange={vi.fn()} />);
    const searchRegion = screen.getByRole('search');
    const clearButton = searchRegion.querySelector('button');
    expect(clearButton).toBeTruthy();
  });

  it('カスタムplaceholderがaria-labelとplaceholder両方に反映される', () => {
    render(<SearchBox value="" onChange={vi.fn()} placeholder="メールで検索" />);
    const input = screen.getByLabelText('メールで検索');
    expect(input).toHaveAttribute('placeholder', 'メールで検索');
  });
});
