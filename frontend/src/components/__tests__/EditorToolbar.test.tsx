import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import EditorToolbar from '../EditorToolbar';

describe('EditorToolbar', () => {
  const defaultProps = {
    onSelectColor: vi.fn(),
    onAlign: vi.fn(),
  };

  it('カラーピッカーが表示される', () => {
    render(<EditorToolbar {...defaultProps} />);
    expect(screen.getByLabelText('赤')).toBeInTheDocument();
  });

  it('配置ボタンが表示される', () => {
    render(<EditorToolbar {...defaultProps} />);
    expect(screen.getByLabelText('左寄せ')).toBeInTheDocument();
    expect(screen.getByLabelText('中央寄せ')).toBeInTheDocument();
    expect(screen.getByLabelText('右寄せ')).toBeInTheDocument();
  });

  it('色選択でonSelectColorが呼ばれる', () => {
    const onSelectColor = vi.fn();
    render(<EditorToolbar {...defaultProps} onSelectColor={onSelectColor} />);
    fireEvent.click(screen.getByLabelText('赤'));
    expect(onSelectColor).toHaveBeenCalled();
  });

  it('配置ボタンクリックでonAlignが呼ばれる', () => {
    const onAlign = vi.fn();
    render(<EditorToolbar {...defaultProps} onAlign={onAlign} />);
    fireEvent.click(screen.getByLabelText('中央寄せ'));
    expect(onAlign).toHaveBeenCalledWith('center');
  });
});
