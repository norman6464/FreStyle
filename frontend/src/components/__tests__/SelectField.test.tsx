import { render, screen, fireEvent } from '@testing-library/react';
import SelectField from '../SelectField';

const OPTIONS = [
  { value: 'a', label: 'オプションA' },
  { value: 'b', label: 'オプションB' },
  { value: 'c', label: 'オプションC' },
];

describe('SelectField', () => {
  it('ラベルが表示される', () => {
    render(<SelectField label="スタイル" name="style" value="a" onChange={vi.fn()} options={OPTIONS} />);
    expect(screen.getByLabelText('スタイル')).toBeInTheDocument();
  });

  it('すべてのオプションが表示される', () => {
    render(<SelectField label="スタイル" name="style" value="a" onChange={vi.fn()} options={OPTIONS} />);
    OPTIONS.forEach((opt) => {
      expect(screen.getByText(opt.label)).toBeInTheDocument();
    });
  });

  it('選択値が反映される', () => {
    render(<SelectField label="スタイル" name="style" value="b" onChange={vi.fn()} options={OPTIONS} />);
    expect(screen.getByLabelText('スタイル')).toHaveValue('b');
  });

  it('変更時にonChangeが呼ばれる', () => {
    const onChange = vi.fn();
    render(<SelectField label="スタイル" name="style" value="a" onChange={onChange} options={OPTIONS} />);
    fireEvent.change(screen.getByLabelText('スタイル'), { target: { value: 'c' } });
    expect(onChange).toHaveBeenCalled();
  });

  it('select要素がレンダリングされる', () => {
    render(<SelectField label="スタイル" name="style" value="a" onChange={vi.fn()} options={OPTIONS} />);
    expect(screen.getByLabelText('スタイル').tagName).toBe('SELECT');
  });

  it('オプションが1つでもレンダリングされる', () => {
    const single = [{ value: 'only', label: '唯一の選択肢' }];
    render(<SelectField label="スタイル" name="style" value="only" onChange={vi.fn()} options={single} />);
    expect(screen.getByText('唯一の選択肢')).toBeInTheDocument();
    expect(screen.getByLabelText('スタイル')).toHaveValue('only');
  });

  it('name属性がselect要素に設定される', () => {
    render(<SelectField label="スタイル" name="myField" value="a" onChange={vi.fn()} options={OPTIONS} />);
    expect(screen.getByLabelText('スタイル')).toHaveAttribute('name', 'myField');
  });

  it('option要素のvalue属性が正しく設定される', () => {
    const { container } = render(
      <SelectField label="スタイル" name="style" value="a" onChange={vi.fn()} options={OPTIONS} />
    );
    const optionElements = container.querySelectorAll('option');
    expect(optionElements).toHaveLength(3);
    expect(optionElements[0]).toHaveValue('a');
    expect(optionElements[1]).toHaveValue('b');
    expect(optionElements[2]).toHaveValue('c');
  });

  it('エラーメッセージが表示される', () => {
    render(<SelectField label="スタイル" name="style" value="a" onChange={vi.fn()} options={OPTIONS} error="選択してください" />);
    expect(screen.getByText('選択してください')).toBeInTheDocument();
  });

  it('エラー時にaria-invalidがtrueになる', () => {
    render(<SelectField label="スタイル" name="style" value="a" onChange={vi.fn()} options={OPTIONS} error="エラー" />);
    expect(screen.getByLabelText('スタイル')).toHaveAttribute('aria-invalid', 'true');
  });

  it('エラー時にボーダーが赤色になる', () => {
    render(<SelectField label="スタイル" name="style" value="a" onChange={vi.fn()} options={OPTIONS} error="エラー" />);
    expect(screen.getByLabelText('スタイル').className).toContain('border-rose-500');
  });

  it('エラーなし時にaria-invalidがfalseになる', () => {
    render(<SelectField label="スタイル" name="style" value="a" onChange={vi.fn()} options={OPTIONS} />);
    expect(screen.getByLabelText('スタイル')).toHaveAttribute('aria-invalid', 'false');
  });

  it('エラーなし時にボーダーが通常色になる', () => {
    render(<SelectField label="スタイル" name="style" value="a" onChange={vi.fn()} options={OPTIONS} />);
    expect(screen.getByLabelText('スタイル').className).toContain('border-surface-3');
    expect(screen.getByLabelText('スタイル').className).not.toContain('border-rose-500');
  });
});
