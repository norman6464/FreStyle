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
});
