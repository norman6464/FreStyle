import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import PageIntro from '../PageIntro';

describe('PageIntro', () => {
  it('タイトルを h1 として描画する（既定）', () => {
    render(<PageIntro title="練習モード" />);
    const heading = screen.getByRole('heading', { name: '練習モード' });
    expect(heading.tagName).toBe('H1');
  });

  it('headingLevel=2 を指定すると h2 になる', () => {
    render(<PageIntro title="プロフィール" headingLevel={2} />);
    const heading = screen.getByRole('heading', { name: 'プロフィール' });
    expect(heading.tagName).toBe('H2');
  });

  it('description が表示される', () => {
    render(<PageIntro title="練習" description="AI とロールプレイ練習ができます" />);
    expect(screen.getByText('AI とロールプレイ練習ができます')).toBeInTheDocument();
  });

  it('description が未指定なら描画しない', () => {
    const { container } = render(<PageIntro title="t" />);
    expect(container.querySelector('p')).toBeNull();
  });

  it('actions を描画する', () => {
    render(
      <PageIntro
        title="t"
        actions={<button type="button">新規作成</button>}
      />
    );
    expect(screen.getByRole('button', { name: '新規作成' })).toBeInTheDocument();
  });

  it('icon を描画する', () => {
    render(
      <PageIntro
        title="t"
        icon={<span data-testid="icon">⭐</span>}
      />
    );
    expect(screen.getByTestId('icon')).toBeInTheDocument();
  });
});
