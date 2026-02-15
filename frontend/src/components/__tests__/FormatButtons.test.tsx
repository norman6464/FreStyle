import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import FormatButtons from '../FormatButtons';

describe('FormatButtons', () => {
  const defaultProps = {
    onBold: vi.fn(),
    onItalic: vi.fn(),
    onUnderline: vi.fn(),
    onStrike: vi.fn(),
  };

  it('4つの書式ボタンが表示される', () => {
    render(<FormatButtons {...defaultProps} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(4);
  });

  it('太字ボタンにaria-labelがある', () => {
    render(<FormatButtons {...defaultProps} />);
    expect(screen.getByLabelText('太字')).toBeInTheDocument();
  });

  it('斜体ボタンにaria-labelがある', () => {
    render(<FormatButtons {...defaultProps} />);
    expect(screen.getByLabelText('斜体')).toBeInTheDocument();
  });

  it('下線ボタンにaria-labelがある', () => {
    render(<FormatButtons {...defaultProps} />);
    expect(screen.getByLabelText('下線')).toBeInTheDocument();
  });

  it('取り消し線ボタンにaria-labelがある', () => {
    render(<FormatButtons {...defaultProps} />);
    expect(screen.getByLabelText('取り消し線')).toBeInTheDocument();
  });

  it('太字クリックでonBoldが呼ばれる', () => {
    const onBold = vi.fn();
    render(<FormatButtons {...defaultProps} onBold={onBold} />);
    fireEvent.click(screen.getByLabelText('太字'));
    expect(onBold).toHaveBeenCalled();
  });

  it('下線クリックでonUnderlineが呼ばれる', () => {
    const onUnderline = vi.fn();
    render(<FormatButtons {...defaultProps} onUnderline={onUnderline} />);
    fireEvent.click(screen.getByLabelText('下線'));
    expect(onUnderline).toHaveBeenCalled();
  });

  it('取り消し線クリックでonStrikeが呼ばれる', () => {
    const onStrike = vi.fn();
    render(<FormatButtons {...defaultProps} onStrike={onStrike} />);
    fireEvent.click(screen.getByLabelText('取り消し線'));
    expect(onStrike).toHaveBeenCalled();
  });
});
