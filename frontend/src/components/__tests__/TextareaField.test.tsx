import { render, screen, fireEvent } from '@testing-library/react';
import TextareaField from '../TextareaField';

describe('TextareaField', () => {
  it('ラベルが表示される', () => {
    render(<TextareaField label="自己紹介" name="bio" value="" onChange={vi.fn()} />);
    expect(screen.getByLabelText('自己紹介')).toBeInTheDocument();
  });

  it('値が反映される', () => {
    render(<TextareaField label="自己紹介" name="bio" value="テスト値" onChange={vi.fn()} />);
    expect(screen.getByDisplayValue('テスト値')).toBeInTheDocument();
  });

  it('入力時にonChangeが呼ばれる', () => {
    const onChange = vi.fn();
    render(<TextareaField label="自己紹介" name="bio" value="" onChange={onChange} />);
    fireEvent.change(screen.getByLabelText('自己紹介'), { target: { value: 'テスト' } });
    expect(onChange).toHaveBeenCalled();
  });

  it('文字数カウントが表示される', () => {
    render(<TextareaField label="自己紹介" name="bio" value="こんにちは" onChange={vi.fn()} maxLength={100} />);
    expect(screen.getByText('5 / 100')).toBeInTheDocument();
  });

  it('maxLength未指定時は文字数カウントが非表示', () => {
    render(<TextareaField label="自己紹介" name="bio" value="テスト" onChange={vi.fn()} />);
    expect(screen.queryByText(/\//)).not.toBeInTheDocument();
  });

  it('プレースホルダーが表示される', () => {
    render(<TextareaField label="自己紹介" name="bio" value="" onChange={vi.fn()} placeholder="入力してください" />);
    expect(screen.getByPlaceholderText('入力してください')).toBeInTheDocument();
  });

  it('rows属性が反映される', () => {
    render(<TextareaField label="自己紹介" name="bio" value="" onChange={vi.fn()} rows={5} />);
    expect(screen.getByLabelText('自己紹介')).toHaveAttribute('rows', '5');
  });
});
