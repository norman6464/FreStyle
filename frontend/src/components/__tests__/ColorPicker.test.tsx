import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ColorPicker from '../ColorPicker';

describe('ColorPicker', () => {
  const onSelectColor = vi.fn();

  it('6色のカラーボタンが表示される', () => {
    render(<ColorPicker onSelectColor={onSelectColor} />);
    const buttons = screen.getAllByRole('button');
    // 6色 + リセットボタン = 7
    expect(buttons).toHaveLength(7);
  });

  it('カラーボタンクリックでonSelectColorが呼ばれる', () => {
    const handler = vi.fn();
    render(<ColorPicker onSelectColor={handler} />);
    const buttons = screen.getAllByRole('button');
    fireEvent.click(buttons[0]);
    expect(handler).toHaveBeenCalled();
  });

  it('リセットボタンで空文字が渡される', () => {
    const handler = vi.fn();
    render(<ColorPicker onSelectColor={handler} />);
    fireEvent.click(screen.getByLabelText('色をリセット'));
    expect(handler).toHaveBeenCalledWith('');
  });

  it('各色にaria-labelがある', () => {
    render(<ColorPicker onSelectColor={onSelectColor} />);
    expect(screen.getByLabelText('赤')).toBeInTheDocument();
    expect(screen.getByLabelText('オレンジ')).toBeInTheDocument();
    expect(screen.getByLabelText('黄')).toBeInTheDocument();
    expect(screen.getByLabelText('緑')).toBeInTheDocument();
    expect(screen.getByLabelText('青')).toBeInTheDocument();
    expect(screen.getByLabelText('紫')).toBeInTheDocument();
  });
});
