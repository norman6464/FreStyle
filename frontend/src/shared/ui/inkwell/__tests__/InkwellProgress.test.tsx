import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import InkwellCircularProgress from '../InkwellCircularProgress';
import InkwellLinearProgress from '../InkwellLinearProgress';
import InkwellSkeleton from '../InkwellSkeleton';

describe('InkwellCircularProgress', () => {
  it('determinate は progressbar として value を持つ', () => {
    render(<InkwellCircularProgress value={40} />);
    const bar = screen.getByRole('progressbar');
    expect(bar).toHaveAttribute('aria-valuenow', '40');
    expect(bar).toHaveAttribute('aria-valuemax', '100');
  });

  it('indeterminate は value を持たず「読み込み中」ラベル + spin', () => {
    render(<InkwellCircularProgress />);
    const bar = screen.getByRole('progressbar', { name: '読み込み中' });
    expect(bar).not.toHaveAttribute('aria-valuenow');
    expect(bar.className).toContain('animate-spin');
  });

  it('value は 0〜100 にクランプされる', () => {
    render(<InkwellCircularProgress value={150} />);
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '100');
  });
});

describe('InkwellLinearProgress', () => {
  it('determinate は value を持つ', () => {
    render(<InkwellLinearProgress value={25} />);
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '25');
  });

  it('indeterminate は流れる帯（animate-inkwell-bar）', () => {
    const { container } = render(<InkwellLinearProgress />);
    expect(container.querySelector('.animate-inkwell-bar')).not.toBeNull();
  });
});

describe('InkwellSkeleton', () => {
  it('status ロールと読み込み中ラベルを持つ', () => {
    render(<InkwellSkeleton />);
    expect(screen.getByRole('status', { name: '読み込み中' })).toBeInTheDocument();
  });

  it('circle バリアントは丸角', () => {
    render(<InkwellSkeleton variant="circle" />);
    expect(screen.getByRole('status').className).toContain('rounded-full');
  });
});
