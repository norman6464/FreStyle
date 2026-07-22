import { describe, it, expect } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useFormField } from '../useFormField';

describe('useFormField', () => {
  it('初期値を返す', () => {
    const { result } = renderHook(() => useFormField({ email: '', password: '' }));
    expect(result.current.form).toEqual({ email: '', password: '' });
  });

  it('handleChangeでフォーム値を更新する', () => {
    const { result } = renderHook(() => useFormField({ email: '', password: '' }));

    act(() => {
      result.current.handleChange({
        target: { name: 'email', value: 'test@example.com' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    expect(result.current.form.email).toBe('test@example.com');
    expect(result.current.form.password).toBe('');
  });

  it('複数フィールドを個別に更新できる', () => {
    const { result } = renderHook(() => useFormField({ name: '', age: '' }));

    act(() => {
      result.current.handleChange({
        target: { name: 'name', value: 'テスト' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    act(() => {
      result.current.handleChange({
        target: { name: 'age', value: '25' },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    expect(result.current.form).toEqual({ name: 'テスト', age: '25' });
  });

  it('setFormで直接フォーム値を設定できる', () => {
    const { result } = renderHook(() => useFormField({ email: '' }));

    act(() => {
      result.current.setForm({ email: 'direct@example.com' });
    });

    expect(result.current.form.email).toBe('direct@example.com');
  });
});
