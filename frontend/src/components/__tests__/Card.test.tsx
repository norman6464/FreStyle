import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import Card from '../Card';

describe('Card', () => {
  it('子要素が表示される', () => {
    render(<Card><p>コンテンツ</p></Card>);
    expect(screen.getByText('コンテンツ')).toBeInTheDocument();
  });

  it('デフォルトのスタイルクラスが適用される', () => {
    const { container } = render(<Card><p>テスト</p></Card>);
    expect(container.firstChild).toHaveClass('bg-surface-1');
    expect(container.firstChild).toHaveClass('rounded-lg');
    expect(container.firstChild).toHaveClass('border');
    expect(container.firstChild).toHaveClass('border-surface-3');
    expect(container.firstChild).toHaveClass('p-4');
  });

  it('追加のclassNameが適用される', () => {
    const { container } = render(<Card className="mt-4"><p>テスト</p></Card>);
    expect(container.firstChild).toHaveClass('mt-4');
    expect(container.firstChild).toHaveClass('bg-surface-1');
  });

  it('classNameが未指定でもデフォルトスタイルが適用される', () => {
    const { container } = render(<Card><p>テスト</p></Card>);
    expect(container.firstChild).toHaveClass('bg-surface-1');
  });

  it('複数の子要素がすべて表示される', () => {
    render(
      <Card>
        <p>要素1</p>
        <p>要素2</p>
        <p>要素3</p>
      </Card>
    );
    expect(screen.getByText('要素1')).toBeInTheDocument();
    expect(screen.getByText('要素2')).toBeInTheDocument();
    expect(screen.getByText('要素3')).toBeInTheDocument();
  });

  it('複数のclassNameが同時に適用される', () => {
    const { container } = render(<Card className="my-3 max-w-[85%]"><p>テスト</p></Card>);
    expect(container.firstChild).toHaveClass('my-3');
    expect(container.firstChild).toHaveClass('max-w-[85%]');
    expect(container.firstChild).toHaveClass('bg-surface-1');
    expect(container.firstChild).toHaveClass('p-4');
  });
});
