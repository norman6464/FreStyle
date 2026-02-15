import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import HorizontalRuleButton from '../HorizontalRuleButton';

describe('HorizontalRuleButton', () => {
  it('水平線ボタンが表示される', () => {
    render(<HorizontalRuleButton onHorizontalRule={vi.fn()} />);
    expect(screen.getByLabelText('水平線')).toBeInTheDocument();
  });

  it('クリックでonHorizontalRuleが呼ばれる', () => {
    const onHorizontalRule = vi.fn();
    render(<HorizontalRuleButton onHorizontalRule={onHorizontalRule} />);
    fireEvent.click(screen.getByLabelText('水平線'));
    expect(onHorizontalRule).toHaveBeenCalledTimes(1);
  });

  it('ボタンがbutton要素である', () => {
    render(<HorizontalRuleButton onHorizontalRule={vi.fn()} />);
    const button = screen.getByLabelText('水平線');
    expect(button.tagName).toBe('BUTTON');
    expect(button).toHaveAttribute('type', 'button');
  });
});
