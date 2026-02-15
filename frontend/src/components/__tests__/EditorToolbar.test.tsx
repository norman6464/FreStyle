import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import EditorToolbar from '../EditorToolbar';

describe('EditorToolbar', () => {
  const defaultProps = {
    onBold: vi.fn(),
    onItalic: vi.fn(),
    onUnderline: vi.fn(),
    onStrike: vi.fn(),
    onSuperscript: vi.fn(),
    onSubscript: vi.fn(),
    onSelectColor: vi.fn(),
    onHighlight: vi.fn(),
    onAlign: vi.fn(),
    onUndo: vi.fn(),
    onRedo: vi.fn(),
    onClearFormatting: vi.fn(),
    onIndent: vi.fn(),
    onOutdent: vi.fn(),
    onBlockquote: vi.fn(),
  };

  it('書式ボタンが表示される', () => {
    render(<EditorToolbar {...defaultProps} />);
    expect(screen.getByLabelText('太字')).toBeInTheDocument();
    expect(screen.getByLabelText('下線')).toBeInTheDocument();
  });

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
