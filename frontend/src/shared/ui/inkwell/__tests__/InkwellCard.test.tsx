import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import InkwellCard, { InkwellCardContent, InkwellCardActions } from '../InkwellCard';

describe('InkwellCard', () => {
  it('既定 elevation=1 は shadow クラスを持つ', () => {
    render(<InkwellCard>本文</InkwellCard>);
    expect(screen.getByText('本文')).toHaveClass('shadow-inkwell-1');
  });

  it('elevation=0 は影ではなく枠線', () => {
    render(<InkwellCard elevation={0}>枠</InkwellCard>);
    const card = screen.getByText('枠');
    expect(card).toHaveClass('border', 'border-inkwell-divider');
    expect(card).not.toHaveClass('shadow-inkwell-1');
  });

  it('elevation=8 は最も強い影クラス', () => {
    render(<InkwellCard elevation={8}>浮遊</InkwellCard>);
    expect(screen.getByText('浮遊')).toHaveClass('shadow-inkwell-8');
  });

  it('Content と Actions を内包できる', () => {
    render(
      <InkwellCard>
        <InkwellCardContent>コンテンツ</InkwellCardContent>
        <InkwellCardActions>
          <button>操作</button>
        </InkwellCardActions>
      </InkwellCard>,
    );
    expect(screen.getByText('コンテンツ')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '操作' })).toBeInTheDocument();
  });
});
