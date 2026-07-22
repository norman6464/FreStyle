import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import InkwellButton from '../InkwellButton';

describe('InkwellButton', () => {
  it('ラベルを描画しクリックで onClick が呼ばれる', () => {
    const onClick = vi.fn();
    render(<InkwellButton onClick={onClick}>保存</InkwellButton>);
    const btn = screen.getByRole('button', { name: '保存' });
    fireEvent.click(btn);
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('contained/primary は塗りの背景クラスを持つ', () => {
    render(<InkwellButton>OK</InkwellButton>);
    expect(screen.getByRole('button')).toHaveClass('bg-inkwell-primary');
  });

  it('outlined は枠線クラスを持ち塗り背景は持たない', () => {
    render(<InkwellButton variant="outlined">OK</InkwellButton>);
    const btn = screen.getByRole('button');
    expect(btn).toHaveClass('border-inkwell-primary/50');
    expect(btn).not.toHaveClass('bg-inkwell-primary');
  });

  it('color=error で error 系クラスに切り替わる', () => {
    render(<InkwellButton color="error">削除</InkwellButton>);
    expect(screen.getByRole('button')).toHaveClass('bg-inkwell-error');
  });

  it('disabled ではクリックが発火しない', () => {
    const onClick = vi.fn();
    render(
      <InkwellButton disabled onClick={onClick}>
        無効
      </InkwellButton>,
    );
    fireEvent.click(screen.getByRole('button'));
    expect(onClick).not.toHaveBeenCalled();
  });

  it('押下で波紋要素が追加される', () => {
    const { container } = render(<InkwellButton>波紋</InkwellButton>);
    const btn = screen.getByRole('button');
    expect(container.querySelector('.animate-inkwell-ripple')).toBeNull();
    fireEvent.pointerDown(btn);
    expect(container.querySelector('.animate-inkwell-ripple')).not.toBeNull();
  });

  it('startIcon / endIcon を描画する', () => {
    render(
      <InkwellButton startIcon={<svg data-testid="start" />} endIcon={<svg data-testid="end" />}>
        送信
      </InkwellButton>,
    );
    expect(screen.getByTestId('start')).toBeInTheDocument();
    expect(screen.getByTestId('end')).toBeInTheDocument();
  });
});
