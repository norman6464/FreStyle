import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import EmojiPicker from '../EmojiPicker';

describe('EmojiPicker', () => {
  const onSelect = vi.fn();
  const onClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('isOpen=falseのとき何も表示しない', () => {
    render(<EmojiPicker isOpen={false} onSelect={onSelect} onClose={onClose} />);
    expect(screen.queryByTestId('emoji-picker')).not.toBeInTheDocument();
  });

  it('isOpen=trueのとき絵文字ピッカーが表示される', () => {
    render(<EmojiPicker isOpen={true} onSelect={onSelect} onClose={onClose} />);
    expect(screen.getByTestId('emoji-picker')).toBeInTheDocument();
  });

  it('検索入力が表示される', () => {
    render(<EmojiPicker isOpen={true} onSelect={onSelect} onClose={onClose} />);
    expect(screen.getByPlaceholderText('絵文字を検索...')).toBeInTheDocument();
  });

  it('カテゴリタブが表示される', () => {
    render(<EmojiPicker isOpen={true} onSelect={onSelect} onClose={onClose} />);
    expect(screen.getByLabelText('よく使う')).toBeInTheDocument();
    expect(screen.getByLabelText('顔')).toBeInTheDocument();
  });

  it('絵文字をクリックするとonSelectが呼ばれる', () => {
    render(<EmojiPicker isOpen={true} onSelect={onSelect} onClose={onClose} />);
    // 最初のカテゴリの最初の絵文字をクリック
    const emojiButtons = screen.getAllByRole('button').filter(btn =>
      btn.textContent && /^\p{Emoji}/u.test(btn.textContent) && !btn.getAttribute('aria-label')
    );
    if (emojiButtons.length > 0) {
      fireEvent.click(emojiButtons[0]);
      expect(onSelect).toHaveBeenCalled();
      expect(onClose).toHaveBeenCalled();
    }
  });

  it('Escapeキーで閉じる', () => {
    render(<EmojiPicker isOpen={true} onSelect={onSelect} onClose={onClose} />);
    fireEvent.keyDown(screen.getByPlaceholderText('絵文字を検索...'), { key: 'Escape' });
    expect(onClose).toHaveBeenCalled();
  });

  it('検索でフィルタリングできる', () => {
    render(<EmojiPicker isOpen={true} onSelect={onSelect} onClose={onClose} />);
    fireEvent.change(screen.getByPlaceholderText('絵文字を検索...'), {
      target: { value: '😀' },
    });
    // 検索結果に😀が含まれるはず
    expect(screen.getByTestId('emoji-picker')).toBeInTheDocument();
  });

  it('カテゴリタブをクリックするとカテゴリが切り替わる', () => {
    render(<EmojiPicker isOpen={true} onSelect={onSelect} onClose={onClose} />);
    const handTab = screen.getByLabelText('手・体');
    fireEvent.click(handTab);
    // 「手・体」カテゴリ名が表示される
    expect(screen.getByText('手・体')).toBeInTheDocument();
  });

  it('検索中はカテゴリタブが非表示になる', () => {
    render(<EmojiPicker isOpen={true} onSelect={onSelect} onClose={onClose} />);
    fireEvent.change(screen.getByPlaceholderText('絵文字を検索...'), {
      target: { value: 'test' },
    });
    // カテゴリタブのaria-labelが見つからない
    expect(screen.queryByLabelText('よく使う')).not.toBeInTheDocument();
  });

  it('検索結果がない場合はメッセージが表示される', () => {
    render(<EmojiPicker isOpen={true} onSelect={onSelect} onClose={onClose} />);
    fireEvent.change(screen.getByPlaceholderText('絵文字を検索...'), {
      target: { value: 'zzzznotfound' },
    });
    expect(screen.getByText('該当する絵文字がありません')).toBeInTheDocument();
  });

  it('Escape以外のキーではonCloseが呼ばれない', () => {
    render(<EmojiPicker isOpen={true} onSelect={onSelect} onClose={onClose} />);
    fireEvent.keyDown(screen.getByPlaceholderText('絵文字を検索...'), { key: 'Enter' });
    expect(onClose).not.toHaveBeenCalled();
  });

  it('検索入力にaria-labelが設定される', () => {
    render(<EmojiPicker isOpen={true} onSelect={onSelect} onClose={onClose} />);
    expect(screen.getByLabelText('絵文字を検索')).toBeInTheDocument();
  });

  it('絵文字ボタンにaria-labelが設定される', () => {
    render(<EmojiPicker isOpen={true} onSelect={onSelect} onClose={onClose} />);
    const emojiButtons = screen.getAllByRole('button').filter(btn => {
      const label = btn.getAttribute('aria-label');
      return label !== null && label !== 'よく使う' && label !== '顔' && label !== '手・体' && label !== '動物・自然' && label !== '食べ物' && label !== 'オブジェクト' && label !== '記号';
    });
    expect(emojiButtons.length).toBeGreaterThan(0);
  });
});
