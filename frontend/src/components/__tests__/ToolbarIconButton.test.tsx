import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ToolbarIconButton from '../ToolbarIconButton';

const MockIcon = (props: React.SVGProps<SVGSVGElement>) => (
  <svg data-testid="mock-icon" {...props} />
);

describe('ToolbarIconButton', () => {
  it('ボタンとアイコンが表示される', () => {
    render(<ToolbarIconButton icon={MockIcon} label="テスト" onClick={vi.fn()} />);
    expect(screen.getByLabelText('テスト')).toBeInTheDocument();
    expect(screen.getByTestId('mock-icon')).toBeInTheDocument();
  });

  it('クリックでonClickが呼ばれる', () => {
    const onClick = vi.fn();
    render(<ToolbarIconButton icon={MockIcon} label="テスト" onClick={onClick} />);
    fireEvent.click(screen.getByLabelText('テスト'));
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('button type="button"である', () => {
    render(<ToolbarIconButton icon={MockIcon} label="テスト" onClick={vi.fn()} />);
    const button = screen.getByLabelText('テスト');
    expect(button.tagName).toBe('BUTTON');
    expect(button).toHaveAttribute('type', 'button');
  });

  it('isActive=trueでtext-primary-500クラスが適用される', () => {
    render(<ToolbarIconButton icon={MockIcon} label="テスト" onClick={vi.fn()} isActive />);
    const button = screen.getByLabelText('テスト');
    expect(button.className).toContain('text-primary-500');
  });

  it('isActive=falseでtext-faintクラスが適用される', () => {
    render(<ToolbarIconButton icon={MockIcon} label="テスト" onClick={vi.fn()} />);
    const button = screen.getByLabelText('テスト');
    expect(button.className).toContain('text-[var(--color-text-faint)]');
  });
});
