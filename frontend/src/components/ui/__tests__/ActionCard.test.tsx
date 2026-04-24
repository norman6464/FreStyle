import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ActionCard from '../ActionCard';

describe('ActionCard', () => {
  it('to prop を渡すと Link として描画される', () => {
    render(
      <MemoryRouter>
        <ActionCard title="練習モード" description="AI と練習" to="/practice" />
      </MemoryRouter>
    );
    const link = screen.getByRole('link', { name: /練習モード/ });
    expect(link).toHaveAttribute('href', '/practice');
  });

  it('onClick prop を渡すと button として描画され、クリックで発火する', () => {
    const onClick = vi.fn();
    render(<ActionCard title="アクション" onClick={onClick} />);
    fireEvent.click(screen.getByRole('button', { name: /アクション/ }));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it('description を描画する', () => {
    render(<ActionCard title="t" onClick={() => {}} description="補足説明" />);
    expect(screen.getByText('補足説明')).toBeInTheDocument();
  });

  it('badge を描画する', () => {
    render(<ActionCard title="t" onClick={() => {}} badge="初心者向け" />);
    expect(screen.getByText('初心者向け')).toBeInTheDocument();
  });

  it('icon を描画する', () => {
    render(
      <ActionCard
        title="t"
        onClick={() => {}}
        icon={<span data-testid="action-icon">★</span>}
      />
    );
    expect(screen.getByTestId('action-icon')).toBeInTheDocument();
  });

  it('emphasis=primary のとき class に primary 系が含まれる', () => {
    render(<ActionCard title="t" onClick={() => {}} emphasis="primary" />);
    const btn = screen.getByRole('button', { name: /t/ });
    expect(btn.className).toMatch(/primary-400\/40|from-primary-500/);
  });
});
