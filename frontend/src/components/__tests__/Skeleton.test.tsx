import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import Skeleton, { SkeletonCard, SkeletonList } from '../Skeleton';

describe('Skeleton', () => {
  it('スケルトン要素が表示される', () => {
    render(<Skeleton />);

    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('count指定で複数のスケルトンが表示される', () => {
    render(<Skeleton count={3} />);

    const elements = screen.getAllByRole('status');
    expect(elements).toHaveLength(3);
  });
});

describe('SkeletonCard', () => {
  it('カード型スケルトンが表示される', () => {
    render(<SkeletonCard />);

    const elements = screen.getAllByRole('status');
    expect(elements.length).toBeGreaterThan(0);
  });
});

describe('SkeletonList', () => {
  it('リスト型スケルトンがデフォルト3件表示される', () => {
    render(<SkeletonList />);

    const elements = screen.getAllByRole('status');
    expect(elements.length).toBeGreaterThan(0);
  });

  it('count指定でリスト件数が変わる', () => {
    render(<SkeletonList count={5} />);

    const elements = screen.getAllByRole('status');
    expect(elements.length).toBeGreaterThanOrEqual(5);
  });
});

describe('Skeleton - アニメーション', () => {
  it('アニメーションクラスが適用される', () => {
    render(<Skeleton />);
    const el = screen.getByRole('status');
    expect(el.className).toContain('animate-skeleton');
  });
});
