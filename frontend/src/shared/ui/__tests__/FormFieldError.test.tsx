import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import FormFieldError from '../FormFieldError';

describe('FormFieldError', () => {
  it('エラーメッセージが表示される', () => {
    render(<FormFieldError name="email" error="メールアドレスが無効です" />);
    expect(screen.getByRole('alert')).toHaveTextContent('メールアドレスが無効です');
  });

  it('errorがundefinedの場合は何も表示しない', () => {
    const { container } = render(<FormFieldError name="email" />);
    expect(container.firstChild).toBeNull();
  });

  it('errorが空文字の場合は何も表示しない', () => {
    const { container } = render(<FormFieldError name="email" error="" />);
    expect(container.firstChild).toBeNull();
  });

  it('正しいid属性が設定される', () => {
    render(<FormFieldError name="password" error="必須です" />);
    expect(screen.getByRole('alert')).toHaveAttribute('id', 'password-error');
  });
});
