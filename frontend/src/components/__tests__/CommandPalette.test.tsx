import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { BrowserRouter } from 'react-router-dom';
import CommandPalette from '../CommandPalette';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const mockToggleTheme = vi.fn();
vi.mock('../../hooks/useTheme', () => ({
  useTheme: () => ({ theme: 'dark', toggleTheme: mockToggleTheme }),
}));

describe('CommandPalette', () => {
  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  function renderPalette(props = {}) {
    return render(
      <BrowserRouter>
        <CommandPalette {...defaultProps} {...props} />
      </BrowserRouter>
    );
  }

  it('isOpen=falseのとき何も表示しない', () => {
    renderPalette({ isOpen: false });
    expect(screen.queryByPlaceholderText('コマンドを検索...')).not.toBeInTheDocument();
  });

  it('isOpen=trueのとき検索入力が表示される', () => {
    renderPalette();
    expect(screen.getByPlaceholderText('コマンドを検索...')).toBeInTheDocument();
  });

  it('全コマンドがデフォルトで表示される', () => {
    renderPalette();
    expect(screen.getByText('ホーム')).toBeInTheDocument();
    expect(screen.getByText('チャット')).toBeInTheDocument();
    expect(screen.getByText('テーマ切替')).toBeInTheDocument();
    expect(screen.getByText('新規ノート作成')).toBeInTheDocument();
  });

  it('カテゴリヘッダーが表示される', () => {
    renderPalette();
    expect(screen.getByText('ページ移動')).toBeInTheDocument();
    expect(screen.getByText('アクション')).toBeInTheDocument();
  });

  it('検索入力で絞り込みができる', () => {
    renderPalette();
    fireEvent.change(screen.getByPlaceholderText('コマンドを検索...'), {
      target: { value: 'ノート' },
    });
    expect(screen.getByText('ノート')).toBeInTheDocument();
    expect(screen.getByText('新規ノート作成')).toBeInTheDocument();
    expect(screen.queryByText('ホーム')).not.toBeInTheDocument();
  });

  it('ナビゲーションコマンドをクリックするとページ移動する', () => {
    renderPalette();
    fireEvent.click(screen.getByText('ノート'));
    expect(mockNavigate).toHaveBeenCalledWith('/notes');
    expect(defaultProps.onClose).toHaveBeenCalled();
  });

  it('テーマ切替コマンドをクリックするとテーマが切り替わる', () => {
    renderPalette();
    fireEvent.click(screen.getByText('テーマ切替'));
    expect(mockToggleTheme).toHaveBeenCalled();
    expect(defaultProps.onClose).toHaveBeenCalled();
  });

  it('背景オーバーレイをクリックするとパレットが閉じる', () => {
    renderPalette();
    fireEvent.click(screen.getByTestId('command-palette-overlay'));
    expect(defaultProps.onClose).toHaveBeenCalled();
  });

  it('Escapeキーでパレットが閉じる', () => {
    renderPalette();
    fireEvent.keyDown(screen.getByPlaceholderText('コマンドを検索...'), {
      key: 'Escape',
    });
    expect(defaultProps.onClose).toHaveBeenCalled();
  });

  it('ArrowDownで選択が移動する', () => {
    renderPalette();
    const input = screen.getByPlaceholderText('コマンドを検索...');
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    // 2番目のアイテムが選択状態になっているか確認
    const items = screen.getAllByRole('option');
    expect(items[1]).toHaveAttribute('aria-selected', 'true');
  });

  it('ArrowUpで選択が戻る', () => {
    renderPalette();
    const input = screen.getByPlaceholderText('コマンドを検索...');
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    fireEvent.keyDown(input, { key: 'ArrowUp' });
    const items = screen.getAllByRole('option');
    expect(items[1]).toHaveAttribute('aria-selected', 'true');
  });

  it('Enterで選択中のコマンドが実行される', () => {
    renderPalette();
    const input = screen.getByPlaceholderText('コマンドを検索...');
    // 最初のアイテム（ホーム）が選択されている状態でEnter
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(mockNavigate).toHaveBeenCalledWith('/');
    expect(defaultProps.onClose).toHaveBeenCalled();
  });

  it('検索結果が空のとき「該当するコマンドがありません」と表示', () => {
    renderPalette();
    fireEvent.change(screen.getByPlaceholderText('コマンドを検索...'), {
      target: { value: 'xxxxxxxxx' },
    });
    expect(screen.getByText('該当するコマンドがありません')).toBeInTheDocument();
  });

  it('新規ノート作成コマンドをクリックするとノートページに移動する', () => {
    renderPalette();
    fireEvent.click(screen.getByText('新規ノート作成'));
    expect(mockNavigate).toHaveBeenCalledWith('/notes');
    expect(defaultProps.onClose).toHaveBeenCalled();
  });
});
