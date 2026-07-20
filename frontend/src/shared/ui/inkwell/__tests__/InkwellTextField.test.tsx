import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import InkwellTextField from '../InkwellTextField';

describe('InkwellTextField', () => {
  it('ラベルと入力が紐づく（label クリックで input にフォーカスできる関連）', () => {
    render(<InkwellTextField label="お名前" />);
    const input = screen.getByLabelText('お名前');
    expect(input).toBeInstanceOf(HTMLInputElement);
  });

  it('helperText を表示し aria-describedby で結びつく', () => {
    render(<InkwellTextField label="メール" helperText="社内アドレス" />);
    const input = screen.getByLabelText('メール');
    const helper = screen.getByText('社内アドレス');
    expect(input.getAttribute('aria-describedby')).toBe(helper.id);
  });

  it('error のとき aria-invalid とエラー色クラスになる', () => {
    render(<InkwellTextField label="パスワード" error helperText="不正" />);
    const input = screen.getByLabelText('パスワード');
    expect(input).toHaveAttribute('aria-invalid', 'true');
    expect(input).toHaveClass('border-inkwell-error');
  });

  it('未入力でも placeholder-shown 判定用に空白 placeholder が入る', () => {
    render(<InkwellTextField label="検索" />);
    expect(screen.getByLabelText('検索')).toHaveAttribute('placeholder', ' ');
  });

  it('disabled を渡せる', () => {
    render(<InkwellTextField label="無効" disabled />);
    expect(screen.getByLabelText('無効')).toBeDisabled();
  });
});
