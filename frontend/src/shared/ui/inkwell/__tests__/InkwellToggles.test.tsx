import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import InkwellCheckbox from '../InkwellCheckbox';
import InkwellSwitch from '../InkwellSwitch';

describe('InkwellCheckbox', () => {
  it('checkbox ロールとラベルを持つ', () => {
    render(<InkwellCheckbox label="同意" />);
    expect(screen.getByRole('checkbox', { name: '同意' })).toBeInTheDocument();
  });

  it('クリックで onChange が発火する', () => {
    const onChange = vi.fn();
    render(<InkwellCheckbox label="同意" onChange={onChange} />);
    fireEvent.click(screen.getByRole('checkbox'));
    expect(onChange).toHaveBeenCalledTimes(1);
  });

  it('checked を反映する', () => {
    render(<InkwellCheckbox label="既定 ON" checked readOnly />);
    expect(screen.getByRole('checkbox')).toBeChecked();
  });

  it('disabled を反映する', () => {
    render(<InkwellCheckbox label="無効" disabled />);
    expect(screen.getByRole('checkbox')).toBeDisabled();
  });

  it('押下で波紋が追加される', () => {
    const { container } = render(<InkwellCheckbox label="波紋" />);
    fireEvent.pointerDown(container.querySelector('span')!);
    expect(container.querySelector('.animate-inkwell-ripple')).not.toBeNull();
  });
});

describe('InkwellSwitch', () => {
  it('checkbox ロール（トグル）とラベルを持つ', () => {
    render(<InkwellSwitch label="通知" />);
    expect(screen.getByRole('checkbox', { name: '通知' })).toBeInTheDocument();
  });

  it('クリックで onChange が発火する', () => {
    const onChange = vi.fn();
    render(<InkwellSwitch label="通知" onChange={onChange} />);
    fireEvent.click(screen.getByRole('checkbox'));
    expect(onChange).toHaveBeenCalledTimes(1);
  });

  it('checked / disabled を反映する', () => {
    render(<InkwellSwitch label="ON 無効" checked disabled readOnly />);
    const el = screen.getByRole('checkbox');
    expect(el).toBeChecked();
    expect(el).toBeDisabled();
  });
});
