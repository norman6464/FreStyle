import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import FormatButtons from '../FormatButtons';

describe('FormatButtons', () => {
  const defaultProps = {
    onBold: vi.fn(),
    onItalic: vi.fn(),
    onUnderline: vi.fn(),
    onStrike: vi.fn(),
    onSuperscript: vi.fn(),
    onSubscript: vi.fn(),
  };

  it('6つの書式ボタンが表示される', () => {
    render(<FormatButtons {...defaultProps} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(6);
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

  it('上付き文字ボタンにaria-labelがある', () => {
    render(<FormatButtons {...defaultProps} />);
    expect(screen.getByLabelText('上付き文字')).toBeInTheDocument();
  });

  it('下付き文字ボタンにaria-labelがある', () => {
    render(<FormatButtons {...defaultProps} />);
    expect(screen.getByLabelText('下付き文字')).toBeInTheDocument();
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

  it('上付き文字クリックでonSuperscriptが呼ばれる', () => {
    const onSuperscript = vi.fn();
    render(<FormatButtons {...defaultProps} onSuperscript={onSuperscript} />);
    fireEvent.click(screen.getByLabelText('上付き文字'));
    expect(onSuperscript).toHaveBeenCalled();
  });

  it('下付き文字クリックでonSubscriptが呼ばれる', () => {
    const onSubscript = vi.fn();
    render(<FormatButtons {...defaultProps} onSubscript={onSubscript} />);
    fireEvent.click(screen.getByLabelText('下付き文字'));
    expect(onSubscript).toHaveBeenCalled();
  });
});
