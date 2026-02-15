import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import HighlightPicker from '../HighlightPicker';

describe('HighlightPicker', () => {
  it('6色のハイライトボタンが表示される', () => {
    render(<HighlightPicker onSelectHighlight={vi.fn()} />);
    const buttons = screen.getAllByRole('button');
    // 6色 + リセットボタン = 7
    expect(buttons).toHaveLength(7);
  });

  it('色ボタンをクリックすると色コードが渡される', () => {
    const handler = vi.fn();
    render(<HighlightPicker onSelectHighlight={handler} />);
    fireEvent.click(screen.getByLabelText('赤ハイライト'));
    expect(handler).toHaveBeenCalledWith('#fecaca');
  });

  it('リセットボタンをクリックすると空文字が渡される', () => {
    const handler = vi.fn();
    render(<HighlightPicker onSelectHighlight={handler} />);
    fireEvent.click(screen.getByLabelText('ハイライトをリセット'));
    expect(handler).toHaveBeenCalledWith('');
  });

  it('各ボタンにaria-labelがある', () => {
    render(<HighlightPicker onSelectHighlight={vi.fn()} />);
    expect(screen.getByLabelText('赤ハイライト')).toBeInTheDocument();
    expect(screen.getByLabelText('オレンジハイライト')).toBeInTheDocument();
    expect(screen.getByLabelText('黄ハイライト')).toBeInTheDocument();
    expect(screen.getByLabelText('緑ハイライト')).toBeInTheDocument();
    expect(screen.getByLabelText('青ハイライト')).toBeInTheDocument();
    expect(screen.getByLabelText('紫ハイライト')).toBeInTheDocument();
  });
});
