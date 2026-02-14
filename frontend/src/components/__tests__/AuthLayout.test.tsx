import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import AuthLayout from '../AuthLayout';

describe('AuthLayout', () => {
  it('FreStyleロゴが表示される', () => {
    render(<AuthLayout><div>テスト</div></AuthLayout>);

    expect(screen.getByText('FreStyle')).toBeInTheDocument();
  });

  it('子要素が表示される', () => {
    render(<AuthLayout><p>子コンテンツ</p></AuthLayout>);

    expect(screen.getByText('子コンテンツ')).toBeInTheDocument();
  });

  it('複数の子要素が表示される', () => {
    render(
      <AuthLayout>
        <p>要素1</p>
        <p>要素2</p>
      </AuthLayout>
    );

    expect(screen.getByText('要素1')).toBeInTheDocument();
    expect(screen.getByText('要素2')).toBeInTheDocument();
  });

  it('プライマリカラーのボーダーが適用される', () => {
    const { container } = render(<AuthLayout><div>テスト</div></AuthLayout>);
    const card = container.querySelector('.border-t-primary-500');
    expect(card).toBeTruthy();
  });
});
