import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import SearchReplaceBar from '../SearchReplaceBar';

function createMockEditor(overrides = {}) {
  return {
    commands: {
      setSearchTerm: vi.fn(),
      setReplaceTerm: vi.fn(),
      setCaseSensitive: vi.fn(),
      findNext: vi.fn(),
      findPrev: vi.fn(),
      replaceCurrent: vi.fn(),
      replaceAll: vi.fn(),
      clearSearch: vi.fn(),
    },
    storage: {
      searchReplace: {
        results: [],
        currentIndex: 0,
        searchTerm: '',
        replaceTerm: '',
        caseSensitive: false,
      },
    },
    ...overrides,
  } as unknown as Parameters<typeof SearchReplaceBar>[0]['editor'];
}

describe('SearchReplaceBar', () => {
  let mockEditor: ReturnType<typeof createMockEditor>;
  const onClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockEditor = createMockEditor();
  });

  it('isOpen=falseのとき何も表示しない', () => {
    render(<SearchReplaceBar editor={mockEditor} isOpen={false} onClose={onClose} />);
    expect(screen.queryByTestId('search-replace-bar')).not.toBeInTheDocument();
  });

  it('isOpen=trueのとき検索バーが表示される', () => {
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    expect(screen.getByTestId('search-replace-bar')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('検索...')).toBeInTheDocument();
  });

  it('検索入力を変更するとsetSearchTermが呼ばれる', () => {
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    fireEvent.change(screen.getByPlaceholderText('検索...'), {
      target: { value: 'テスト' },
    });
    expect(mockEditor!.commands.setSearchTerm).toHaveBeenCalledWith('テスト');
  });

  it('Enterキーで次を検索する', () => {
    mockEditor = createMockEditor({
      storage: {
        searchReplace: { results: [{ from: 0, to: 3 }], currentIndex: 0 },
      },
    });
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    fireEvent.keyDown(screen.getByPlaceholderText('検索...'), { key: 'Enter' });
    expect(mockEditor!.commands.findNext).toHaveBeenCalled();
  });

  it('Shift+Enterキーで前を検索する', () => {
    mockEditor = createMockEditor({
      storage: {
        searchReplace: { results: [{ from: 0, to: 3 }], currentIndex: 0 },
      },
    });
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    fireEvent.keyDown(screen.getByPlaceholderText('検索...'), { key: 'Enter', shiftKey: true });
    expect(mockEditor!.commands.findPrev).toHaveBeenCalled();
  });

  it('Escapeキーで閉じる', () => {
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    fireEvent.keyDown(screen.getByPlaceholderText('検索...'), { key: 'Escape' });
    expect(onClose).toHaveBeenCalled();
    expect(mockEditor!.commands.clearSearch).toHaveBeenCalled();
  });

  it('×ボタンで閉じる', () => {
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    fireEvent.click(screen.getByLabelText('検索を閉じる'));
    expect(onClose).toHaveBeenCalled();
  });

  it('次を検索ボタンをクリックするとfindNextが呼ばれる', () => {
    mockEditor = createMockEditor({
      storage: {
        searchReplace: { results: [{ from: 0, to: 3 }], currentIndex: 0 },
      },
    });
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    fireEvent.click(screen.getByLabelText('次を検索'));
    expect(mockEditor!.commands.findNext).toHaveBeenCalled();
  });

  it('前を検索ボタンをクリックするとfindPrevが呼ばれる', () => {
    mockEditor = createMockEditor({
      storage: {
        searchReplace: { results: [{ from: 0, to: 3 }], currentIndex: 0 },
      },
    });
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    fireEvent.click(screen.getByLabelText('前を検索'));
    expect(mockEditor!.commands.findPrev).toHaveBeenCalled();
  });

  it('デフォルトでは置換入力が非表示', () => {
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    expect(screen.queryByPlaceholderText('置換...')).not.toBeInTheDocument();
  });

  it('置換ボタンをクリックすると置換入力が表示される', () => {
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    fireEvent.click(screen.getByLabelText('置換を表示'));
    expect(screen.getByPlaceholderText('置換...')).toBeInTheDocument();
    expect(screen.getByLabelText('置換')).toBeInTheDocument();
    expect(screen.getByLabelText('全て置換')).toBeInTheDocument();
  });

  it('置換入力を変更するとsetReplaceTermが呼ばれる', () => {
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    fireEvent.click(screen.getByLabelText('置換を表示'));
    fireEvent.change(screen.getByPlaceholderText('置換...'), {
      target: { value: '新テスト' },
    });
    expect(mockEditor!.commands.setReplaceTerm).toHaveBeenCalledWith('新テスト');
  });

  it('置換ボタンをクリックするとreplaceCurrentが呼ばれる', () => {
    mockEditor = createMockEditor({
      storage: {
        searchReplace: { results: [{ from: 0, to: 3 }], currentIndex: 0 },
      },
    });
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    fireEvent.click(screen.getByLabelText('置換を表示'));
    fireEvent.click(screen.getByLabelText('置換'));
    expect(mockEditor!.commands.replaceCurrent).toHaveBeenCalled();
  });

  it('全置換ボタンをクリックするとreplaceAllが呼ばれる', () => {
    mockEditor = createMockEditor({
      storage: {
        searchReplace: { results: [{ from: 0, to: 3 }], currentIndex: 0 },
      },
    });
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    fireEvent.click(screen.getByLabelText('置換を表示'));
    fireEvent.click(screen.getByLabelText('全て置換'));
    expect(mockEditor!.commands.replaceAll).toHaveBeenCalled();
  });

  it('結果数が表示される', () => {
    mockEditor = createMockEditor({
      storage: {
        searchReplace: {
          results: [{ from: 0, to: 3 }, { from: 5, to: 8 }, { from: 10, to: 13 }],
          currentIndex: 1,
          searchTerm: 'テスト',
        },
      },
    });
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    // 検索語を設定してカウンター表示をトリガー
    fireEvent.change(screen.getByPlaceholderText('検索...'), {
      target: { value: 'テスト' },
    });
    expect(screen.getByText('2 / 3')).toBeInTheDocument();
  });

  it('結果0件のときナビゲーションボタンが無効化される', () => {
    render(<SearchReplaceBar editor={mockEditor} isOpen={true} onClose={onClose} />);
    expect(screen.getByLabelText('次を検索')).toBeDisabled();
    expect(screen.getByLabelText('前を検索')).toBeDisabled();
  });
});
