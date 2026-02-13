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
});
