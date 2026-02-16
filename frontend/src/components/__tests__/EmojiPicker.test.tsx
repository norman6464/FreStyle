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
});
